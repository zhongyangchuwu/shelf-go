package manager

import "html/template"

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Shelf Vault Manager</title>
<style>
:root {
  color-scheme: light;
  --ink: #18211f;
  --slate: #40504c;
  --paper: #f6f2e8;
  --panel: #fffaf0;
  --seal: #2f5d50;
  --wax: #b45f3a;
  --cipher: #d7c7a4;
  --glow: #dbe8d2;
  --shadow: 0 20px 50px rgba(24, 33, 31, .12);
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  background: var(--paper);
  color: var(--ink);
}
* { box-sizing: border-box; }
body { margin: 0; min-height: 100vh; background: radial-gradient(circle at top left, rgba(47, 93, 80, .12), transparent 34rem), var(--paper); }
button, input, textarea { font: inherit; }
button { cursor: pointer; }
button:disabled { cursor: not-allowed; opacity: .55; }
:focus-visible { outline: 3px solid rgba(47, 93, 80, .35); outline-offset: 2px; }
.shell { min-height: 100vh; padding: 28px; }
.header { display: flex; justify-content: space-between; gap: 20px; align-items: flex-start; margin: 0 auto 22px; max-width: 1240px; }
.eyebrow { margin: 0 0 6px; color: var(--seal); font: 700 12px/1 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; letter-spacing: .14em; text-transform: uppercase; }
h1 { margin: 0; font: 700 clamp(32px, 5vw, 54px)/.95 ui-serif, Georgia, Cambria, "Times New Roman", serif; letter-spacing: -.04em; }
.header p { margin: 10px 0 0; color: var(--slate); max-width: 58ch; }
.safety { display: grid; gap: 7px; padding: 14px 16px; min-width: 230px; background: rgba(255, 250, 240, .82); border: 1px solid var(--cipher); border-radius: 18px; box-shadow: var(--shadow); }
.safety span { display: flex; align-items: center; gap: 8px; color: var(--slate); font-size: 13px; }
.safety span::before { content: ""; width: 8px; height: 8px; border-radius: 999px; background: var(--seal); }
.workspace { max-width: 1240px; margin: 0 auto; display: grid; grid-template-columns: minmax(320px, 390px) 1fr; min-height: 680px; border: 1px solid var(--cipher); border-radius: 28px; overflow: hidden; background: rgba(255, 250, 240, .55); box-shadow: var(--shadow); }
.rail { border-right: 1px solid var(--cipher); padding: 20px; background: rgba(255, 250, 240, .72); }
.rail-actions { display: flex; gap: 10px; margin-bottom: 14px; }
.search { margin-bottom: 16px; }
.field { display: grid; gap: 7px; margin-bottom: 14px; }
label { color: var(--slate); font-size: 12px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
input, textarea { width: 100%; border: 1px solid var(--cipher); border-radius: 14px; padding: 11px 12px; color: var(--ink); background: #fffdf8; }
textarea { min-height: 96px; resize: vertical; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.btn { border: 1px solid var(--cipher); border-radius: 999px; padding: 10px 14px; color: var(--ink); background: #fffdf8; font-weight: 700; }
.btn.primary { border-color: var(--seal); color: #fffdf8; background: var(--seal); }
.btn.danger { border-color: rgba(180, 95, 58, .5); color: var(--wax); background: rgba(180, 95, 58, .08); }
.btn.ghost { background: transparent; }
.list { display: grid; gap: 9px; max-height: 510px; overflow: auto; padding-right: 4px; }
.row { width: 100%; text-align: left; border: 1px solid var(--cipher); border-radius: 18px; padding: 13px; background: rgba(255, 253, 248, .72); color: var(--ink); }
.row[aria-selected="true"] { border-color: var(--seal); background: rgba(219, 232, 210, .55); }
.row-title { display: flex; align-items: center; justify-content: space-between; gap: 10px; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-weight: 800; }
.row-meta { color: var(--slate); font-size: 13px; margin-top: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chips { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 10px; }
.chip { display: inline-flex; align-items: center; border-radius: 999px; padding: 4px 8px; background: rgba(215, 199, 164, .45); color: var(--slate); font: 700 12px/1 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.status { color: var(--seal); font-size: 12px; font-weight: 800; }
.bench { padding: 28px; display: grid; grid-template-rows: auto 1fr; gap: 20px; }
.seal { padding: 20px; border: 1px solid var(--cipher); border-radius: 24px; background: var(--panel); }
.seal-path { margin: 0 0 8px; font: 700 clamp(26px, 4vw, 42px)/1 ui-serif, Georgia, Cambria, "Times New Roman", serif; letter-spacing: -.035em; overflow-wrap: anywhere; }
.group { color: var(--seal); }
.key { color: var(--ink); }
.colon { color: var(--wax); padding: 0 .18em; }
.seal p { margin: 0; color: var(--slate); }
.panel { border: 1px solid var(--cipher); border-radius: 24px; background: rgba(255, 250, 240, .74); padding: 20px; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.actions { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 16px; }
.value-box { margin-top: 16px; padding: 14px; min-height: 54px; border: 1px dashed var(--cipher); border-radius: 16px; background: #fffdf8; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; white-space: pre-wrap; overflow-wrap: anywhere; }
.value-box.locked { color: var(--slate); font-family: inherit; }
.toast { position: fixed; right: 22px; bottom: 22px; max-width: 360px; padding: 13px 16px; border-radius: 16px; background: var(--ink); color: #fffdf8; box-shadow: var(--shadow); transform: translateY(20px); opacity: 0; pointer-events: none; transition: opacity .16s ease, transform .16s ease; }
.toast.show { transform: translateY(0); opacity: 1; }
.empty { color: var(--slate); padding: 24px; border: 1px dashed var(--cipher); border-radius: 18px; background: rgba(255, 253, 248, .56); }
@media (max-width: 860px) {
  .shell { padding: 18px; }
  .header { display: block; }
  .safety { margin-top: 16px; min-width: 0; }
  .workspace { grid-template-columns: 1fr; }
  .rail { border-right: 0; border-bottom: 1px solid var(--cipher); }
  .grid { grid-template-columns: 1fr; }
}
@media (prefers-reduced-motion: reduce) { .toast { transition: none; } }
</style>
</head>
<body>
<div class="shell">
  <header class="header">
    <div>
      <p class="eyebrow">Local encrypted vault</p>
      <h1>Shelf manager</h1>
      <p>Edit the records you use every day. Values stay locked until you reveal or copy them.</p>
    </div>
    <aside class="safety" aria-label="Safety boundaries">
      <span>Loopback only</span>
      <span>No cached secret responses</span>
      <span>Values reveal by action</span>
    </aside>
  </header>
  <main class="workspace">
    <section class="rail" aria-label="Vault rail">
      <div class="rail-actions">
        <button class="btn primary" id="addSecret" type="button">Add secret</button>
        <button class="btn ghost" id="refresh" type="button">Refresh</button>
      </div>
      <form class="search" id="searchForm">
        <label for="query">Search vault</label>
        <input id="query" name="q" autocomplete="off" placeholder="Path, env, description, tag">
      </form>
      <div class="list" id="results" role="listbox" aria-label="Secrets"></div>
    </section>
    <section class="bench" aria-label="Inspection bench">
      <div class="seal">
        <p class="seal-path" id="sealPath"><span class="group">Select</span><span class="colon">:</span><span class="key">secret</span></p>
        <p id="sealHint">Choose a record from the vault rail, or add a new one.</p>
      </div>
      <form class="panel" id="editor">
        <div class="grid">
          <div class="field"><label for="path">Path</label><input id="path" name="path" autocomplete="off" placeholder="providers/openai:token" required></div>
          <div class="field"><label for="env">Env name</label><input id="env" name="env" autocomplete="off" placeholder="OPENAI_API_KEY"></div>
        </div>
        <div class="field"><label for="description">Description</label><input id="description" name="description" autocomplete="off" placeholder="Used by local agent experiments"></div>
        <div class="field"><label for="tags">Tags</label><input id="tags" name="tags" autocomplete="off" placeholder="ai local prod"></div>
        <div class="field"><label for="value">Value</label><textarea id="value" name="value" placeholder="Leave empty while editing metadata, or enter a replacement value."></textarea></div>
        <div class="actions">
          <button class="btn primary" type="submit">Save secret</button>
          <button class="btn" id="reveal" type="button">Reveal value</button>
          <button class="btn" id="copy" type="button">Copy value</button>
          <button class="btn" id="hide" type="button">Hide value</button>
          <button class="btn danger" id="delete" type="button">Delete secret</button>
        </div>
        <div class="value-box locked" id="valueBox">Value locked. Reveal or copy only when you need it.</div>
      </form>
    </section>
  </main>
</div>
<div class="toast" id="toast" role="status" aria-live="polite"></div>
<script>
const state = { secrets: [], selected: null, oldPath: '', revealed: '', hideTimer: 0 };
const els = {
  results: document.getElementById('results'), query: document.getElementById('query'), searchForm: document.getElementById('searchForm'),
  addSecret: document.getElementById('addSecret'), refresh: document.getElementById('refresh'), editor: document.getElementById('editor'),
  path: document.getElementById('path'), env: document.getElementById('env'), description: document.getElementById('description'), tags: document.getElementById('tags'), value: document.getElementById('value'),
  reveal: document.getElementById('reveal'), copy: document.getElementById('copy'), hide: document.getElementById('hide'), delete: document.getElementById('delete'), valueBox: document.getElementById('valueBox'),
  sealPath: document.getElementById('sealPath'), sealHint: document.getElementById('sealHint'), toast: document.getElementById('toast')
};
function toast(message) { els.toast.textContent = message; els.toast.classList.add('show'); setTimeout(() => els.toast.classList.remove('show'), 2200); }
function escapeHTML(value) { return String(value || '').replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch])); }
function splitPath(path) { const idx = path.indexOf(':'); return idx < 0 ? [path, ''] : [path.slice(0, idx), path.slice(idx + 1)]; }
function renderSeal(path) { const [group, key] = splitPath(path || 'Select:secret'); els.sealPath.innerHTML = '<span class="group">' + escapeHTML(group) + '</span><span class="colon">:</span><span class="key">' + escapeHTML(key) + '</span>'; }
function selectedPath() { return state.oldPath || els.path.value.trim(); }
function clearReveal() { state.revealed = ''; clearTimeout(state.hideTimer); els.valueBox.textContent = 'Value locked. Reveal or copy only when you need it.'; els.valueBox.classList.add('locked'); }
async function api(path, options = {}) {
  const res = await fetch(path, { ...options, headers: { 'Content-Type': 'application/json', ...(options.headers || {}) } });
  if (!res.ok) throw new Error((await res.text()).trim() || 'Request failed');
  return res.json();
}
async function load() {
  const data = await api('/api/secrets?q=' + encodeURIComponent(els.query.value.trim()), { headers: {} });
  state.secrets = data.secrets || [];
  renderList();
}
function renderList() {
  els.results.innerHTML = '';
  if (state.secrets.length === 0) { const empty = document.createElement('div'); empty.className = 'empty'; empty.textContent = 'No secrets match this view. Add a secret or clear the search.'; els.results.appendChild(empty); return; }
  for (const item of state.secrets) {
    const row = document.createElement('button'); row.type = 'button'; row.className = 'row'; row.setAttribute('role', 'option'); row.setAttribute('aria-selected', item.path === state.oldPath ? 'true' : 'false');
    row.innerHTML = '<div class="row-title"><span>' + escapeHTML(item.path) + '</span><span class="status">' + (item.value_set ? 'value set' : 'empty') + '</span></div>' +
      '<div class="row-meta">' + escapeHTML([item.env, item.description].filter(Boolean).join(' · ') || 'No env or description') + '</div>' +
      '<div class="chips">' + (item.tags || []).map(tag => '<span class="chip">' + escapeHTML(tag) + '</span>').join('') + '</div>';
    row.addEventListener('click', () => selectSecret(item));
    els.results.appendChild(row);
  }
}
function selectSecret(item) {
  state.selected = item; state.oldPath = item.path;
  els.path.value = item.path || ''; els.env.value = item.env || ''; els.description.value = item.description || ''; els.tags.value = (item.tags || []).join(' '); els.value.value = '';
  renderSeal(item.path); els.sealHint.textContent = item.value_set ? 'Value is stored. Reveal only when needed.' : 'No value stored for this record.'; clearReveal(); renderList();
}
function startCreate() { state.selected = null; state.oldPath = ''; els.editor.reset(); renderSeal('New:secret'); els.sealHint.textContent = 'Create a record with a path, value, and optional env, description, and tags.'; clearReveal(); renderList(); els.path.focus(); }
function tagsFromInput() { return els.tags.value.split(/[\s,]+/).map(tag => tag.trim()).filter(Boolean); }
async function saveSecret(event) {
  event.preventDefault();
  const body = { old_path: state.oldPath || undefined, path: els.path.value.trim(), env: els.env.value.trim(), description: els.description.value.trim(), tags: tagsFromInput() };
  if (els.value.value !== '') body.value = els.value.value;
  const method = state.oldPath ? 'PUT' : 'POST';
  await api('/api/secrets', { method, body: JSON.stringify(body) });
  toast(state.oldPath ? 'Saved changes' : 'Saved secret');
  state.oldPath = body.path; clearReveal(); await load();
  const item = state.secrets.find(secret => secret.path === body.path); if (item) selectSecret(item);
}
async function revealValue() {
  const path = selectedPath(); if (!path) { toast('Select a secret first'); return ''; }
  const data = await api('/api/reveal', { method: 'POST', body: JSON.stringify({ path }) });
  state.revealed = data.value || ''; els.valueBox.textContent = state.revealed; els.valueBox.classList.remove('locked');
  clearTimeout(state.hideTimer); state.hideTimer = setTimeout(clearReveal, 60000);
  return state.revealed;
}
async function copyValue() {
  const value = state.revealed || await revealValue();
  await navigator.clipboard.writeText(value);
  toast('Copied value');
}
async function deleteSecret() {
  const path = selectedPath(); if (!path) { toast('Select a secret first'); return; }
  if (!confirm('Delete ' + path + '? This removes it from the vault.')) return;
  await api('/api/secrets?path=' + encodeURIComponent(path), { method: 'DELETE', body: '' });
  toast('Deleted secret'); startCreate(); await load();
}
els.searchForm.addEventListener('submit', event => { event.preventDefault(); load().catch(err => toast(err.message)); });
els.query.addEventListener('input', () => load().catch(err => toast(err.message)));
els.addSecret.addEventListener('click', startCreate);
els.refresh.addEventListener('click', () => load().then(() => toast('Refreshed vault')).catch(err => toast(err.message)));
els.editor.addEventListener('submit', event => saveSecret(event).catch(err => toast(err.message)));
els.reveal.addEventListener('click', () => revealValue().catch(err => toast(err.message)));
els.copy.addEventListener('click', () => copyValue().catch(err => toast(err.message)));
els.hide.addEventListener('click', clearReveal);
els.delete.addEventListener('click', () => deleteSecret().catch(err => toast(err.message)));
startCreate(); load().catch(err => toast(err.message));
</script>
</body>
</html>`))
