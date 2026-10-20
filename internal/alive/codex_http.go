package alive

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed codex_request_template.json
var codexRequestTemplate []byte

const (
	capturedCodexTUIVersion    = "0.163.0"
	codexTUIUserAgentFormat    = "codex-tui/%s (Ubuntu 24.4.0; aarch64) tmux-256color (codex-tui; %s)"
	codexVersionCommandTimeout = 3 * time.Second
	cliVersionCacheTTL         = 24 * time.Hour
	maxHTTPErrorChars          = 2000
	maxHTTPResponseBody        = 8 << 20
)

var cliVersionMu sync.Mutex

type cliVersionCache struct {
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type attemptExecution struct {
	Output     string
	FirstError string
	LastError  string
	ExitCode   *int
	Err        error
}

type codexTUIIDs struct {
	SessionID       string
	ThreadID        string
	TurnID          string
	RootTurnID      string
	ContextWindowID string
	WindowID        string
	InstallationID  string
}

func runCodexHTTPAttempt(ctx context.Context, baseURL, model string, prompt PromptCase, onFirstError func(string)) attemptExecution {
	body, headers, err := codexTUIRequest(baseURL, model, prompt.Input)
	if err != nil {
		return attemptExecution{LastError: err.Error(), Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, codexResponsesEndpoint(baseURL), bytes.NewReader(body))
	if err != nil {
		return attemptExecution{LastError: err.Error(), Err: err}
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			ForceAttemptHTTP2:  true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return attemptExecution{Err: ctx.Err()}
		}
		message := "request failed: " + err.Error()
		return attemptExecution{LastError: message, Err: errors.New(message)}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPResponseBody))
		message := formatHTTPFailure(resp.StatusCode, body)
		if onFirstError != nil {
			onFirstError(message)
		}
		return attemptExecution{FirstError: message, LastError: message, Err: errors.New(message)}
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/event-stream") {
		return consumeCodexSSE(ctx, resp.Body, onFirstError)
	}
	body, err = io.ReadAll(io.LimitReader(resp.Body, maxHTTPResponseBody))
	if err != nil {
		if ctx.Err() != nil {
			return attemptExecution{Err: ctx.Err()}
		}
		message := "read response: " + err.Error()
		return attemptExecution{LastError: message, Err: errors.New(message)}
	}
	if message := responseErrorMessage(body); message != "" {
		if onFirstError != nil {
			onFirstError(message)
		}
		return attemptExecution{FirstError: message, LastError: message, Err: errors.New(message)}
	}
	return attemptExecution{Output: responseOutputText(body)}
}

func codexResponsesEndpoint(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	switch {
	case strings.HasSuffix(base, "/v1/responses"):
		return base
	case strings.HasSuffix(base, "/responses"):
		return base
	case strings.HasSuffix(base, "/v1"):
		return base + "/responses"
	default:
		return base + "/v1/responses"
	}
}

func codexTUIRequest(baseURL, model, input string) ([]byte, map[string]string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, nil, errors.New("codex base URL is empty")
	}
	var template map[string]any
	if err := json.Unmarshal(codexRequestTemplate, &template); err != nil {
		return nil, nil, fmt.Errorf("decode embedded Codex request template: %w", err)
	}
	ids, err := newCodexTUIIDs()
	if err != nil {
		return nil, nil, err
	}
	token, err := readCodexAccessToken()
	if err != nil {
		return nil, nil, err
	}
	template["model"] = model
	template["prompt_cache_key"] = ids.SessionID
	inputItems, _ := template["input"].([]any)
	inputItems = append(inputItems, map[string]any{
		"type": "message",
		"id":   "msg_" + newUUIDv7(),
		"role": "user",
		"content": []any{map[string]any{
			"type": "input_text",
			"text": input,
		}},
	})
	template["input"] = inputItems
	metadata := codexTurnMetadata(template, ids, model)
	template["client_metadata"] = codexClientMetadata(metadata, ids)
	body, err := json.Marshal(template)
	if err != nil {
		return nil, nil, fmt.Errorf("encode Codex request: %w", err)
	}
	headers := map[string]string{
		"Authorization":                          "Bearer " + token,
		"Content-Type":                           "application/json",
		"Accept":                                 "text/event-stream",
		"Originator":                             "codex-tui",
		"User-Agent":                             codexTUIUserAgent(),
		"X-Codex-Beta-Features":                  "remote_compaction_v2",
		"X-Codex-Window-Id":                      ids.WindowID,
		"X-Codex-Turn-Metadata":                  metadata,
		"X-OpenAI-Internal-Codex-Responses-Lite": "true",
		"X-Client-Request-Id":                    ids.ThreadID,
		"Session-Id":                             ids.SessionID,
		"Thread-Id":                              ids.ThreadID,
	}
	return body, headers, nil
}

func codexTUIUserAgent() string {
	return formatCodexTUIUserAgent(codexTUIVersion())
}

func codexTUIVersion() string {
	cliVersionMu.Lock()
	defer cliVersionMu.Unlock()
	return refreshCLIVersion(codexVersionCachePath(), "codex", capturedCodexTUIVersion, time.Now().UTC())
}

