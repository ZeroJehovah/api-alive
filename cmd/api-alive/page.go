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
    .add-form { display: grid; grid-template-columns: minmax(180px, 1fr) auto; gap: 8px; margin-bottom: 12px; }
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
    .log-list {
      display: grid;
      gap: 8px;
      max-height: 340px;
      overflow: auto;
      margin: 0;
      padding: 0 4px 0 0;
      list-style: none;
    }
    .log-empty {
      padding: 22px 12px;
      text-align: center;
      color: var(--muted);
      border: 1px dashed var(--line);
      border-radius: 8px;
      background: #fbfcfe;
      font-size: 13px;
    }
    .log-entry {
      display: grid;
      grid-template-columns: auto 1fr auto;
      gap: 10px;
      align-items: start;
      padding: 10px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
      font-size: 12px;
    }
    .log-entry.ok { border-color: rgba(6, 118, 71, .24); background: rgba(6, 118, 71, .045); }
    .log-entry.bad { border-color: rgba(180, 35, 24, .22); background: rgba(180, 35, 24, .045); }
    .log-icon { line-height: 1.7; }
    .log-main { min-width: 0; display: grid; gap: 4px; }
    .log-meta { display: flex; flex-wrap: wrap; gap: 8px; color: var(--muted); }
    .log-model { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; color: var(--text); overflow-wrap: anywhere; }
    .log-error { color: var(--danger); overflow-wrap: anywhere; }
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
          <form class="add-form" id="addForm">
            <input id="newModel" placeholder="gpt-5 or vendor/gpt-5.5">
            <button type="submit">Add</button>
          </form>
          <div class="toolbar">
            <label style="display:flex; align-items:center; gap:8px; color:var(--text);">
              <input id="selectAll" type="checkbox" style="width:16px; min-height:16px;"> Select all
            </label>
            <div class="actions">
              <button class="secondary" id="orderModeBtn" type="button">Adjust order</button>
              <button id="runSelectedBtn">Run selected</button>
              <button class="secondary" id="deleteSelectedBtn">Delete selected</button>
            </div>
          </div>
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
        <ul class="log-list" id="logList"></ul>
      </div>
    </section>
  </main>

  <script>
    const state = { config: null, task: {}, selected: new Set(), results: new Map(), running: false, runningModels: new Set(), logEntries: [], ordering: false };
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
    function schedulePoll() {
      clearTimeout(pollTimer);
      pollTimer = setTimeout(() => pollState().catch(err => {
        setMessage(err.message);
        schedulePoll();
      }), state.running ? runningPollMS : idlePollMS);
    }
    function setBusy(value) {
      state.running = value;
      if (value && state.ordering) state.ordering = false;
      $('runSelectedBtn').disabled = value || state.selected.size === 0;
      $('deleteSelectedBtn').disabled = value || state.selected.size === 0;
      $('orderModeBtn').disabled = value || (state.config?.models || []).length < 2;
      $('orderModeBtn').textContent = state.ordering ? 'Done ordering' : 'Adjust order';
      $('stopProbeBtn').disabled = !value || state.task?.stopping;
      $('stopProbeBtn').textContent = state.task?.stopping ? 'Stopping...' : 'Stop task';
      document.querySelectorAll('[data-run-one]').forEach(btn => btn.disabled = value);
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
        host.innerHTML = '<li class="log-empty">Ready.</li>';
        return;
      }
      host.innerHTML = state.logEntries.map(entry => {
        const res = entry.result || entry;
        const status = res.success ? 'success' : 'failed';
        const cls = res.success ? 'ok' : 'bad';
        const icon = res.success ? '✅' : '❌';
        const error = res.success ? '' : '<div class="log-error">error=' + escapeText(res.error || 'unknown error') + '</div>';
        return '<li class="log-entry ' + cls + '">' +
          '<div class="log-icon">' + icon + '</div>' +
          '<div class="log-main">' +
            '<div><strong>' + status + '</strong> <span class="log-model">' + escapeText(res.model) + '</span></div>' +
            '<div class="log-meta"><span>attempt=' + escapeText(res.attempts) + '</span><span>seconds=' + escapeText(((res.duration_ms || 0) / 1000).toFixed(3)) + '</span></div>' + error +
          '</div><time class="log-time">' + escapeText(displayTime(entry.time)) + '</time></li>';
      }).join('');
    }
    function updateSelectedCount() {
      $('selectedCount').textContent = state.selected.size + ' selected';
      $('runSelectedBtn').disabled = state.running || state.selected.size === 0;
      $('deleteSelectedBtn').disabled = state.running || state.selected.size === 0;
      const models = state.config?.models || [];
      $('orderModeBtn').disabled = state.running || models.length < 2;
      $('orderModeBtn').textContent = state.ordering ? 'Done ordering' : 'Adjust order';
      $('selectAll').checked = models.length > 0 && state.selected.size === models.length;
      $('selectAll').indeterminate = state.selected.size > 0 && state.selected.size < models.length;
    }
    function statusPill(model) {
      if (state.runningModels.has(model)) return '<span class="pill run">Running</span>';
      const res = resultFor(model);
      if (!res) return '<span class="pill">Idle</span>';
      return res.success ? '<span class="pill ok">Success</span>' : '<span class="pill bad">Failed</span>';
    }
    function renderModels() {
      const models = state.config?.models || [];
      if (state.running || models.length < 2) state.ordering = false;
      updateSelectedCount();
      if (!models.length) {
        $('modelHost').innerHTML = '<div class="empty">No models configured.</div>';
        return;
      }
      const orderHead = state.ordering ? '<th class="order">Order</th>' : '';
      $('modelHost').innerHTML = '<table><thead><tr><th class="check"></th><th>Model</th>' + orderHead + '<th class="result">Result</th><th class="attempts">Attempts</th><th class="duration">Seconds</th><th class="row-actions"></th></tr></thead><tbody>' + models.map((model, index) => {
        const res = resultFor(model);
        const checked = state.selected.has(model) ? 'checked' : '';
        const seconds = res && !state.runningModels.has(model) ? ((res.duration_ms || 0) / 1000).toFixed(3) : '';
        const attempts = res && !state.runningModels.has(model) ? res.attempts : '';
        const upDisabled = index === 0 ? 'disabled data-boundary="true"' : 'data-boundary="false"';
        const downDisabled = index === models.length - 1 ? 'disabled data-boundary="true"' : 'data-boundary="false"';
        const orderCell = state.ordering ? '<td class="order"><div class="move-actions">' +
            '<button type="button" class="secondary table-action" data-move="up" data-model="' + escapeText(model) + '" ' + upDisabled + '>Up</button>' +
            '<button type="button" class="secondary table-action" data-move="down" data-model="' + escapeText(model) + '" ' + downDisabled + '>Down</button>' +
          '</div></td>' : '';
        return '<tr><td class="check"><input data-select="' + escapeText(model) + '" type="checkbox" ' + checked + '></td>' +
          '<td class="model" title="' + escapeText(model) + '">' + escapeText(model) + '</td>' +
          orderCell +
          '<td class="result">' + statusPill(model) + '</td><td class="attempts">' + escapeText(attempts) + '</td><td class="duration">' + escapeText(seconds) + '</td>' +
          '<td class="row-actions"><button type="button" class="secondary table-action" data-run-one="' + escapeText(model) + '">Run</button></td></tr>';
      }).join('') + '</tbody></table>';
      document.querySelectorAll('[data-select]').forEach(input => input.addEventListener('change', () => {
        input.checked ? state.selected.add(input.dataset.select) : state.selected.delete(input.dataset.select);
        updateSelectedCount();
      }));
      document.querySelectorAll('[data-run-one]').forEach(button => button.addEventListener('click', () => startProbe([button.dataset.runOne]).catch(err => setMessage(err.message))));
      document.querySelectorAll('[data-move]').forEach(button => button.addEventListener('click', () => moveModel(button.dataset.model, button.dataset.move)));
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
      state.selected = new Set([...state.selected].filter(model => state.config.models.includes(model)));
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
      if (!name) return;
      state.config = await request('/api/models', { method: 'POST', body: JSON.stringify({ models: [name] }) });
      $('newModel').value = '';
      renderModels();
      setMessage('Added ' + name);
      schedulePoll();
    }
    async function deleteSelected() {
      const models = [...state.selected];
      if (!models.length || state.running) return;
      state.config = await request('/api/models', { method: 'DELETE', body: JSON.stringify({ models }) });
      models.forEach(model => state.selected.delete(model));
      if ((state.config.models || []).length < 2) state.ordering = false;
      renderModels();
      setMessage('Deleted ' + models.length + ' model(s)');
      schedulePoll();
    }
    async function moveModel(model, direction) {
      if (state.running || !state.config) return;
      const models = [...(state.config.models || [])];
      const index = models.indexOf(model);
      const nextIndex = direction === 'up' ? index - 1 : index + 1;
      if (index < 0 || nextIndex < 0 || nextIndex >= models.length) return;
      [models[index], models[nextIndex]] = [models[nextIndex], models[index]];
      state.config = await request('/api/config', { method: 'POST', body: JSON.stringify({ ...state.config, models }) });
      renderModels();
      setMessage('Model order saved.');
      schedulePoll();
    }
    async function startProbe(models) {
      models = [...new Set(models)].filter(Boolean);
      if (!models.length || state.running) return;
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
      const models = state.config?.models || [];
      state.selected = $('selectAll').checked ? new Set(models) : new Set();
      renderModels();
    });
    $('runSelectedBtn').addEventListener('click', () => startProbe([...state.selected]).catch(err => setMessage(err.message)));
    $('deleteSelectedBtn').addEventListener('click', () => deleteSelected().catch(err => setMessage(err.message)));
    $('orderModeBtn').addEventListener('click', () => {
      if (state.running || (state.config?.models || []).length < 2) return;
      state.ordering = !state.ordering;
      renderModels();
    });
    renderLog();
    renderRunningModels();
    loadState().catch(err => setMessage(err.message));
  </script>
</body>
</html>`
