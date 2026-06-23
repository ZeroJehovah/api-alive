package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"test-api-alive/internal/alive"
)

func TestDecodeWSLOutputHandlesUTF16LE(t *testing.T) {
	input := []byte{'U', 0, 'b', 0, 'u', 0, 'n', 0, 't', 0, 'u', 0, '\r', 0, '\n', 0}
	if got, want := decodeWSLOutput(input), "Ubuntu\r\n"; got != want {
		t.Fatalf("decodeWSLOutput() = %q, want %q", got, want)
	}
}

func TestServerStateCreatesDefaultCodexWSLConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	srv.mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}

	var state appState
	if err := json.Unmarshal(res.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Config.Provider != "codex-wsl" {
		t.Fatalf("provider = %q, want codex-wsl", state.Config.Provider)
	}
	if state.Config.WSLCommand != "wsl.exe" {
		t.Fatalf("wsl command = %q, want wsl.exe", state.Config.WSLCommand)
	}
}

func TestServerModelsAPIAddsAndRemovesModels(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := alive.DefaultConfig()
	cfg.Provider = "codex-wsl"
	cfg.Models = []string{"gpt-5"}
	if err := alive.SaveConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	srv := newServer(configPath)

	body := bytes.NewBufferString(`{"models":["gpt-5-mini","gpt-5"]}`)
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/models", body)
	srv.mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("add status = %d, body = %s", res.Code, res.Body.String())
	}

	loaded, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(loaded.Models, ","); got != "gpt-5,gpt-5-mini" {
		t.Fatalf("models after add = %q", got)
	}

	body = bytes.NewBufferString(`{"models":["gpt-5"]}`)
	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/models", body)
	srv.mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", res.Code, res.Body.String())
	}

	loaded, err = alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(loaded.Models, ","); got != "gpt-5-mini" {
		t.Fatalf("models after delete = %q", got)
	}
}