func formatCodexTUIUserAgent(version string) string {
	return fmt.Sprintf(codexTUIUserAgentFormat, version, version)
}

func refreshCLIVersion(cachePath, command, fallback string, now time.Time) string {
	cache := readVersionCache(cachePath)
	if cache.Version != "" && !cache.UpdatedAt.IsZero() && now.Sub(cache.UpdatedAt) < cliVersionCacheTTL {
		return cache.Version
	}
	version, err := readCLIVersion(command)
	if err == nil {
		if cachePath != "" {
			_ = writeVersionCache(cachePath, cliVersionCache{Version: version, UpdatedAt: now})
		}
		return version
	}
	if cache.Version != "" {
		return cache.Version
	}
	return fallback
}

func codexVersionCachePath() string {
	if path := strings.TrimSpace(os.Getenv("API_ALIVE_CODEX_VERSION_CACHE")); path != "" {
		return path
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cacheDir, "api-alive", "codex-version.json")
}

func readVersionCache(path string) cliVersionCache {
	if path == "" {
		return cliVersionCache{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cliVersionCache{}
	}
	var cache cliVersionCache
	if json.Unmarshal(data, &cache) != nil {
		return cliVersionCache{}
	}
	cache.Version = strings.TrimSpace(cache.Version)
	return cache
}

func writeVersionCache(path string, cache cliVersionCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".codex-version-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func readCLIVersion(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), codexVersionCommandTimeout)
	defer cancel()
	output, err := exec.CommandContext(ctx, command, "--version").CombinedOutput()
	if err != nil {
		return "", err
	}
	version := parseCLIVersion(string(output))
	if version == "" {
		return "", errors.New("codex --version output did not contain a version")
	}
	return version, nil
}

func parseCLIVersion(output string) string {
	for _, field := range strings.Fields(output) {
		candidate := strings.TrimPrefix(field, "v")
		base := strings.SplitN(candidate, "-", 2)[0]
		parts := strings.Split(base, ".")
		if len(parts) < 2 {
			continue
		}
		valid := true
		for _, part := range parts {
			if _, err := strconv.Atoi(part); err != nil {
				valid = false
				break
			}
		}
		if valid {
			return candidate
		}
	}
	return ""
}

func codexTurnMetadata(template map[string]any, ids codexTUIIDs, model string) string {
	metadata := map[string]any{}
	if clientMetadata, ok := template["client_metadata"].(map[string]any); ok {
		if encoded, ok := clientMetadata["x-codex-turn-metadata"].(string); ok {
			_ = json.Unmarshal([]byte(encoded), &metadata)
		}
	}
	metadata["installation_id"] = ids.InstallationID
	metadata["session_id"] = ids.SessionID
	metadata["thread_id"] = ids.ThreadID
	metadata["agent_name"] = "/root"
	metadata["turn_id"] = ids.TurnID
	metadata["window_id"] = ids.WindowID
	metadata["window_number"] = 0
	metadata["context_window_id"] = ids.ContextWindowID
	metadata["request_kind"] = "turn"
	metadata["history_initialization"] = "cold_resume"
	metadata["root_turn_id"] = ids.RootTurnID
	metadata["thread_source"] = "user"
	metadata["turn_trigger"] = "user"
	metadata["sandbox"] = "seccomp"
	metadata["sandbox_mode"] = "read-only"
	metadata["auto_review_enabled"] = false
	metadata["node_repl_auto_review_required"] = true
	metadata["node_repl_disabled"] = false
	metadata["turn_started_at_unix_ms"] = time.Now().UnixMilli()
	metadata["analytics_enabled"] = true
	metadata["model"] = model
	metadata["reasoning_effort"] = "max"
	body, _ := json.Marshal(metadata)
	return string(body)
}

func codexClientMetadata(turnMetadata string, ids codexTUIIDs) map[string]any {
	return map[string]any{
		"x-codex-installation-id": ids.InstallationID,
		"turn_id":                 ids.TurnID,
		"thread_id":               ids.ThreadID,
		"x-codex-turn-metadata":   turnMetadata,
		"x-codex-window-id":       ids.WindowID,
		"session_id":              ids.SessionID,
		"root_turn_id":            ids.RootTurnID,
	}
}

func newCodexTUIIDs() (codexTUIIDs, error) {
	installationID, err := readInstallationID()
	if err != nil {
		return codexTUIIDs{}, err
	}
	threadID := newUUIDv7()
	turnID := newUUIDv7()
	return codexTUIIDs{
		SessionID:       threadID,
		ThreadID:        threadID,
		TurnID:          turnID,
		RootTurnID:      turnID,
		ContextWindowID: newUUIDv7(),
		WindowID:        threadID + ":0",
		InstallationID:  installationID,
	}, nil
}

func readCodexAccessToken() (string, error) {
	path := os.Getenv("CODEX_AUTH_FILE")
	if path == "" {
		codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME"))
		if codexHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("locate Codex auth file: %w", err)
			}
			codexHome = filepath.Join(home, ".codex")
		}
		path = filepath.Join(codexHome, "auth.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read Codex auth file: %w", err)
	}
	var auth struct {
		APIKey string `json:"OPENAI_API_KEY"`
		Tokens struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(data, &auth); err != nil {
		return "", fmt.Errorf("decode Codex auth file: %w", err)
	}
	token := strings.TrimSpace(auth.APIKey)
	if token == "" {
		token = strings.TrimSpace(auth.Tokens.AccessToken)
	}
	if token == "" {
		return "", errors.New("Codex auth file does not contain an API key or access token")
	}
	return token, nil
}

