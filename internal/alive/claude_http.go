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
	"path/filepath"
	"strings"
	"time"
)

//go:embed claude_request_template.json
var claudeRequestTemplate []byte

const (
	capturedClaudeTUIVersion      = "2.1.301"
	capturedClaudeBillingVersion  = "2.1.301.c3d"
	claudeTUIUserAgentFormat      = "claude-cli/%s (external, sdk-cli)"
	claudeAnthropicBeta           = "claude-code-20250219,interleaved-thinking-2025-05-14,thinking-token-count-2026-05-13,context-management-2025-06-27,prompt-caching-scope-2026-01-05,mid-conversation-tool-changes-2026-07-01,advisor-tool-2026-03-01,effort-2025-11-24"
	claudeStainlessPackageVersion = "0.128.0"
	claudeStainlessRuntimeVersion = "v26.3.0"
)

func runClaudeHTTPAttempt(ctx context.Context, baseURL, model string, prompt PromptCase, onFirstError func(string)) attemptExecution {
	body, headers, err := claudeCodeRequest(baseURL, model, prompt.Input)
	if err != nil {
		return attemptExecution{LastError: err.Error(), Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeMessagesEndpoint(baseURL), bytes.NewReader(body))
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
		message := formatClaudeHTTPFailure(resp.StatusCode, body)
		if onFirstError != nil {
			onFirstError(message)
		}
		return attemptExecution{FirstError: message, LastError: message, Err: errors.New(message)}
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/event-stream") {
		return consumeClaudeSSE(ctx, resp.Body, onFirstError)
	}
	body, err = io.ReadAll(io.LimitReader(resp.Body, maxHTTPResponseBody))
	if err != nil {
		if ctx.Err() != nil {
			return attemptExecution{Err: ctx.Err()}
		}
		message := "read response: " + err.Error()
		return attemptExecution{LastError: message, Err: errors.New(message)}
	}
	if message := claudeErrorMessage(body); message != "" {
		if onFirstError != nil {
			onFirstError(message)
		}
		return attemptExecution{FirstError: message, LastError: message, Err: errors.New(message)}
	}
	return attemptExecution{Output: claudeResponseOutputText(body)}
}

func claudeCodeRequest(baseURL, model, input string) ([]byte, map[string]string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, nil, errors.New("Claude base URL is empty")
	}
	var template map[string]any
	if err := json.Unmarshal(claudeRequestTemplate, &template); err != nil {
		return nil, nil, fmt.Errorf("decode embedded Claude request template: %w", err)
	}
	token, err := readClaudeAccessToken()
	if err != nil {
		return nil, nil, err
	}
	sessionID := newUUIDv4()
	deviceID := readClaudeUserID()
	version := claudeTUIVersion()

	template["model"] = model
	updateClaudeBillingVersion(template, version)
	if err := appendClaudePrompt(template, input); err != nil {
		return nil, nil, err
	}
	if err := updateClaudeMetadata(template, deviceID, sessionID); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(template)
	if err != nil {
		return nil, nil, fmt.Errorf("encode Claude request: %w", err)
	}
	headers := map[string]string{
		"Authorization":                             claudeAuthorization(token),
		"Accept":                                    "application/json",
		"Content-Type":                              "application/json",
		"User-Agent":                                fmt.Sprintf(claudeTUIUserAgentFormat, version),
		"X-Claude-Code-Session-Id":                  sessionID,
		"X-Stainless-Arch":                          "arm64",
		"X-Stainless-Lang":                          "js",
		"X-Stainless-OS":                            "Linux",
		"X-Stainless-Package-Version":               claudeStainlessPackageVersion,
		"X-Stainless-Retry-Count":                   "0",
		"X-Stainless-Runtime":                       "node",
		"X-Stainless-Runtime-Version":               claudeStainlessRuntimeVersion,
		"X-Stainless-Timeout":                       "600",
		"Anthropic-Beta":                            claudeAnthropicBeta,
		"Anthropic-Dangerous-Direct-Browser-Access": "true",
		"Anthropic-Version":                         "2023-06-01",
		"X-App":                                     "cli",
	}
	return body, headers, nil
}

func claudeMessagesEndpoint(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	switch {
	case strings.HasSuffix(base, "/v1/messages"):
		return base + "?beta=true"
	case strings.HasSuffix(base, "/v1"):
		return base + "/messages?beta=true"
	default:
		return base + "/v1/messages?beta=true"
	}
}

func appendClaudePrompt(template map[string]any, input string) error {
	messages, ok := template["messages"].([]any)
	if !ok || len(messages) == 0 {
		return errors.New("Claude request template has no messages")
	}
	message, ok := messages[0].(map[string]any)
	if !ok {
		return errors.New("Claude request template first message is invalid")
	}
	content, ok := message["content"].([]any)
	if !ok {
		return errors.New("Claude request template first message has no content")
	}
	message["content"] = append(content, map[string]any{"type": "text", "text": input})
	return nil
}

func updateClaudeMetadata(template map[string]any, deviceID, sessionID string) error {
	metadata, ok := template["metadata"].(map[string]any)
	if !ok {
		return errors.New("Claude request template has no metadata")
	}
	userID, err := json.Marshal(struct {
		DeviceID    string `json:"device_id"`
		AccountUUID string `json:"account_uuid"`
		SessionID   string `json:"session_id"`
	}{DeviceID: deviceID, SessionID: sessionID})
	if err != nil {
		return err
	}
	metadata["user_id"] = string(userID)
	return nil
}

