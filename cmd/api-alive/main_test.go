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
	"time"

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
		`<th>Time</th><th>Status</th><th>Model</th><th>Try</th><th>Seconds</th><th>Error</th>`,
		`function displayDateTime`,
		`displayDateTime(res.updated_at)`,
		`<th class="result-time">Result time</th>`,
		`<th class="progress">Progress</th>`,
		`function activeLoopCountFor`,
		`function displayProgress`,
		`function displaySeconds`,
		`Success (' + escapeText(displaySeconds(res.duration_ms)) + 's)`,
		`idlePollMS = 60000`,
		`runningPollMS = 5000`,
		`id="notificationBtn"`,
		`Allow alerts`,
		`class="alert-button alert-request"`,
		`.alert-button.alert-blocked`,
		`Notifications are blocked. Allow them in browser site settings, then refresh.`,
		`successNotificationsPrimed: false`,
		`notifiedSuccessLogKeys: new Set()`,
		`function updateNotificationButton`,
		`async function requestNotificationPermission`,
		`async function autoRequestNotificationPermission`,
		`function installNotificationPermissionPrompt`,
		`notificationPermission() !== 'default'`,
		`document.addEventListener('pointerdown', requestAfterGesture`,
		`installNotificationPermissionPrompt();`,
		`function handleSuccessNotifications`,
		`function notifySuccessEntries`,
		`function isPageForeground`,
		`document.visibilityState === 'visible' && document.hasFocus()`,
		`isPageForeground()`,
		`API Alive: ' + entries.length + ' successes`,
		`New successes: ' + models.join(', ')`,
		`requireInteraction: true`,
		`new Notification(title`,
		`async function startProbe`,
		`function runnableSelectedModels`,
		`return [...state.selected].filter(model => model);`,
		`btn.disabled = state.editing || !state.runningModels.has(btn.dataset.stopOne)`,
		`async function stopProbe`,
		`id="clearResultsBtn"`,
		`data-stop-one`,
		`data-select-all`,
		`Clear results`,
		`/api/probe/results/clear`,
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
		`.add-form input, .add-form button`,
		`flex: 1 1 320px`,
		`editing: false`,
		`draftModels: []`,
		`draftModelLoopCounts: {}`,
		`dirtyLoopCounts: new Set()`,
		`editLockedModels: new Set()`,
		`state.editLockedModels = new Set(state.runningModels)`,
		`state.editLockedModels.has(model)`,
		`Running when editing started`,
		`function loopCountInput`,
		`function mergeDirtyLoopCounts`,
		`state.config = mergeDirtyLoopCounts(data.config);`,
		`state.dirtyLoopCounts.add(model);`,
		`replaceChildrenWithHTML(resultCell, statusPill(model));`,
		`document.activeElement !== retryInput`,
		`clearDirtyLoopCounts(models);`,
		`maxlength="2"`,
		`loopCountLabel`,
		`displayAttemptProgress`,
		`.loop-count-input`,
		`model_loop_counts`,
		`<table class="model-table">`,
		`<table class="model-table model-editor">`,
		`.model-table .model-heading { width: 30ch; }`,
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

func TestIndexHTMLAllowsRetryInputWhileRunning(t *testing.T) {
	if strings.Contains(indexHTML, `state.runningModels.has(input.dataset.loopCount)`) {
		t.Fatal("running models must not disable retries input")
	}
}

