package main

import "encoding/json"

// ---------------------------------------------------
// KEYBOARD SHORTCUTS REGISTRY
// Actual key-capture happens in JS (keydown listener);
// this is the canonical list used to generate the
// sparrow://shortcuts help page and keep bindings consistent.
// ---------------------------------------------------

type Shortcut struct {
	Keys        string `json:"Keys"`
	Description string `json:"Description"`
	Category    string `json:"Category"`
}

var shortcutList = []Shortcut{
	{"Ctrl + T", "Open a new tab", "Tabs"},
	{"Ctrl + W", "Close current tab", "Tabs"},
	{"Ctrl + Shift + T", "Reopen last closed tab", "Tabs"},
	{"Ctrl + Shift + W", "Close all tabs and show the Sparrow home page", "Tabs"},
	{"Ctrl + Tab", "Switch to next tab", "Tabs"},
	{"Ctrl + Shift + Tab", "Switch to previous tab", "Tabs"},
	{"Ctrl + L", "Focus address bar", "Navigation"},
	{"Ctrl + R", "Reload current page", "Navigation"},
	{"Alt + Left", "Go back", "Navigation"},
	{"Alt + Right", "Go forward", "Navigation"},
	{"Ctrl + D", "Add current page to favorites", "Bookmarks"},
	{"Ctrl + Shift + O", "Open bookmarks page", "Bookmarks"},
	{"Ctrl + H", "Open history page", "History"},
	{"Ctrl + J", "Open downloads page", "Downloads"},
	{"Ctrl + Shift + N", "Open an incognito tab", "Privacy"},
	{"Ctrl + Shift + Delete", "Clear browsing history", "Privacy"},
	{"Ctrl + F", "Find in page", "Page"},
	{"Ctrl + Plus", "Zoom in", "Page"},
	{"Ctrl + Minus", "Zoom out", "Page"},
	{"Ctrl + 0", "Reset zoom", "Page"},
	{"F11", "Toggle fullscreen", "Window"},
	{"Ctrl + Shift + R", "Toggle reader mode", "Page"},
	{"Ctrl + ,", "Open settings", "Settings"},
}

func getAllShortcuts() []Shortcut {
	return shortcutList
}

func shortcutsJSON() string {
	data, err := json.Marshal(shortcutList)
	if err != nil {
		return "[]"
	}
	return string(data)
}

const shortcutsPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Shortcuts</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0;
    background: #0b0b0d;
    color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 18px; }
  .panel {
    background: #17181a;
    border: 1px solid #212225;
    border-radius: 12px;
    padding: 18px;
    margin-bottom: 26px;
  }
  .row {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: 12px;
  }
  input {
    background: #111315;
    border: 1px solid #303235;
    border-radius: 8px;
    color: #f0f0f0;
    padding: 10px 12px;
    font-size: 13px;
    min-width: 220px;
    flex: 1;
  }
  button {
    background: #2f74ff;
    border: none;
    border-radius: 8px;
    color: white;
    padding: 10px 14px;
    cursor: pointer;
    font-weight: 600;
  }
  button.danger {
    background: #3b1d1d;
    color: #ff9d9d;
  }
  .group { margin-bottom: 26px; }
  .group-title {
    font-size: 13px;
    color: #4d90fe;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 12px;
  }
  .sc-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    padding: 10px 16px;
    background: #17181a;
    border: 1px solid #202022;
    border-radius: 8px;
    margin-bottom: 6px;
    font-size: 13px;
  }
  .shortcut-link {
    color: #dfe9ff;
    cursor: pointer;
    text-decoration: none;
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .keys {
    background: #232323;
    padding: 3px 10px;
    border-radius: 6px;
    font-family: monospace;
    color: #ccc;
    box-shadow: 0 2px 4px rgba(0,0,0,0.3);
  }
  .empty {
    color: #a9a9b5;
    font-size: 13px;
    opacity: 0.9;
  }
</style>
</head>
<body>
  <h1>Shortcuts</h1>

  <div class="panel">
    <div class="group-title">Add Website Shortcut</div>
    <div class="row">
      <input id="shortcutName" placeholder="Shortcut name (optional)">
      <input id="shortcutURL" placeholder="https://example.com">
      <button id="addShortcutBtn">Add</button>
    </div>
  </div>

  <div class="group">
    <div class="group-title">Saved Shortcuts</div>
    <div id="savedShortcuts"></div>
  </div>

  <div class="group">
    <div class="group-title">Keyboard Shortcuts</div>
    <div id="scContainer"></div>
  </div>

<script>
  const shortcuts = window.__SHORTCUTS__ || [];

  function normalizeURL(raw) {
    const value = (raw || '').trim();
    if (!value) return '';
    return /^https?:\/\//i.test(value) ? value : 'https://' + value;
  }

  function renderSavedShortcuts() {
    if (!window.getShortcutsNewTab) {
      document.getElementById('savedShortcuts').innerHTML = '<div class="empty">Shortcut API unavailable.</div>';
      return;
    }

    window.getShortcutsNewTab().then(function(list) {
      const host = document.getElementById('savedShortcuts');
      if (!list || list.length === 0) {
        host.innerHTML = '<div class="empty">No saved shortcuts yet. Add one above.</div>';
        return;
      }

      host.innerHTML = list.map(function(item, index) {
        const label = (item.Title && item.Title.trim()) || item.URL;
        return '<div class="sc-row">' +
          '<span class="shortcut-link" onclick="window.go(\'' + item.URL + '\')">' + label + '</span>' +
          '<button class="danger" onclick="removeShortcut(' + index + ')">Remove</button>' +
          '</div>';
      }).join('');
    });
  }

  function renderKeyboardShortcuts() {
    const grouped = {};
    shortcuts.forEach(function(s) {
      if (!grouped[s.Category]) grouped[s.Category] = [];
      grouped[s.Category].push(s);
    });
    const container = document.getElementById('scContainer');
    const entries = Object.keys(grouped).map(function(cat) {
      const rows = grouped[cat].map(function(s) {
        return '<div class="sc-row"><span>' + s.Description + '</span><span class="keys">' + s.Keys + '</span></div>';
      }).join('');
      return '<div class="group"><div class="group-title">' + cat + '</div>' + rows + '</div>';
    }).join('');
    container.innerHTML = entries;
  }

  async function addShortcut() {
    const title = document.getElementById('shortcutName').value;
    const url = normalizeURL(document.getElementById('shortcutURL').value);
    if (!url) {
      alert('Please enter a valid website URL.');
      return;
    }

    if (window.addShortcut) {
      await window.addShortcut(title, url);
    }

    document.getElementById('shortcutName').value = '';
    document.getElementById('shortcutURL').value = '';
    renderSavedShortcuts();
  }

  function removeShortcut(index) {
    if (!window.removeShortcut) return;
    window.removeShortcut(index).then(function() {
      renderSavedShortcuts();
    });
  }

  document.getElementById('addShortcutBtn').onclick = addShortcut;
  document.getElementById('shortcutURL').addEventListener('keydown', function(e) {
    if (e.key === 'Enter') addShortcut();
  });

  renderSavedShortcuts();
  renderKeyboardShortcuts();
</script>
</body>
</html>
`
