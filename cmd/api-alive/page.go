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
      min-height: 52px;
      padding: 10px 16px;
      border-bottom: 1px solid var(--line);
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
    }
    .panel h2 { margin: 0; font-size: 15px; }
    .panel .body { padding: 16px; }
    .header-actions { display: flex; flex-wrap: wrap; align-items: center; justify-content: flex-end; gap: 8px; }
    .header-actions button {
      min-height: 30px;
      padding: 0 10px;
      font-size: 12px;
    }
    .alert-button {
      border-color: var(--line);
      background: #fff;
      color: var(--text);
    }
    .alert-button::before {
      content: "";
      width: 7px;
      height: 7px;
      border-radius: 999px;
      background: var(--muted);
      flex: 0 0 auto;
    }
    .alert-button:hover { background: #f8fafc; }
    .alert-button.alert-on {
      color: var(--ok);
      border-color: transparent;
      background: transparent;
      min-height: 0;
      padding: 0;
      cursor: default;
      pointer-events: none;
    }
    .alert-button.alert-on:hover { background: transparent; }
    .alert-button.alert-on::before { background: var(--ok); }
    .alert-button.alert-blocked {
      color: var(--warn);
      border-color: rgba(181, 71, 8, .35);
      background: rgba(181, 71, 8, .08);
    }
    .alert-button.alert-blocked::before { background: var(--warn); }
    .alert-button.alert-request::before { background: #2563eb; }
    .alert-button.alert-unavailable { color: var(--muted); }
    .models-actions { flex: 1 1 auto; min-width: 0; }
    .selection-count { margin-right: auto; }
    .settings { display: grid; gap: 10px; }
    .settings .row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
    .settings label { gap: 5px; }
    .settings input, .settings button {
      min-height: 30px;
      padding: 0 10px;
      font-size: 12px;
    }
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
    .add-form { display: grid; grid-template-columns: minmax(180px, 1fr) auto; gap: 8px; flex: 1 1 320px; }
    .add-form input, .add-form button {
      min-height: 30px;
      padding: 0 10px;
      font-size: 12px;
    }
    [hidden] { display: none !important; }
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
    .model-table .model-heading { width: 30ch; }
    .model-table input[type="checkbox"] { width: 16px; height: 16px; min-height: 16px; padding: 0; display: block; }
    .model-table button.table-action { min-height: 24px; padding: 0 7px; }
    .loop-count-input { width: 5.25ch; min-width: 5.25ch; min-height: 24px; padding: 0 8px; text-align: center; font-size: 12px; }
    .model-editor .model-heading { width: auto; }
    .editor-actions { width: 150px; }
    .check { width: 34px; }
    .order { width: 92px; }
    .result { width: 118px; }
    .progress { width: 78px; }
    .retry-count { width: 66px; }
    .result-time { width: 154px; }
    .row-actions { width: 108px; }
    .row-action-group { display: flex; gap: 4px; align-items: center; }
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
      gap: 6px;
    }
    .pill.ok { color: var(--ok); border-color: rgba(6, 118, 71, .25); background: rgba(6, 118, 71, .08); }
    .pill.bad { color: var(--danger); border-color: rgba(180, 35, 24, .24); background: rgba(180, 35, 24, .08); }
    .pill.run { color: #1d4ed8; border-color: rgba(37, 99, 235, .28); background: rgba(37, 99, 235, .08); }
    .pill.run .live-dot {
      width: 7px;
      height: 7px;
      background: #2563eb;
      box-shadow: 0 0 0 0 rgba(37, 99, 235, .55);
      animation: blue-breathe 1.4s ease-in-out infinite;
    }
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
    @keyframes blue-breathe {
      0%, 100% { transform: scale(.82); box-shadow: 0 0 0 0 rgba(37, 99, 235, .45); opacity: .75; }
      50% { transform: scale(1.08); box-shadow: 0 0 0 6px rgba(37, 99, 235, 0); opacity: 1; }
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
    .log-time-col { width: 154px; }
    .log-status-col { width: 88px; }
    .log-model-col { width: 30ch; }
    .log-attempt-col { width: 72px; }
    .log-seconds-col { width: 80px; }
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
      th.progress, td.progress, th.result-time, td.result-time { display: none; }
      .order { width: 86px; }
      .row-actions { width: 108px; }
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
        <header>
          <h2>Runtime</h2>
          <div class="header-actions">
            <button class="secondary" id="reloadBtn">Refresh</button>
            <button id="saveConfigBtn">Save runtime</button>
          </div>
        </header>
        <div class="body settings">
          <label>Codex command
            <input id="codexCommand" placeholder="codex">
          </label>
          <label>Listen address
            <input id="listenAddr" placeholder="0.0.0.0:8080">
          </label>
          <label>Timeout seconds
            <input id="timeoutSeconds" type="number" min="1" step="1">
          </label>
          <label>Max output chars
            <input id="maxOutputChars" type="number" min="1" step="1">
          </label>
        </div>
      </aside>

      <section class="panel">
        <header>
          <h2>Models</h2>
          <div class="header-actions models-actions">
            <span class="pill selection-count" id="selectedCount">0/0</span>
            <form class="add-form" id="addForm" hidden>
              <input id="newModel" placeholder="gpt-5 or vendor/gpt-5.5">
              <button type="submit">Add</button>
            </form>
            <button class="secondary" id="clearResultsBtn" type="button">Clear results</button>
            <button class="secondary" id="editModelsBtn" type="button">Edit</button>
            <button class="secondary" id="cancelEditBtn" type="button" hidden>Cancel</button>
            <button id="runSelectedBtn">Run selected</button>
          </div>
        </header>
        <div class="body">
          <div id="modelHost"></div>
        </div>
      </section>
    </section>

    <section class="panel log-panel">
      <header>
        <div class="log-title">
          <h2>Log</h2>
        </div>
        <div class="header-actions">
          <button class="alert-button alert-request" id="notificationBtn" type="button">Allow alerts</button>
        </div>
      </header>
      <div class="body">
        <div class="log-table-wrap">
          <table class="log-table">
            <colgroup>
              <col class="log-time-col">
              <col class="log-status-col">
              <col class="log-model-col">
              <col class="log-attempt-col">
              <col class="log-seconds-col">
              <col class="log-error-col">
            </colgroup>
            <thead><tr><th>Time</th><th>Status</th><th>Model</th><th>Try</th><th>Seconds</th><th>Error</th></tr></thead>
            <tbody id="logList"></tbody>
          </table>
        </div>
      </div>
    </section>
  </main>

  <script>
    const state = { config: null, task: {}, selected: new Set(), results: new Map(), running: false, runningModels: new Set(), logEntries: [], editing: false, draftModels: [], draftModelLoopCounts: {}, dirtyLoopCounts: new Set(), editLockedModels: new Set(), successNotificationsPrimed: false, notifiedSuccessLogKeys: new Set(), modelTableMode: '', modelRowOrder: [] };
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
    function twoDigits(value) { return String(value).padStart(2, '0'); }
    function displayDateTime(value) {
      if (!value) return '';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return date.getFullYear() + '-' + twoDigits(date.getMonth() + 1) + '-' + twoDigits(date.getDate()) + ' ' +
        twoDigits(date.getHours()) + ':' + twoDigits(date.getMinutes()) + ':' + twoDigits(date.getSeconds());
    }
    function visibleModels() { return state.editing ? state.draftModels : (state.config?.models || []); }
    function text(value) { return document.createTextNode(String(value ?? '')); }
    function replaceChildrenWithHTML(element, html) {
      if (element.innerHTML !== html) element.innerHTML = html;
    }
    function notificationsSupported() { return 'Notification' in window; }
    function notificationPermission() { return notificationsSupported() ? Notification.permission : 'unsupported'; }
    function updateNotificationButton() {
      const btn = $('notificationBtn');
      if (!btn) return;
      const permission = notificationPermission();
      btn.className = 'alert-button';
      if (permission === 'unsupported') {
        btn.classList.add('alert-unavailable');
        btn.textContent = 'No alerts';
        btn.disabled = true;
        btn.title = 'Browser notifications are unavailable in this browser or context.';
      } else if (permission === 'granted') {
        btn.classList.add('alert-on');
        btn.textContent = 'Alerts on';
        btn.disabled = false;
        btn.title = 'Background success notifications are enabled.';
      } else if (permission === 'denied') {
        btn.classList.add('alert-blocked');
        btn.textContent = 'Alerts blocked';
        btn.disabled = false;
        btn.title = 'Notifications are blocked for this site. Use browser site settings to allow them.';
      } else {
        btn.classList.add('alert-request');
        btn.textContent = 'Allow alerts';
        btn.disabled = false;
        btn.title = 'Enable browser notifications for background successes.';
      }
    }
    async function requestNotificationPermission(options = {}) {
      const quiet = !!options.quiet;
      if (!notificationsSupported()) {
        if (!quiet) setMessage('Browser notifications are unavailable.');
        updateNotificationButton();
        return false;
      }
      if (Notification.permission === 'granted') {
        updateNotificationButton();
        return true;
      }
      if (Notification.permission === 'denied') {
        if (!quiet) setMessage('Notifications are blocked. Allow them in browser site settings, then refresh.');
        updateNotificationButton();
        return false;
      }
      const permission = await Notification.requestPermission();
      updateNotificationButton();
      if (!quiet) setMessage(permission === 'granted' ? 'Background alerts enabled.' : 'Background alerts not enabled.');
      return permission === 'granted';
    }
    async function autoRequestNotificationPermission() {
      if (notificationPermission() !== 'default') {
        updateNotificationButton();
        return false;
      }
      return requestNotificationPermission({ quiet: true }).catch(err => {
        console.warn('notification permission request failed', err);
        updateNotificationButton();
        return false;
      });
    }
    function installNotificationPermissionPrompt() {
      autoRequestNotificationPermission();
      const requestAfterGesture = (event) => {
        if (event.target?.closest?.('#notificationBtn')) return;
        if (notificationPermission() === 'default') autoRequestNotificationPermission();
      };
      document.addEventListener('pointerdown', requestAfterGesture, { once: true, capture: true });
      document.addEventListener('keydown', requestAfterGesture, { once: true, capture: true });
    }
    function successLogKey(entry) {
      const res = entry.result || entry;
      return [entry.time || '', res.model || '', res.attempts || '', res.duration_ms || ''].join('|');
    }
    function successfulLogEntries(task) {
      return (task?.logs || []).filter(entry => {
        const res = entry.result || entry;
        return res?.success && res.model;
      });
    }
    function recordSuccessLogEntries(entries) {
      entries.forEach(entry => state.notifiedSuccessLogKeys.add(successLogKey(entry)));
    }
    function isPageForeground() {
      return document.visibilityState === 'visible' && document.hasFocus();
    }
    function notifySuccessEntries(entries) {
      if (!entries.length || isPageForeground() || notificationPermission() !== 'granted') return;
      const newest = entries[0].result || entries[0];
      const models = entries.map(entry => (entry.result || entry).model).filter(Boolean);
      const title = entries.length === 1 ? 'API Alive success' : 'API Alive: ' + entries.length + ' successes';
      const body = entries.length === 1
        ? newest.model + ' succeeded in ' + displaySeconds(newest.duration_ms) + 's'
        : 'New successes: ' + models.join(', ');
      try {
        new Notification(title, {
          body,
          requireInteraction: true,
        });
      } catch (err) {
        console.warn('notification failed', err);
      }
    }
    function handleSuccessNotifications(task) {
      const successes = successfulLogEntries(task);
      if (!state.successNotificationsPrimed) {
        recordSuccessLogEntries(successes);
        state.successNotificationsPrimed = true;
        return;
      }
      const fresh = successes.filter(entry => !state.notifiedSuccessLogKeys.has(successLogKey(entry)));
      recordSuccessLogEntries(fresh);
      notifySuccessEntries(fresh);
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
      $('runSelectedBtn').disabled = state.editing || runnableSelectedModels().length === 0;
      $('runSelectedBtn').hidden = state.editing;
      $('clearResultsBtn').hidden = state.editing;
      $('addForm').hidden = !state.editing;
      $('cancelEditBtn').hidden = !state.editing;
      $('cancelEditBtn').disabled = false;
      $('editModelsBtn').disabled = !state.config;
      $('editModelsBtn').textContent = state.editing ? 'Save' : 'Edit';
      $('editModelsBtn').className = state.editing ? '' : 'secondary';
      document.querySelectorAll('[data-run-one]').forEach(btn => btn.disabled = state.editing);
      document.querySelectorAll('[data-stop-one]').forEach(btn => btn.disabled = state.editing || !state.runningModels.has(btn.dataset.stopOne));
      document.querySelectorAll('[data-move]').forEach(btn => { btn.disabled = btn.dataset.boundary === 'true'; });
      renderRunningModels();
    }
    function renderRunningModels() {
      const host = $('runningModels');
      if (!host) return;
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
        const seconds = Math.round((res.duration_ms || 0) / 1000);
        const time = displayDateTime(entry.time);
        const errorText = res.success ? '' : (res.error || 'unknown error');
        return '<tr class="log-entry ' + cls + '">' +
          '<td class="log-time">' + escapeText(time) + '</td>' +
          '<td class="log-status">' + icon + ' ' + status + '</td>' +
          '<td class="log-model">' + escapeText(res.model) + '</td>' +
          '<td>' + escapeText(displayAttemptProgress(entry)) + '</td>' +
          '<td>' + escapeText(seconds) + '</td>' +
          '<td class="log-error" title="' + escapeText(errorText) + '">' + escapeText(errorText) + '</td></tr>';
      }).join('');
    }
    function updateSelectedCount() {
      const models = visibleModels();
      $('selectedCount').textContent = state.editing ? 'Editing' : state.selected.size + '/' + models.length;
      $('runSelectedBtn').disabled = state.editing || runnableSelectedModels().length === 0;
      $('runSelectedBtn').hidden = state.editing;
      $('clearResultsBtn').hidden = state.editing;
      $('addForm').hidden = !state.editing;
      $('cancelEditBtn').hidden = !state.editing;
      $('cancelEditBtn').disabled = false;
      $('editModelsBtn').disabled = !state.config;
      $('editModelsBtn').textContent = state.editing ? 'Save' : 'Edit';
      $('editModelsBtn').className = state.editing ? '' : 'secondary';
      const selectAll = document.querySelector('[data-select-all]');
      if (selectAll) {
        selectAll.checked = models.length > 0 && state.selected.size === models.length;
        selectAll.indeterminate = state.selected.size > 0 && state.selected.size < models.length;
      }
    }
    function runnableSelectedModels() {
      return [...state.selected].filter(model => model);
    }
    function statusPill(model) {
      if (state.runningModels.has(model)) return '<span class="pill run"><span class="live-dot"></span>Running</span>';
      const res = resultFor(model);
      if (!res) return '<span class="pill">Idle</span>';
      return res.success ? '<span class="pill ok">Success (' + escapeText(displaySeconds(res.duration_ms)) + 's)</span>' : '<span class="pill bad">Failed</span>';
    }
    function activeLoopCountFor(model) {
      const counts = state.task?.loop_counts || {};
      return normalizeLoopCount(Object.prototype.hasOwnProperty.call(counts, model) ? counts[model] : loopCountFor(model));
    }
    function displayProgress(model, res) {
      if (!res) return '';
      return displayAttemptNumber(res.attempts || 1) + '/' + loopCountLabel(activeLoopCountFor(model));
    }
    function displayAttemptProgress(entry) {
      const res = entry.result || entry;
      if (!res?.attempts) return '';
      return displayAttemptNumber(res.attempts) + '/' + loopCountLabel(entry.loop_count ?? activeLoopCountFor(res.model));
    }
    function displaySeconds(durationMS) {
      return Math.round((durationMS || 0) / 1000);
    }
    function displayAttemptNumber(value) {
      const count = Number(value) || 1;
      return Math.max(1, Math.floor(count));
    }
    function normalizeLoopCount(value) {
      const digits = String(value ?? '').replace(/\D/g, '').slice(0, 2);
      if (digits === '') return 1;
      return Math.max(0, Math.min(99, Number(digits)));
    }
    function loopCountInputValue(value) {
      return String(normalizeLoopCount(value));
    }
    function loopCountLabel(count) {
      return normalizeLoopCount(count) === 0 ? '∞' : String(normalizeLoopCount(count));
    }
    function modelLoopCounts() { return state.config?.model_loop_counts || {}; }
    function loopCountFor(model) {
      const counts = modelLoopCounts();
      return normalizeLoopCount(Object.prototype.hasOwnProperty.call(counts, model) ? counts[model] : 1);
    }
    function draftLoopCountFor(model) {
      return normalizeLoopCount(Object.prototype.hasOwnProperty.call(state.draftModelLoopCounts, model) ? state.draftModelLoopCounts[model] : 1);
    }
    function mergeDirtyLoopCounts(config) {
      if (!config) return config;
      const merged = { ...config, models: [...(config.models || [])], model_loop_counts: { ...(config.model_loop_counts || {}) } };
      const currentCounts = state.config?.model_loop_counts || {};
      const models = new Set(merged.models || []);
      state.dirtyLoopCounts.forEach(model => {
        if (!models.has(model)) {
          state.dirtyLoopCounts.delete(model);
          return;
        }
        if (currentCounts[model] !== undefined) {
          merged.model_loop_counts[model] = normalizeLoopCount(currentCounts[model]);
        }
      });
      return merged;
    }
    function clearDirtyLoopCounts(models) {
      if (!models) {
        state.dirtyLoopCounts.clear();
        return;
      }
      models.forEach(model => state.dirtyLoopCounts.delete(model));
    }
    function loopCountsForModels(models, source) {
      const out = {};
      models.forEach(model => {
        out[model] = normalizeLoopCount(Object.prototype.hasOwnProperty.call(source || {}, model) ? source[model] : 1);
      });
      return out;
    }
    function loopCountInput(model, value, draft) {
      const attr = draft ? 'data-draft-loop-count' : 'data-loop-count';
      return '<input class="loop-count-input" type="text" inputmode="numeric" pattern="[0-9]*" maxlength="2" ' + attr + '="' + escapeText(model) + '" value="' + escapeText(loopCountInputValue(value)) + '" title="Max attempts; 0 means unlimited">';
    }
    function modelTableNeedsReset(mode, models) {
      return state.modelTableMode !== mode || state.modelRowOrder.join('\n') !== models.join('\n');
    }
    function rememberModelTable(mode, models) {
      state.modelTableMode = mode;
      state.modelRowOrder = [...models];
    }
    function resetModelTableState() {
      state.modelTableMode = '';
      state.modelRowOrder = [];
    }
    function modelRow(model) {
      return [...document.querySelectorAll('[data-model-row]')].find(row => row.dataset.modelRow === model) || null;
    }
    function renderModelEditor(models) {
      if (!models.length) {
        $('modelHost').innerHTML = '<div class="empty">No models configured.</div>';
        resetModelTableState();
        return;
      }
      if (modelTableNeedsReset('edit', models)) {
        $('modelHost').innerHTML = '<table class="model-table model-editor"><thead><tr><th class="model-heading">Model</th><th class="retry-count">Retries</th><th class="editor-actions">Actions</th></tr></thead><tbody id="modelRows"></tbody></table>';
        const tbody = $('modelRows');
        models.forEach(model => tbody.appendChild(createModelEditorRow(model)));
        rememberModelTable('edit', models);
      }
      models.forEach((model, index) => updateModelEditorRow(model, index, models.length));
    }
    function createModelEditorRow(model) {
      const row = document.createElement('tr');
      row.dataset.modelRow = model;

      const modelCell = document.createElement('td');
      modelCell.className = 'model';
      modelCell.title = model;
      modelCell.appendChild(text(model));

      const retryCell = document.createElement('td');
      retryCell.className = 'retry-count';
      retryCell.innerHTML = loopCountInput(model, draftLoopCountFor(model), true);
      const retryInput = retryCell.querySelector('[data-draft-loop-count]');
      retryInput.addEventListener('input', () => {
        retryInput.value = String(retryInput.value).replace(/\D/g, '').slice(0, 2);
        state.draftModelLoopCounts[model] = normalizeLoopCount(retryInput.value);
      });

      const actionCell = document.createElement('td');
      actionCell.className = 'editor-actions';
      actionCell.innerHTML = '<div class="move-actions">' +
        '<button type="button" class="secondary table-action" data-move="up" data-model="' + escapeText(model) + '">Up</button>' +
        '<button type="button" class="secondary table-action" data-move="down" data-model="' + escapeText(model) + '">Down</button>' +
        '<button type="button" class="secondary table-action" data-delete-model="' + escapeText(model) + '">Delete</button>' +
        '</div>';
      actionCell.querySelector('[data-move="up"]').addEventListener('click', () => moveModel(model, 'up'));
      actionCell.querySelector('[data-move="down"]').addEventListener('click', () => moveModel(model, 'down'));
      actionCell.querySelector('[data-delete-model]').addEventListener('click', () => deleteModel(model));

      row.append(modelCell, retryCell, actionCell);
      return row;
    }
    function updateModelEditorRow(model, index, total) {
      const row = modelRow(model);
      if (!row) return;
      const retryInput = row.querySelector('[data-draft-loop-count]');
      if (document.activeElement !== retryInput && retryInput.value !== loopCountInputValue(draftLoopCountFor(model))) {
        retryInput.value = loopCountInputValue(draftLoopCountFor(model));
      }
      const up = row.querySelector('[data-move="up"]');
      const down = row.querySelector('[data-move="down"]');
      const del = row.querySelector('[data-delete-model]');
      up.disabled = index === 0;
      up.dataset.boundary = String(index === 0);
      down.disabled = index === total - 1;
      down.dataset.boundary = String(index === total - 1);
      del.disabled = state.editLockedModels.has(model);
      del.title = state.editLockedModels.has(model) ? 'Running when editing started' : '';
    }
    function createModelsTable(models) {
      $('modelHost').innerHTML = '<table class="model-table"><thead><tr>' +
        '<th class="check"><input data-select-all type="checkbox" title="Select all"></th>' +
        '<th class="model-heading">Model</th><th class="result">Result</th><th class="progress">Progress</th>' +
        '<th class="result-time">Result time</th><th class="retry-count">Retries</th><th class="row-actions"></th>' +
        '</tr></thead><tbody id="modelRows"></tbody></table>';
      const tbody = $('modelRows');
      models.forEach(model => tbody.appendChild(createModelRow(model)));
      document.querySelector('[data-select-all]').addEventListener('change', event => {
        state.selected = event.target.checked ? new Set(visibleModels()) : new Set();
        renderModels();
      });
      rememberModelTable('view', models);
    }
    function createModelRow(model) {
      const row = document.createElement('tr');
      row.dataset.modelRow = model;

      const selectCell = document.createElement('td');
      selectCell.className = 'check';
      const select = document.createElement('input');
      select.type = 'checkbox';
      select.dataset.select = model;
      select.addEventListener('change', () => {
        select.checked ? state.selected.add(model) : state.selected.delete(model);
        updateSelectedCount();
      });
      selectCell.appendChild(select);

      const modelCell = document.createElement('td');
      modelCell.className = 'model';
      modelCell.title = model;
      modelCell.appendChild(text(model));

      const resultCell = document.createElement('td');
      resultCell.className = 'result';
      const progressCell = document.createElement('td');
      progressCell.className = 'progress';
      const timeCell = document.createElement('td');
      timeCell.className = 'result-time';

      const retryCell = document.createElement('td');
      retryCell.className = 'retry-count';
      retryCell.innerHTML = loopCountInput(model, loopCountFor(model), false);
      const retryInput = retryCell.querySelector('[data-loop-count]');
      retryInput.addEventListener('input', () => {
        retryInput.value = String(retryInput.value).replace(/\D/g, '').slice(0, 2);
        if (!state.config.model_loop_counts) state.config.model_loop_counts = {};
        state.config.model_loop_counts[model] = normalizeLoopCount(retryInput.value);
        state.dirtyLoopCounts.add(model);
      });

      const actionCell = document.createElement('td');
      actionCell.className = 'row-actions';
      actionCell.innerHTML = '<div class="row-action-group">' +
        '<button type="button" class="secondary table-action" data-run-one="' + escapeText(model) + '">Run</button>' +
        '<button type="button" class="secondary table-action" data-stop-one="' + escapeText(model) + '">Stop</button>' +
        '</div>';
      actionCell.querySelector('[data-run-one]').addEventListener('click', () => startProbe([model]).catch(err => setMessage(err.message)));
      actionCell.querySelector('[data-stop-one]').addEventListener('click', () => stopProbe(model).catch(err => setMessage(err.message)));

      row.append(selectCell, modelCell, resultCell, progressCell, timeCell, retryCell, actionCell);
      return row;
    }
    function updateModelRow(model) {
      const row = modelRow(model);
      if (!row) return;
      const res = resultFor(model);
      const select = row.querySelector('[data-select]');
      select.checked = state.selected.has(model);
      const resultCell = row.querySelector('.result');
      replaceChildrenWithHTML(resultCell, statusPill(model));
      row.querySelector('.progress').textContent = displayProgress(model, res);
      row.querySelector('.result-time').textContent = res ? displayDateTime(res.updated_at) : '';
      const retryInput = row.querySelector('[data-loop-count]');
      if (document.activeElement !== retryInput && retryInput.value !== loopCountInputValue(loopCountFor(model))) {
        retryInput.value = loopCountInputValue(loopCountFor(model));
      }
      row.querySelector('[data-run-one]').disabled = state.editing;
      row.querySelector('[data-stop-one]').disabled = state.editing || !state.runningModels.has(model);
    }
    function renderModels() {
      const models = visibleModels();
      updateSelectedCount();
      if (state.editing) {
        renderModelEditor(models);
        setBusy(state.running);
        return;
      }
      if (!models.length) {
        $('modelHost').innerHTML = '<div class="empty">No models configured.</div>';
        resetModelTableState();
        return;
      }
      if (modelTableNeedsReset('view', models)) createModelsTable(models);
      models.forEach(updateModelRow);
      updateSelectedCount();
      setBusy(state.running);
    }
    function fillForm(config) {
      $('codexCommand').value = config.codex_command || 'codex';
      $('listenAddr').value = config.listen_addr || '0.0.0.0:8080';
      $('timeoutSeconds').value = config.timeout_seconds || 120;
      $('maxOutputChars').value = config.max_output_chars || 4000;
    }
    function applyTask(task) {
      handleSuccessNotifications(task || {});
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
      state.config = mergeDirtyLoopCounts(data.config);
      state.selected = new Set([...state.selected].filter(model => visibleModels().includes(model)));
      $('configPath').textContent = data.config_path || 'config.json';
      fillForm(state.config);
      applyTask(data.task || {});
      if (state.running) setMessage('Running ' + state.runningModels.size + ' model(s)...');
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
      cfg.model_loop_counts = loopCountsForModels(cfg.models || [], cfg.model_loop_counts || {});
      cfg.max_output_chars = Number($('maxOutputChars').value) || 4000;
      state.config = await request('/api/config', { method: 'POST', body: JSON.stringify(cfg) });
      clearDirtyLoopCounts();
      fillForm(state.config);
      renderModels();
      setMessage('Runtime saved.');
      schedulePoll();
    }
    async function addModel(name) {
      name = name.trim();
      if (!name || !state.editing) return;
      if (!state.draftModels.includes(name)) {
        state.draftModels.push(name);
        state.draftModelLoopCounts[name] = 1;
      }
      $('newModel').value = '';
      renderModels();
      setMessage('Added ' + name);
    }
    function deleteModel(model) {
      if (!state.editing || state.editLockedModels.has(model)) return;
      state.draftModels = state.draftModels.filter(existing => existing !== model);
      delete state.draftModelLoopCounts[model];
      state.selected.delete(model);
      renderModels();
      setMessage('Removed ' + model);
    }
    async function moveModel(model, direction) {
      if (!state.editing) return;
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
      if (!state.config) return;
      if (!state.editing) {
        state.editing = true;
        state.draftModels = [...(state.config.models || [])];
        state.draftModelLoopCounts = loopCountsForModels(state.draftModels, state.config.model_loop_counts || {});
        state.editLockedModels = new Set(state.runningModels);
        state.selected = new Set();
        renderModels();
        setMessage('Editing models.');
        return;
      }
      state.config = await request('/api/config', { method: 'POST', body: JSON.stringify({ ...state.config, models: state.draftModels, model_loop_counts: loopCountsForModels(state.draftModels, state.draftModelLoopCounts) }) });
      clearDirtyLoopCounts();
      state.editing = false;
      state.draftModels = [];
      state.draftModelLoopCounts = {};
      state.editLockedModels = new Set();
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
      state.draftModelLoopCounts = {};
      state.editLockedModels = new Set();
      $('newModel').value = '';
      renderModels();
      setMessage('Edit cancelled.');
    }
    async function startProbe(models) {
      models = [...new Set(models)].filter(model => model);
      if (!models.length || state.editing) return;
      const data = await request('/api/probe', { method: 'POST', body: JSON.stringify({ models, model_loop_counts: loopCountsForModels(models, state.config.model_loop_counts || {}) }) });
      clearDirtyLoopCounts(models);
      if (data.config) state.config = mergeDirtyLoopCounts(data.config);
      applyTask(data.task);
      setMessage('Probe task started.');
      schedulePoll();
    }
    async function stopProbe(model) {
      if (!state.runningModels.has(model)) return;
      const data = await request('/api/probe/stop', { method: 'POST', body: JSON.stringify({ model }) });
      applyTask(data.task);
      setMessage('Stopping ' + model + '...');
      schedulePoll();
    }
    async function clearResults() {
      const data = await request('/api/probe/results/clear', { method: 'POST', body: '{}' });
      applyTask(data.task);
      setMessage('Results cleared.');
      schedulePoll();
    }

    $('reloadBtn').addEventListener('click', () => loadState().catch(err => setMessage(err.message)));
    $('notificationBtn').addEventListener('click', () => requestNotificationPermission().catch(err => setMessage(err.message)));
    $('saveConfigBtn').addEventListener('click', () => saveRuntime().catch(err => setMessage(err.message)));
    $('clearResultsBtn').addEventListener('click', () => clearResults().catch(err => setMessage(err.message)));
    $('addForm').addEventListener('submit', event => { event.preventDefault(); addModel($('newModel').value).catch(err => setMessage(err.message)); });
    $('runSelectedBtn').addEventListener('click', () => startProbe([...state.selected]).catch(err => setMessage(err.message)));
    $('editModelsBtn').addEventListener('click', () => toggleModelEdit().catch(err => setMessage(err.message)));
    $('cancelEditBtn').addEventListener('click', () => cancelModelEdit());
    renderLog();
    renderRunningModels();
    updateNotificationButton();
    installNotificationPermissionPrompt();
    loadState().catch(err => setMessage(err.message));
  </script>
</body>
</html>`
