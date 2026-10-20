package alive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodexTUIRequestBuildsCapturedRequestShape(t *testing.T) {
	setupCodexTestHome(t)
	body, headers, err := codexTUIRequest("https://cpa.example/v1", "vendor/gpt-test", "Say OK.")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Model string `json:"model"`
		Input []struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"input"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Model != "vendor/gpt-test" {
		t.Fatalf("model = %q", payload.Model)
	}
	if len(payload.Input) < 7 {
		t.Fatalf("input count = %d, want captured context plus the user turn", len(payload.Input))
	}
	last := payload.Input[len(payload.Input)-1]
	if last.Type != "message" || last.Role != "user" || len(last.Content) != 1 || last.Content[0].Text != "Say OK." {
		t.Fatalf("last input = %#v", last)
	}
	for _, key := range []string{"Authorization", "X-Codex-Turn-Metadata", "X-Client-Request-Id", "Session-Id", "Thread-Id", "Originator"} {
		if headers[key] == "" {
			t.Fatalf("missing header %s", key)
		}
	}
}

func TestRunCodexHTTPAttemptReportsFirstHTTPError(t *testing.T) {
	setupCodexTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("originator"); got != "codex-tui" {
			t.Fatalf("originator = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"get_channel_failed","message":"capacity full","type":"new_api_error"}}`))
	}))
	defer server.Close()

	var progress []string
	result := runCodexHTTPAttempt(context.Background(), server.URL+"/v1", "vendor/gpt-test", PromptCase{Input: "Say OK."}, func(message string) {
		progress = append(progress, message)
	})
	if result.Err == nil {
		t.Fatal("expected HTTP failure")
	}
	for _, want := range []string{"HTTP 500", "capacity full", "get_channel_failed"} {
		if !strings.Contains(result.LastError, want) {
			t.Fatalf("error %q does not contain %q", result.LastError, want)
		}
	}
	if len(progress) != 1 || progress[0] != result.LastError {
		t.Fatalf("progress = %#v", progress)
	}
}

func TestRunCodexHTTPAttemptReadsSSEOutput(t *testing.T) {
	setupCodexTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"type":"response.output_text.delta","delta":"O"}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"type":"response.output_text.delta","delta":"K"}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"type":"response.completed","response":{"output":[{"type":"message","content":[{"type":"output_text","text":"OK"}]}]}}`)
		fmt.Fprintln(w)
	}))
	defer server.Close()

	result := runCodexHTTPAttempt(context.Background(), server.URL+"/v1", "vendor/gpt-test", PromptCase{Input: "Say OK."}, nil)
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if result.Output != "OK" {
		t.Fatalf("output = %q", result.Output)
	}
}

func TestProbeEndpointsUseUnifiedBaseURL(t *testing.T) {
	for base, want := range map[string]string{
		"https://cpa.example":              "https://cpa.example/v1/responses",
		"https://cpa.example/v1":           "https://cpa.example/v1/responses",
		"https://cpa.example/v1/responses": "https://cpa.example/v1/responses",
	} {
		if got := codexResponsesEndpoint(base); got != want {
			t.Fatalf("codex endpoint for %q = %q, want %q", base, got, want)
		}
	}
	for base, want := range map[string]string{
		"https://cpa.example":             "https://cpa.example/v1/messages?beta=true",
		"https://cpa.example/v1":          "https://cpa.example/v1/messages?beta=true",
		"https://cpa.example/v1/messages": "https://cpa.example/v1/messages?beta=true",
	} {
		if got := claudeMessagesEndpoint(base); got != want {
			t.Fatalf("Claude endpoint for %q = %q, want %q", base, got, want)
		}
	}
}

func TestRunCodexHTTPAttemptReportsSSEFailureAsFirstError(t *testing.T) {
	setupCodexTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"type":"response.failed","response":{"status":"failed","error":{"code":"upstream_unavailable","message":"provider rejected request"}}}`)
		fmt.Fprintln(w)
	}))
	defer server.Close()

	var progress []string
	result := runCodexHTTPAttempt(context.Background(), server.URL+"/v1", "vendor/gpt-test", PromptCase{Input: "Say OK."}, func(message string) {
		progress = append(progress, message)
	})
	if result.Err == nil {
		t.Fatal("expected SSE failure")
	}
	for _, want := range []string{"provider rejected request", "upstream_unavailable"} {
		if !strings.Contains(result.FirstError, want) {
			t.Fatalf("first error %q does not contain %q", result.FirstError, want)
		}
	}
	if len(progress) != 1 || progress[0] != result.FirstError {
		t.Fatalf("progress = %#v", progress)
	}
}

func TestReadCodexCLIVersionFromVersionCommand(t *testing.T) {
	command := filepath.Join(t.TempDir(), "fake-codex")
	if err := os.WriteFile(command, []byte("#!/bin/sh\nprintf 'codex-cli 0.163.0\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := readCLIVersion(command)
	if err != nil {
		t.Fatal(err)
	}
	if got != "0.163.0" {
		t.Fatalf("version = %q", got)
	}
}

func TestCodexVersionCacheRefreshPolicy(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "codex-version.json")
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	oldVersion := cliVersionCache{Version: "1.2.3", UpdatedAt: now.Add(-25 * time.Hour)}
	if err := writeVersionCache(cachePath, oldVersion); err != nil {
		t.Fatal(err)
	}

	missingCommand := filepath.Join(dir, "missing-codex")
	if got := refreshCLIVersion(cachePath, missingCommand, capturedCodexTUIVersion, now); got != "1.2.3" {
		t.Fatalf("stale cache fallback = %q", got)
	}

	command := filepath.Join(dir, "fake-codex")
	if err := os.WriteFile(command, []byte("#!/bin/sh\nprintf 'codex-cli 9.8.7\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := refreshCLIVersion(cachePath, command, capturedCodexTUIVersion, now); got != "9.8.7" {
		t.Fatalf("refreshed version = %q", got)
	}
	cache := readVersionCache(cachePath)
	if cache.Version != "9.8.7" || !cache.UpdatedAt.Equal(now) {
		t.Fatalf("refreshed cache = %#v", cache)
	}
	if got := refreshCLIVersion(cachePath, missingCommand, capturedCodexTUIVersion, now.Add(23*time.Hour)); got != "9.8.7" {
		t.Fatalf("fresh cache = %q", got)
	}
}

func TestParseCodexCLIVersion(t *testing.T) {
	for input, want := range map[string]string{
		"codex-cli 0.163.0\n":        "0.163.0",
		"codex v1.2.3\n":             "1.2.3",
		"codex-cli 0.164.0-beta.1\n": "0.164.0-beta.1",
		"invalid":                    "",
	} {
		if got := parseCLIVersion(input); got != want {
			t.Fatalf("parseCodexCLIVersion(%q) = %q, want %q", input, got, want)
		}
	}
	if got := formatCodexTUIUserAgent("0.163.0"); got != "codex-tui/0.163.0 (Ubuntu 24.4.0; aarch64) tmux-256color (codex-tui; 0.163.0)" {
		t.Fatalf("user agent = %q", got)
	}
}

func setupCodexTestHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	authPath := filepath.Join(home, "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"OPENAI_API_KEY":"test-key"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "installation_id"), []byte("test-installation"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", home)
	t.Setenv("CODEX_AUTH_FILE", authPath)
	cachePath := filepath.Join(home, "cache", "codex-version.json")
	if err := writeVersionCache(cachePath, cliVersionCache{Version: capturedCodexTUIVersion, UpdatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("API_ALIVE_CODEX_VERSION_CACHE", cachePath)
}
