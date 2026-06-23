package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"test-api-alive/internal/alive"
)

func TestStateCreatesDefaultConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/state", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	var state appState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Config.CodexCommand != "codex" {
		t.Fatalf("codex command = %q", state.Config.CodexCommand)
	}
	if state.Config.ListenAddr != "0.0.0.0:8080" {
		t.Fatalf("listen addr = %q", state.Config.ListenAddr)
	}
}

func TestModelsEndpointAddsAndDeletesModels(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)

	postBody := strings.NewReader(`{"models":["gpt-5"," gpt-5-mini ","gpt-5"]}`)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/models", postBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("add status = %d, body = %q", rec.Code, rec.Body.String())
	}
	assertConfigModels(t, configPath, []string{"gpt-5", "gpt-5-mini"})

	deleteBody := strings.NewReader(`{"models":["gpt-5"]}`)
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/models", deleteBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %q", rec.Code, rec.Body.String())
	}
	assertConfigModels(t, configPath, []string{"gpt-5-mini"})
}

func TestConfigEndpointUpdatesRuntime(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)
	body := bytes.NewBufferString(`{"models":["gpt-5"],"timeout_seconds":30,"loop_count":2,"codex_command":"codex-beta","listen_addr":"127.0.0.1:0","max_output_chars":1234}`)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/config", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	cfg, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TimeoutSeconds != 30 || cfg.LoopCount != 2 || cfg.CodexCommand != "codex-beta" || cfg.ListenAddr != "127.0.0.1:0" || cfg.MaxOutputChars != 1234 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestRunRejectsOldCLIArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"list"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unexpected argument") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func assertConfigModels(t *testing.T, path string, want []string) {
	t.Helper()
	cfg, err := alive.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Models, want) {
		t.Fatalf("models = %#v, want %#v", cfg.Models, want)
	}
}