func updateClaudeBillingVersion(template map[string]any, version string) {
	system, ok := template["system"].([]any)
	if !ok || len(system) == 0 {
		return
	}
	block, ok := system[0].(map[string]any)
	if !ok {
		return
	}
	text := firstString(block["text"])
	const prefix = "x-anthropic-billing-header: cc_version="
	if !strings.HasPrefix(text, prefix) {
		return
	}
	end := strings.IndexByte(text[len(prefix):], ';')
	if end < 0 {
		return
	}
	billingVersion := version
	if version == capturedClaudeTUIVersion {
		billingVersion = capturedClaudeBillingVersion
	}
	block["text"] = prefix + billingVersion + text[len(prefix)+end:]
}

func claudeTUIVersion() string {
	cliVersionMu.Lock()
	defer cliVersionMu.Unlock()
	return refreshCLIVersion(claudeVersionCachePath(), "claude", capturedClaudeTUIVersion, time.Now().UTC())
}

func claudeVersionCachePath() string {
	if path := strings.TrimSpace(os.Getenv("API_ALIVE_CLAUDE_VERSION_CACHE")); path != "" {
		return path
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cacheDir, "api-alive", "claude-version.json")
}

func readClaudeAccessToken() (string, error) {
	for _, key := range []string{"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY"} {
		if token := strings.TrimSpace(os.Getenv(key)); token != "" {
			return token, nil
		}
	}
	configDir := strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR"))
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate Claude settings: %w", err)
		}
		configDir = filepath.Join(home, ".claude")
	}
	data, err := os.ReadFile(filepath.Join(configDir, "settings.json"))
	if err != nil {
		return "", fmt.Errorf("read Claude settings: %w", err)
	}
	var settings struct {
		Env struct {
			AuthToken string `json:"ANTHROPIC_AUTH_TOKEN"`
			APIKey    string `json:"ANTHROPIC_API_KEY"`
		} `json:"env"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", fmt.Errorf("decode Claude settings: %w", err)
	}
	token := strings.TrimSpace(settings.Env.AuthToken)
	if token == "" {
		token = strings.TrimSpace(settings.Env.APIKey)
	}
	if token == "" {
		return "", errors.New("Claude settings do not contain ANTHROPIC_AUTH_TOKEN or ANTHROPIC_API_KEY")
	}
	return token, nil
}

func claudeAuthorization(token string) string {
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token
	}
	return "Bearer " + token
}

func readClaudeUserID() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return ""
	}
	var config struct {
		UserID string `json:"userID"`
	}
	if json.Unmarshal(data, &config) != nil {
		return ""
	}
	return strings.TrimSpace(config.UserID)
}

func consumeClaudeSSE(ctx context.Context, body io.Reader, onFirstError func(string)) attemptExecution {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var dataLines []string
	var output strings.Builder
	var preview strings.Builder
	sawData := false
	var firstError, lastError string
	handle := func() (bool, error) {
		if len(dataLines) == 0 {
			return false, nil
		}
		data := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		var event map[string]any
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return false, nil
		}
		switch firstString(event["type"]) {
		case "content_block_start":
			if block, ok := event["content_block"].(map[string]any); ok && firstString(block["type"]) == "text" {
				output.WriteString(firstString(block["text"]))
			}
		case "content_block_delta":
			if delta, ok := event["delta"].(map[string]any); ok {
				output.WriteString(firstString(delta["text"]))
			}
		case "message_stop":
			return true, nil
		case "error":
			message := claudeErrorFromValue(event)
			if message == "" {
				message = "Claude response failed"
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
			sawData = true
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			continue
		}
		if line != "" && !strings.HasPrefix(line, "event:") && !strings.HasPrefix(line, ":") {
			if preview.Len() < 500 {
				preview.WriteString(strings.TrimSpace(line))
				preview.WriteByte(' ')
			}
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
	if !sawData {
		message := "response was not SSE"
		if text := strings.TrimSpace(preview.String()); text != "" {
			message += ": " + truncateText(text, 300)
		}
		if onFirstError != nil {
			onFirstError(message)
		}
		return attemptExecution{Output: output.String(), FirstError: message, LastError: message, Err: errors.New(message)}
	}
	return attemptExecution{Output: output.String(), FirstError: firstError, LastError: lastError}
}

func formatClaudeHTTPFailure(status int, body []byte) string {
	message := claudeErrorMessage(body)
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	if message == "" {
		message = http.StatusText(status)
	}
	return truncateText(fmt.Sprintf("HTTP %d: %s", status, message), maxHTTPErrorChars)
}

func claudeErrorMessage(body []byte) string {
	var value any
	if json.Unmarshal(body, &value) != nil {
		return ""
	}
	return claudeErrorFromValue(value)
}

func claudeErrorFromValue(value any) string {
	object, ok := value.(map[string]any)
	if !ok {
		return errorMessageFromAny(value)
	}
	if nested, ok := object["error"]; ok {
		message := errorMessageFromAny(nested)
		if message == "" {
			return ""
		}
		if errorObject, ok := nested.(map[string]any); ok {
			for _, key := range []string{"type", "code"} {
				if detail := firstString(errorObject[key]); detail != "" && !strings.Contains(message, detail) {
					message += " [" + detail + "]"
				}
			}
		}
		return message
	}
	return errorMessageFromAny(value)
}

func claudeResponseOutputText(body []byte) string {
	var value any
	if json.Unmarshal(body, &value) != nil {
		return ""
	}
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	content, ok := object["content"].([]any)
	if !ok {
		return ""
	}
	var output strings.Builder
	for _, item := range content {
		block, _ := item.(map[string]any)
		if firstString(block["type"]) == "text" {
			output.WriteString(firstString(block["text"]))
		}
	}
	return output.String()
}

func newUUIDv4() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", data[0:4], data[4:6], data[6:8], data[8:10], data[10:16])
}
