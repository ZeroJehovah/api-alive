package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"api-alive/internal/alive"
)

func TestIndexHTMLContainsModelOrderAndStandaloneLogPanel(t *testing.T) {
	for _, want := range []string{
		`class="panel log-panel"`,
		`<table class="log-table">`,
		`<tbody id="logList"></tbody>`,
		`maxLogEntries = 100`,
		`text-overflow: ellipsis`,
		`.log-model-col { width: 30ch; }`,
		`<td class="log-error" title="' + escapeText(errorText) + '">`,
		`<th>Status</th><th>Model</th><th>Try</th><th>Seconds</th><th>Time</th><th>Error</th>`,
		`idlePollMS = 60000`,
		`runningPollMS = 5000`,
		`async function startProbe`,
		`async function stopProbe`,
		`id="stopProbeBtn"`,
		`class="header-actions"`,
		`class="header-actions models-actions"`,
		`.models-actions { flex: 1 1 auto; min-width: 0; }`,
		`.settings input, .settings button`,
		`min-height: 52px`,
		`class="live-dot"`,
		`blue-breathe`,
		`<span class="pill run"><span class="live-dot"></span>Running</span>`,
		`id="editModelsBtn"`,
		`id="cancelEditBtn"`,
		`id="selectAllLabel"`,
		`.select-all-label`,
		`.add-form input, .add-form button`,
		`flex: 1 1 320px`,
		`editing: false`,
		`draftModels: []`,
		`<table class="model-table">`,
		`<table class="model-table model-editor">`,
		`.model-table .model-heading { width: 34%; }`,
		`.model-table input[type="checkbox"]`,
		`data-move="up"`,
		`data-move="down"`,
		`data-delete-model`,
		`async function toggleModelEdit`,
		`function cancelModelEdit`,
		`async function moveModel`,
	} {
		if !strings.Contains(indexHTML, want) {
			t.Fatalf("indexHTML missing %q", want)
		}
	}
}

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

func TestConfigEndpointPreservesModelOrder(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)
	body := bytes.NewBufferString(`{"models":["model-b","model-a","model-b","model-c"],"timeout_seconds":30,"loop_count":1,"codex_command":"codex","listen_addr":"127.0.0.1:0","max_output_chars":4000}`)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/config", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	assertConfigModels(t, configPath, []string{"model-b", "model-a", "model-c"})
}

func TestProbeTaskLifecyclePersistsStateAndRejectsConcurrentStart(t *testing.T) {
	store := &taskStore{}
	task, ctx, err := store.start([]string{"model-a", "model-b"}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !task.Running || task.ID == 0 || len(task.RunningModels) != 2 {
		t.Fatalf("started task = %#v", task)
	}
	if len(task.Results) != 2 || task.Results[0].Attempts != 1 || task.Results[1].Attempts != 1 {
		t.Fatalf("initial results = %#v", task.Results)
	}
	if _, _, err := store.start([]string{"model-c"}, 2); err == nil {
		t.Fatal("second start succeeded while task was running")
	}
	store.applyEvent(task.ID, alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}})
	store.applyEvent(task.ID, alive.Event{Type: alive.EventResult, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10, AttemptResults: []alive.Result{{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}}}})
	snap := store.snapshot()
	if !snap.Running || len(snap.Logs) != 1 || resultByModel(snap.Results, "model-a").Model != "model-a" {
		t.Fatalf("snapshot after event = %#v", snap)
	}
	if len(snap.RunningModels) != 1 || snap.RunningModels[0] != "model-b" {
		t.Fatalf("running models = %#v", snap.RunningModels)
	}
	stopping, err := store.stop()
	if err != nil {
		t.Fatal(err)
	}
	if !stopping.Stopping {
		t.Fatalf("stop did not mark stopping: %#v", stopping)
	}
	select {
	case <-ctx.Done():
	default:
		t.Fatal("stop did not cancel task context")
	}
	store.finish(task.ID)
	if store.snapshot().Running {
		t.Fatal("task still running after finish")
	}
	if _, _, err := store.start([]string{"model-c"}, 2); err != nil {
		t.Fatalf("start after finish failed: %v", err)
	}
}

func TestProbeTaskAttemptEventsUpdateDisplayedAttempt(t *testing.T) {
	store := &taskStore{}
	task, _, err := store.start([]string{"model-a"}, 3)
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(task.ID, alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: false, DurationMS: 10}})
	if got := resultByModel(store.snapshot().Results, "model-a").Attempts; got != 2 {
		t.Fatalf("attempt after first failure = %d, want 2", got)
	}
	store.applyEvent(task.ID, alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 3, Success: false, DurationMS: 30}})
	if got := resultByModel(store.snapshot().Results, "model-a").Attempts; got != 3 {
		t.Fatalf("attempt after final failure = %d, want 3", got)
	}
}

func TestProbeTaskStartKeepsPreviousLogs(t *testing.T) {
	store := &taskStore{}
	first, _, err := store.start([]string{"model-a"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(first.ID, alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}})
	store.finish(first.ID)

	second, _, err := store.start([]string{"model-b"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	snap := store.snapshot()
	if len(snap.Logs) != 1 || snap.Logs[0].Result.Model != "model-a" {
		t.Fatalf("logs after second start = %#v", snap.Logs)
	}

	store.applyEvent(second.ID, alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-b", Attempts: 1, Success: false, DurationMS: 20}})
	snap = store.snapshot()
	if len(snap.Logs) != 2 || snap.Logs[0].Result.Model != "model-b" || snap.Logs[1].Result.Model != "model-a" {
		t.Fatalf("logs after second event = %#v", snap.Logs)
	}
}

func TestProbeEndpointStartsTaskRejectsConcurrentRequestAndStops(t *testing.T) {
	dir := t.TempDir()
	commandPath := filepath.Join(dir, "fake-codex")
	if err := os.WriteFile(commandPath, []byte("#!/bin/sh\nsleep 2\necho OK\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config.json")
	cfg := alive.DefaultConfig()
	cfg.Models = []string{"model-a"}
	cfg.CodexCommand = commandPath
	if err := alive.SaveConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	srv := newServer(configPath)
	body := strings.NewReader(`{"models":["model-a"]}`)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("first probe status = %d, body = %q", rec.Code, rec.Body.String())
	}
	body = strings.NewReader(`{"models":["model-a"]}`)
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe", body))
	if rec.Code != http.StatusConflict {
		t.Fatalf("second probe status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe/stop", strings.NewReader(`{}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("stop status = %d, body = %q", rec.Code, rec.Body.String())
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

func resultByModel(results []alive.Result, model string) alive.Result {
	for _, res := range results {
		if res.Model == model {
			return res
		}
	}
	return alive.Result{}
}
