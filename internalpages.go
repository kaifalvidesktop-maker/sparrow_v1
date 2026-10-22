package main

import (
	"html"
	"strings"
)

// ---------------------------------------------------
// SPARROW INTERNAL PAGES (sparrow://settings, sparrow://bookmarks, etc.)
// ---------------------------------------------------

func pdfViewerHTMLForURL(rawURL string) string {
	if rawURL == "" {
		return strings.ReplaceAll(pdfViewerPageHTML, "__SPARROW_PDF_URL__", "")
	}
	return strings.ReplaceAll(pdfViewerPageHTML, "__SPARROW_PDF_URL__", html.EscapeString(rawURL))
}

const settingsPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Settings</title>
<style>
  * { box-sizing: border-box; }

  body {
    margin: 0;
    background: #0b0b0d;
    color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif;
    display: flex;
    min-height: 100vh;
  }

  body[data-theme="light"] {
    background: #f5f7fa;
    color: #202124;
  }

  body[data-theme="light"] .sidebar {
    background: #ffffff;
    border-right: 1px solid #e1e4e8;
  }

  body[data-theme="light"] .sidebar .brand,
  body[data-theme="light"] .sidebar .item,
  body[data-theme="light"] .card-title,
  body[data-theme="light"] .row-label,
  body[data-theme="light"] h1,
  body[data-theme="light"] .card-desc,
  body[data-theme="light"] .stat,
  body[data-theme="light"] .row-label {
    color: #202124;
  }

  body[data-theme="light"] .sidebar .item {
    color: #5f6368;
  }

  body[data-theme="light"] .sidebar .item.active {
    background: #e8f0fe;
    color: #1967d2;
    border-left: 3px solid #4d90fe;
  }

  body[data-theme="light"] .content,
  body[data-theme="light"] .card,
  body[data-theme="light"] .btn,
  body[data-theme="light"] select,
  body[data-theme="light"] input[type=text],
  body[data-theme="light"] .switch .knob {
    background: #ffffff;
  }

  body[data-theme="light"] .card,
  body[data-theme="light"] select,
  body[data-theme="light"] input[type=text],
  body[data-theme="light"] .btn {
    border-color: #dadce0;
    color: #202124;
  }

  body[data-theme="light"] .card-desc,
  body[data-theme="light"] .sidebar .item,
  body[data-theme="light"] .row-label,
  body[data-theme="light"] .btn,
  body[data-theme="light"] select,
  body[data-theme="light"] input[type=text] {
    color: #3c4043;
  }

  body[data-theme="light"] .sidebar .item:hover,
  body[data-theme="light"] .btn:hover {
    background: #f1f3f4;
  }

  body[data-theme="light"] .card:hover {
    box-shadow: 0 5px 16px rgba(0,0,0,0.09);
  }

  body[data-theme="light"] .row {
    border-bottom: 1px solid #edf0f2;
  }

  body[data-theme="light"] .switch {
    background: #dadce0;
  }

  body[data-theme="light"] .switch.on {
    background: #4d90fe;
  }

  body[data-theme="light"] .btn.danger {
    border-color: #f1b5b5;
    color: #d93025;
  }

  body[data-theme="light"] .btn.danger:hover {
    background: #fce8e6;
  }

  body[data-theme="light"] .stat {
    color: #188038;
  }

  .sidebar {
    width: 220px;
    background: #111315;
    border-right: 1px solid #232323;
    padding: 24px 0;
    flex-shrink: 0;
  }

  .sidebar .brand {
    font-size: 20px;
    font-weight: 700;
    letter-spacing: 2px;
    padding: 0 24px 24px 24px;
    color: #f0f0f0;
  }

  .sidebar .item {
    padding: 12px 24px;
    color: #b4b4b4;
    cursor: pointer;
    font-size: 13px;
    border-left: 3px solid transparent;
    transition: 0.15s;
  }

  .sidebar .item:hover {
    background: #1a1c1f;
    color: #ffffff;
  }

  .sidebar .item.active {
    background: #1a2638;
    color: #9ec2ff;
    border-left: 3px solid #4d90fe;
  }

  .content {
    flex: 1;
    padding: 40px 60px;
    max-width: 720px;
  }

  h1 {
    font-size: 26px;
    margin-bottom: 30px;
    font-weight: 700;
    color: #f4f4f4;
  }

  .card {
    background: #17181a;
    border: 1px solid #232323;
    border-radius: 14px;
    padding: 22px 26px;
    margin-bottom: 20px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.24);
    transition: transform 0.15s, box-shadow 0.15s;
  }

  .card:hover {
    box-shadow: 0 5px 16px rgba(0,0,0,0.25);
    transform: translateY(-1px);
  }

  .card-title {
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 4px;
    color: #f2f2f2;
  }

  .card-desc {
    font-size: 12px;
    color: #a0a0a0;
    margin-bottom: 16px;
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 0;
    border-bottom: 1px solid #232323;
  }

  .row:last-child { border-bottom: none; }

  .row-label {
    font-size: 13px;
    color: #d7d7d7;
  }

  .switch {
    position: relative;
    width: 42px;
    height: 24px;
    background: #2e2f33;
    border-radius: 20px;
    cursor: pointer;
    transition: 0.2s;
  }

  .switch.on {
    background: #4d90fe;
    box-shadow: 0 0 10px rgba(77,144,254,0.25);
  }

  .switch .knob {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 20px;
    height: 20px;
    background: #ffffff;
    border-radius: 50%;
    transition: 0.2s;
    box-shadow: 0 2px 4px rgba(0,0,0,0.18);
  }

  .switch.on .knob {
    left: 20px;
  }

  select {
    background: #111315;
    color: #f0f0f0;
    border: 1px solid #353535;
    border-radius: 8px;
    padding: 8px 12px;
    font-size: 13px;
    outline: none;
    cursor: pointer;
  }

  input[type=text] {
    background: #111315;
    color: #f0f0f0;
    border: 1px solid #353535;
    border-radius: 8px;
    padding: 8px 12px;
    font-size: 13px;
    outline: none;
    width: 260px;
  }

  input[type=text]:focus,
  select:focus {
    border-color: #4d90fe;
  }

  .btn {
    background: #111315;
    color: #dcdcdc;
    border: 1px solid #353535;
    border-radius: 8px;
    padding: 9px 18px;
    font-size: 12px;
    cursor: pointer;
    transition: 0.15s;
  }

  .btn:hover {
    background: #1a1d21;
    color: #ffffff;
  }

  .btn.danger {
    border-color: #5a2d2d;
    color: #ff8a80;
  }

  .btn.danger:hover {
    background: #2a1b1a;
  }

  .stat {
    font-size: 12px;
    color: #7fe3a0;
    margin-top: 6px;
  }