func TestIndexHTMLMovesNotificationsToLogHeaderAndRemovesGlobalStop(t *testing.T) {
	runtimeHeader := `<h2>Runtime</h2>
          <div class="header-actions">
            <button class="secondary" id="reloadBtn">Refresh</button>
            <button id="saveConfigBtn">Save runtime</button>`
	if !strings.Contains(indexHTML, runtimeHeader) {
		t.Fatal("runtime header should contain refresh and save runtime only")
	}
	logHeader := `<h2>Log</h2>
        </div>
        <div class="header-actions">
          <button class="alert-button alert-request" id="notificationBtn" type="button">Allow alerts</button>`
	if !strings.Contains(indexHTML, logHeader) {
		t.Fatal("notification button should be in the log header")
	}
	if strings.Contains(indexHTML, `id="stopProbeBtn"`) || strings.Contains(indexHTML, `Stop task`) {
		t.Fatal("global stop task control must be removed")
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
	body := bytes.NewBufferString(`{"models":["gpt-5"],"model_loop_counts":{"gpt-5":2},"timeout_seconds":30,"codex_command":"codex-beta","listen_addr":"127.0.0.1:0","max_output_chars":1234}`)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/config", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	cfg, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TimeoutSeconds != 30 || cfg.ModelLoopCounts["gpt-5"] != 2 || cfg.CodexCommand != "codex-beta" || cfg.ListenAddr != "127.0.0.1:0" || cfg.MaxOutputChars != 1234 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestConfigEndpointPreservesModelOrder(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)
	body := bytes.NewBufferString(`{"models":["model-b","model-a","model-b","model-c"],"model_loop_counts":{"model-b":2,"model-a":3,"model-c":4},"timeout_seconds":30,"codex_command":"codex","listen_addr":"127.0.0.1:0","max_output_chars":4000}`)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/config", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	assertConfigModels(t, configPath, []string{"model-b", "model-a", "model-c"})
}

func TestProbeTaskLifecyclePersistsStateAndAllowsConcurrentDistinctModels(t *testing.T) {
	store := &taskStore{}
	task, runs, err := store.start([]string{"model-a", "model-b"}, map[string]int{"model-a": 2, "model-b": 2})
	if err != nil {
		t.Fatal(err)
	}
	if !task.Running || task.ID == 0 || len(task.RunningModels) != 2 {
		t.Fatalf("started task = %#v", task)
	}
	if len(task.Results) != 2 || task.Results[0].Attempts != 1 || task.Results[1].Attempts != 1 {
		t.Fatalf("initial results = %#v", task.Results)
	}
	ctx := runCtxByModel(t, runs, "model-b")
	secondTask, secondRuns, err := store.start([]string{"model-c"}, map[string]int{"model-c": 1})
	if err != nil {
		t.Fatalf("distinct concurrent start failed: %v", err)
	}
	if secondTask.ID != task.ID {
		t.Fatalf("concurrent start created task id %d, want %d", secondTask.ID, task.ID)
	}
	if len(secondTask.RunningModels) != 3 {
		t.Fatalf("running models after concurrent start = %#v", secondTask.RunningModels)
	}
	modelBFirstCtx := runCtxByModel(t, runs, "model-b")
	restartedTask, restartedRuns, err := store.start([]string{"model-b"}, map[string]int{"model-b": 2})
	if err != nil {
		t.Fatalf("restart running model failed: %v", err)
	}
	if got := resultByModel(restartedTask.Results, "model-b").Attempts; got != 1 {
		t.Fatalf("restarted model attempts = %d, want 1", got)
	}
	select {
	case <-modelBFirstCtx:
	default:
		t.Fatal("restart did not cancel previous model-b context")
	}
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-b"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-b", Attempts: 20, Success: false, DurationMS: 200}})
	if got := resultByModel(store.snapshot().Results, "model-b").Attempts; got != 1 {
		t.Fatalf("stale model-b event changed attempts to %d, want 1", got)
	}
	runs = append(runs, restartedRuns...)
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}})
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-a"), alive.Event{Type: alive.EventResult, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10, AttemptResults: []alive.Result{{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}}}})
	store.finishRun(task.ID, runIDByModel(t, runs, "model-a"), "model-a")
	snap := store.snapshot()
	modelAResult := resultByModel(snap.Results, "model-a")
	if !snap.Running || len(snap.Logs) != 1 || modelAResult.Model != "model-a" {
		t.Fatalf("snapshot after event = %#v", snap)
	}
	if _, err := time.Parse(time.RFC3339, modelAResult.UpdatedAt); err != nil {
		t.Fatalf("updated_at = %q, parse error = %v", modelAResult.UpdatedAt, err)
	}
	if !reflect.DeepEqual(snap.RunningModels, []string{"model-b", "model-c"}) {
		t.Fatalf("running models = %#v", snap.RunningModels)
	}
	stopping, err := store.stopModels([]string{"model-b"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx:
	default:
		t.Fatal("stop did not cancel selected model-b context")
	}
	select {
	case <-runCtxByModel(t, secondRuns, "model-c"):
		t.Fatal("stop canceled unselected model-c context")
	default:
	}
	if !reflect.DeepEqual(stopping.RunningModels, []string{"model-b", "model-c"}) {
		t.Fatalf("running models before canceled run finishes = %#v", stopping.RunningModels)
	}
	store.finishRun(task.ID, runIDByModel(t, runs, "model-b"), "model-b")
	if !store.snapshot().Running {
		t.Fatal("task stopped before all active runs finished")
	}
	store.finishRun(secondTask.ID, runIDByModel(t, secondRuns, "model-c"), "model-c")
	if store.snapshot().Running {
		t.Fatal("task still running after all active runs finished")
	}
	if _, _, err := store.start([]string{"model-c"}, map[string]int{"model-c": 2}); err != nil {
		t.Fatalf("start after finish failed: %v", err)
	}
}

