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
      padding: 10px 8px;
      border-bottom: 1px solid var(--line);
      text-align: left;
      vertical-align: middle;
      font-size: 13px;
    }
    th { color: var(--muted); font-weight: 600; background: #fbfcfe; }
    td.model { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; overflow-wrap: anywhere; }
    .check { width: 34px; }
    .result { width: 110px; }
    .attempts { width: 90px; }
    .duration { width: 86px; }
    .row-actions { width: 132px; }
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
    .log {
      margin-top: 12px;
      min-height: 92px;
      max-height: 180px;
      overflow: auto;
      white-space: pre-wrap;
      overflow-wrap: anywhere;
      background: #111827;
      color: #e5e7eb;
      border-radius: 8px;
      padding: 12px;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 12px;
    }
    @media (max-width: 860px) {
      .app { padding: 16px; }
      .top { display: grid; }
      .status { text-align: left; min-width: 0; }
      .grid { grid-template-columns: 1fr; }
      .toolbar { align-items: stretch; }
      .toolbar .actions { width: 100%; }
      .toolbar .actions button { flex: 1; }
      th.attempts, td.attempts, th.duration, td.duration { display: none; }
      .row-actions { width: 92px; }
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
              <button id="runSelectedBtn">Run selected</button>
              <button class="secondary" id="deleteSelectedBtn">Delete selected</button>
            </div>
          </div>
          <div id="modelHost"></div>
          <div class="log" id="log">Ready.</div>
        </div>
      </section>
    </section>
  </main>

  <script>
    const state = { config: null, selected: new Set(), results: new Map(), running: false };
    const $ = (id) => document.getElementById(id);
    function setMessage(text) { $("message").textContent = text; }
    function resultFor(model) { return state.results.get(model) || null; }
    function escapeText(value) {
      return String(value ?? "").replace(/[&<>"']/g, c => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", "\"":"&quot;", "'":"&#39;" }[c]));
    }
    function setBusy(value) {
      state.running = value;
      $("runSelectedBtn").disabled = value || state.selected.size === 0;
      $("deleteSelectedBtn").disabled = value || state.selected.size === 0;
      document.querySelectorAll("[data-run-one]").forEach(btn => btn.disabled = value);
    }
    function updateSelectedCount() {
      $("selectedCount").textContent = state.selected.size + " selected";
      $("runSelectedBtn").disabled = state.running || state.selected.size === 0;
      $("deleteSelectedBtn").disabled = state.running || state.selected.size === 0;
      const models = state.config?.models || [];
      $("selectAll").checked = models.length > 0 && state.selected.size === models.length;
      $("selectAll").indeterminate = state.selected.size > 0 && state.selected.size < models.length;
    }
    function statusPill(model) {
      const res = resultFor(model);
      if (!res) return "<span class=\"pill\">Idle</span>";
      if (res.running) return "<span class=\"pill run\">Running</span>";
      if (res.success) return "<span class=\"pill ok\">Success</span>";
      return "<span class=\"pill bad\">Failed</span>";
    }
    function renderModels() {
      const models = state.config?.models || [];
      updateSelectedCount();
      if (!models.length) {
        $("modelHost").innerHTML = "<div class=\"empty\">No models configured.</div>";
        return;
      }
      $("modelHost").innerHTML = "<table><thead><tr>" +
        "<th class=\"check\"></th><th>Model</th><th class=\"result\">Result</th><th class=\"attempts\">Attempts</th><th class=\"duration\">Seconds</th><th class=\"row-actions\"></th>" +
        "</tr></thead><tbody>" + models.map(model => {
          const res = resultFor(model);
          const checked = state.selected.has(model) ? "checked" : "";
          const seconds = res && !res.running ? (res.duration_ms / 1000).toFixed(3) : "";
          const attempts = res && !res.running ? res.attempts : "";
          return "<tr>" +
            "<td class=\"check\"><input data-select=\"" + escapeText(model) + "\" type=\"checkbox\" " + checked + "></td>" +
            "<td class=\"model\" title=\"" + escapeText(model) + "\">" + escapeText(model) + "</td>" +
            "<td class=\"result\">" + statusPill(model) + "</td>" +
            "<td class=\"attempts\">" + escapeText(attempts) + "</td>" +
            "<td class=\"duration\">" + escapeText(seconds) + "</td>" +
            "<td class=\"row-actions\"><button class=\"secondary\" data-run-one=\"" + escapeText(model) + "\">Run</button></td>" +
          "</tr>";
        }).join("") + "</tbody></table>";
      document.querySelectorAll("[data-select]").forEach(input => {
        input.addEventListener("change", () => {
          input.checked ? state.selected.add(input.dataset.select) : state.selected.delete(input.dataset.select);
          updateSelectedCount();
        });
      });
      document.querySelectorAll("[data-run-one]").forEach(button => {
        button.addEventListener("click", () => runModels([button.dataset.runOne]));
      });
      updateSelectedCount();
      setBusy(state.running);
    }
    function fillForm(config) {
      $("codexCommand").value = config.codex_command || "codex";
      $("listenAddr").value = config.listen_addr || "0.0.0.0:8080";
      $("timeoutSeconds").value = config.timeout_seconds || 120;
      $("loopCount").value = config.loop_count || 1;
      $("maxOutputChars").value = config.max_output_chars || 4000;
    }
    async function request(path, options = {}) {
      const res = await fetch(path, { headers: { "Content-Type": "application/json" }, ...options });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || ("HTTP " + res.status));
      return data;
    }
    async function loadState() {
      setMessage("Loading...");
      const data = await request("/api/state");
      state.config = data.config;
      state.selected = new Set([...state.selected].filter(model => state.config.models.includes(model)));
      $("configPath").textContent = data.config_path || "config.json";
      fillForm(state.config);
      renderModels();
      setMessage(state.config.models.length + " configured model(s)");
    }
    async function saveRuntime() {
      const cfg = { ...state.config };
      cfg.codex_command = $("codexCommand").value.trim() || "codex";
      cfg.listen_addr = $("listenAddr").value.trim() || "0.0.0.0:8080";
      cfg.timeout_seconds = Number($("timeoutSeconds").value) || 120;
      cfg.loop_count = Number($("loopCount").value) || 1;
      cfg.max_output_chars = Number($("maxOutputChars").value) || 4000;
      state.config = await request("/api/config", { method: "POST", body: JSON.stringify(cfg) });
      fillForm(state.config);
      setMessage("Runtime saved.");
    }
    async function addModel(name) {
      name = name.trim();
      if (!name) return;
      state.config = await request("/api/models", { method: "POST", body: JSON.stringify({ models: [name] }) });
      $("newModel").value = "";
      renderModels();
      setMessage("Added " + name);
    }
    async function deleteSelected() {
      const models = [...state.selected];
      if (!models.length) return;
      state.config = await request("/api/models", { method: "DELETE", body: JSON.stringify({ models }) });
      models.forEach(model => { state.selected.delete(model); state.results.delete(model); });
      renderModels();
      setMessage("Deleted " + models.length + " model(s)");
    }
    async function runModels(models) {
      models = [...new Set(models)].filter(Boolean);
      if (!models.length) return;
      models.forEach(model => state.results.set(model, { running: true }));
      $("log").textContent = "Running " + models.length + " model(s)...";
      renderModels();
      setBusy(true);
      try {
        const data = await request("/api/probe", { method: "POST", body: JSON.stringify({ models }) });
        for (const res of data.results || []) state.results.set(res.model, res);
        const failed = (data.results || []).filter(res => !res.success).length;
        $("log").textContent = (data.results || []).map(res => {
          const status = res.success ? "success" : "failed error=" + (res.error || "");
          return res.model + " " + status + " attempts=" + res.attempts + " seconds=" + (res.duration_ms / 1000).toFixed(3);
        }).join("\n") || "No results.";
        setMessage(failed ? failed + " model(s) failed." : "All selected probes succeeded.");
      } catch (err) {
        $("log").textContent = err.message;
        models.forEach(model => state.results.set(model, { success: false, attempts: 0, duration_ms: 0, error: err.message }));
        setMessage("Probe failed.");
      } finally {
        setBusy(false);
        renderModels();
      }
    }
    $("reloadBtn").addEventListener("click", () => loadState().catch(err => setMessage(err.message)));
    $("saveConfigBtn").addEventListener("click", () => saveRuntime().catch(err => setMessage(err.message)));
    $("addForm").addEventListener("submit", event => {
      event.preventDefault();
      addModel($("newModel").value).catch(err => setMessage(err.message));
    });
    $("selectAll").addEventListener("change", () => {
      const models = state.config?.models || [];
      state.selected = $("selectAll").checked ? new Set(models) : new Set();
      renderModels();
    });
    $("runSelectedBtn").addEventListener("click", () => runModels([...state.selected]));
    $("deleteSelectedBtn").addEventListener("click", () => deleteSelected().catch(err => setMessage(err.message)));
    loadState().catch(err => setMessage(err.message));
  </script>
</body>
</html>`
