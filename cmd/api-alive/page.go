package main

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link id="favicon" rel="icon" type="image/svg+xml">
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
    button, input, select { font: inherit; }
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
    input, select {
      width: 100%;
      min-height: 36px;
      border: 1px solid var(--line);
      border-radius: 6px;
      background: #fff;
      color: var(--text);
      padding: 7px 10px;
      outline: none;
    }
    input:focus, select:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(15, 118, 110, .12); }
    label { font-size: 12px; color: var(--muted); display: grid; gap: 6px; }
    .app { max-width: 1280px; margin: 0 auto; padding: 24px; }
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
      grid-template-columns: 320px minmax(0, 1fr);
      gap: 18px;
      align-items: start;
    }
    .grid > .panel { min-width: 0; }
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
    .add-form { display: grid; grid-template-columns: minmax(180px, 1fr) minmax(96px, 140px) auto; gap: 8px; flex: 1 1 430px; }
    .add-form input, .add-form button, .add-form select {
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
    .attempt-limit-select { width: 62px; min-width: 62px; min-height: 24px; padding: 0 22px 0 10px; text-align: left; text-align-last: left; font-size: 12px; background: #fff; }
    .model-editor .model-heading { width: auto; }
    .editor-actions { width: 76px; }
    .check { width: 34px; }
    .drag { width: 42px; }
    .result { width: 118px; }
    .progress { width: 78px; }
    .attempt-limit { width: 76px; }
    .result-time { width: 154px; }
    .row-actions { width: 108px; }
    .row-action-group { display: flex; gap: 4px; align-items: center; }
    .model-groups { display: grid; gap: 12px; }
    .model-group {
      border: 1px solid var(--line);
      border-radius: 8px;
      overflow: hidden;
      background: #fff;
      transition: border-color .18s ease, box-shadow .18s ease;
    }
    .model-group.drop-target {
      border-color: rgba(37, 99, 235, .45);
      box-shadow: inset 0 0 0 1px rgba(37, 99, 235, .18);
    }
    .model-group.group-drop-target {
      border-color: rgba(15, 118, 110, .48);
      box-shadow: inset 0 0 0 1px rgba(15, 118, 110, .18);
    }
    .model-group.group-dragging { opacity: .55; }
    .model-group-header {
      min-height: 38px;
      padding: 6px 8px;
      display: flex;
      align-items: center;
      gap: 8px;
      background: #fbfcfe;
      border-bottom: 1px solid var(--line);
    }
    .model-group-header.collapsed { border-bottom: 0; }
    .model-group-body {
      height: auto;
      opacity: 1;
      overflow: hidden;
      transition: height .22s ease, opacity .18s ease;
    }
    .model-group-body.collapsed {
      opacity: 0;
    }
    .model-group-body.animating { will-change: height, opacity; }
    .model-group-body-inner {
      min-height: 0;
      overflow: hidden;
    }
    .group-toggle {
      width: 26px;
      min-width: 26px;
      min-height: 26px;
      padding: 0;
      border-color: var(--line);
      background: #fff;
      color: var(--text);
    }
    .group-toggle:hover { background: #f8fafc; }
    .group-drag-handle {
      width: 26px;
      height: 24px;
      border: 1px solid var(--line);
      border-radius: 6px;
      background: #fff;
      color: var(--muted);
      cursor: grab;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      user-select: none;
      touch-action: none;
      font-size: 14px;
    }
    .group-drag-handle:active { cursor: grabbing; }
    .group-title {
      min-width: 0;
      flex: 1 1 auto;
      display: flex;
      align-items: center;
      gap: 8px;
      font-weight: 700;
      font-size: 13px;
    }
    .group-title span:first-child {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .group-count { color: var(--muted); font-weight: 500; }
    .group-select {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      color: var(--muted);
      font-size: 12px;
      white-space: nowrap;
    }
    .group-select input { width: 16px; height: 16px; min-height: 16px; padding: 0; }
    .group-name-input { max-width: 260px; min-height: 28px; padding: 0 8px; font-size: 12px; font-weight: 600; }
    .drag-handle {
      width: 26px;
      height: 24px;
      border: 1px solid var(--line);
      border-radius: 6px;
      background: #fff;
      color: var(--muted);
      cursor: grab;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      user-select: none;
      touch-action: none;
    }
    .drag-handle:active { cursor: grabbing; }
    .model-editor tr {
      transition: transform .18s ease, background-color .18s ease, opacity .18s ease;
    }
    .model-editor tr.dragging { opacity: .5; background: rgba(37, 99, 235, .08); }
    .model-editor tbody.drag-over { outline: 2px dashed rgba(37, 99, 235, .35); outline-offset: -3px; }
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
      .drag { width: 42px; }
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
              <select id="newModelGroup" title="Target group"></select>
              <button type="submit">Add</button>
            </form>
            <button class="secondary" id="addGroupBtn" type="button" hidden>Add group</button>
            <button class="secondary" id="editModelsBtn" type="button">Edit</button>
            <button class="secondary" id="cancelEditBtn" type="button" hidden>Cancel</button>
            <button id="runSelectedBtn">Run selected</button>
            <button class="secondary" id="clearResultsBtn" type="button">Clear results</button>
            <button class="secondary" id="stopAllBtn" type="button">Stop all</button>
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
    const state = { config: null, task: {}, selected: new Set(), results: new Map(), running: false, runningModels: new Set(), logEntries: [], editing: false, draftGroups: [], draftModels: [], draftModelLoopCounts: {}, nextDraftGroupID: 1, dirtyLoopCounts: new Set(), editLockedModels: new Set(), collapsedGroups: new Set(), dragModel: '', dragGroupID: '', optimisticStarts: new Map(), optimisticStops: new Map(), localStopLogs: new Map(), optimisticClearedResults: new Map(), successNotificationsPrimed: false, notifiedSuccessLogKeys: new Set(), failureFaviconPrimed: false, backgroundSuccessFavicon: false, backgroundFailedFavicon: false, handledFailureFaviconKey: '', faviconStatus: '', modelTableMode: '', modelRowOrder: [] };
    const maxLogEntries = 100;
    const idlePollMS = 60000;
    const runningPollMS = 5000;
    const localTrustWindowMS = 12000;
    let pollTimer = null;
    let taskEventSource = null;
    const faviconColors = {
      idle: { fill: '#94a3b8', highlight: '#cbd5e1' },
      running: { fill: '#2563eb', highlight: '#93c5fd' },
      success: { fill: '#22c55e', highlight: '#86efac' },
      failed: { fill: '#ef4444', highlight: '#fca5a5' },
    };
    const faviconCache = {};
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
    function visibleModels() { return state.editing ? state.draftModels : flattenGroups(visibleGroups()); }
    function visibleGroups() { return state.editing ? state.draftGroups : configGroups(state.config); }
    function defaultGroupName() { return 'Default'; }
    function configGroups(config) {
      const groups = (config?.model_groups || []).map(group => ({
        name: String(group.name || defaultGroupName()).trim() || defaultGroupName(),
        models: [...(group.models || [])],
      })).filter(group => group.models.length > 0);
      if (groups.length) return groups;
      const models = [...(config?.models || [])];
      return models.length ? [{ name: defaultGroupName(), models }] : [];
    }
    function flattenGroups(groups) {
      const seen = new Set();
      const models = [];
      (groups || []).forEach(group => (group.models || []).forEach(model => {
        if (!seen.has(model)) {
          seen.add(model);
          models.push(model);
        }
      }));
      return models;
    }
    function groupKey(index, group) { return String(index) + ':' + (group?.name || defaultGroupName()); }
    function newDraftGroupID() { return 'group-' + state.nextDraftGroupID++; }
    function ensureDraftGroupIDs() {
      state.draftGroups.forEach(group => {
        if (!group._id) group._id = newDraftGroupID();
      });
    }
    function text(value) { return document.createTextNode(String(value ?? '')); }
    function replaceChildrenWithHTML(element, html) {
      if (element.innerHTML !== html) element.innerHTML = html;
    }
    function faviconHref(status) {
      if (faviconCache[status]) return faviconCache[status];
      const color = faviconColors[status] || faviconColors.idle;
      const svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="32" height="32">' +
        '<rect x="0" y="0" width="32" height="32" rx="7" fill="#1e293b"/>' +
        '<circle cx="16" cy="16" r="9" fill="' + color.fill + '"/>' +
        '<circle cx="13.5" cy="13.5" r="3" fill="' + color.highlight + '" opacity="0.8"/>' +
        '</svg>';
      faviconCache[status] = 'data:image/svg+xml,' + encodeURIComponent(svg);
      return faviconCache[status];
    }
    function latestTaskResult(task) {
      const logs = task?.logs || [];
      if (logs.length) return logs[0].result || logs[0];
      return (task?.results || []).filter(res => res.updated_at).sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))[0] || null;
    }
    function taskFinishedWithFailure(task) {
      if (!task || task.running || !task.finished_at) return false;
      const latest = latestTaskResult(task);
      return !!latest && latest.success === false;
    }
    function failureFaviconKey(task) {
      if (!taskFinishedWithFailure(task)) return '';
      const latest = latestTaskResult(task);
      return [task.id || '', task.finished_at || '', latest.model || '', latest.updated_at || '', latest.attempts || '', latest.duration_ms || '', latest.error || ''].join('|');
    }
    function desiredFaviconStatus() {
      if (isPageForeground()) return state.running ? 'running' : 'idle';
      if (state.backgroundSuccessFavicon) return 'success';
      if (state.running) return 'running';
      if (state.backgroundFailedFavicon) return 'failed';
      return 'idle';
    }
    function updateFavicon() {
      const status = desiredFaviconStatus();
      if (state.faviconStatus === status) return;
      const icon = $('favicon');
      if (icon) icon.href = faviconHref(status);
      state.faviconStatus = status;
    }
    function handlePageAttentionChange() {
      if (isPageForeground()) {
        state.backgroundSuccessFavicon = false;
        state.backgroundFailedFavicon = false;
      }
      updateFavicon();
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
      if (fresh.length && !isPageForeground()) state.backgroundSuccessFavicon = true;
      notifySuccessEntries(fresh);
    }
    function handleFailureFavicon(task) {
      const key = failureFaviconKey(task);
      if (!state.failureFaviconPrimed) {
        state.handledFailureFaviconKey = key;
        state.failureFaviconPrimed = true;
        return;
      }
      if (!key || key === state.handledFailureFaviconKey) return;
      state.handledFailureFaviconKey = key;
      if (!isPageForeground()) state.backgroundFailedFavicon = true;
    }
    function localTrustExpiresAt() {
      return Date.now() + localTrustWindowMS;
    }
    function isTrustedLocalEntry(entry) {
      return !!entry && Number(entry.expires_at || 0) > Date.now();
    }
    function pruneLocalTrust() {
      state.optimisticStarts.forEach((entry, model) => {
        if (!isTrustedLocalEntry(entry)) state.optimisticStarts.delete(model);
      });
      state.optimisticStops.forEach((entry, model) => {
        if (!isTrustedLocalEntry(entry)) {
          state.optimisticStops.delete(model);
          state.localStopLogs.delete(model);
        }
      });
      state.localStopLogs.forEach((entry, model) => {
        if (!state.optimisticStops.has(model)) state.localStopLogs.delete(model);
      });
      state.optimisticClearedResults.forEach((entry, model) => {
        if (!isTrustedLocalEntry(entry)) state.optimisticClearedResults.delete(model);
      });
    }
    function hasLocalTrust() {
      pruneLocalTrust();
      return state.optimisticStarts.size > 0 || state.optimisticStops.size > 0 || state.optimisticClearedResults.size > 0;
    }
    function resultStateKey(res) {
      if (!res) return '';
      return [res.model || '', res.updated_at || '', res.attempts || '', res.success ? '1' : '0', res.duration_ms || '', res.error || ''].join('|');
    }
    function taskResultFor(task, model) {
      return (task?.results || []).find(res => res.model === model) || null;
    }
    function schedulePoll() {
      clearTimeout(pollTimer);
      pollTimer = setTimeout(() => pollState().catch(err => {
        setMessage(err.message);
        schedulePoll();
      }), state.running || hasLocalTrust() ? runningPollMS : idlePollMS);
    }
    function connectTaskStream() {
      if (taskEventSource || !state.config || !('EventSource' in window)) return;
      taskEventSource = new EventSource('/api/events');
      taskEventSource.onmessage = event => {
        if (!event.data) return;
        try {
          applyTask(JSON.parse(event.data) || {});
        } catch (err) {
          console.warn('task stream parse failed', err);
        }
      };
      taskEventSource.onerror = () => {
        if (taskEventSource?.readyState === EventSource.CLOSED) taskEventSource = null;
      };
    }
    function closeTaskStream() {
      if (!taskEventSource) return;
      taskEventSource.close();
      taskEventSource = null;
    }
    function setBusy(value) {
      state.running = value;
      updateModelHeaderControls();
      document.querySelectorAll('[data-run-one]').forEach(btn => btn.disabled = state.editing);
      document.querySelectorAll('[data-stop-one]').forEach(btn => btn.disabled = state.editing || !state.runningModels.has(btn.dataset.stopOne));
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
        const status = res.success ? 'success' : (isCanceledResult(res) ? 'canceled' : 'failed');
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
      updateModelHeaderControls();
      document.querySelectorAll('[data-select-group]').forEach(selectAll => {
        const group = visibleGroups()[Number(selectAll.dataset.selectGroup)];
        const groupModels = group?.models || [];
        const selectedCount = groupModels.filter(model => state.selected.has(model)).length;
        selectAll.checked = groupModels.length > 0 && selectedCount === groupModels.length;
        selectAll.indeterminate = selectedCount > 0 && selectedCount < groupModels.length;
      });
    }
    function updateModelHeaderControls() {
      $('runSelectedBtn').disabled = state.editing || runnableSelectedModels().length === 0;
      $('runSelectedBtn').hidden = state.editing;
      $('clearResultsBtn').disabled = state.editing || !hasClearableResults();
      $('clearResultsBtn').hidden = state.editing;
      $('stopAllBtn').disabled = state.editing || stoppableModels().length === 0;
      $('stopAllBtn').hidden = state.editing;
      $('addForm').hidden = !state.editing;
      $('addGroupBtn').hidden = !state.editing;
      $('cancelEditBtn').hidden = !state.editing;
      $('cancelEditBtn').disabled = false;
      $('editModelsBtn').disabled = !state.config;
      $('editModelsBtn').textContent = state.editing ? 'Save' : 'Edit';
      $('editModelsBtn').className = state.editing ? '' : 'secondary';
    }
    function runnableSelectedModels() {
      return [...state.selected].filter(model => model);
    }
    function hasClearableResults() {
      return [...state.results.values()].some(res => res?.model && !state.runningModels.has(res.model));
    }
    function stoppableModels() {
      return orderedModelNames([...state.runningModels]).filter(model => state.runningModels.has(model));
    }
    function statusPill(model) {
      if (state.runningModels.has(model)) return '<span class="pill run"><span class="live-dot"></span>Running</span>';
      const res = resultFor(model);
      if (!res) return '<span class="pill">Idle</span>';
      if (res.success) return '<span class="pill ok">Success (' + escapeText(displaySeconds(res.duration_ms)) + 's)</span>';
      return '<span class="pill bad">' + (isCanceledResult(res) ? 'Canceled' : 'Failed') + '</span>';
    }
    function isCanceledResult(res) {
      return res?.success === false && res.error === 'context canceled';
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
    const loopCountChoices = [0, 1, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99];
    function normalizeLoopCountChoice(value) {
      const count = normalizeLoopCount(value);
      let best = loopCountChoices[0];
      loopCountChoices.forEach(choice => {
        const bestDistance = Math.abs(best - count);
        const choiceDistance = Math.abs(choice - count);
        if (choiceDistance < bestDistance || (choiceDistance === bestDistance && choice > best)) best = choice;
      });
      return best;
    }
    function loopCountValue(value) {
      return String(normalizeLoopCountChoice(value));
    }
    function loopCountLabel(count) {
      return normalizeLoopCount(count) === 0 ? '∞' : String(normalizeLoopCount(count));
    }
    function modelLoopCounts() { return state.config?.model_loop_counts || {}; }
    function loopCountFor(model) {
      const counts = modelLoopCounts();
      return normalizeLoopCountChoice(Object.prototype.hasOwnProperty.call(counts, model) ? counts[model] : 1);
    }
    function draftLoopCountFor(model) {
      return normalizeLoopCountChoice(Object.prototype.hasOwnProperty.call(state.draftModelLoopCounts, model) ? state.draftModelLoopCounts[model] : 1);
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
          merged.model_loop_counts[model] = normalizeLoopCountChoice(currentCounts[model]);
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
        out[model] = normalizeLoopCountChoice(Object.prototype.hasOwnProperty.call(source || {}, model) ? source[model] : 1);
      });
      return out;
    }
    function loopCountOptions(value) {
      const selected = normalizeLoopCountChoice(value);
      let html = '';
      loopCountChoices.forEach(count => {
        html += '<option value="' + count + '"' + (count === selected ? ' selected' : '') + '>' + count + '</option>';
      });
      return html;
    }
    function loopCountControl(model, value, draft) {
      const attr = draft ? 'data-draft-attempt-limit' : 'data-attempt-limit';
      return '<select class="attempt-limit-select" autocomplete="off" data-keepassxc-ignore="true" data-lpignore="true" data-1p-ignore="true" data-bwignore="true" data-form-type="other" ' + attr + '="' + escapeText(model) + '" title="0 means unlimited">' + loopCountOptions(value) + '</select>';
    }
    function modelTableNeedsReset(mode) {
      return state.modelTableMode !== mode;
    }
    function rememberModelTable(mode, models = []) {
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
    function captureModelRects() {
      const first = new Map();
      document.querySelectorAll('[data-model-row]').forEach(row => {
        first.set('model:' + row.dataset.modelRow, row.getBoundingClientRect());
      });
      document.querySelectorAll('[data-edit-group-id]').forEach(section => {
        first.set('group:' + section.dataset.editGroupId, section.getBoundingClientRect());
      });
      return first;
    }
    function animateModelRows(first) {
      requestAnimationFrame(() => {
        document.querySelectorAll('[data-model-row], [data-edit-group-id]').forEach(row => {
          const key = row.dataset.modelRow ? 'model:' + row.dataset.modelRow : 'group:' + row.dataset.editGroupId;
          const before = first.get(key);
          if (!before) return;
          const after = row.getBoundingClientRect();
          const dx = before.left - after.left;
          const dy = before.top - after.top;
          if (!dx && !dy) return;
          row.style.transition = 'none';
          row.style.transform = 'translate(' + dx + 'px, ' + dy + 'px)';
          requestAnimationFrame(() => {
            row.style.transition = 'transform .18s ease, background-color .18s ease, opacity .18s ease, border-color .18s ease, box-shadow .18s ease';
            row.style.transform = '';
          });
        });
      });
    }
    function renderModelsAnimated() {
      const first = captureModelRects();
      renderModels();
      animateModelRows(first);
    }
    function updateAddGroupSelect() {
      const select = $('newModelGroup');
      if (!select) return;
      const current = select.value;
      ensureDraftGroupIDs();
      select.innerHTML = state.draftGroups.map(group => '<option value="' + escapeText(group._id) + '">' + escapeText(group.name || defaultGroupName()) + '</option>').join('');
      if (state.draftGroups.some(group => group._id === current)) select.value = current;
      else if (state.draftGroups.length) select.value = state.draftGroups[state.draftGroups.length - 1]._id;
    }
    function setGroupBodyCollapsed(body, collapsed) {
      const initialized = body.dataset.collapseInitialized === 'true';
      const wasCollapsed = body.dataset.collapsed === 'true';
      body.dataset.collapsed = String(collapsed);
      body.setAttribute('aria-hidden', String(collapsed));

      if (!initialized) {
        body.dataset.collapseInitialized = 'true';
        body.classList.toggle('collapsed', collapsed);
        body.style.height = collapsed ? '0px' : '';
        return;
      }
      if (wasCollapsed === collapsed) {
        body.classList.toggle('collapsed', collapsed);
        if (!body.classList.contains('animating')) body.style.height = collapsed ? '0px' : '';
        return;
      }

      if (body._collapseFinish) body.removeEventListener('transitionend', body._collapseFinish);
      const startHeight = body.getBoundingClientRect().height;
      body.classList.add('animating');
      body.style.height = startHeight + 'px';
      body.offsetHeight;
      body.classList.toggle('collapsed', collapsed);
      const targetHeight = collapsed ? 0 : body.scrollHeight;
      body.style.height = targetHeight + 'px';

      body._collapseFinish = event => {
        if (event.target !== body || event.propertyName !== 'height') return;
        body.classList.remove('animating');
        body.style.height = collapsed ? '0px' : '';
        body.removeEventListener('transitionend', body._collapseFinish);
        body._collapseFinish = null;
      };
      body.addEventListener('transitionend', body._collapseFinish);
    }
    function renderModelEditor(models) {
      const groups = visibleGroups();
      ensureDraftGroupIDs();
      updateAddGroupSelect();
      if (!groups.length) {
        $('modelHost').innerHTML = '<div class="empty">No models configured.</div>';
        resetModelTableState();
        return;
      }
      if (modelTableNeedsReset('edit')) {
        $('modelHost').innerHTML = '<div class="model-groups model-editor" id="modelGroups"></div>';
        rememberModelTable('edit', models);
      }
      const host = $('modelGroups');
      const keep = new Set(groups.map(group => group._id));
      [...host.querySelectorAll('[data-edit-group-id]')].forEach(section => {
        if (!keep.has(section.dataset.editGroupId)) section.remove();
      });
      groups.forEach((group, groupIndex) => {
        let section = host.querySelector('[data-edit-group-id="' + group._id + '"]');
        if (!section) {
          section = createModelEditorGroup(group);
        }
        host.appendChild(section);
        updateModelEditorGroup(section, group, groupIndex);
        const tbody = section.querySelector('tbody');
        group.models.forEach(model => {
          let row = modelRow(model);
          if (!row) row = createModelEditorRow(model);
          tbody.appendChild(row);
          updateModelEditorRow(model);
        });
        [...tbody.querySelectorAll('[data-model-row]')].forEach(row => {
          if (!group.models.includes(row.dataset.modelRow)) row.remove();
        });
      });
      models.forEach(updateModelEditorRow);
    }
    function createModelEditorGroup(group) {
      const section = document.createElement('section');
      section.className = 'model-group';
      section.dataset.editGroupId = group._id;
      section.innerHTML = '<div class="model-group-header">' +
        '<span class="group-drag-handle" draggable="true" data-group-drag-handle title="Drag group">≡</span>' +
        '<button type="button" class="group-toggle" data-toggle-group="0">▾</button>' +
        '<input class="group-name-input" data-group-name="0" maxlength="32">' +
        '<span class="group-count"></span>' +
        '</div><div class="model-group-body"><div class="model-group-body-inner"><table class="model-table model-editor"><thead><tr>' +
        '<th class="drag"></th><th class="model-heading">Model</th><th class="attempt-limit">Retries</th><th class="editor-actions">Actions</th>' +
        '</tr></thead><tbody data-drop-group="0"></tbody></table></div></div>';
      section.querySelector('[data-toggle-group]').addEventListener('click', event => toggleGroupCollapsed(Number(event.currentTarget.dataset.toggleGroup || 0)));
      const nameInput = section.querySelector('[data-group-name]');
      nameInput.addEventListener('input', () => {
        const groupIndex = Number(nameInput.dataset.groupName || 0);
        const group = state.draftGroups[groupIndex];
        if (!group) return;
        group.name = nameInput.value.trim() || defaultGroupName();
        updateAddGroupSelect();
      });
      const tbody = section.querySelector('[data-drop-group]');
      installDropTarget(tbody);
      installGroupDropTarget(section);
      installGroupDrag(section);
      return section;
    }
    function updateModelEditorGroup(section, group, groupIndex) {
      section.dataset.editGroup = String(groupIndex);
      section.dataset.editGroupId = group._id;
      const key = groupKey(groupIndex, group);
      const collapsed = state.collapsedGroups.has(key);
      const header = section.querySelector('.model-group-header');
      header.classList.toggle('collapsed', collapsed);
      const toggle = section.querySelector('[data-toggle-group]');
      toggle.dataset.toggleGroup = String(groupIndex);
      toggle.textContent = collapsed ? '▸' : '▾';
      const input = section.querySelector('[data-group-name]');
      input.dataset.groupName = String(groupIndex);
      if (document.activeElement !== input && input.value !== (group.name || defaultGroupName())) input.value = group.name || defaultGroupName();
      section.querySelector('.group-count').textContent = group.models.length + ' model(s)';
      const body = section.querySelector('.model-group-body');
      setGroupBodyCollapsed(body, collapsed);
      const tbody = section.querySelector('[data-drop-group]');
      tbody.dataset.dropGroup = String(groupIndex);
      section.classList.toggle('group-dragging', state.dragGroupID === group._id);
    }
    function createModelEditorRow(model) {
      const row = document.createElement('tr');
      row.dataset.modelRow = model;
      row.draggable = true;

      const dragCell = document.createElement('td');
      dragCell.className = 'drag';
      dragCell.innerHTML = '<span class="drag-handle" title="Drag to reorder">≡</span>';

      const modelCell = document.createElement('td');
      modelCell.className = 'model';
      modelCell.title = model;
      modelCell.appendChild(text(model));

      const retryCell = document.createElement('td');
      retryCell.className = 'attempt-limit';
      retryCell.setAttribute('data-keepassxc-ignore', 'true');
      retryCell.innerHTML = loopCountControl(model, draftLoopCountFor(model), true);
      const retryControl = retryCell.querySelector('[data-draft-attempt-limit]');
      retryControl.addEventListener('change', () => {
        state.draftModelLoopCounts[model] = normalizeLoopCountChoice(retryControl.value);
      });

      const actionCell = document.createElement('td');
      actionCell.className = 'editor-actions';
      actionCell.innerHTML = '<button type="button" class="secondary table-action" data-delete-model="' + escapeText(model) + '">Delete</button>';
      actionCell.querySelector('[data-delete-model]').addEventListener('click', () => deleteModel(model));

      row.addEventListener('dragstart', event => {
        state.dragModel = model;
        row.classList.add('dragging');
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('text/plain', model);
      });
      row.addEventListener('dragover', event => {
        if (!state.dragModel || state.dragModel === model) return;
        event.preventDefault();
        event.stopPropagation();
        event.dataTransfer.dropEffect = 'move';
        const targetGroupIndex = Number(row.closest('[data-drop-group]')?.dataset.dropGroup || 0);
        const targetGroup = state.draftGroups[targetGroupIndex];
        if (!targetGroup) return;
        const rect = row.getBoundingClientRect();
        const targetIndex = targetGroup.models.indexOf(model) + (event.clientY > rect.top + rect.height / 2 ? 1 : 0);
        if (moveDraftModelTo(state.dragModel, targetGroupIndex, targetIndex)) renderModelsAnimated();
      });
      row.addEventListener('dragend', cleanupDragState);

      row.append(dragCell, modelCell, retryCell, actionCell);
      return row;
    }
    function updateModelEditorRow(model) {
      const row = modelRow(model);
      if (!row) return;
      const retryControl = row.querySelector('[data-draft-attempt-limit]');
      if (document.activeElement !== retryControl && retryControl.value !== loopCountValue(draftLoopCountFor(model))) {
        retryControl.value = loopCountValue(draftLoopCountFor(model));
      }
      const del = row.querySelector('[data-delete-model]');
      del.disabled = state.editLockedModels.has(model);
      del.title = state.editLockedModels.has(model) ? 'Running when editing started' : '';
      row.classList.toggle('dragging', state.dragModel === model);
    }
    function installDropTarget(tbody) {
      tbody.addEventListener('dragover', event => {
        if (!state.dragModel || tbody.querySelector('[data-model-row]')) return;
        event.preventDefault();
        event.stopPropagation();
        const groupIndex = Number(tbody.dataset.dropGroup || 0);
        tbody.classList.add('drag-over');
        if (moveDraftModelTo(state.dragModel, groupIndex, 0)) renderModelsAnimated();
      });
      tbody.addEventListener('dragleave', () => tbody.classList.remove('drag-over'));
      tbody.addEventListener('drop', event => {
        event.preventDefault();
        tbody.classList.remove('drag-over');
        cleanupDragState();
      });
    }
    function installGroupDropTarget(section) {
      section.addEventListener('dragover', event => {
        if (!state.dragModel || event.target.closest('[data-model-row]')) return;
        event.preventDefault();
        const groupIndex = Number(section.dataset.editGroup || 0);
        const group = state.draftGroups[groupIndex];
        if (!group) return;
        section.classList.add('drop-target');
        event.dataTransfer.dropEffect = 'move';
        if (moveDraftModelTo(state.dragModel, groupIndex, group.models.length)) renderModelsAnimated();
      });
      section.addEventListener('dragleave', event => {
        if (!section.contains(event.relatedTarget)) section.classList.remove('drop-target');
      });
      section.addEventListener('drop', event => {
        event.preventDefault();
        section.classList.remove('drop-target');
        cleanupDragState();
      });
    }
    function installGroupDrag(section) {
      const handle = section.querySelector('[data-group-drag-handle]');
      handle.addEventListener('dragstart', event => {
        state.dragGroupID = section.dataset.editGroupId || '';
        section.classList.add('group-dragging');
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('text/plain', state.dragGroupID);
      });
      section.addEventListener('dragover', event => {
        if (!state.dragGroupID || state.dragGroupID === section.dataset.editGroupId) return;
        event.preventDefault();
        event.stopPropagation();
        event.dataTransfer.dropEffect = 'move';
        section.classList.add('group-drop-target');
        const targetGroupIndex = Number(section.dataset.editGroup || 0);
        const rect = section.getBoundingClientRect();
        const targetIndex = targetGroupIndex + (event.clientY > rect.top + rect.height / 2 ? 1 : 0);
        if (moveDraftGroupTo(state.dragGroupID, targetIndex)) renderModelsAnimated();
      });
      section.addEventListener('dragleave', event => {
        if (!section.contains(event.relatedTarget)) section.classList.remove('group-drop-target');
      });
      section.addEventListener('drop', event => {
        if (!state.dragGroupID) return;
        event.preventDefault();
        section.classList.remove('group-drop-target');
        cleanupDragState();
      });
      handle.addEventListener('dragend', cleanupDragState);
    }
    function cleanupDragState() {
      state.dragModel = '';
      state.dragGroupID = '';
      document.querySelectorAll('.dragging').forEach(row => row.classList.remove('dragging'));
      document.querySelectorAll('.drag-over').forEach(row => row.classList.remove('drag-over'));
      document.querySelectorAll('.drop-target').forEach(row => row.classList.remove('drop-target'));
      document.querySelectorAll('.group-dragging').forEach(row => row.classList.remove('group-dragging'));
      document.querySelectorAll('.group-drop-target').forEach(row => row.classList.remove('group-drop-target'));
    }
    function moveDraftModelTo(model, targetGroupIndex, targetIndex) {
      if (!state.editing || !model) return false;
      const groups = state.draftGroups;
      const sourceGroupIndex = groups.findIndex(group => group.models.includes(model));
      const targetGroup = groups[targetGroupIndex];
      if (sourceGroupIndex < 0 || !targetGroup) return false;
      const sourceGroup = groups[sourceGroupIndex];
      const sourceIndex = sourceGroup.models.indexOf(model);
      if (sourceGroupIndex === targetGroupIndex && (targetIndex === sourceIndex || targetIndex === sourceIndex + 1)) return false;
      sourceGroup.models.splice(sourceIndex, 1);
      if (sourceGroupIndex === targetGroupIndex && targetIndex > sourceIndex) targetIndex--;
      targetIndex = Math.max(0, Math.min(targetIndex, targetGroup.models.length));
      targetGroup.models.splice(targetIndex, 0, model);
      state.draftModels = flattenGroups(groups);
      return true;
    }
    function moveDraftGroupTo(groupID, targetIndex) {
      if (!state.editing || !groupID) return false;
      const groups = state.draftGroups;
      const sourceIndex = groups.findIndex(group => group._id === groupID);
      if (sourceIndex < 0) return false;
      if (targetIndex === sourceIndex || targetIndex === sourceIndex + 1) return false;
      const [group] = groups.splice(sourceIndex, 1);
      if (targetIndex > sourceIndex) targetIndex--;
      targetIndex = Math.max(0, Math.min(targetIndex, groups.length));
      groups.splice(targetIndex, 0, group);
      state.draftModels = flattenGroups(groups);
      return true;
    }
    function createModelsTable(models) {
      const groups = visibleGroups();
      if (modelTableNeedsReset('view')) {
        $('modelHost').innerHTML = '<div class="model-groups" id="modelGroups"></div>';
        rememberModelTable('view', models);
      }
      const host = $('modelGroups');
      const keep = new Set(groups.map((group, index) => String(index)));
      [...host.querySelectorAll('[data-view-group]')].forEach(section => {
        if (!keep.has(section.dataset.viewGroup)) section.remove();
      });
      groups.forEach((group, groupIndex) => {
        const groupID = String(groupIndex);
        let section = host.querySelector('[data-view-group="' + groupID + '"]');
        if (!section) section = createModelViewGroup(groupIndex);
        host.appendChild(section);
        updateModelViewGroup(section, group, groupIndex);
        const tbody = section.querySelector('tbody');
        group.models.forEach(model => {
          let row = modelRow(model);
          if (!row) row = createModelRow(model);
          tbody.appendChild(row);
        });
        [...tbody.querySelectorAll('[data-model-row]')].forEach(row => {
          if (!group.models.includes(row.dataset.modelRow)) row.remove();
        });
      });
      rememberModelTable('view', models);
    }
    function createModelViewGroup(groupIndex) {
      const section = document.createElement('section');
      section.className = 'model-group';
      section.dataset.viewGroup = String(groupIndex);
      section.innerHTML = '<div class="model-group-header">' +
        '<button type="button" class="group-toggle" data-toggle-group="' + groupIndex + '">▾</button>' +
        '<div class="group-title"><span></span><span class="group-count"></span></div>' +
        '<label class="group-select"><input data-select-group="' + groupIndex + '" type="checkbox">Select group</label>' +
        '</div><div class="model-group-body"><div class="model-group-body-inner"><table class="model-table"><thead><tr>' +
        '<th class="check"></th><th class="model-heading">Model</th><th class="result">Result</th><th class="progress">Progress</th>' +
        '<th class="result-time">Result time</th><th class="attempt-limit">Retries</th><th class="row-actions"></th>' +
        '</tr></thead><tbody></tbody></table></div></div>';
      section.querySelector('[data-toggle-group]').addEventListener('click', () => toggleGroupCollapsed(groupIndex));
      section.querySelector('[data-select-group]').addEventListener('change', event => toggleGroupSelection(groupIndex, event.target.checked));
      return section;
    }
    function updateModelViewGroup(section, group, groupIndex) {
      section.dataset.viewGroup = String(groupIndex);
      const key = groupKey(groupIndex, group);
      const collapsed = state.collapsedGroups.has(key);
      const header = section.querySelector('.model-group-header');
      header.classList.toggle('collapsed', collapsed);
      const toggle = section.querySelector('[data-toggle-group]');
      toggle.dataset.toggleGroup = String(groupIndex);
      toggle.textContent = collapsed ? '▸' : '▾';
      section.querySelector('.group-title span:first-child').textContent = group.name || defaultGroupName();
      section.querySelector('.group-count').textContent = group.models.length + ' model(s)';
      const select = section.querySelector('[data-select-group]');
      select.dataset.selectGroup = String(groupIndex);
      const selectedCount = group.models.filter(model => state.selected.has(model)).length;
      select.checked = group.models.length > 0 && selectedCount === group.models.length;
      select.indeterminate = selectedCount > 0 && selectedCount < group.models.length;
      const body = section.querySelector('.model-group-body');
      setGroupBodyCollapsed(body, collapsed);
    }
    function toggleGroupCollapsed(groupIndex) {
      const group = visibleGroups()[groupIndex];
      if (!group) return;
      const key = groupKey(groupIndex, group);
      if (state.collapsedGroups.has(key)) state.collapsedGroups.delete(key);
      else state.collapsedGroups.add(key);
      renderModels();
    }
    function toggleGroupSelection(groupIndex, checked) {
      const group = visibleGroups()[groupIndex];
      if (!group || state.editing) return;
      group.models.forEach(model => checked ? state.selected.add(model) : state.selected.delete(model));
      renderModels();
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
      retryCell.className = 'attempt-limit';
      retryCell.setAttribute('data-keepassxc-ignore', 'true');
      retryCell.innerHTML = loopCountControl(model, loopCountFor(model), false);
      const retryControl = retryCell.querySelector('[data-attempt-limit]');
      retryControl.addEventListener('change', () => {
        if (!state.config.model_loop_counts) state.config.model_loop_counts = {};
        state.config.model_loop_counts[model] = normalizeLoopCountChoice(retryControl.value);
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
      const retryControl = row.querySelector('[data-attempt-limit]');
      if (document.activeElement !== retryControl && retryControl.value !== loopCountValue(loopCountFor(model))) {
        retryControl.value = loopCountValue(loopCountFor(model));
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
      createModelsTable(models);
      models.forEach(updateModelRow);
      updateSelectedCount();
      setBusy(state.running);
    }
    function renderTaskViews() {
      renderLog();
      renderRunningModels();
      renderModels();
      updateFavicon();
    }
    function fillForm(config) {
      $('codexCommand').value = config.codex_command || 'codex';
      $('listenAddr').value = config.listen_addr || '0.0.0.0:8080';
      $('timeoutSeconds').value = config.timeout_seconds || 120;
      $('maxOutputChars').value = config.max_output_chars || 4000;
    }
    function applyTask(task) {
      pruneLocalTrust();
      task = reconcileOptimisticStarts(task || {});
      task = reconcileOptimisticStops(task || {});
      task = reconcileOptimisticClearedResults(task || {});
      handleSuccessNotifications(task || {});
      handleFailureFavicon(task || {});
      state.task = task || {};
      state.running = !!state.task.running;
      state.runningModels = new Set(state.task.running_models || []);
      state.results = new Map((state.task.results || []).map(res => [res.model, res]));
      state.logEntries = (state.task.logs || []).slice(0, maxLogEntries);
      renderTaskViews();
      connectTaskStream();
      schedulePoll();
    }
    function snapshotClientState() {
      return {
        task: state.task,
        running: state.running,
        runningModels: new Set(state.runningModels),
        results: new Map(state.results),
        logEntries: [...state.logEntries],
        optimisticStarts: new Map(state.optimisticStarts),
        optimisticStops: new Map(state.optimisticStops),
        localStopLogs: new Map(state.localStopLogs),
        optimisticClearedResults: new Map(state.optimisticClearedResults),
      };
    }
    function restoreClientState(snapshot) {
      state.task = snapshot.task;
      state.running = snapshot.running;
      state.runningModels = new Set(snapshot.runningModels);
      state.results = new Map(snapshot.results);
      state.logEntries = [...snapshot.logEntries];
      state.optimisticStarts = new Map(snapshot.optimisticStarts);
      state.optimisticStops = new Map(snapshot.optimisticStops);
      state.localStopLogs = new Map(snapshot.localStopLogs);
      state.optimisticClearedResults = new Map(snapshot.optimisticClearedResults);
      renderLog();
      renderRunningModels();
      renderModels();
      updateFavicon();
    }
    function applyServerState(data) {
      state.config = mergeDirtyLoopCounts(data.config);
      state.selected = new Set([...state.selected].filter(model => visibleModels().includes(model)));
      $('configPath').textContent = data.config_path || 'config.json';
      fillForm(state.config);
      applyTask(data.task || {});
      if (state.running) setMessage('Running ' + state.runningModels.size + ' model(s)...');
      else if (state.optimisticStops.size > 0) setMessage('Stopping ' + state.optimisticStops.size + ' model(s)...');
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
    function orderedModelNames(models) {
      const wanted = new Set(models);
      const ordered = [];
      (state.config?.models || []).forEach(model => {
        if (wanted.has(model)) {
          ordered.push(model);
          wanted.delete(model);
        }
      });
      models.forEach(model => {
        if (wanted.has(model)) {
          ordered.push(model);
          wanted.delete(model);
        }
      });
      return ordered;
    }
    function optimisticStartProbe(models, loopCounts) {
      const currentTask = state.task || {};
      const taskModels = [...(currentTask.models || [])];
      const expiresAt = localTrustExpiresAt();
      models.forEach(model => {
        state.optimisticClearedResults.delete(model);
        state.optimisticStops.delete(model);
        state.localStopLogs.delete(model);
        if (!taskModels.includes(model)) taskModels.push(model);
      });
      const runningModels = orderedModelNames([...state.runningModels, ...models]);
      const nextResults = new Map(state.results);
      models.forEach(model => {
        const result = { model, attempts: 1 };
        state.optimisticStarts.set(model, {
          expires_at: expiresAt,
          loop_count: normalizeLoopCount(loopCounts[model]),
          base_result_key: resultStateKey(state.results.get(model)),
          result,
        });
        nextResults.set(model, result);
      });
      const nextLoopCounts = { ...(currentTask.loop_counts || {}) };
      models.forEach(model => {
        nextLoopCounts[model] = normalizeLoopCount(loopCounts[model]);
      });
      state.task = { ...currentTask, running: true, models: taskModels, running_models: runningModels, results: [...nextResults.values()], logs: currentTask.logs || [], loop_counts: nextLoopCounts, finished_at: '', error: '' };
      state.running = true;
      state.runningModels = new Set(runningModels);
      state.results = nextResults;
      renderRunningModels();
      renderModels();
      updateFavicon();
    }
    function clientTaskCopy(task) {
      return {
        ...(task || {}),
        models: [...(task?.models || [])],
        running_models: [...(task?.running_models || [])],
        results: [...(task?.results || [])],
        logs: [...(task?.logs || [])],
        loop_counts: { ...(task?.loop_counts || {}) },
      };
    }
    function reconcileOptimisticStarts(task) {
      const next = clientTaskCopy(task);
      if (state.optimisticStarts.size === 0) return next;
      state.optimisticStarts.forEach((entry, model) => {
        if (!isTrustedLocalEntry(entry)) {
          state.optimisticStarts.delete(model);
          return;
        }
        const serverRunning = (next.running_models || []).includes(model);
        const serverResult = taskResultFor(next, model);
        const serverHasFreshResult = !!serverResult && !!serverResult.updated_at && resultStateKey(serverResult) !== (entry.base_result_key || '');
        if (!serverRunning && serverHasFreshResult) {
          state.optimisticStarts.delete(model);
          return;
        }
        if (!next.models.includes(model)) next.models.push(model);
        if (!next.running_models.includes(model)) next.running_models = orderedModelNames([...next.running_models, model]);
        if (!serverHasFreshResult) next.results = upsertClientResult(next.results, entry.result);
        next.loop_counts[model] = normalizeLoopCount(entry.loop_count);
      });
      next.running = next.running_models.length > 0;
      return next;
    }
    function optimisticClearResults() {
      const running = new Set(state.runningModels);
      const currentTask = clientTaskCopy(state.task || {});
      const keptResults = currentTask.results.filter(res => running.has(res.model));
      const expiresAt = localTrustExpiresAt();
      currentTask.results.forEach(res => {
        if (!running.has(res.model)) state.optimisticClearedResults.set(res.model, { expires_at: expiresAt, base_result_key: resultStateKey(res) });
      });
      state.task = { ...currentTask, results: keptResults };
      state.results = new Map(keptResults.map(res => [res.model, res]));
      renderModels();
      updateFavicon();
    }
    function canceledResultFor(model) {
      const res = resultFor(model) || {};
      const timestamp = new Date().toISOString();
      return {
        model,
        success: false,
        attempts: displayAttemptNumber(res.attempts || 1),
        duration_ms: 0,
        error: 'context canceled',
        updated_at: timestamp,
      };
    }
    function upsertClientResult(results, result) {
      const next = [...(results || [])];
      const index = next.findIndex(res => res.model === result.model);
      if (index >= 0) next[index] = result;
      else next.push(result);
      return next;
    }
    function isCanceledLogFor(entry, model) {
      const res = entry?.result || entry;
      return res?.model === model && isCanceledResult(res);
    }
    function reconcileOptimisticStops(task) {
      const next = clientTaskCopy(task);
      state.localStopLogs.forEach((entry, model) => {
        if (!state.optimisticStops.has(model)) {
          state.localStopLogs.delete(model);
          return;
        }
        if (next.logs.some(log => isCanceledLogFor(log, model))) {
          state.localStopLogs.delete(model);
          return;
        }
        if (!state.optimisticClearedResults.has(model)) next.results = upsertClientResult(next.results, entry.result);
        next.logs = [entry, ...next.logs.filter(log => !isCanceledLogFor(log, model))].slice(0, maxLogEntries);
      });
      state.optimisticStops.forEach((entry, model) => {
        if (!isTrustedLocalEntry(entry)) {
          state.optimisticStops.delete(model);
          state.localStopLogs.delete(model);
          return;
        }
        next.running_models = next.running_models.filter(runningModel => runningModel !== model);
        if (!state.optimisticClearedResults.has(model)) next.results = upsertClientResult(next.results, entry.result);
      });
      next.running = next.running_models.length > 0;
      return next;
    }
    function reconcileOptimisticClearedResults(task) {
      const next = clientTaskCopy(task);
      if (state.optimisticClearedResults.size === 0) return next;
      const running = new Set(next.running_models || []);
      next.results = next.results.filter(res => {
        const entry = state.optimisticClearedResults.get(res.model);
        if (!entry) return true;
        if (running.has(res.model) || !isTrustedLocalEntry(entry)) {
          state.optimisticClearedResults.delete(res.model);
          return true;
        }
        return false;
      });
      state.optimisticClearedResults.forEach((entry, model) => {
        if (running.has(model) || !isTrustedLocalEntry(entry)) {
          state.optimisticClearedResults.delete(model);
        }
      });
      return next;
    }
    function optimisticStopProbe(model) {
      optimisticStopProbes([model]);
    }
    function optimisticStopProbes(models) {
      const expiresAt = localTrustExpiresAt();
      orderedModelNames([...new Set(models)]).forEach(model => {
        const result = canceledResultFor(model);
        const entry = { time: result.updated_at, loop_count: activeLoopCountFor(model), expires_at: expiresAt, result };
        state.optimisticStarts.delete(model);
        state.optimisticStops.set(model, entry);
        state.localStopLogs.set(model, entry);
      });
      applyTask(state.task || {});
    }
    async function saveRuntime() {
      const cfg = { ...state.config };
      cfg.model_groups = configGroups(cfg);
      cfg.models = flattenGroups(cfg.model_groups);
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
        if (!state.draftGroups.length) state.draftGroups.push({ _id: newDraftGroupID(), name: defaultGroupName(), models: [] });
        const selectedGroupID = $('newModelGroup').value;
        let groupIndex = state.draftGroups.findIndex(group => group._id === selectedGroupID);
        if (groupIndex < 0) groupIndex = Math.max(0, state.draftGroups.length - 1);
        state.draftGroups[groupIndex].models.push(name);
        state.draftModels = flattenGroups(state.draftGroups);
        state.draftModelLoopCounts[name] = 1;
      }
      $('newModel').value = '';
      renderModels();
      setMessage('Added ' + name);
    }
    function addModelGroup() {
      if (!state.editing) return;
      const base = 'Group';
      let index = state.draftGroups.length + 1;
      let name = base + ' ' + index;
      const names = new Set(state.draftGroups.map(group => group.name));
      while (names.has(name)) {
        index++;
        name = base + ' ' + index;
      }
      state.draftGroups.push({ _id: newDraftGroupID(), name, models: [] });
      updateAddGroupSelect();
      $('newModelGroup').value = state.draftGroups[state.draftGroups.length - 1]._id;
      renderModels();
      setMessage('Added group ' + name);
    }
    function deleteModel(model) {
      if (!state.editing || state.editLockedModels.has(model)) return;
      state.draftGroups.forEach(group => {
        group.models = group.models.filter(existing => existing !== model);
      });
      state.draftGroups = state.draftGroups.filter(group => group.models.length > 0 || state.draftGroups.length === 1);
      state.draftModels = flattenGroups(state.draftGroups);
      delete state.draftModelLoopCounts[model];
      state.selected.delete(model);
      renderModels();
      setMessage('Removed ' + model);
    }
    async function toggleModelEdit() {
      if (!state.config) return;
      if (!state.editing) {
        state.editing = true;
        state.nextDraftGroupID = 1;
        state.draftGroups = configGroups(state.config).map(group => ({ _id: newDraftGroupID(), name: group.name, models: [...group.models] }));
        if (!state.draftGroups.length) state.draftGroups = [{ _id: newDraftGroupID(), name: defaultGroupName(), models: [] }];
        state.draftModels = flattenGroups(state.draftGroups);
        state.draftModelLoopCounts = loopCountsForModels(state.draftModels, state.config.model_loop_counts || {});
        state.editLockedModels = new Set(state.runningModels);
        state.selected = new Set();
        renderModels();
        setMessage('Editing models.');
        return;
      }
      const groups = state.draftGroups
        .map(group => ({ name: (group.name || defaultGroupName()).trim() || defaultGroupName(), models: [...group.models] }))
        .filter(group => group.models.length > 0);
      const models = flattenGroups(groups);
      state.config = await request('/api/config', { method: 'POST', body: JSON.stringify({ ...state.config, model_groups: groups, models, model_loop_counts: loopCountsForModels(models, state.draftModelLoopCounts) }) });
      clearDirtyLoopCounts();
      state.editing = false;
      state.draftGroups = [];
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
      state.draftGroups = [];
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
      const previous = snapshotClientState();
      const loopCounts = loopCountsForModels(models, state.config.model_loop_counts || {});
      optimisticStartProbe(models, loopCounts);
      setMessage('Starting ' + models.length + ' model(s)...');
      schedulePoll();
      try {
        const data = await request('/api/probe', { method: 'POST', body: JSON.stringify({ models, model_loop_counts: loopCounts }) });
        clearDirtyLoopCounts(models);
        if (data.config) state.config = mergeDirtyLoopCounts(data.config);
        applyTask(data.task);
        setMessage('Probe task started.');
        schedulePoll();
      } catch (err) {
        restoreClientState(previous);
        await pollState().catch(refreshErr => setMessage(refreshErr.message));
        throw err;
      }
    }
    async function stopProbe(models) {
      models = Array.isArray(models) ? models : [models];
      models = orderedModelNames([...new Set(models)]).filter(model => state.runningModels.has(model));
      if (!models.length || state.editing) return;
      const previous = snapshotClientState();
      optimisticStopProbes(models);
      const label = models.length === 1 ? models[0] : models.length + ' models';
      setMessage('Stopping ' + label + '...');
      schedulePoll();
      try {
        const data = await request('/api/probe/stop', { method: 'POST', body: JSON.stringify({ models }) });
        applyTask(data.task);
        setMessage('Stopping ' + label + '...');
        schedulePoll();
      } catch (err) {
        restoreClientState(previous);
        await pollState().catch(refreshErr => setMessage(refreshErr.message));
        throw err;
      }
    }
    async function stopAllProbes() {
      return stopProbe(stoppableModels());
    }
    async function clearResults() {
      const previous = snapshotClientState();
      optimisticClearResults();
      setMessage('Results cleared.');
      schedulePoll();
      try {
        const data = await request('/api/probe/results/clear', { method: 'POST', body: '{}' });
        applyTask(data.task);
        setMessage('Results cleared.');
        schedulePoll();
      } catch (err) {
        restoreClientState(previous);
        await pollState().catch(refreshErr => setMessage(refreshErr.message));
        throw err;
      }
    }

    $('reloadBtn').addEventListener('click', () => loadState().catch(err => setMessage(err.message)));
    $('notificationBtn').addEventListener('click', () => requestNotificationPermission().catch(err => setMessage(err.message)));
    $('saveConfigBtn').addEventListener('click', () => saveRuntime().catch(err => setMessage(err.message)));
    $('clearResultsBtn').addEventListener('click', () => clearResults().catch(err => setMessage(err.message)));
    $('addForm').addEventListener('submit', event => { event.preventDefault(); addModel($('newModel').value).catch(err => setMessage(err.message)); });
    $('addGroupBtn').addEventListener('click', () => addModelGroup());
    $('runSelectedBtn').addEventListener('click', () => startProbe([...state.selected]).catch(err => setMessage(err.message)));
    $('stopAllBtn').addEventListener('click', () => stopAllProbes().catch(err => setMessage(err.message)));
    $('editModelsBtn').addEventListener('click', () => toggleModelEdit().catch(err => setMessage(err.message)));
    $('cancelEditBtn').addEventListener('click', () => cancelModelEdit());
    document.addEventListener('visibilitychange', handlePageAttentionChange);
    window.addEventListener('focus', handlePageAttentionChange);
    window.addEventListener('blur', handlePageAttentionChange);
    window.addEventListener('beforeunload', closeTaskStream);
    renderLog();
    renderRunningModels();
    updateFavicon();
    updateNotificationButton();
    installNotificationPermissionPrompt();
    loadState().catch(err => setMessage(err.message));
  </script>
</body>
</html>`
