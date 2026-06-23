package main

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>API Alive</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f5f7fb;
      --panel: #ffffff;
      --text: #1b2330;
      --muted: #667085;
      --line: #d9e0ea;
      --accent: #0f766e;
      --accent-dark: #115e59;
      --danger: #b42318;
      --ok: #067647;
      --warn: #b54708;
      --shadow: 0 18px 50px rgba(31, 41, 55, .10);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: var(--bg);
      color: var(--text);
      letter-spacing: 0;
    }
    button, input { font: inherit; }
    button {
      border: 1px solid transparent;
      background: var(--accent);
      color: white;
      border-radius: 6px;
      min-height: 36px;
      padding: 0 12px;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 6px;
      white-space: nowrap;
    }
    button:hover { background: var(--accent-dark); }
    button.secondary {
      background: #fff;
      color: var(--text);
      border-color: var(--line);
    }
    button.secondary:hover { background: #f8fafc; }
    button:disabled { opacity: .55; cursor: default; }
    button.table-action {
      min-height: 26px;
      padding: 0 8px;
      font-size: 12px;
    }
    input {
      width: 100%;
      min-height: 36px;
      border: 1px solid var(--line);
      border-radius: 6px;
      background: #fff;
      color: var(--text);
      padding: 7px 10px;
      outline: none;
    }
    input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(15, 118, 110, .12); }
    label { font-size: 12px; color: var(--muted); display: grid; gap: 6px; }
    .app { max-width: 1180px; margin: 0 auto; padding: 24px; }
    .top {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 18px;
      margin-bottom: 18px;
    }
    .brand h1 { margin: 0; font-size: 26px; line-height: 1.15; }
    .brand p { margin: 6px 0 0; color: var(--muted); font-size: 13px; }
    .status { min-width: 240px; text-align: right; color: var(--muted); font-size: 13px; }
    .status strong { color: var(--text); }
    .grid {
      display: grid;
      grid-template-columns: 320px 1fr;
      gap: 18px;
      align-items: start;
    }
    .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      box-shadow: var(--shadow);
    }
    .panel header {
      padding: 14px 16px;
      border-bottom: 1px solid var(--line);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
    }
    .panel h2 { margin: 0; font-size: 15px; }
    .panel .body { padding: 16px; }
    .settings { display: grid; gap: 12px; }
    .settings .row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
    .toolbar {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
      margin-bottom: 12px;
    }
    .toolbar .actions { display: flex; flex-wrap: wrap; gap: 8px; }
    .toolbar button {
      min-height: 30px;
      padding: 0 10px;
      font-size: 12px;
    }
    .add-form { display: grid; grid-template-columns: minmax(180px, 1fr) auto; gap: 8px; margin-bottom: 12px; }
    .add-form[hidden], button[hidden], label[hidden] { display: none; }
    table { width: 100%; border-collapse: collapse; table-layout: fixed; }
    th, td {
      padding: 6px 8px;
      border-bottom: 1px solid var(--line);
      text-align: left;
      vertical-align: middle;
      font-size: 12px;
    }
    th { color: var(--muted); font-weight: 600; background: #fbfcfe; }
    td.model { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; overflow-wrap: anywhere; }
    .model-table th, .model-table td { padding: 4px 6px; height: 30px; }
    .model-table .model-heading { width: 34%; }
    .model-table input[type="checkbox"] { width: 16px; height: 16px; min-height: 16px; padding: 0; display: block; }
    .model-table button.table-action { min-height: 24px; padding: 0 7px; }
    .model-editor .model-heading { width: auto; }
    .editor-actions { width: 150px; }
    .check { width: 34px; }
    .order { width: 92px; }
    .result { width: 96px; }
    .attempts { width: 78px; }
    .duration { width: 78px; }
    .row-actions { width: 70px; }
    .move-actions { display: flex; gap: 4px; }
    .pill {
      display: inline-flex;
      align-items: center;
      border-radius: 999px;
      padding: 3px 8px;
      font-size: 12px;
      border: 1px solid var(--line);
      background: #fff;
      color: var(--muted);
      white-space: nowrap;
    }
    .pill.ok { color: var(--ok); border-color: rgba(6, 118, 71, .25); background: rgba(6, 118, 71, .08); }
    .pill.bad { color: var(--danger); border-color: rgba(180, 35, 24, .24); background: rgba(180, 35, 24, .08); }
    .pill.run { color: var(--warn); border-color: rgba(181, 71, 8, .24); background: rgba(181, 71, 8, .08); }
    .empty {
      padding: 36px 12px;
      text-align: center;
      color: var(--muted);
      border: 1px dashed var(--line);
      border-radius: 8px;
      background: #fbfcfe;
    }
    .log-panel { margin-top: 18px; }
    .log-title {
      display: flex;
      align-items: center;
      gap: 10px;
      flex-wrap: wrap;
      min-width: 0;
    }
    .running-models {
      display: flex;
      align-items: center;
      gap: 6px;
      flex-wrap: wrap;
      min-width: 0;
    }
    .running-model {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      max-width: 240px;
      padding: 3px 8px;
      border: 1px solid rgba(6, 118, 71, .28);
      border-radius: 999px;
      background: rgba(6, 118, 71, .07);
      color: var(--ok);
      font-size: 12px;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    }
    .running-model span:last-child {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .live-dot {
      width: 8px;
      height: 8px;
      border-radius: 999px;
      background: #12b76a;
      box-shadow: 0 0 0 0 rgba(18, 183, 106, .55);
      animation: breathe 1.4s ease-in-out infinite;
      flex: 0 0 auto;
    }
    @keyframes breathe {
      0%, 100% { transform: scale(.82); box-shadow: 0 0 0 0 rgba(18, 183, 106, .45); opacity: .75; }
      50% { transform: scale(1.08); box-shadow: 0 0 0 6px rgba(18, 183, 106, 0); opacity: 1; }
    }
    .log-table-wrap {
      max-height: 340px;
      overflow: auto;
    }
    .log-table { font-size: 12px; }
    .log-table th {
      position: sticky;
      top: 0;
      z-index: 1;
    }
    .log-table th, .log-table td {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .log-status-col { width: 88px; }
    .log-model-col { width: 28%; }
    .log-attempt-col { width: 72px; }
    .log-seconds-col { width: 80px; }
    .log-time-col { width: 86px; }
    .log-error-col { width: auto; }
    .log-empty {
      padding: 22px 12px;
      text-align: center;
      color: var(--muted);
      border: 1px dashed var(--line);
      border-radius: 8px;
      background: #fbfcfe;
      font-size: 13px;
    }
    .log-entry.ok { background: rgba(6, 118, 71, .045); }
    .log-entry.bad { background: rgba(180, 35, 24, .045); }
    .log-status { font-weight: 700; }
    .log-model { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; color: var(--text); }
    .log-error { color: var(--danger); }
    .log-time { color: var(--muted); white-space: nowrap; }
    @media (max-width: 860px) {
      .app { padding: 16px; }
      .top { display: grid; }
      .status { text-align: left; min-width: 0; }
      .grid { grid-template-columns: 1fr; }
      .toolbar { align-items: stretch; }
      .toolbar .actions { width: 100%; }
      .toolbar .actions button { flex: 1; }
      th.attempts, td.attempts, th.duration, td.duration { display: none; }
      .order { width: 86px; }
      .row-actions { width: 68px; }
    }
  </style>
</head>
<body>
  <main class="app">
    <section class="top">
      <div class="brand">
        <h1>API Alive</h1>
        <p>VPS Codex model liveness dashboard</p>
      </div>
      <div class="status">
        <div><strong id="configPath">config.json</strong></div>
        <div id="message">Loading...</div>
      </div>
    </section>

    <section class="grid">
      <aside class="panel">
        <header><h2>Runtime</h2><button class="secondary" id="reloadBtn">Refresh</button></header>
        <div class="body settings">
          <label>Codex command
            <input id="codexCommand" placeholder="codex">
          </label>
          <label>Listen address
            <input id="listenAddr" placeholder="0.0.0.0:8080">
          </label>
          <div class="row">
            <label>Timeout seconds
              <input id="timeoutSeconds" type="number" min="1" step="1">
            </label>
            <label>Loop count
              <input id="loopCount" type="number" min="1" step="1">
            </label>
          </div>
          <label>Max output chars
            <input id="maxOutputChars" type="number" min="1" step="1">
          </label>
          <button id="saveConfigBtn">Save runtime</button>
        </div>
      </aside>

      <section class="panel">
        <header>
          <h2>Models</h2>
          <span class="pill" id="selectedCount">0 selected</span>
        </header>
        <div class="body">
          <div class="toolbar">
            <label id="selectAllLabel" style="display:flex; align-items:center; gap:8px; color:var(--text);">
              <input id="selectAll" type="checkbox" style="width:16px; min-height:16px;"> Select all
            </label>
            <div class="actions">
              <button class="secondary" id="editModelsBtn" type="button">Edit</button>
              <button class="secondary" id="cancelEditBtn" type="button" hidden>Cancel</button>
              <button id="runSelectedBtn">Run selected</button>
            </div>
          </div>
          <form class="add-form" id="addForm" hidden>
            <input id="newModel" placeholder="gpt-5 or vendor/gpt-5.5">
            <button type="submit">Add</button>
          </form>
          <div id="modelHost"></div>
        </div>
      </section>
    </section>

    <section class="panel log-panel">
      <header>
        <div class="log-title">
          <h2>Log</h2>
          <div class="running-models" id="runningModels"></div>
        </div>
        <button class="secondary" id="stopProbeBtn" disabled>Stop task</button>
      </header>
      <div class="body">
        <div class="log-table-wrap">
          <table class="log-table">
            <colgroup>
              <col class="log-status-col">
              <col class="log-model-col">
              <col class="log-attempt-col">
              <col class="log-seconds-col">
              <col class="log-time-col">
              <col class="log-error-col">
            </colgroup>
            <thead><tr><th>Status</th><th>Model</th><th>Try</th><th>Seconds</th><th>Time</th><th>Error</th></tr></thead>
            <tbody id="logList"></tbody>
          </table>
        </div>
      </div>
    </section>
  </main>

  <script>
    const state = { config: null, task: {}, selected: new Set(), results: new Map(), running: false, runningModels: new Set(), logEntries: [], editing: false, draftModels: [] };
    const maxLogEntries = 100;
    const idlePollMS = 60000;
    const runningPollMS = 5000;
    let pollTimer = null;
    const $ = (id) => document.getElementById(id);

    function setMessage(text) { $('message').textContent = text; }
    function resultFor(model) { return state.results.get(model) || null; }
    function escapeText(value) {
      return String(value ?? '').replace(/[&<>"']/g, c => ({ '&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;' }[c]));
    }
    function displayTime(value) {
      if (!value) return '';
      const date = new Date(value);
      return Number.isNaN(date.getTime()) ? value : date.toLocaleTimeString();
    }
    function visibleModels() { return state.editing ? state.draftModels : (state.config?.models || []); }
    function schedulePoll() {
      clearTimeout(pollTimer);
      pollTimer = setTimeout(() => pollState().catch(err => {
        setMessage(err.message);
        schedulePoll();
      }), state.running ? runningPollMS : idlePollMS);
    }
    function setBusy(value) {
      state.running = value;
      if (value && state.editing) {
        state.editing = false;
        state.draftModels = [];
      }
      $('runSelectedBtn').disabled = value || state.editing || state.selected.size === 0;
      $('runSelectedBtn').hidden = state.editing;
      $('selectAllLabel').hidden = state.editing;
      $('addForm').hidden = !state.editing;
      $('cancelEditBtn').hidden = !state.editing;
      $('cancelEditBtn').disabled = value;
      $('editModelsBtn').disabled = value || !state.config;
      $('editModelsBtn').textContent = state.editing ? 'Save' : 'Edit';
      $('editModelsBtn').className = state.editing ? '' : 'secondary';
      $('stopProbeBtn').disabled = !value || state.task?.stopping;
      $('stopProbeBtn').textContent = state.task?.stopping ? 'Stopping...' : 'Stop task';
      document.querySelectorAll('[data-run-one]').forEach(btn => btn.disabled = value || state.editing);
      document.querySelectorAll('[data-move]').forEach(btn => { btn.disabled = value || btn.dataset.boundary === 'true'; });
      renderRunningModels();
    }
    function renderRunningModels() {
      const host = $('runningModels');
      const models = [...state.runningModels];
      host.innerHTML = models.map(model => '<span class="running-model" title="' + escapeText(model) + '"><span class="live-dot"></span><span>' + escapeText(model) + '</span></span>').join('');
    }
    function renderLog() {
      const host = $('logList');
      if (!state.logEntries.length) {
        host.innerHTML = '<tr><td class="log-empty" colspan="6">Ready.</td></tr>';
        return;
      }
      host.innerHTML = state.logEntries.map(entry => {
        const res = entry.result || entry;
        const status = res.success ? 'success' : 'failed';
        const cls = res.success ? 'ok' : 'bad';
        const icon = res.success ? '✅' : '❌';
        const seconds = ((res.duration_ms || 0) / 1000).toFixed(3);
        const time = displayTime(entry.time);
        const errorText = res.success ? '' : (res.error || 'unknown error');
        const logTitle = status + ' model=' + (res.model || '') + ' attempt=' + (res.attempts ?? '') + ' seconds=' + seconds + ' time=' + time + (errorText ? ' error=' + errorText : '');
        return '<tr class="log-entry ' + cls + '" title="' + escapeText(logTitle) + '">' +
          '<td class="log-status">' + icon + ' ' + status + '</td>' +
          '<td class="log-model">' + escapeText(res.model) + '</td>' +
          '<td>' + escapeText(res.attempts) + '</td>' +
          '<td>' + escapeText(seconds) + '</td>' +
          '<td class="log-time">' + escapeText(time) + '</td>' +
          '<td class="log-error">' + escapeText(errorText) + '</td></tr>';
      }).join('');
    }
    function updateSelectedCount() {
      $('selectedCount').textContent = state.editing ? 'Editing' : state.selected.size + ' selected';
      $('runSelectedBtn').disabled = state.running || state.editing || state.selected.size === 0;
      $('runSelectedBtn').hidden = state.editing;
      $('selectAllLabel').hidden = state.editing;
      $('addForm').hidden = !state.editing;
      $('cancelEditBtn').hidden = !state.editing;
      $('cancelEditBtn').disabled = state.running;
      $('editModelsBtn').disabled = state.running || !state.config;
      $('editModelsBtn').textContent = state.editing ? 'Save' : 'Edit';
      $('editModelsBtn').className = state.editing ? '' : 'secondary';
      const models = visibleModels();
      $('selectAll').checked = models.length > 0 && state.selected.size === models.length;
      $('selectAll').indeterminate = state.selected.size > 0 && state.selected.size < models.length;
    }
    function statusPill(model) {
      if (state.runningModels.has(model)) return '<span class="pill run">Running</span>';
      const res = resultFor(model);
      if (!res) return '<span class="pill">Idle</span>';
      return res.success ? '<span class="pill ok">Success</span>' : '<span class="pill bad">Failed</span>';
    }
    function renderModelEditor(models) {
      if (!models.length) {
        $('modelHost').innerHTML = '<div class="empty">No models configured.</div>';
        return;
      }
      $('modelHost').innerHTML = '<table class="model-table model-editor"><thead><tr><th class="model-heading">Model</th><th class="editor-actions">Actions</th></tr></thead><tbody>' + models.map((model, index) => {
        const upDisabled = index === 0 ? 'disabled data-boundary="true"' : 'data-boundary="false"';
        const downDisabled = index === models.length - 1 ? 'disabled data-boundary="true"' : 'data-boundary="false"';
        return '<tr>' +
          '<td class="model" title="' + escapeText(model) + '">' + escapeText(model) + '</td>' +
          '<td class="editor-actions"><div class="move-actions">' +
            '<button type="button" class="secondary table-action" data-move="up" data-model="' + escapeText(model) + '" ' + upDisabled + '>Up</button>' +
            '<button type="button" class="secondary table-action" data-move="down" data-model="' + escapeText(model) + '" ' + downDisabled + '>Down</button>' +
            '<button type="button" class="secondary table-action" data-delete-model="' + escapeText(model) + '">Delete</button>' +
          '</div></td></tr>';
      }).join('') + '</tbody></table>';
      document.querySelectorAll('[data-move]').forEach(button => button.addEventListener('click', () => moveModel(button.dataset.model, button.dataset.move)));
      document.querySelectorAll('[data-delete-model]').forEach(button => button.addEventListener('click', () => deleteModel(button.dataset.deleteModel)));
    }
    function renderModels() {
      if (state.running && state.editing) {
        state.editing = false;
        state.draftModels = [];
      }
      const models = visibleModels();
      updateSelectedCount();
      if (state.editing) {
        renderModelEditor(models);
        setBusy(state.running);
        return;
      }
      if (!models.length) {
        $('modelHost').innerHTML = '<div class="empty">No models configured.</div>';
        return;
      }
      $('modelHost').innerHTML = '<table class="model-table"><thead><tr><th class="check"></th><th class="model-heading">Model</th><th class="result">Result</th><th class="attempts">Attempts</th><th class="duration">Seconds</th><th class="row-actions"></th></tr></thead><tbody>' + models.map((model) => {
        const res = resultFor(model);
        const checked = state.selected.has(model) ? 'checked' : '';
        const seconds = res && !state.runningModels.has(model) ? ((res.duration_ms || 0) / 1000).toFixed(3) : '';
        const attempts = res && !state.runningModels.has(model) ? res.attempts : '';
        return '<tr><td class="check"><input data-select="' + escapeText(model) + '" type="checkbox" ' + checked + '></td>' +
          '<td class="model" title="' + escapeText(model) + '">' + escapeText(model) + '</td>' +
          '<td class="result">' + statusPill(model) + '</td><td class="attempts">' + escapeText(attempts) + '</td><td class="duration">' + escapeText(seconds) + '</td>' +
          '<td class="row-actions"><button type="button" class="secondary table-action" data-run-one="' + escapeText(model) + '">Run</button></td></tr>';
      }).join('') + '</tbody></table>';
      document.querySelectorAll('[data-select]').forEach(input => input.addEventListener('change', () => {
        input.checked ? state.selected.add(input.dataset.select) : state.selected.delete(input.dataset.select);
        updateSelectedCount();
      }));
      document.querySelectorAll('[data-run-one]').forEach(button => button.addEventListener('click', () => startProbe([button.dataset.runOne]).catch(err => setMessage(err.message))));
      updateSelectedCount();
      setBusy(state.running);
    }
    function fillForm(config) {
      $('codexCommand').value = config.codex_command || 'codex';
      $('listenAddr').value = config.listen_addr || '0.0.0.0:8080';
      $('timeoutSeconds').value = config.timeout_seconds || 120;
      $('loopCount').value = config.loop_count || 1;
      $('maxOutputChars').value = config.max_output_chars || 4000;
    }
    function applyTask(task) {
      state.task = task || {};
      state.running = !!state.task.running;
      state.runningModels = new Set(state.task.running_models || []);
      state.results = new Map((state.task.results || []).map(res => [res.model, res]));
      state.logEntries = (state.task.logs || []).slice(0, maxLogEntries);
      renderLog();
      renderRunningModels();
      renderModels();
    }
    function applyServerState(data) {
      state.config = data.config;
      state.selected = new Set([...state.selected].filter(model => visibleModels().includes(model)));
      $('configPath').textContent = data.config_path || 'config.json';
      fillForm(state.config);
      applyTask(data.task || {});
      if (state.running) setMessage((state.task.stopping ? 'Stopping' : 'Running') + ' ' + state.runningModels.size + ' model(s)...');
      else if (state.task?.id && state.task.error) setMessage('Last task failed: ' + state.task.error);
      else if (state.task?.id && state.task.finished_at) setMessage('Last task finished. ' + state.config.models.length + ' configured model(s)');
      else setMessage(state.config.models.length + ' configured model(s)');
      schedulePoll();
    }
    async function request(path, options = {}) {
      const res = await fetch(path, { headers: { 'Content-Type': 'application/json' }, ...options });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || ('HTTP ' + res.status));
      return data;
    }
    async function loadState() { setMessage('Loading...'); applyServerState(await request('/api/state')); }
    async function pollState() { applyServerState(await request('/api/state')); }
    async function saveRuntime() {
      const cfg = { ...state.config };
      cfg.codex_command = $('codexCommand').value.trim() || 'codex';
      cfg.listen_addr = $('listenAddr').value.trim() || '0.0.0.0:8080';
      cfg.timeout_seconds = Number($('timeoutSeconds').value) || 120;
      cfg.loop_count = Number($('loopCount').value) || 1;
      cfg.max_output_chars = Number($('maxOutputChars').value) || 4000;
      state.config = await request('/api/config', { method: 'POST', body: JSON.stringify(cfg) });
      fillForm(state.config);
      renderModels();
      setMessage('Runtime saved.');
      schedulePoll();
    }
    async function addModel(name) {
      name = name.trim();
      if (!name || !state.editing) return;
      if (!state.draftModels.includes(name)) state.draftModels.push(name);
      $('newModel').value = '';
      renderModels();
      setMessage('Added ' + name);
    }
    function deleteModel(model) {
      if (state.running || !state.editing) return;
      state.draftModels = state.draftModels.filter(existing => existing !== model);
      state.selected.delete(model);
      renderModels();
      setMessage('Removed ' + model);
    }
    async function moveModel(model, direction) {
      if (state.running || !state.editing) return;
      const models = [...state.draftModels];
      const index = models.indexOf(model);
      const nextIndex = direction === 'up' ? index - 1 : index + 1;
      if (index < 0 || nextIndex < 0 || nextIndex >= models.length) return;
      [models[index], models[nextIndex]] = [models[nextIndex], models[index]];
      state.draftModels = models;
      renderModels();
      setMessage('Model order updated.');
    }
    async function toggleModelEdit() {
      if (state.running || !state.config) return;
      if (!state.editing) {
        state.editing = true;
        state.draftModels = [...(state.config.models || [])];
        state.selected = new Set();
        renderModels();
        setMessage('Editing models.');
        return;
      }
      state.config = await request('/api/config', { method: 'POST', body: JSON.stringify({ ...state.config, models: state.draftModels }) });
      state.editing = false;
      state.draftModels = [];
      state.selected = new Set([...state.selected].filter(model => state.config.models.includes(model)));
      $('newModel').value = '';
      renderModels();
      setMessage('Models saved.');
      schedulePoll();
    }
    function cancelModelEdit() {
      if (!state.editing) return;
      state.editing = false;
      state.draftModels = [];
      $('newModel').value = '';
      renderModels();
      setMessage('Edit cancelled.');
    }
    async function startProbe(models) {
      models = [...new Set(models)].filter(Boolean);
      if (!models.length || state.running || state.editing) return;
      const data = await request('/api/probe', { method: 'POST', body: JSON.stringify({ models }) });
      applyTask(data.task);
      setMessage('Probe task started.');
      schedulePoll();
    }
    async function stopProbe() {
      if (!state.running) return;
      const data = await request('/api/probe/stop', { method: 'POST', body: '{}' });
      applyTask(data.task);
      setMessage('Stopping probe task...');
      schedulePoll();
    }

    $('reloadBtn').addEventListener('click', () => loadState().catch(err => setMessage(err.message)));
    $('saveConfigBtn').addEventListener('click', () => saveRuntime().catch(err => setMessage(err.message)));
    $('stopProbeBtn').addEventListener('click', () => stopProbe().catch(err => setMessage(err.message)));
    $('addForm').addEventListener('submit', event => { event.preventDefault(); addModel($('newModel').value).catch(err => setMessage(err.message)); });
    $('selectAll').addEventListener('change', () => {
      const models = visibleModels();
      state.selected = $('selectAll').checked ? new Set(models) : new Set();
      renderModels();
    });
    $('runSelectedBtn').addEventListener('click', () => startProbe([...state.selected]).catch(err => setMessage(err.message)));
    $('editModelsBtn').addEventListener('click', () => toggleModelEdit().catch(err => setMessage(err.message)));
    $('cancelEditBtn').addEventListener('click', () => cancelModelEdit());
    renderLog();
    renderRunningModels();
    loadState().catch(err => setMessage(err.message));
  </script>
</body>
</html>`
