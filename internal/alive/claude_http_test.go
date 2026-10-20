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

func TestClaudeCodeRequestBuildsCapturedRequestShape(t *testing.T) {
	setupClaudeTestHome(t)
	body, headers, err := claudeCodeRequest("https://cpa.example", "vendor/claude-test", "Say OK.")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Model    string `json:"model"`
		Messages []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"messages"`
		System []struct {
			Text string `json:"text"`
		} `json:"system"`
		Metadata struct {
			UserID string `json:"user_id"`
		} `json:"metadata"`
		Tools []any `json:"tools"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Model != "vendor/claude-test" {
		t.Fatalf("model = %q", payload.Model)
	}
	if len(payload.Messages) != 2 || len(payload.Messages[0].Content) < 4 {
		t.Fatalf("messages = %#v", payload.Messages)
	}
	last := payload.Messages[0].Content[len(payload.Messages[0].Content)-1]
	if last.Type != "text" || last.Text != "Say OK." {
		t.Fatalf("last content = %#v", last)
	}
	if len(payload.Tools) != 23 {
		t.Fatalf("tools = %d", len(payload.Tools))
	}
	if !strings.Contains(payload.System[0].Text, "cc_version=2.1.301.c3d;") {
		t.Fatalf("billing header = %q", payload.System[0].Text)
	}
	var userID struct {
		DeviceID  string `json:"device_id"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal([]byte(payload.Metadata.UserID), &userID); err != nil {
		t.Fatal(err)
	}
	if userID.DeviceID != "test-device" || userID.SessionID != headers["X-Claude-Code-Session-Id"] {
		t.Fatalf("user_id = %#v, session header = %q", userID, headers["X-Claude-Code-Session-Id"])
	}
	for key, want := range map[string]string{
		"Authorization":               "Bearer test-token",
		"User-Agent":                  "claude-cli/2.1.301 (external, sdk-cli)",
		"X-App":                       "cli",
		"Anthropic-Version":           "2023-06-01",
		"X-Stainless-Package-Version": "0.128.0",
	} {
		if headers[key] != want {
			t.Fatalf("header %s = %q, want %q", key, headers[key], want)
		}
	}
}

func TestRunClaudeHTTPAttemptReportsHTTPError(t *testing.T) {
	setupClaudeTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.URL.Query().Get("beta") != "true" {
			t.Fatalf("request URL = %q", r.URL.String())
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"overloaded_error","message":"provider busy"}}`))
	}))
	defer server.Close()

	var progress []string
	result := runClaudeHTTPAttempt(context.Background(), server.URL, "vendor/claude-test", PromptCase{Input: "Say OK."}, func(message string) {
		progress = append(progress, message)
	})
	if result.Err == nil {
		t.Fatal("expected HTTP failure")
	}
	for _, want := range []string{"HTTP 500", "provider busy", "overloaded_error"} {
		if !strings.Contains(result.LastError, want) {
			t.Fatalf("error %q does not contain %q", result.LastError, want)
		}
	}
	if len(progress) != 1 || progress[0] != result.LastError {
		t.Fatalf("progress = %#v", progress)
	}
}

func TestRunClaudeHTTPAttemptReadsSSEOutput(t *testing.T) {
	setupClaudeTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `event: message_start`)
		fmt.Fprintln(w, `data: {"type":"message_start","message":{"content":[]}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `event: content_block_delta`)
		fmt.Fprintln(w, `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"O"}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"K"}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"type":"message_stop"}`)
		fmt.Fprintln(w)
	}))
	defer server.Close()

	result := runClaudeHTTPAttempt(context.Background(), server.URL, "vendor/claude-test", PromptCase{Input: "Say OK."}, nil)
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if result.Output != "OK" {
		t.Fatalf("output = %q", result.Output)
	}
}

func TestRunClaudeHTTPAttemptReportsSSEFailure(t *testing.T) {
	setupClaudeTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `event: error`)
		fmt.Fprintln(w, `data: {"type":"error","error":{"type":"rate_limit_error","message":"quota exhausted"}}`)
		fmt.Fprintln(w)
	}))
	defer server.Close()

	var progress []string
	result := runClaudeHTTPAttempt(context.Background(), server.URL, "vendor/claude-test", PromptCase{Input: "Say OK."}, func(message string) {
		progress = append(progress, message)
	})
	if result.Err == nil {
		t.Fatal("expected SSE failure")
	}
	for _, want := range []string{"quota exhausted", "rate_limit_error"} {
		if !strings.Contains(result.FirstError, want) {
			t.Fatalf("first error %q does not contain %q", result.FirstError, want)
		}
	}
	if len(progress) != 1 || progress[0] != result.FirstError {
		t.Fatalf("progress = %#v", progress)
	}
}

func TestRunClaudeHTTPAttemptRejectsHTMLAsSSE(t *testing.T) {
	setupClaudeTestHome(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, "<!doctype html>")
		fmt.Fprintln(w, "<html><head><title>New API</title></head></html>")
	}))
	defer server.Close()

	var progress []string
	result := runClaudeHTTPAttempt(context.Background(), server.URL, "vendor/claude-test", PromptCase{Input: "Say OK."}, func(message string) {
		progress = append(progress, message)
	})
	if result.Err == nil {
		t.Fatal("expected malformed SSE failure")
	}
	for _, want := range []string{"not SSE", "New API"} {
		if !strings.Contains(result.LastError, want) {
			t.Fatalf("error %q does not contain %q", result.LastError, want)
		}
	}
	if len(progress) != 1 || progress[0] != result.LastError {
		t.Fatalf("progress = %#v", progress)
	}
}

func setupClaudeTestHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	configDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	settings := []byte(`{"env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`)
	if err := os.WriteFile(filepath.Join(configDir, "settings.json"), settings, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"userID":"test-device"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(home, "cache", "claude-version.json")
	if err := writeVersionCache(cachePath, cliVersionCache{Version: capturedClaudeTUIVersion, UpdatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", configDir)
	t.Setenv("API_ALIVE_CLAUDE_VERSION_CACHE", cachePath)
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
}