</style>
</head>
<body>

  <div class="sidebar">
    <div class="brand">SPARROW</div>
    <div class="item active">General</div>
    <div class="item" onclick="window.realNavigate('sparrow://bookmarks')">Bookmarks</div>
    <div class="item" onclick="window.realNavigate('sparrow://downloads')">Downloads</div>
    <div class="item" onclick="window.realNavigate('sparrow://history')">History</div>
  </div>

  <div class="content">
    <h1>Settings</h1>

    <div class="card">
      <div class="card-title">Appearance</div>
      <div class="card-desc">Choose how SPARROW looks</div>
      <div class="row">
        <div class="row-label">Dark Theme</div>
        <div class="switch on" id="themeSwitch" onclick="toggleTheme()"><div class="knob"></div></div>
      </div>
    </div>

    <div class="card">
      <div class="card-title">Search Engine</div>
      <div class="card-desc">Default engine used when typing a search in the address bar</div>
      <div class="row">
        <div class="row-label">Engine</div>
        <select id="engineSelect" onchange="changeEngine()">
          <option value="google">Google</option>
          <option value="bing">Bing</option>
          <option value="brave">Brave Search</option>
          <option value="duckduckgo">DuckDuckGo</option>
          <option value="yahoo">Yahoo</option>
        </select>
      </div>
    </div>

    <div class="card">
      <div class="card-title">Homepage</div>
      <div class="card-desc">Page shown when you open a new tab</div>
      <div class="row">
        <input type="text" id="homepageInput" placeholder="sparrow://home">
        <button class="btn" onclick="saveHomepage()">Save</button>
      </div>
    </div>

    <div class="card">
      <div class="card-title">Privacy & Security Shield</div>
      <div class="card-desc">Control what SPARROW blocks automatically</div>

      <div class="row">
        <div class="row-label">Block Ads</div>
        <div class="switch on" id="adSwitch" onclick="toggleShield('ad')"><div class="knob"></div></div>
      </div>

      <div class="row">
        <div class="row-label">Block Trackers</div>
        <div class="switch on" id="trackerSwitch" onclick="toggleShield('tracker')"><div class="knob"></div></div>
      </div>

      <div class="row">
        <div class="row-label">HTTPS Only Mode</div>
        <div class="switch on" id="httpsSwitch" onclick="toggleShield('https')"><div class="knob"></div></div>
      </div>

      <div class="stat" id="blockedStat">0 threats blocked this session</div>
    </div>

    <div class="card">
      <div class="card-title">Content</div>
      <div class="card-desc">Control what pages are allowed to run</div>

      <div class="row">
        <div class="row-label">JavaScript</div>
        <div class="switch on" id="jsSwitch" onclick="toggleContent('js')"><div class="knob"></div></div>
      </div>

      <div class="row">
        <div class="row-label">Images</div>
        <div class="switch on" id="imgSwitch" onclick="toggleContent('img')"><div class="knob"></div></div>
      </div>
    </div>

    <div class="card">
      <div class="card-title">Data & Privacy</div>
      <div class="card-desc">SPARROW keeps nothing on disk. History and cache live only in RAM.</div>

      <div class="row">
        <div class="row-label">Clear all browsing history</div>
        <button class="btn danger" onclick="clearAllHistory()">Clear Now</button>
      </div>

      <div class="row">
        <div class="row-label">Reset all settings to default</div>
        <button class="btn danger" onclick="resetAll()">Reset</button>
      </div>
    </div>

  </div>

<script>
  function applyThemeState(theme) {
    const t = theme === 'light' ? 'light' : 'dark';
    try {
      localStorage.setItem('sparrow-theme', t);
    } catch (e) {}
    document.body.setAttribute('data-theme', t);
  }

  async function loadSettings() {
    const s = await window.getSettings();
    const savedTheme = (s && s.theme) ? s.theme : (localStorage.getItem('sparrow-theme') || 'dark');

    applyThemeState(savedTheme);
    document.getElementById('themeSwitch').className =
      (savedTheme === 'dark' ? 'switch on' : 'switch');

    document.getElementById('engineSelect').value = s.searchEngine;
    document.getElementById('homepageInput').value = s.homepage;

    document.getElementById('jsSwitch').className =
      s.jsEnabled ? 'switch on' : 'switch';

    document.getElementById('imgSwitch').className =
      s.imagesEnabled ? 'switch on' : 'switch';

    const stats = await window.getShieldStats();

    document.getElementById('adSwitch').className =
      stats.adBlock ? 'switch on' : 'switch';

    document.getElementById('trackerSwitch').className =
      stats.trackerBlock ? 'switch on' : 'switch';

    document.getElementById('httpsSwitch').className =
      stats.httpsOnly ? 'switch on' : 'switch';

    document.getElementById('blockedStat').innerText =
      stats.blockedCount + ' threats blocked this session';
  }

  async function toggleTheme() {
    const el = document.getElementById('themeSwitch');
    const isOn = el.classList.contains('on');
    const newTheme = isOn ? 'light' : 'dark';

    await window.setTheme(newTheme);
    applyThemeState(newTheme);
    el.className = !isOn ? 'switch on' : 'switch';
  }

  async function changeEngine() {
    const val = document.getElementById('engineSelect').value;
    await window.setSearchEngine(val);
    if (window.updateSearchPlaceholder) {
      window.updateSearchPlaceholder(val);
    }
  }

  async function saveHomepage() {
    const val = document.getElementById('homepageInput').value;

    await window.setHomepage(val);

    alert('Homepage saved');
  }

  async function toggleShield(kind) {
    let id, fn;

    if (kind === 'ad') {
      id = 'adSwitch';
      fn = window.toggleAdBlock;
    }

    if (kind === 'tracker') {
      id = 'trackerSwitch';
      fn = window.toggleTrackerBlock;
    }

    if (kind === 'https') {
      id = 'httpsSwitch';
      fn = window.toggleHTTPSOnly;
    }

    const result = await fn();

    document.getElementById(id).className =
      result ? 'switch on' : 'switch';
  }

  async function toggleContent(kind) {
    let id, fn;

    if (kind === 'js') {
      id = 'jsSwitch';
      fn = window.toggleJS;
    }

    if (kind === 'img') {
      id = 'imgSwitch';
      fn = window.toggleImages;
    }

    const result = await fn();

    document.getElementById(id).className =
      result ? 'switch on' : 'switch';
  }

  async function clearAllHistory() {
    if (!confirm('Clear all browsing history?')) return;

    await window.clearHistory('all');

    alert('History cleared');
  }

  async function resetAll() {
    if (!confirm('Reset all settings to default?')) return;

    try {
      localStorage.setItem('sparrow-theme', 'dark');
      localStorage.setItem('sparrow-home-theme', 'dark');
    } catch (e) {}

    await window.resetSettings();

    applyThemeState('dark');
    document.getElementById('themeSwitch').className = 'switch on';
    loadSettings();
  }

  loadSettings();
