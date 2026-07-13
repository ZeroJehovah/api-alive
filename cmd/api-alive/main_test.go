package main

import (
	"bytes"
	"context"
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
		`.log-status-col { width: 112px; }`,
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
		`const status = res.success ? 'success' : (isCanceledResult(res) ? 'canceled' : 'failed');`,
		`return '<span class="pill bad">' + (isCanceledResult(res) ? 'Canceled' : 'Failed') + '</span>';`,
		`function isCanceledResult`,
		`idlePollMS = 60000`,
		`runningPollMS = 5000`,
		`let taskEventSource = null`,
		`function connectTaskStream`,
		`new EventSource('/api/events')`,
		`acceptServerTask(JSON.parse(event.data) || {})`,
		`lastServerTask: {}`,
		`lastServerRevision: 0`,
		`function acceptServerTask`,
		`revision < state.lastServerRevision`,
		`applyTask(clientTaskCopy(state.lastServerTask));`,
		`connectTaskStream();`,
		`window.addEventListener('beforeunload', closeTaskStream);`,
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
		`<link id="favicon" rel="icon" type="image/svg+xml">`,
		`backgroundSuccessFavicon: false`,
		`failureFaviconPrimed: false`,
		`backgroundFailedFavicon: false`,
		`handledFailureFaviconKey: ''`,
		`faviconStatus: ''`,
		`const faviconColors = {`,
		`idle: { fill: '#94a3b8', highlight: '#cbd5e1' }`,
		`running: { fill: '#2563eb', highlight: '#93c5fd' }`,
		`success: { fill: '#22c55e', highlight: '#86efac' }`,
		`failed: { fill: '#ef4444', highlight: '#fca5a5' }`,
		`function faviconHref`,
		`<rect x="0" y="0" width="32" height="32" rx="7" fill="#1e293b"/>`,
		`<circle cx="16" cy="16" r="9" fill="' + color.fill + '"/>`,
		`<circle cx="13.5" cy="13.5" r="3" fill="' + color.highlight + '" opacity="0.8"/>`,
		`function latestTaskResult`,
		`function taskFinishedWithFailure`,
		`function failureFaviconKey`,
		`function desiredFaviconStatus`,
		`function updateFavicon`,
		`function handlePageAttentionChange`,
		`if (isPageForeground()) return state.running ? 'running' : 'idle';`,
		`if (state.backgroundSuccessFavicon) return 'success';`,
		`if (state.running) return 'running';`,
		`if (state.backgroundFailedFavicon) return 'failed';`,
		`state.backgroundSuccessFavicon = false;`,
		`state.backgroundFailedFavicon = false;`,
		`if (fresh.length && !isPageForeground()) state.backgroundSuccessFavicon = true;`,
		`function handleFailureFavicon`,
		`if (!state.failureFaviconPrimed)`,
		`state.failureFaviconPrimed = true;`,
		`if (!isPageForeground()) state.backgroundFailedFavicon = true;`,
		`handleFailureFavicon(task || {});`,
		`document.addEventListener('visibilitychange', handlePageAttentionChange);`,
		`window.addEventListener('focus', handlePageAttentionChange);`,
		`window.addEventListener('blur', handlePageAttentionChange);`,
		`isPageForeground()`,
		`API Alive: ' + entries.length + ' successes`,
		`New successes: ' + models.join(', ')`,
		`requireInteraction: true`,
		`new Notification(title`,
		`function orderedModelNames`,
		`function optimisticStartProbe`,
		`const runningModels = orderedModelNames([...state.runningModels, ...models]);`,
		`const result = { model, attempts: 1 };`,
		`state.optimisticStarts.set(model, {`,
		`state.task = { ...currentTask, running: true`,
		`optimisticStartProbe(models, loopCounts);`,
		`setMessage('Starting ' + models.length + ' model(s)...');`,
		`await pollState().catch(refreshErr => setMessage(refreshErr.message));`,
		`async function startProbe`,
		`localTrustWindowMS = 12000`,
		`function hasLocalTrust`,
		`state.running || hasLocalTrust() ? runningPollMS : idlePollMS`,
		`optimisticStarts: new Map()`,
		`optimisticStops: new Map()`,
		`localStopLogs: new Map()`,
		`function updateModelHeaderControls`,
		`function reconcileOptimisticStarts`,
		`function snapshotClientState`,
		`function restoreClientState`,
		`function canceledResultFor`,
		`error: 'context canceled'`,
		`function reconcileOptimisticStops`,
		`state.localStopLogs.forEach((entry, model) =>`,
		`state.localStopLogs.delete(model);`,
		`expires_at: expiresAt`,
		`next.running_models = next.running_models.filter(runningModel => runningModel !== model);`,
		`function optimisticStopProbe`,
		`function optimisticStopProbes`,
		`state.optimisticStops.set(model, entry);`,
		`state.localStopLogs.set(model, entry);`,
		`state.localStopLogs.delete(model);`,
		`optimisticStopProbes(models);`,
		`restoreClientState(previous);`,
		`function runnableSelectedModels`,
		`function hasClearableResults`,
		`function stoppableModels`,
		`$('clearResultsBtn').disabled = state.editing || !hasClearableResults();`,
		`$('stopAllBtn').disabled = state.editing || stoppableModels().length === 0;`,
		`return [...state.selected].filter(model => model);`,
		`btn.disabled = state.editing || !state.runningModels.has(btn.dataset.stopOne)`,
		`async function stopProbe`,
		`async function stopAllProbes`,
		`id="clearResultsBtn"`,
		`id="stopAllBtn"`,
		`data-stop-one`,
		`data-select-group`,
		`Clear results`,
		`Stop all`,
		`/api/probe/results/clear`,
		`/api/probe/stop`,
		`JSON.stringify({ models })`,
		`class="header-actions"`,
		`class="header-actions models-actions"`,
		`.app { max-width: 1280px; margin: 0 auto; padding: 24px; }`,
		`grid-template-columns: 320px minmax(0, 1fr);`,
		`.grid > .panel { min-width: 0; }`,
		`.models-actions { flex: 1 1 auto; min-width: 0; }`,
		`.settings input, .settings button`,
		`<option value="codex">Codex</option>`,
		`<option value="claude">Claude</option>`,
		`id="claudeCommand"`,
		`function providerValue`,
		`function providerLabel`,
		`function providerIconHTML`,
		`provider: providerValue(group.provider)`,
		`class="provider-icon"`,
		`data-provider-icon="codex"`,
		`data-provider-icon="claude"`,
		`viewBox="0 0 100 100"`,
		`function groupResultStats`,
		`function groupStatsHTML`,
		`group-stat-empty`,
		`function groupSelectionLabel`,
		`selectedCount + '/' + groupModels.length + ' selected'`,
		`class="group-stats"`,
		`group-stat-success`,
		`statHTML('Success', stats.success`,
		`statHTML('Failed', stats.failed`,
		`statHTML('Running', stats.running`,
		`class="group-provider-select"`,
		`config.claude_command || 'claude'`,
		`delete cfg.provider;`,
		`cfg.claude_command = $('claudeCommand').value.trim() || 'claude';`,
		`min-height: 52px`,
		`class="live-dot"`,
		`blue-breathe`,
		`<span class="pill run"><span class="live-dot"></span>Running</span>`,
		`id="editModelsBtn"`,
		`id="cancelEditBtn"`,
		`.add-form input, .add-form button, .add-form select`,
		`flex: 1 1 430px`,
		`editing: false`,
		`draftGroups: []`,
		`draftModels: []`,
		`collapsedGroups: new Set()`,
		`groupsCollapsedInitialized: false`,
		`dragModel: ''`,
		`draftModelLoopCounts: {}`,
		`dirtyLoopCounts: new Set()`,
		`editLockedModels: new Set()`,
		`state.editLockedModels = new Set(state.runningModels)`,
		`state.editLockedModels.has(model)`,
		`Running when editing started`,
		`function loopCountControl`,
		`function loopCountOptions`,
		`data-keepassxc-ignore="true"`,
		`data-lpignore="true"`,
		`data-1p-ignore="true"`,
		`data-bwignore="true"`,
		`data-form-type="other"`,
		`<select class="attempt-limit-select"`,
		`data-attempt-limit`,
		`const loopCountChoices = [0, 1, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99]`,
		`function normalizeLoopCountChoice`,
		`loopCountChoices.forEach(count =>`,
		`function mergeDirtyLoopCounts`,
		`state.config = mergeDirtyLoopCounts(data.config);`,
		`state.dirtyLoopCounts.add(model);`,
		`replaceChildrenWithHTML(resultCell, statusPill(model));`,
		`document.activeElement !== retryControl`,
		`function renderTaskViews`,
		`clearDirtyLoopCounts(models);`,
		`addEventListener('change'`,
		`loopCountLabel`,
		`displayAttemptProgress`,
		`.attempt-limit-select`,
		`model_loop_counts`,
		`model_groups`,
		`class="model-groups"`,
		`model-group-body-inner`,
		`transition: height .22s ease`,
		`.model-group-header.collapsed { border-bottom-color: transparent; }`,
		`.group-title`,
		`line-height: 18px`,
		`.provider-icon-host`,
		`transform: translateY(1px)`,
		`function setGroupBodyCollapsed`,
		`body.style.height = targetHeight + 'px'`,
		`className = 'model-group'`,
		`data-toggle-header`,
		`data-toggle-group`,
		`data-select-group`,
		`class="group-select-checkbox" data-select-group`,
		`id="newModelGroup"`,
		`id="addGroupBtn"`,
		`function initializeCollapsedGroups`,
		`state.collapsedGroups = new Set(configGroups(state.config).map((group, index) => groupKey(index, group)))`,
		`function configGroups`,
		`function flattenGroups`,
		`function toggleGroupCollapsed`,
		`function toggleGroupSelection`,
		`function renderModelsAnimated`,
		`function placeChildAt`,
		`placeChildAt(tbody, row, modelIndex);`,
		`function modelInsertionIndex`,
		`function unanimatedRowRect`,
		`function isGroupEndDropArea`,
		`const animation = row.animate`,
		`row.dataset.modelRow === state.dragModel`,
		`function moveDraftModelTo`,
		`function moveDraftGroupTo`,
		`function installGroupDropTarget`,
		`function installModelEditorDragSurface`,
		`installModelEditorDragSurface(host);`,
		`{ capture: true }`,
		`host.dataset.dragSurfaceInstalled`,
		`function keepActiveEditorDragAllowed`,
		`function finishActiveEditorDrag`,
		`function installEditorDragGuard`,
		`document.addEventListener('dragenter', keepActiveEditorDragAllowed, { capture: true });`,
		`document.addEventListener('dragover', keepActiveEditorDragAllowed, { capture: true });`,
		`document.addEventListener('drop', finishActiveEditorDrag, { capture: true });`,
		`installEditorDragGuard();`,
		`function installGroupDrag`,
		`data-group-drag-handle`,
		`drop-target`,
		`group-drop-target`,
		`group-dragging`,
		`draggable = true`,
		`dragstart`,
		`dragover`,
		`dragend`,
		`drag-handle`,
		`dragging`,
		`.model-table .model-heading { width: 30ch; }`,
		`.model-table th.check, .model-table td.check`,
		`.model-table input[type="checkbox"]`,
		`data-delete-model`,
		`async function toggleModelEdit`,
		`function cancelModelEdit`,
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

func TestIndexHTMLOptimisticallyClearsResults(t *testing.T) {
	for _, want := range []string{
		`optimisticClearedResults: new Map()`,
		`function optimisticClearResults`,
		`const running = new Set(state.runningModels);`,
		`const keptResults = currentTask.results.filter(res => running.has(res.model));`,
		`state.optimisticClearedResults.set(res.model, { expires_at: expiresAt, action_id: actionID, base_result_key: resultStateKey(res) });`,
		`function reconcileOptimisticClearedResults`,
		`task = reconcileOptimisticClearedResults(task || {});`,
		`state.optimisticClearedResults.delete(model);`,
		`optimisticClearResults();`,
		`restoreClientState(previous);`,
		`await pollState().catch(refreshErr => setMessage(refreshErr.message));`,
	} {
		if !strings.Contains(indexHTML, want) {
			t.Fatalf("indexHTML missing optimistic clear behavior %q", want)
		}
	}
}

func TestIndexHTMLMovesNotificationsToLogHeaderAndAddsStopAll(t *testing.T) {
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
	editIndex := strings.Index(indexHTML, `id="editModelsBtn"`)
	runIndex := strings.Index(indexHTML, `id="runSelectedBtn"`)
	clearIndex := strings.Index(indexHTML, `id="clearResultsBtn"`)
	stopIndex := strings.Index(indexHTML, `id="stopAllBtn"`)
	if editIndex < 0 || runIndex < 0 || clearIndex < 0 || stopIndex < 0 {
		t.Fatal("models header controls are missing")
	}
	if !(editIndex < runIndex && runIndex < clearIndex && clearIndex < stopIndex) {
		t.Fatalf("models header controls are in the wrong order: edit=%d run=%d clear=%d stop=%d", editIndex, runIndex, clearIndex, stopIndex)
	}
	if strings.Contains(indexHTML, `id="stopProbeBtn"`) || strings.Contains(indexHTML, `Stop task`) {
		t.Fatal("old global stop task control must remain removed")
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
	if state.Config.ClaudeCommand != "claude" {
		t.Fatalf("claude command = %q", state.Config.ClaudeCommand)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`"provider"`)) {
		t.Fatalf("state JSON should not expose top-level provider: %s", rec.Body.String())
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
	body := bytes.NewBufferString(`{"models":["sonnet"],"model_groups":[{"name":"Claude","provider":"claude","models":["sonnet"]}],"model_loop_counts":{"sonnet":2},"timeout_seconds":30,"codex_command":"codex-beta","claude_command":"claude-beta","listen_addr":"127.0.0.1:0","max_output_chars":1234}`)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/config", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	cfg, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TimeoutSeconds != 30 || cfg.ModelLoopCounts["sonnet"] != 2 || cfg.CodexCommand != "codex-beta" || cfg.ClaudeCommand != "claude-beta" || cfg.ListenAddr != "127.0.0.1:0" || cfg.MaxOutputChars != 1234 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if got := cfg.ProviderForModel("sonnet"); got != alive.ProviderClaude {
		t.Fatalf("provider for sonnet = %q, want claude", got)
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

func TestConfigEndpointPersistsModelGroups(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	srv := newServer(configPath)
	body := bytes.NewBufferString(`{"models":["model-b","model-a","model-c"],"model_groups":[{"name":"fast","provider":"codex","models":["model-b","model-a"]},{"name":"slow","provider":"claude","models":["model-c"]}],"model_loop_counts":{"model-b":2,"model-a":3,"model-c":4},"timeout_seconds":30,"codex_command":"codex","listen_addr":"127.0.0.1:0","max_output_chars":4000}`)

	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/config", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	cfg, err := alive.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	wantGroups := []alive.ModelGroup{
		{Name: "fast", Provider: alive.ProviderCodex, Models: []string{"model-b", "model-a"}},
		{Name: "slow", Provider: alive.ProviderClaude, Models: []string{"model-c"}},
	}
	if !reflect.DeepEqual(cfg.ModelGroups, wantGroups) {
		t.Fatalf("model groups = %#v, want %#v", cfg.ModelGroups, wantGroups)
	}
	if !reflect.DeepEqual(cfg.Models, []string{"model-b", "model-a", "model-c"}) {
		t.Fatalf("models = %#v", cfg.Models)
	}
}

func TestProbeRunnerConfigDoesNotExpandBatchGroups(t *testing.T) {
	cfg := alive.DefaultConfig()
	cfg.ModelGroups = []alive.ModelGroup{
		{Name: "fast", Provider: alive.ProviderCodex, Models: []string{"model-a", "model-b"}},
		{Name: "slow", Provider: alive.ProviderClaude, Models: []string{"model-c"}},
	}
	cfg.ModelLoopCounts = map[string]int{"model-a": 2, "model-b": 5, "model-c": 10}
	cfg.ApplyDefaults()

	runnerCfg := probeRunnerConfig(cfg, []string{"model-b", "model-c"})
	if !reflect.DeepEqual(runnerCfg.Models, []string{"model-b", "model-c"}) {
		t.Fatalf("runner models = %#v, want selected models only", runnerCfg.Models)
	}
	if _, ok := runnerCfg.ModelLoopCounts["model-a"]; ok {
		t.Fatalf("runner loop counts kept unselected model: %#v", runnerCfg.ModelLoopCounts)
	}
	wantRunnerGroups := []alive.ModelGroup{
		{Name: "fast", Provider: alive.ProviderCodex, Models: []string{"model-b"}},
		{Name: "slow", Provider: alive.ProviderClaude, Models: []string{"model-c"}},
	}
	if !reflect.DeepEqual(runnerCfg.ModelGroups, wantRunnerGroups) {
		t.Fatalf("runner groups = %#v, want %#v", runnerCfg.ModelGroups, wantRunnerGroups)
	}

	modelCfg := singleModelRunnerConfig(runnerCfg, "model-b")
	modelCfg.ApplyDefaults()
	if !reflect.DeepEqual(modelCfg.Models, []string{"model-b"}) {
		t.Fatalf("single-model runner models = %#v, want model-b only", modelCfg.Models)
	}
	if modelCfg.Provider != alive.ProviderCodex {
		t.Fatalf("single-model provider = %q, want codex", modelCfg.Provider)
	}
	if !reflect.DeepEqual(modelCfg.ModelGroups, []alive.ModelGroup{{Name: "Default", Provider: alive.ProviderCodex, Models: []string{"model-b"}}}) {
		t.Fatalf("single-model runner groups = %#v", modelCfg.ModelGroups)
	}
	if !reflect.DeepEqual(modelCfg.ModelLoopCounts, map[string]int{"model-b": 5}) {
		t.Fatalf("single-model loop counts = %#v", modelCfg.ModelLoopCounts)
	}

	claudeCfg := singleModelRunnerConfig(runnerCfg, "model-c")
	claudeCfg.ApplyDefaults()
	if claudeCfg.Provider != alive.ProviderClaude {
		t.Fatalf("single-model provider = %q, want claude", claudeCfg.Provider)
	}
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
	if !reflect.DeepEqual(stopping.RunningModels, []string{"model-c"}) {
		t.Fatalf("running models after immediate cancellation = %#v", stopping.RunningModels)
	}
	modelBResult := resultByModel(stopping.Results, "model-b")
	if modelBResult.Error != context.Canceled.Error() || modelBResult.Success {
		t.Fatalf("stopped model result = %#v", modelBResult)
	}
	if len(stopping.Logs) == 0 || stopping.Logs[0].Result.Model != "model-b" || stopping.Logs[0].Result.Error != context.Canceled.Error() {
		t.Fatalf("stopped model log = %#v", stopping.Logs)
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

func TestProbeTaskStartDoesNotReviveCompletedRuns(t *testing.T) {
	store := &taskStore{}
	task, runs, err := store.start([]string{"model-a", "model-b"}, map[string]int{"model-a": 1, "model-b": 1})
	if err != nil {
		t.Fatal(err)
	}
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-b"), alive.Event{Type: alive.EventResult, Result: alive.Result{
		Model:          "model-b",
		Success:        true,
		Attempts:       1,
		DurationMS:     10,
		AttemptResults: []alive.Result{{Model: "model-b", Success: true, Attempts: 1, DurationMS: 10}},
	}})
	if snap := store.snapshot(); !reflect.DeepEqual(snap.RunningModels, []string{"model-a"}) {
		t.Fatalf("running models after model-b completed = %#v", snap.RunningModels)
	}
	next, _, err := store.start([]string{"model-c"}, map[string]int{"model-c": 1})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(next.RunningModels, []string{"model-a", "model-c"}) {
		t.Fatalf("completed model was revived as running: %#v", next.RunningModels)
	}
}

func TestProbeTaskRecoversRunsWhenRunningFlagIsStale(t *testing.T) {
	store := &taskStore{}
	task, runs, err := store.start([]string{"backup-a"}, map[string]int{"backup-a": 10})
	if err != nil {
		t.Fatal(err)
	}
	backupCtx := runCtxByModel(t, runs, "backup-a")

	store.mu.Lock()
	store.task.Running = false
	store.task.RunningModels = nil
	store.mu.Unlock()

	snap := store.snapshot()
	if !snap.Running || !reflect.DeepEqual(snap.RunningModels, []string{"backup-a"}) {
		t.Fatalf("snapshot did not recover active run: %#v", snap)
	}

	next, nextRuns, err := store.start([]string{"free-a"}, map[string]int{"free-a": 1})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-backupCtx:
		t.Fatal("starting another model canceled or lost the stale active run")
	default:
	}
	if !reflect.DeepEqual(next.RunningModels, []string{"backup-a", "free-a"}) {
		t.Fatalf("running models after start = %#v", next.RunningModels)
	}

	if _, err := store.stopModels([]string{"backup-a"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-backupCtx:
	case <-time.After(time.Second):
		t.Fatal("stop did not cancel recovered active run")
	}
	store.finishRun(task.ID, runIDByModel(t, runs, "backup-a"), "backup-a")
	if got := store.snapshot().RunningModels; !reflect.DeepEqual(got, []string{"free-a"}) {
		t.Fatalf("running models after stale run finished = %#v", got)
	}
	store.finishRun(next.ID, runIDByModel(t, nextRuns, "free-a"), "free-a")
}

func TestProbeTaskFinishRemovesRunWhenTaskIDIsStale(t *testing.T) {
	store := &taskStore{}
	task, runs, err := store.start([]string{"backup-a"}, map[string]int{"backup-a": 10})
	if err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.task.ID++
	store.mu.Unlock()

	store.finishRun(task.ID, runIDByModel(t, runs, "backup-a"), "backup-a")
	snap := store.snapshot()
	if snap.Running || len(snap.RunningModels) != 0 {
		t.Fatalf("stale run remained visible after finish: %#v", snap)
	}
}

func TestCurrentRunContextCancelsWhenRunBecomesStale(t *testing.T) {
	store := &taskStore{}
	srv := &server{tasks: store}
	task, runs, err := store.start([]string{"backup-a"}, map[string]int{"backup-a": 10})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := srv.currentRunContext(context.Background(), task.ID, runIDByModel(t, runs, "backup-a"), "backup-a")
	defer cancel()

	store.mu.Lock()
	store.task.ID++
	store.mu.Unlock()

	select {
	case <-ctx.Done():
	case <-time.After(2*runLivenessCheckInterval + 500*time.Millisecond):
		t.Fatal("run context was not canceled after run became stale")
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
	if !strings.Contains(string(body), `"revision":1`) {
		t.Fatalf("task JSON missing revision: %s", body)
	}
}

func TestProbeTaskRevisionStaysMonotonicAcrossStopAndRestart(t *testing.T) {
	store := &taskStore{}
	first, firstRuns, err := store.start([]string{"model-a"}, map[string]int{"model-a": 1})
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision < 1 {
		t.Fatalf("first revision = %d", first.Revision)
	}

	stopped, err := store.stopModels([]string{"model-a"})
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Revision <= first.Revision || stopped.Running {
		t.Fatalf("stopped task = %#v", stopped)
	}

	restarted, restartedRuns, err := store.start([]string{"model-a"}, map[string]int{"model-a": 1})
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Revision <= stopped.Revision || !restarted.Running {
		t.Fatalf("restarted task = %#v", restarted)
	}

	store.applyEvent(first.ID, runIDByModel(t, firstRuns, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 9, DurationMS: 90}})
	store.finishRun(first.ID, runIDByModel(t, firstRuns, "model-a"), "model-a")
	snap := store.snapshot()
	if snap.Revision != restarted.Revision || !reflect.DeepEqual(snap.RunningModels, []string{"model-a"}) {
		t.Fatalf("stale stopped run changed restarted task: %#v", snap)
	}
	store.stopModels([]string{"model-a"})
	store.finishRun(restarted.ID, runIDByModel(t, restartedRuns, "model-a"), "model-a")
}

func TestProbeTaskSubscribersReceiveTaskUpdates(t *testing.T) {
	store := &taskStore{}
	updates, unsubscribe := store.subscribe()
	defer unsubscribe()

	if initial := readTaskUpdate(t, updates); initial.Running || initial.ID != 0 {
		t.Fatalf("initial task update = %#v", initial)
	}
	task, runs, err := store.start([]string{"model-a"}, map[string]int{"model-a": 2})
	if err != nil {
		t.Fatal(err)
	}
	started := readTaskUpdate(t, updates)
	if !started.Running || started.ID != task.ID || !reflect.DeepEqual(started.RunningModels, []string{"model-a"}) {
		t.Fatalf("started update = %#v", started)
	}
	store.applyEvent(task.ID, runIDByModel(t, runs, "model-a"), alive.Event{Type: alive.EventAttempt, Result: alive.Result{Model: "model-a", Attempts: 1, Success: false, DurationMS: 10}})
	attempted := readTaskUpdate(t, updates)
	if len(attempted.Logs) != 1 || attempted.Logs[0].Result.Model != "model-a" {
		t.Fatalf("attempt update = %#v", attempted)
	}
	store.finishRun(task.ID, runIDByModel(t, runs, "model-a"), "model-a")
	finished := readTaskUpdate(t, updates)
	if finished.Running || finished.FinishedAt == "" {
		t.Fatalf("finished update = %#v", finished)
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

func TestProbeTaskStartPreservesPreviousLogsAndUnselectedResults(t *testing.T) {
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
	if modelAResult := resultByModel(snap.Results, "model-a"); modelAResult.Model != "model-a" || !modelAResult.Success {
		t.Fatalf("unselected model result was not preserved: %#v", snap.Results)
	}
	modelBResult := resultByModel(snap.Results, "model-b")
	if modelBResult.Model != "model-b" || modelBResult.Attempts != 1 || modelBResult.Success {
		t.Fatalf("selected model did not get initial running result: %#v", snap.Results)
	}
	if snap.LoopCounts["model-a"] != 1 || snap.LoopCounts["model-b"] != 1 {
		t.Fatalf("loop counts after second start = %#v", snap.LoopCounts)
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
	srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/probe/stop", strings.NewReader(`{"models":["model-b"]}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("stop all status = %d, body = %q", rec.Code, rec.Body.String())
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

func readTaskUpdate(t *testing.T, updates <-chan probeTask) probeTask {
	t.Helper()
	select {
	case task, ok := <-updates:
		if !ok {
			t.Fatal("task update channel closed")
		}
		return task
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for task update")
		return probeTask{}
	}
}
