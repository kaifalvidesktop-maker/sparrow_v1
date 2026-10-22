package main

const linkSaverPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Saved Links - SPARROW</title>
<style>
  * { box-sizing: border-box; }
  :root {
    color-scheme: dark;
    --bg: #0b0b0d;
    --panel: #17181a;
    --border: #292a2e;
    --text: #eee;
    --muted: #999;
    --accent: #4d90fe;
  }
  html[data-theme="light"] {
    color-scheme: light;
    --bg: #f5f7fa;
    --panel: #fff;
    --border: #dfe3e8;
    --text: #1f2328;
    --muted: #647080;
  }
  body { margin: 0; padding: 38px 24px; min-height: 100vh; background: var(--bg); color: var(--text); font: 14px 'Segoe UI', sans-serif; }
  main { width: min(760px, 100%); margin: 0 auto; }
  header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 24px; }
  h1 { margin: 0 0 5px; font-size: 26px; }
  .subtitle { color: var(--muted); }
  button, input { font: inherit; }
  button { cursor: pointer; }
  .primary { padding: 10px 16px; border: 0; border-radius: 9px; background: var(--accent); color: white; font-weight: 600; }
  .header-actions { display: flex; align-items: center; gap: 10px; }
  .home-button { padding: 10px 16px; border: 1px solid var(--border); border-radius: 9px; background: var(--panel); color: var(--text); font-weight: 600; }
  .panel { padding: 20px; border: 1px solid var(--border); border-radius: 14px; background: var(--panel); margin-bottom: 24px; }
  .form-row { display: grid; grid-template-columns: 1fr 1.4fr auto; gap: 10px; }
  input { min-width: 0; padding: 11px 12px; border: 1px solid var(--border); border-radius: 8px; outline: none; background: var(--bg); color: var(--text); }
  input:focus { border-color: var(--accent); }
  #formError { min-height: 18px; margin-top: 9px; color: #ff8a80; }
  h2 { margin: 0 0 12px; font-size: 15px; }
  #links { display: grid; grid-template-columns: repeat(auto-fill, minmax(190px, 1fr)); gap: 12px; }
  .link-card { position: relative; min-width: 0; padding: 15px; padding-right: 44px; border: 1px solid var(--border); border-radius: 12px; background: var(--panel); }
  .link-title { overflow: hidden; color: var(--text); font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
  .link-url { overflow: hidden; margin-top: 6px; color: var(--muted); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
  .link-open { display: block; color: inherit; text-decoration: none; }
  .more { position: absolute; top: 8px; right: 8px; width: 30px; height: 30px; border: 0; border-radius: 7px; background: transparent; color: var(--muted); font-size: 20px; line-height: 1; }
  .more:hover, .menu button:hover { background: color-mix(in srgb, var(--accent) 15%, transparent); color: var(--text); }
  .menu { position: absolute; z-index: 2; top: 40px; right: 8px; display: none; min-width: 112px; padding: 5px; border: 1px solid var(--border); border-radius: 9px; background: var(--panel); box-shadow: 0 8px 24px #0005; }
  .menu.open { display: grid; }
  .menu button { padding: 8px 10px; border: 0; border-radius: 6px; background: transparent; color: var(--text); text-align: left; }
  .menu .delete { color: #ff8a80; }
  .empty { grid-column: 1 / -1; padding: 22px; color: var(--muted); text-align: center; }
  @media (max-width: 620px) { .form-row { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<main>
  <header>
    <div><h1>Saved links</h1><div class="subtitle">Keep your favorite websites one click away on the home screen.</div></div>
    <div class="header-actions">
      <button class="home-button" id="homeButton" type="button">← Home</button>
      <button class="primary" id="saveButton">Add link</button>
    </div>
  </header>
  <section class="panel" aria-label="Add or edit a link">
    <div class="form-row">
      <input id="titleInput" maxlength="120" placeholder="Title">
      <input id="urlInput" type="url" placeholder="https://example.com">
      <button class="primary" id="submitButton">Save link</button>
    </div>
    <div id="formError" role="status"></div>
  </section>
  <h2>Your links</h2>
  <section id="links" aria-live="polite"></section>
</main>
<script>
  let editingID = 0;
  let savedLinks = [];

  function normalizeURL(value) {
    value = (value || '').trim();
    if (!value) return '';
    return /^https?:\/\//i.test(value) ? value : 'https://' + value;
  }

  function resetForm() {
    editingID = 0;
    document.getElementById('titleInput').value = '';
    document.getElementById('urlInput').value = '';
    document.getElementById('submitButton').textContent = 'Save link';
    document.getElementById('formError').textContent = '';
  }

  function editLink(link) {
    editingID = link.ID;
    document.getElementById('titleInput').value = link.Title || '';
    document.getElementById('urlInput').value = link.URL || '';
    document.getElementById('submitButton').textContent = 'Update link';
    document.getElementById('formError').textContent = '';
    document.getElementById('titleInput').focus();
  }

  function makeCard(link) {
    const card = document.createElement('article');
    card.className = 'link-card';
    const open = document.createElement('a');
    open.className = 'link-open';
    open.href = link.URL;
    open.addEventListener('click', function (event) {
      event.preventDefault();
      window.go(link.URL);
    });
    const title = document.createElement('div');
    title.className = 'link-title';
    title.textContent = link.Title;
    const url = document.createElement('div');
    url.className = 'link-url';
    url.textContent = link.URL;
    open.append(title, url);

    const more = document.createElement('button');
    more.className = 'more';
    more.type = 'button';
    more.textContent = '⋮';
    more.setAttribute('aria-label', 'Link options');
    const menu = document.createElement('div');
    menu.className = 'menu';
    const edit = document.createElement('button');
    edit.type = 'button';
    edit.textContent = 'Edit';
    edit.addEventListener('click', function () { editLink(link); });
    const remove = document.createElement('button');
    remove.type = 'button';
    remove.className = 'delete';
    remove.textContent = 'Delete';
    remove.addEventListener('click', async function () {
      if (!window.confirm('Delete "' + link.Title + '"?')) return;
      try {
        await window.deleteLink(link.ID);
        await renderLinks();
      } catch (error) {
        document.getElementById('formError').textContent = String(error);
      }
    });
    more.addEventListener('click', function (event) {
      event.stopPropagation();
      menu.classList.toggle('open');
    });
    menu.append(edit, remove);
    card.append(open, more, menu);
    return card;
  }

  async function renderLinks() {
    savedLinks = await window.getShortcutsNewTab();
    const host = document.getElementById('links');
    host.replaceChildren();
    if (!savedLinks || savedLinks.length === 0) {
      const empty = document.createElement('div');
      empty.className = 'empty';
      empty.textContent = 'No saved links yet. Add your first link above.';
      host.appendChild(empty);
      return;
    }
    savedLinks.forEach(function (link) { host.appendChild(makeCard(link)); });
  }

  async function submitLink() {
    const title = document.getElementById('titleInput').value.trim();
    const url = normalizeURL(document.getElementById('urlInput').value);
    const error = document.getElementById('formError');
    if (!title || !url) {
      error.textContent = 'Enter both a title and a website address.';
      return;
    }
    try {
      if (editingID) {
        await window.updateLink(editingID, title, url);
      } else {
        await window.saveLink(title, url);
      }
      resetForm();
      await renderLinks();
    } catch (err) {
      error.textContent = String(err);
    }
  }

  document.getElementById('submitButton').addEventListener('click', submitLink);
  document.getElementById('saveButton').addEventListener('click', function () {
    resetForm();
    document.getElementById('titleInput').focus();
  });
  document.getElementById('homeButton').addEventListener('click', function () { window.go('sparrow://home'); });
  document.getElementById('urlInput').addEventListener('keydown', function (event) {
    if (event.key === 'Enter') submitLink();
  });
  document.addEventListener('click', function (event) {
    if (!event.target.closest('.link-card')) {
      document.querySelectorAll('.menu.open').forEach(function (menu) { menu.classList.remove('open'); });
    }
  });
  window.getSettings && window.getSettings().then(function (settings) {
    if (settings && settings.theme === 'light') document.documentElement.setAttribute('data-theme', 'light');
  });
  renderLinks();
</script>
</body>
</html>
`