func readInstallationID() (string, error) {
	codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if codexHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate Codex installation id: %w", err)
		}
		codexHome = filepath.Join(home, ".codex")
	}
	data, err := os.ReadFile(filepath.Join(codexHome, "installation_id"))
	if err != nil {
		return "", fmt.Errorf("read Codex installation id: %w", err)
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", errors.New("Codex installation id is empty")
	}
	return id, nil
}

func consumeCodexSSE(ctx context.Context, body io.Reader, onFirstError func(string)) attemptExecution {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var dataLines []string
	var output strings.Builder
	var firstError, lastError string
	handle := func() (bool, error) {
		if len(dataLines) == 0 {
			return false, nil
		}
		data := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		if data == "[DONE]" {
			return true, nil
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return false, nil
		}
		eventType := firstString(event["type"])
		switch eventType {
		case "response.output_text.delta":
			output.WriteString(firstString(event["delta"]))
		case "response.output_text.done":
			if text := firstString(event["text"]); text != "" {
				output.WriteString(text)
			}
		case "response.completed", "response.done":
			if text := responseOutputTextFromAny(event); text != "" {
				output.Reset()
				output.WriteString(text)
			}
			return true, nil
		case "response.failed", "response.incomplete", "error":
			message := errorMessageFromAny(event)
			if message == "" {
				message = "Codex response failed"
			}
			lastError = message
			if firstError == "" {
				firstError = message
				if onFirstError != nil {
					onFirstError(message)
				}
			}
			return true, errors.New(message)
		}
		return false, nil
	}
	for scanner.Scan() {
		if ctx.Err() != nil {
			return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError, Err: ctx.Err()}
		}
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			continue
		}
		if line != "" {
			continue
		}
		done, err := handle()
		if err != nil {
			return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError, Err: err}
		}
		if done {
			return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError}
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError, Err: ctx.Err()}
		}
		message := "read response stream: " + err.Error()
		return attemptExecution{Output: output.String(), FirstError: firstError, LastError: message, Err: errors.New(message)}
	}
	if ctx.Err() != nil {
		return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError, Err: ctx.Err()}
	}
	done, err := handle()
	if err != nil {
		return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError, Err: err}
	}
	_ = done
	return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError}
}

func formatHTTPFailure(status int, body []byte) string {
	message := responseErrorMessage(body)
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	if message == "" {
		message = http.StatusText(status)
	}
	return truncateText(fmt.Sprintf("HTTP %d: %s", status, message), maxHTTPErrorChars)
}

func responseErrorMessage(body []byte) string {
	var value any
	if json.Unmarshal(body, &value) != nil {
		return ""
	}
	return errorMessageFromAny(value)
}

func errorMessageFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		for _, key := range []string{"error", "response", "detail", "message"} {
			if nested, ok := v[key]; ok {
				if message := errorMessageFromAny(nested); message != "" {
					if code := firstString(v["code"]); code != "" && !strings.Contains(message, code) {
						message += " [" + code + "]"
					}
					return message
				}
			}
		}
	case []any:
		for _, item := range v {
			if message := errorMessageFromAny(item); message != "" {
				return message
			}
		}
	}
	return ""
}

func responseOutputText(body []byte) string {
	var value any
	if json.Unmarshal(body, &value) != nil {
		return ""
	}
	return responseOutputTextFromAny(value)
}

func responseOutputTextFromAny(value any) string {
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	if response, ok := object["response"].(map[string]any); ok {
		if text := responseOutputTextFromAny(response); text != "" {
			return text
		}
	}
	if text := firstString(object["output_text"]); text != "" {
		return text
	}
	if output, ok := object["output"].([]any); ok {
		var builder strings.Builder
		for _, item := range output {
			entry, _ := item.(map[string]any)
			contents, _ := entry["content"].([]any)
			for _, content := range contents {
				part, _ := content.(map[string]any)
				builder.WriteString(firstString(part["text"]))
			}
		}
		return builder.String()
	}
	return ""
}

func firstString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func truncateText(value string, max int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if max <= 0 || len(runes) <= max {
		return value
	}
	if max <= 20 {
		return string(runes[:max])
	}
	return string(runes[:max-20]) + "...<truncated>"
}

func newUUIDv7() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	millis := uint64(time.Now().UnixMilli())
	for i := 0; i < 6; i++ {
		data[5-i] = byte(millis >> (8 * i))
	}
	data[6] = (data[6] & 0x0f) | 0x70
	data[8] = (data[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", data[0:4], data[4:6], data[6:8], data[8:10], data[10:16])
}