</script>

</body>
</html>
`

const pdfViewerPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Sparrow PDF Viewer</title>
<style>
  * { box-sizing: border-box; }
  html, body {
    margin: 0;
    width: 100%;
    min-height: 100%;
    background: #111214;
    color: #e9eaee;
    font-family: "Segoe UI", sans-serif;
  }
  body.light {
    background: #f3f5f8;
    color: #20242a;
  }
  .pdf-page {
    min-height: calc(100vh - 116px);
    background: #202226;
    display: flex;
    flex-direction: column;
  }
  body.light .pdf-page { background: #e8ebef; }

  .pdf-toolbar {
    height: 52px;
    flex: 0 0 52px;
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 7px 10px;
    background: rgba(25,27,30,.98);
    border-bottom: 1px solid rgba(255,255,255,.08);
    box-shadow: 0 2px 10px rgba(0,0,0,.16);
    position: sticky;
    top: 0;
    z-index: 10;
  }
  body.light .pdf-toolbar {
    background: rgba(255,255,255,.98);
    border-bottom-color: rgba(0,0,0,.08);
  }
  .pdf-btn {
    border: 1px solid rgba(255,255,255,.08);
    background: #303237;
    color: #e9eaee;
    height: 34px;
    min-width: 34px;
    padding: 0 11px;
    border-radius: 9px;
    cursor: pointer;
    font: 600 12px "Segoe UI", sans-serif;
    transition: .15s ease;
  }
  .pdf-btn:hover { background: #3a3c42; transform: translateY(-1px); }
  body.light .pdf-btn {
    background: #f6f7f9;
    color: #252a31;
    border-color: #dfe3e8;
  }
  body.light .pdf-btn:hover { background: #e9edf2; }
  .pdf-btn.primary { background: #4d90fe; border-color: #4d90fe; color: white; }
  .pdf-btn.icon { width: 34px; padding: 0; font-size: 15px; }
  .pdf-title {
    flex: 1;
    min-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #d9dde5;
    font-size: 12px;
    padding: 0 7px;
  }
  body.light .pdf-title { color: #3a414b; }

  .page-controls, .zoom-controls {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px;
    border: 1px solid rgba(255,255,255,.07);
    background: rgba(255,255,255,.035);
    border-radius: 10px;
  }
  body.light .page-controls, body.light .zoom-controls {
    background: #f2f4f7;
    border-color: #e0e4e9;
  }
  .page-number {
    width: 48px;
    height: 28px;
    border: 0;
    outline: 0;
    border-radius: 7px;
    background: transparent;
    color: inherit;
    text-align: center;
    font-size: 12px;
  }
  .page-total, .zoom-value {
    font-size: 12px;
    color: #aeb4be;
    padding: 0 5px;
  }
  body.light .page-total, body.light .zoom-value { color: #66707d; }

  .pdf-workspace {
    flex: 1;
    min-height: 0;
    position: relative;
    display: flex;
  }
  .pdf-canvas {
    flex: 1;
    overflow: auto;
    padding: 28px 24px 50px;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    background:
      radial-gradient(circle at 50% 0%, rgba(255,255,255,.045), transparent 35%),
      #202226;
  }
  body.light .pdf-canvas {
    background:
      radial-gradient(circle at 50% 0%, rgba(255,255,255,.9), transparent 40%),
      #e8ebef;
  }
  .pdf-canvas canvas {
    display: block;
    background: white;
    box-shadow: 0 12px 36px rgba(0,0,0,.28);
    border-radius: 3px;
  }
  .pdf-empty {
    display: none;
    align-items: center;
    justify-content: center;
    flex: 1;
    color: #aeb4be;
    font-size: 13px;
  }
  .pdf-empty.show { display: flex; }
  .loading {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #aeb4be;
    font-size: 13px;
    background: rgba(20,22,25,.4);
    z-index: 5;
  }
  .loading.hide { display: none; }
</style>
</head>
<body>
<div id="pdf-config" data-pdf-url="__SPARROW_PDF_URL__" hidden></div>
<div class="pdf-page">
  <div class="pdf-toolbar">
    <button class="pdf-btn icon" onclick="goBack()" title="Back">←</button>
    <button class="pdf-btn icon" onclick="reloadPDF()" title="Reload">↻</button>

    <div class="page-controls">
      <button class="pdf-btn icon" onclick="prevPage()" title="Previous page">‹</button>
      <input class="page-number" id="pageNumber" value="1" aria-label="Page number"
             onkeydown="if(event.key==='Enter') jumpPage()">
      <span class="page-total">/ <span id="pageTotal">—</span></span>
      <button class="pdf-btn icon" onclick="nextPage()" title="Next page">›</button>
    </div>

    <div class="zoom-controls">
      <button class="pdf-btn icon" onclick="zoomOut()" title="Zoom out">−</button>
      <span class="zoom-value" id="zoomValue">120%</span>
      <button class="pdf-btn icon" onclick="zoomIn()" title="Zoom in">+</button>
    </div>

    <button class="pdf-btn" onclick="fitToWidth()">Fit width</button>
    <button class="pdf-btn" onclick="fitToPage()">Fit page</button>
    <button class="pdf-btn" onclick="rotatePage()">Rotate</button>
    <button class="pdf-btn" id="bookmarkBtn" onclick="toggleBookmark()">Bookmark</button>
    <button class="pdf-btn" onclick="searchPDF()">Find</button>
    <button class="pdf-btn" onclick="downloadPDF()">Download</button>
    <button class="pdf-btn primary" onclick="window.print()">Print</button>
  </div>

  <div class="pdf-workspace">
    <div class="pdf-canvas" id="pdfCanvas"></div>
    <div class="pdf-empty" id="pdfEmpty">No PDF loaded.</div>
    <div class="loading" id="loading">Loading PDF…</div>
  </div>
</div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js"></script>
<script>
  var pdfjsLib = window.pdfjsLib;
  var pdfDoc = null;
  var currentPage = 1;
  var pageScale = 1.2;
  var rotation = 0;
  var currentUrl = '';

  function getURLParam(name) {
    try {
      return new URLSearchParams(window.location.search).get(name) || '';
    } catch (e) { return ''; }
  }

  function getOriginalURL() {
    var config = document.getElementById('pdf-config');
    return config ? config.getAttribute('data-pdf-url') || '' : '';
  }

  function applyTheme() {
    try {
      var theme = localStorage.getItem('sparrow-shell-theme') || 'dark';
      document.body.classList.toggle('light', theme === 'light');
    } catch (e) {}
  }

  function setLoading(on) {
    var el = document.getElementById('loading');
    if (el) el.classList.toggle('hide', !on);
  }

  function updateUI() {
    document.getElementById('pageNumber').value = currentPage;
    document.getElementById('pageTotal').textContent = pdfDoc ? pdfDoc.numPages : '—';
    document.getElementById('zoomValue').textContent = Math.round(pageScale * 100) + '%';
  }

  function renderPage(pageNum) {
    if (!pdfDoc || pageNum < 1 || pageNum > pdfDoc.numPages) return;
    currentPage = pageNum;
    updateUI();

    pdfDoc.getPage(pageNum).then(function(page) {
      var viewport = page.getViewport({ scale: pageScale, rotation: rotation });
      var canvas = document.createElement('canvas');
      var ctx = canvas.getContext('2d', { alpha: false });
      canvas.width = Math.floor(viewport.width);
      canvas.height = Math.floor(viewport.height);

      setLoading(true);
      return page.render({ canvasContext: ctx, viewport: viewport }).promise.then(function() {
        var container = document.getElementById('pdfCanvas');
        container.innerHTML = '';
        container.appendChild(canvas);
        setLoading(false);
        container.scrollTop = 0;
      });
    }).catch(function() {
      setLoading(false);
      document.getElementById('pdfCanvas').innerHTML =
        '<div style="padding:30px;color:#aeb4be">This page could not be rendered.</div>';
    });
  }

  function prevPage() {
    if (currentPage > 1) renderPage(currentPage - 1);
  }
  function nextPage() {
    if (pdfDoc && currentPage < pdfDoc.numPages) renderPage(currentPage + 1);
  }
  function jumpPage() {
    var n = parseInt(document.getElementById('pageNumber').value, 10);
    if (isFinite(n)) renderPage(Math.max(1, Math.min(pdfDoc ? pdfDoc.numPages : 1, n)));
  }

  function zoomIn() {
    pageScale = Math.min(4, pageScale + 0.1);
    renderPage(currentPage);
    updateUI();
  }
  function zoomOut() {
    pageScale = Math.max(0.4, pageScale - 0.1);
    renderPage(currentPage);
    updateUI();
  }

  function fitToWidth() {
    if (!pdfDoc) return;
    pdfDoc.getPage(currentPage).then(function(page) {
      var base = page.getViewport({scale: 1, rotation: rotation});
      var box = document.getElementById('pdfCanvas');
      pageScale = Math.max(0.4, Math.min(4, (box.clientWidth - 48) / base.width));
      renderPage(currentPage);
      updateUI();
    });
  }

  function fitToPage() {
    if (!pdfDoc) return;
    pdfDoc.getPage(currentPage).then(function(page) {
      var base = page.getViewport({scale: 1, rotation: rotation});
      var box = document.getElementById('pdfCanvas');
      var availableH = box.clientHeight - 48;
      pageScale = Math.max(0.4, Math.min(4, Math.min(
        (box.clientWidth - 48) / base.width,
        availableH / base.height
      )));
      renderPage(currentPage);
      updateUI();
    });
  }

  function rotatePage() {
    rotation = (rotation + 90) % 360;
    renderPage(currentPage);
  }

  function goBack() {
    if (window.goBackNav) {
      window.goBackNav();
    } else {
      history.back();
    }
  }

  function reloadPDF() {
    loadPDF(getOriginalURL());
  }

  function setBookmarkState(saved) {
    var btn = document.getElementById('bookmarkBtn');
    if (!btn) return;
    btn.textContent = saved ? 'Bookmarked' : 'Bookmark';
    btn.classList.toggle('primary', saved);
  }

  async function toggleBookmark() {
    if (!currentUrl || !window.toggleBookmark) return;
    try {
      var saved = await window.toggleBookmark(currentUrl, currentUrl);
      setBookmarkState(saved);
    } catch (e) {}
  }

  function downloadPDF() {
    if (!currentUrl) return;
    if (window.downloadFile) window.downloadFile(currentUrl);
    else window.open(currentUrl, '_blank');
  }

  function renderFallback(url) {
    var frame = document.createElement('iframe');
    frame.src = url;
    frame.style.cssText = 'width:100%;height:100%;border:0;background:white;';
    var container = document.getElementById('pdfCanvas');
    container.innerHTML = '';
    container.appendChild(frame);
    setLoading(false);
  }

  function loadPDF(url) {
    currentUrl = url;
    document.title = (url.split('/').pop() || 'PDF') + ' — Sparrow PDF Viewer';
    document.getElementById('pdfEmpty').classList.remove('show');
    setLoading(true);

    if (!url) {
      setLoading(false);
      document.getElementById('pdfEmpty').classList.add('show');
      return;
    }

    if (!pdfjsLib) {
      renderFallback(url);
      return;
    }

    pdfjsLib.GlobalWorkerOptions.workerSrc =
      'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';

    pdfDoc = null;
    pdfjsLib.getDocument({url: url, withCredentials: false}).promise.then(function(doc) {
      pdfDoc = doc;
      currentPage = 1;
      rotation = 0;
      pageScale = 1.2;
      updateUI();
      renderPage(1);
      if (window.isBookmarked) {
        window.isBookmarked(url).then(setBookmarkState).catch(function(){});
      }
    }).catch(function() {
      pdfDoc = null;
      updateUI();
      renderFallback(url);
      if (window.isBookmarked) {
        window.isBookmarked(url).then(setBookmarkState).catch(function(){});
      }
    });
  }

  function searchPDF() {
    if (!pdfDoc) return;
    var query = prompt('Find text in this PDF', '');
    if (!query || !query.trim()) return;
    var term = query.trim().toLowerCase();

    function checkPage(i) {
      if (i > pdfDoc.numPages) {
        alert('No matching text found in this PDF.');
        return;
      }
      pdfDoc.getPage(i).then(function(page) {
        return page.getTextContent();
      }).then(function(content) {
        var text = content.items.map(function(x){ return x.str || ''; }).join(' ');
        if (text.toLowerCase().indexOf(term) !== -1) {
          renderPage(i);
          alert('Match found on page ' + i + '.');
        } else {
          checkPage(i + 1);
        }
      }).catch(function(){ checkPage(i + 1); });
    }
    checkPage(1);
  }

  document.addEventListener('keydown', function(e) {
    if ((e.ctrlKey || e.metaKey) && e.key === '+') { e.preventDefault(); zoomIn(); }
    if ((e.ctrlKey || e.metaKey) && e.key === '-') { e.preventDefault(); zoomOut(); }
    if ((e.ctrlKey || e.metaKey) && e.key === '0') {
      e.preventDefault(); pageScale = 1.2; renderPage(currentPage); updateUI();
    }
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
      e.preventDefault(); searchPDF();
    }
    if (e.key === 'ArrowRight') nextPage();
    if (e.key === 'ArrowLeft') prevPage();
  });

  applyTheme();
  updateUI();
  loadPDF(getOriginalURL());
</script>
</body>
</html>
`

const bookmarksPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Bookmarks</title>
<style>
  * { box-sizing: border-box; }

  body {
    margin: 0;
    background: var(--shell-bg, #0b0b0d);
    color: var(--chrome-text, #e8e8e8);
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }

  h1 {
    font-size: 26px;
    margin-bottom: 6px;
    color: var(--chrome-text, #f2f2f2);
  }

  .page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .home-btn { flex: 0 0 auto; padding: 8px 14px; border: 1px solid var(--menu-border, #353535); border-radius: 8px; background: var(--chrome-panel, #111315); color: var(--chrome-text, #e6e6e6); cursor: pointer; font: 12px 'Segoe UI', sans-serif; }
  .home-btn:hover { background: var(--chrome-panel-2, #1a1d21); }

  .sub {
    color: var(--chrome-muted, #a0a0a0);
    font-size: 13px;
    margin-bottom: 26px;
  }

  .search {
    width: 320px;
    padding: 10px 16px;
    border-radius: 20px;
    border: 1px solid var(--menu-border, #2f2f2f);
    background: var(--chrome-panel, #111315);
    color: var(--chrome-text, #f5f5f5);
    outline: none;
    margin-bottom: 24px;
    font-size: 13px;
  }

  .search:focus {
    border-color: var(--chrome-accent, #4d90fe);
  }

  .bm-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--chrome-panel, #17181a);
    border: 1px solid var(--menu-border, #232323);
    border-radius: 12px;
    padding: 14px 18px;
    margin-bottom: 10px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.24);
    transition: 0.15s;
  }

  .bm-item:hover {
    transform: translateY(-1px);
    box-shadow: 0 5px 16px rgba(0,0,0,0.28);
    border-color: var(--chrome-accent, #2e2e30);
  }

  .bm-info {
    cursor: pointer;
  }

  .bm-title {
    font-size: 14px;
    color: var(--chrome-text, #f5f5f5);
    margin-bottom: 3px;
  }

  .bm-url {
    font-size: 11px;
    color: var(--chrome-muted, #a7a7a7);
  }

  .bm-remove {
    color: var(--chrome-text, #d9d9d9);
    cursor: pointer;
    font-size: 12px;
    padding: 6px 10px;
    border-radius: 6px;
  }

  .bm-remove:hover {
    color: #ff8a80;
    background: color-mix(in srgb, #ff8a80 14%, var(--chrome-panel, #2b1b1a));
  }

  .empty {
    color: var(--chrome-muted, #a3a3a3);
    font-size: 13px;
    margin-top: 40px;
    text-align: center;
  }

  .action-btn {
    background: var(--chrome-panel, #111315);
    color: var(--chrome-text, #e6e6e6);
    border: 1px solid var(--menu-border, #353535);
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 12px;
  }

  .action-btn:hover {
    background: var(--chrome-panel-2, #1a1d21);
  }
</style>
</head>
<body>

<header class="page-header">
  <h1>Bookmarks</h1>
  <button class="home-btn" type="button" onclick="window.realNavigate('sparrow://home')">← Home</button>
</header>

<div class="sub" id="countText">Loading...</div>

<input
  class="search"
  placeholder="Search bookmarks..."
  oninput="filterBookmarks(this.value)"
>

<div style="display:flex; gap:10px; margin-bottom:20px;">
  <button class="action-btn" onclick="exportBm()">Export Bookmarks</button>

  <button
    class="action-btn"
    onclick="document.getElementById('importFile').click()"
  >
    Import Bookmarks
  </button>

  <input
    type="file"
    id="importFile"
    accept=".json"
    style="display:none;"
    onchange="importBm(event)"
  >
</div>

<div id="bmList"></div>

<script>
  let allBookmarks = [];

  async function exportBm() {
    const json = await window.exportBookmarks();

    const blob = new Blob(
      [json],
      { type: 'application/json' }
    );

    const url = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;
    a.download = 'sparrow-bookmarks.json';
    a.click();

    URL.revokeObjectURL(url);
  }

  function importBm(event) {
    const file = event.target.files[0];

    if (!file) return;

    const reader = new FileReader();

    reader.onload = async function(e) {
      const count =
        await window.importBookmarks(e.target.result);

      alert(count + ' bookmark(s) imported');

      loadBookmarks();
    };

    reader.readAsText(file);
  }

  async function loadBookmarks() {
    allBookmarks =
      await window.getBookmarks() || [];

    document.getElementById('countText').innerText =
      allBookmarks.length + ' saved pages';

    renderList(allBookmarks);
  }

  function renderList(list) {
    const container =
      document.getElementById('bmList');

    if (list.length === 0) {
      container.innerHTML =
        '<div class="empty">No bookmarks yet. Click the star icon on any page to save it here.</div>';

      return;
    }

    container.innerHTML = list.map(function(b) {
      return '<div class="bm-item">' +
        '<div class="bm-info" onclick="window.realNavigate(\'' + b.URL + '\')">' +
          '<div class="bm-title">' + b.Title + '</div>' +
          '<div class="bm-url">' + b.URL + '</div>' +
        '</div>' +
        '<div class="bm-remove" onclick="removeBm(' + b.ID + ')">Remove</div>' +
      '</div>';
    }).join('');
  }

  async function filterBookmarks(keyword) {
    if (keyword.trim() === '') {
      renderList(allBookmarks);
      return;
    }

    const results =
      await window.searchBookmarks(keyword);

    renderList(results || []);
  }

  async function removeBm(id) {
    await window.removeBookmark(id);
    loadBookmarks();
  }

  loadBookmarks();
</script>

</body>
</html>
`

const favoritesPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Favorites</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0;
    background: var(--shell-bg, #0b0b0d);
    color: var(--chrome-text, #e8e8e8);
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; color: var(--chrome-text, #f2f2f2); }
  .page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .home-btn { flex: 0 0 auto; padding: 8px 14px; border: 1px solid var(--menu-border, #353535); border-radius: 8px; background: var(--chrome-panel, #111315); color: var(--chrome-text, #e6e6e6); cursor: pointer; font: 12px 'Segoe UI', sans-serif; }
  .home-btn:hover { background: var(--chrome-panel-2, #1a1d21); }
  .sub { color: var(--chrome-muted, #a0a0a0); font-size: 13px; margin-bottom: 26px; }
  .fav-item {
    display: flex; align-items: center; justify-content: space-between;
    background: var(--chrome-panel, #17181a); border: 1px solid var(--menu-border, #232323); border-radius: 12px;
    padding: 14px 18px; margin-bottom: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.24);
  }
  .fav-info { cursor: pointer; }
  .fav-title { font-size: 14px; color: var(--chrome-text, #f5f5f5); margin-bottom: 3px; }
  .fav-url { font-size: 11px; color: var(--chrome-muted, #a7a7a7); }
  .fav-remove { color: var(--chrome-text, #d9d9d9); cursor: pointer; font-size: 12px; padding: 6px 10px; border-radius: 6px; }
  .fav-remove:hover { color: #ff8a80; background: color-mix(in srgb, #ff8a80 14%, var(--chrome-panel, #2b1b1a)); }
  .empty { color: var(--chrome-muted, #a3a3a3); font-size: 13px; margin-top: 40px; text-align: center; }
</style>
</head>
<body>
<header class="page-header">
  <h1>Favorites</h1>
  <button class="home-btn" type="button" onclick="window.realNavigate('sparrow://home')">← Home</button>
</header>
<div class="sub" id="countText">Loading...</div>
<div id="favList"></div>
<script>
  let allFavorites = [];

  async function loadFavorites() {
    allFavorites = await window.getFavorites() || [];
    document.getElementById('countText').innerText = allFavorites.length + ' saved favorite sites';
    renderList(allFavorites);
  }

  function renderList(list) {
    const container = document.getElementById('favList');
    if (list.length === 0) {
      container.innerHTML = '<div class="empty">No favorites saved yet. Click the star icon in the address bar to save a website.</div>';
      return;
    }
    container.innerHTML = '';
    list.forEach(function(f) {
      var item = document.createElement('div');
      item.className = 'fav-item';
      var info = document.createElement('div');
      info.className = 'fav-info';
      info.onclick = function() { window.realNavigate(f.url); };
      var t = document.createElement('div');
      t.className = 'fav-title';
      t.textContent = f.title;
      var u = document.createElement('div');
      u.className = 'fav-url';
      u.textContent = f.url;
      info.appendChild(t);
      info.appendChild(u);
      var rm = document.createElement('div');
      rm.className = 'fav-remove';
      rm.textContent = 'Remove';
      rm.setAttribute('role', 'button');
      rm.tabIndex = 0;
      rm.onclick = function(event) {
        event.preventDefault();
        event.stopPropagation();
        removeFavorite(f.url);
      };
      rm.onkeydown = function(event) {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          removeFavorite(f.url);
        }
      };
      item.appendChild(info);
      item.appendChild(rm);
      container.appendChild(item);
    });
  }

  async function removeFavorite(url) {
    const countText = document.getElementById('countText');
    countText.innerText = 'Removing favorite...';
    try {
      const removed = await window.removeFavoriteByURL(url);
      if (!removed) {
        countText.innerText = 'Could not remove favorite. Please refresh and try again.';
        return;
      }
      allFavorites = await window.getFavorites() || [];
      try {
        localStorage.setItem('sparrow-favorites', JSON.stringify(allFavorites));
      } catch (storageError) {}
      countText.innerText = allFavorites.length + ' saved favorite sites';
      renderList(allFavorites);
    } catch (error) {
      countText.innerText = 'Could not remove favorite: ' + (error && error.message ? error.message : String(error));
    }
  }

  loadFavorites();
</script>
</body>
</html>
`

const downloadsPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Downloads</title>
<style>
  * { box-sizing: border-box; }

  body {
    margin: 0;
    background: var(--shell-bg, #0b0b0d);
    color: var(--chrome-text, #e8e8e8);
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }

  h1 {
    font-size: 26px;
    margin: 0 0 6px;
    color: var(--chrome-text, #f2f2f2);
  }

  .page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .home-btn { flex: 0 0 auto; padding: 8px 14px; border: 1px solid var(--menu-border, #353535); border-radius: 8px; background: var(--chrome-panel, #111315); color: var(--chrome-text, #e6e6e6); cursor: pointer; font: 12px 'Segoe UI', sans-serif; }
  .home-btn:hover { background: var(--chrome-panel-2, #1a1d21); }

  .sub {
    color: var(--chrome-muted, #a0a0a0);
    font-size: 13px;
    margin-bottom: 20px;
  }

  .btnbar {
    display: flex;
    gap: 10px;
    margin-bottom: 24px;
  }

  .btn {
    background: var(--chrome-panel, #111315);
    color: var(--chrome-text, #e6e6e6);
    border: 1px solid var(--menu-border, #353535);
    border-radius: 8px;
    padding: 8px 16px;
    font-size: 12px;
    cursor: pointer;
  }

  .btn:hover { background: var(--chrome-panel-2, #1a1d21); }

  .day-group { margin-bottom: 26px; }

  .day-title {
    color: var(--chrome-muted, #a7a7a7);
    font-size: 12px;
    font-weight: 600;
    margin: 0 0 10px 4px;
  }

  .dl-item {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 18px;
    background: var(--chrome-panel, #17181a);
    border: 1px solid var(--menu-border, #232323);
    border-radius: 12px;
    padding: 16px 20px;
    margin-bottom: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.24);
  }

  .dl-info { min-width: 0; flex: 1; }

  .dl-name {
    font-size: 14px;
    color: var(--chrome-text, #f4f4f4);
    margin-bottom: 6px;
    overflow-wrap: anywhere;
  }

  .dl-meta {
    font-size: 11px;
    color: var(--chrome-muted, #bcbcbc);
    margin-bottom: 5px;
  }

  .dl-url {
    font-size: 11px;
    color: var(--chrome-muted, #a7a7a7);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
    margin-bottom: 10px;
  }

  .bar-bg {
    width: 100%;
    height: 6px;
    background: var(--chrome-panel-2, #2b2d30);
    border-radius: 4px;
    overflow: hidden;
  }

  .bar-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--chrome-accent, #4d90fe), color-mix(in srgb, var(--chrome-accent, #4d90fe) 65%, white));
    border-radius: 4px;
    transition: width 0.3s;
  }

  .dl-status {
    flex: 0 0 auto;
    font-size: 11px;
    color: var(--chrome-muted, #bcbcbc);
    text-transform: capitalize;
    text-align: right;
  }

  .status-completed { color: #7fe3a0; }
  .status-failed { color: #ff8a80; }
  .status-cancelled { color: #b5b5b5; }

  .empty {
    color: var(--chrome-muted, #a3a3a3);
    font-size: 13px;
    margin-top: 40px;
    text-align: center;
  }

  @media (max-width: 700px) {
    body { padding: 28px 20px; }
    .dl-item { padding: 14px; gap: 10px; }
  }
</style>
</head>
<body>

<header class="page-header">
  <h1>Downloads</h1>
  <button class="home-btn" type="button" onclick="window.realNavigate('sparrow://home')">← Home</button>
</header>
<div class="sub">Stored only in RAM — cleared automatically when SPARROW closes</div>
<div class="btnbar">
  <button class="btn" type="button" onclick="clearFinished()">Clear Finished</button>
</div>

<div id="dlList"></div>

<script>
  function formatDate(date) {
    var today = new Date();
    var startOfToday = new Date(today.getFullYear(), today.getMonth(), today.getDate());
    var startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    var dayDifference = Math.round((startOfToday - startOfDate) / 86400000);
    if (dayDifference === 0) return 'Today';
    if (dayDifference === 1) return 'Yesterday';
    return date.toLocaleDateString();
  }

  function makeElement(tag, className, text) {
    var element = document.createElement(tag);
    if (className) element.className = className;
    if (text != null) element.textContent = text;
    return element;
  }

  async function loadDownloads() {
    var container = document.getElementById('dlList');
    try {
      var items = await window.getDownloads() || [];
      container.textContent = '';
      container.className = '';

      if (items.length === 0) {
        container.appendChild(makeElement('div', 'empty', 'No downloads yet.'));
        return;
      }

      var groups = new Map();
      items.slice().sort(function(a, b) { return b.ID - a.ID; }).forEach(function(download) {
        var startedAt = new Date(download.StartedAt);
        var validDate = !isNaN(startedAt.getTime());
        var groupTitle = validDate ? formatDate(startedAt) : 'Date unknown';
        var group = groups.get(groupTitle);
        if (!group) {
          group = makeElement('section', 'day-group');
          group.appendChild(makeElement('h2', 'day-title', groupTitle));
          groups.set(groupTitle, group);
          container.appendChild(group);
        }

        var item = makeElement('article', 'dl-item');
        var info = makeElement('div', 'dl-info');
        info.appendChild(makeElement('div', 'dl-name', download.FileName || 'Download'));
        var timeText = validDate ? 'Started ' + startedAt.toLocaleTimeString() : 'Start time unavailable';
        var finishedAt = new Date(download.FinishedAt);
        if (download.Status !== 'downloading' && !isNaN(finishedAt.getTime()) && finishedAt.getUTCFullYear() > 1) {
          timeText += ' · Finished ' + finishedAt.toLocaleTimeString();
        }
        info.appendChild(makeElement('div', 'dl-meta', timeText));
        info.appendChild(makeElement('div', 'dl-url', download.URL || 'Source URL unavailable'));
        var progress = Math.max(0, Math.min(100, Number(download.Progress) || 0));
        var bar = makeElement('div', 'bar-bg');
        var fill = makeElement('div', 'bar-fill');
        fill.style.width = progress + '%';
        bar.appendChild(fill);
        info.appendChild(bar);

        var status = String(download.Status || 'unknown');
        var statusElement = makeElement('div', 'dl-status status-' + status, status + ' · ' + progress + '%');
        item.appendChild(info);
        item.appendChild(statusElement);
        if (status === 'downloading') {
          var cancelBtn = makeElement('button', 'btn', 'Cancel');
          cancelBtn.type = 'button';
          cancelBtn.onclick = function () {
            Promise.resolve(window.cancelDownload(download.ID)).then(loadDownloads);
          };
          item.appendChild(cancelBtn);
        }
        group.appendChild(item);
      });
    } catch (error) {
      console.error('SPARROW could not load downloads:', error);
      container.textContent = 'Could not load downloads. Please refresh and try again.';
      container.className = 'empty';
    }
  }

  async function clearFinished() {
    try {
      await window.clearFinishedDownloads();
      await loadDownloads();
    } catch (error) {
      console.error('SPARROW could not clear finished downloads:', error);
    }
  }

  loadDownloads();

  setInterval(
    loadDownloads,
    1500
  );
</script>

</body>
</html>
`

const historyPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW History</title>
<style>
  * { box-sizing: border-box; }

  body {
    margin: 0;
    background: var(--shell-bg, #0b0b0d);
    color: var(--chrome-text, #e8e8e8);
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }

  h1 {
    font-size: 26px;
    margin-bottom: 6px;
    color: var(--chrome-text, #f2f2f2);
  }

  .page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .home-btn { flex: 0 0 auto; padding: 8px 14px; border: 1px solid var(--menu-border, #353535); border-radius: 8px; background: var(--chrome-panel, #111315); color: var(--chrome-text, #e6e6e6); cursor: pointer; font: 12px 'Segoe UI', sans-serif; }
  .home-btn:hover { background: var(--chrome-panel-2, #1a1d21); }

  .sub {
    color: var(--chrome-muted, #a0a0a0);
    font-size: 13px;
    margin-bottom: 24px;
  }

  .btnbar {
    display: flex;
    gap: 10px;
    margin-bottom: 24px;
  }

  .btn {
    background: var(--chrome-panel, #111315);
    color: var(--chrome-text, #e6e6e6);
    border: 1px solid var(--menu-border, #353535);
    border-radius: 8px;
    padding: 8px 16px;
    font-size: 12px;
    cursor: pointer;
  }

  .btn:hover {
    background: var(--chrome-panel-2, #1a1d21);
    color: var(--chrome-text, #ffffff);
  }

  .hist-item {
    display: flex;
    justify-content: space-between;
    padding: 12px 16px;
    background: var(--chrome-panel, #17181a);
    border: 1px solid var(--menu-border, #232323);
    border-radius: 10px;
    margin-bottom: 8px;
    cursor: pointer;
    font-size: 13px;
    transition: 0.15s;
  }

  .hist-item:hover {
    background: var(--chrome-panel-2, #1c1e21);
    box-shadow: 0 2px 8px rgba(0,0,0,0.18);
  }

  .hist-url {
    color: var(--chrome-text, #eeeeee);
  }

  .hist-time {
    color: var(--chrome-muted, #a7a7a7);
    font-size: 11px;
  }

  .empty {
    color: var(--chrome-muted, #a3a3a3);
    font-size: 13px;
    margin-top: 40px;
    text-align: center;
  }
</style>
</head>
<body>

<header class="page-header">
  <h1>History</h1>
  <button class="home-btn" type="button" onclick="window.realNavigate('sparrow://home')">← Home</button>
</header>

<div class="sub">
  Stored only in RAM — cleared automatically when SPARROW closes
</div>

<div class="btnbar">
  <button class="btn" onclick="clearScope('1hour')">
    Clear Last Hour
  </button>

  <button class="btn" onclick="clearScope('today')">
    Clear Today
  </button>

  <button class="btn" onclick="clearScope('all')">
    Clear All
  </button>
</div>

<div id="histList"></div>

<script>
  async function loadHistory() {
    const items =
      await window.getRecentHistory(200) || [];

    const container =
      document.getElementById('histList');

    if (items.length === 0) {
      container.innerHTML =
        '<div class="empty">No history yet.</div>';

      return;
    }

    container.innerHTML =
      items.slice().reverse().map(function(h) {

        return '<div class="hist-item" onclick="window.realNavigate(\'' +
          h.URL +
          '\')">' +

          '<span class="hist-url">' +
            h.URL +
          '</span>' +

          '<span class="hist-time">' +
            h.Time +
          '</span>' +

        '</div>';

      }).join('');
  }

  async function clearScope(scope) {
    await window.clearHistory(scope);
    loadHistory();
  }

  loadHistory();
</script>

</body>
</html>
`