func TestProbeTaskJSONIncludesLoopCounts(t *testing.T) {
	store := &taskStore{}
	task, _, err := store.start([]string{"model-a"}, map[string]int{"model-a": 3})
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"loop_counts":{"model-a":3}`) {
		t.Fatalf("task JSON missing loop_counts: %s", body)
	}
}

func TestProbeTaskAttemptEventsUpdateDisplayedAttempt(t *testing.T) {
	store := &taskStore{}
	task, runs, err := store.start([]string{"model-a"}, map[string]int{"model-a": 3})
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: false, DurationMS: 10}})
	if got := resultByModel(store.snapshot().Results, "model-a").Attempts; got != 2 {
		t.Fatalf("attempt after first failure = %d, want 2", got)
	}
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 3, Success: false, DurationMS: 30}})
	if got := resultByModel(store.snapshot().Results, "model-a").Attempts; got != 3 {
		t.Fatalf("attempt after final failure = %d, want 3", got)
	}
}

func TestProbeTaskUnlimitedAttemptEventsAdvanceDisplayedAttempt(t *testing.T) {
	store := &taskStore{}
	task, runs, err := store.start([]string{"model-a"}, map[string]int{"model-a": 0})
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: false, DurationMS: 10}})
	if got := resultByModel(store.snapshot().Results, "model-a").Attempts; got != 2 {
		t.Fatalf("unlimited attempt after first failure = %d, want 2", got)
	}
	if got := store.snapshot().Logs[0].LoopCount; got != 0 {
		t.Fatalf("log loop count = %d, want 0", got)
	}
}

func TestProbeTaskClearResultsKeepsRunningModels(t *testing.T) {
	store := &taskStore{}
	first, firstRuns, err := store.start([]string{"model-a"}, map[string]int{"model-a": 1})
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(first.ID, runIDByModel(t, firstRuns, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}})
	store.finishRun(first.ID, runIDByModel(t, firstRuns, "model-a"), "model-a")

	second, _, err := store.start([]string{"model-b"}, map[string]int{"model-b": 2})
	if err != nil {
		t.Fatal(err)
	}
	snap := store.clearResults()
	if resultByModel(snap.Results, "model-a").Model != "" {
		t.Fatalf("cleared finished model result: %#v", snap.Results)
	}
	if got := resultByModel(snap.Results, "model-b"); got.Model != "model-b" || got.Attempts != 1 {
		t.Fatalf("running model result not kept after clear, task %d: %#v", second.ID, snap.Results)
	}
}

func TestProbeTaskStartKeepsPreviousLogsAndUnselectedResults(t *testing.T) {
	store := &taskStore{}
	first, firstRuns, err := store.start([]string{"model-a"}, map[string]int{"model-a": 1})
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(first.ID, runIDByModel(t, firstRuns, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}})
	store.applyEvent(first.ID, runIDByModel(t, firstRuns, "model-a"), alive.Event{Type: alive.EventResult, Result: alive.Result{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10, AttemptResults: []alive.Result{{Model: "model-a", Attempts: 1, Success: true, DurationMS: 10}}}})
	store.finishRun(first.ID, runIDByModel(t, firstRuns, "model-a"), "model-a")

	second, secondRuns, err := store.start([]string{"model-b"}, map[string]int{"model-b": 1})
	if err != nil {
		t.Fatal(err)
	}
	snap := store.snapshot()
	if len(snap.Logs) != 1 || snap.Logs[0].Result.Model != "model-a" {
		t.Fatalf("logs after second start = %#v", snap.Logs)
	}
	modelAResult := resultByModel(snap.Results, "model-a")
	if !modelAResult.Success || modelAResult.DurationMS != 10 {
		t.Fatalf("unselected model result was not preserved: %#v", snap.Results)
	}
	modelBResult := resultByModel(snap.Results, "model-b")
	if modelBResult.Model != "model-b" || modelBResult.Attempts != 1 || modelBResult.Success {
		t.Fatalf("selected model did not get initial running result: %#v", snap.Results)
	}

	store.applyEvent(second.ID, runIDByModel(t, secondRuns, "model-b"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-b", Attempts: 1, Success: false, DurationMS: 20}})
	snap = store.snapshot()
	if len(snap.Logs) != 2 || snap.Logs[0].Result.Model != "model-b" || snap.Logs[1].Result.Model != "model-a" {
		t.Fatalf("logs after second event = %#v", snap.Logs)
	}
}

func TestProbeEndpointAllowsDistinctConcurrentRequestsRejectsDuplicateAndStops(t *testing.T) {
	dir := t.TempDir()
	commandPath := filepath.Join(dir, "fake-codex")
	if err := os.WriteFile(commandPath, []byte("#!/bin/sh\nsleep 2\necho OK\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config.json")
	cfg := alive.DefaultConfig()
	cfg.Models = []string{"model-a", "model-b"}
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
	body = strings.NewReader(`{"models":["model-b"]}`)
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("distinct concurrent probe status = %d, body = %q", rec.Code, rec.Body.String())
	}
	body = strings.NewReader(`{"models":["model-a"]}`)
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("duplicate restart probe status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var restarted probeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &restarted); err != nil {
		t.Fatal(err)
	}
	if got := resultByModel(restarted.Task.Results, "model-a").Attempts; got != 1 {
		t.Fatalf("restarted model attempts = %d, want 1", got)
	}
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe/stop", strings.NewReader(`{"model":"model-a"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("stop status = %d, body = %q", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe/stop", strings.NewReader(`{}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty stop status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe/results/clear", strings.NewReader(`{}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("clear results status = %d, body = %q", rec.Code, rec.Body.String())
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

func runIDByModel(t *testing.T, runs []modelRun, model string) int64 {
	t.Helper()
	for i := len(runs) - 1; i >= 0; i-- {
		if runs[i].Model == model {
			return runs[i].ID
		}
	}
	t.Fatalf("run for model %s not found in %#v", model, runs)
	return 0
}

func runCtxByModel(t *testing.T, runs []modelRun, model string) <-chan struct{} {
	t.Helper()
	for i := len(runs) - 1; i >= 0; i-- {
		if runs[i].Model == model {
			return runs[i].Ctx.Done()
		}
	}
	t.Fatalf("run for model %s not found in %#v", model, runs)
	return nil
}

func resultByModel(results []alive.Result, model string) alive.Result {
	for _, res := range results {
		if res.Model == model {
			return res
		}
	}
	return alive.Result{}
}
