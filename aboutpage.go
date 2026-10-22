package main

const sparrowVersion = "1.0.0"
const sparrowBuildDate = "2026-10-12"

const aboutPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>About SPARROW</title>
<style>
  * { box-sizing: border-box; }
  :root {
    color-scheme: dark;
    --about-bg: #0b0b0d;
    --about-panel: #17181a;
    --about-panel-2: #1c1d20;
    --about-border: #292a2e;
    --about-text: #eee;
    --about-muted: #a5a6ad;
    --about-subtle: #777a83;
    --about-accent: #4d90fe;
  }
  html[data-theme="light"] {
    color-scheme: light;
    --about-bg: #f5f7fa;
    --about-panel: #fff;
    --about-panel-2: #f1f3f5;
    --about-border: #dfe3e8;
    --about-text: #202124;
    --about-muted: #5f6368;
    --about-subtle: #73777f;
    --about-accent: #1967d2;
  }
  body {
    margin: 0; min-height: 100vh; background: var(--about-bg); color: var(--about-text);
    font-family: 'Segoe UI', sans-serif; padding: 38px 24px 56px;
  }
  .page { width: min(100%, 1020px); margin: 0 auto; }
  .hero {
    position: relative; overflow: hidden; padding: 34px 38px; margin-bottom: 24px;
    border: 1px solid var(--about-border); border-radius: 18px;
    background: linear-gradient(125deg, var(--about-panel), var(--about-panel-2));
    box-shadow: 0 10px 32px rgba(0,0,0,.16);
  }
  .hero:after {
    content: ''; position: absolute; width: 300px; height: 300px; right: -90px; top: -180px;
    border-radius: 50%; background: var(--about-accent); opacity: .10; filter: blur(2px);
  }
  .brand { font-size: 38px; font-weight: 750; letter-spacing: 5px; margin: 0 0 6px; }
  .tagline { margin: 0; color: var(--about-muted); font-size: 15px; }
  .version {
    display: inline-flex; gap: 8px; align-items: center; margin-top: 20px; padding: 6px 10px;
    border: 1px solid var(--about-border); border-radius: 20px; background: var(--about-bg);
    color: var(--about-muted); font-size: 12px;
  }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--about-accent); }
  .intro { color: var(--about-muted); font-size: 14px; line-height: 1.7; margin: 0 0 30px; max-width: 850px; }
  h2 { font-size: 19px; margin: 30px 0 13px; }
  .grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
  .card {
    padding: 20px 22px; border: 1px solid var(--about-border); border-radius: 13px;
    background: var(--about-panel); box-shadow: 0 3px 12px rgba(0,0,0,.10);
  }
  .card h3 { font-size: 14px; margin: 0 0 8px; }
  .card p { color: var(--about-muted); font-size: 12.5px; line-height: 1.65; margin: 0; }
  .card ul { padding-left: 18px; margin: 10px 0 0; color: var(--about-muted); font-size: 12.5px; line-height: 1.8; }
  .wide { grid-column: 1 / -1; }
  .key {
    display: inline-block; padding: 2px 7px; border: 1px solid var(--about-border);
    border-radius: 5px; background: var(--about-panel-2); color: var(--about-text);
    font: 11px Consolas, monospace; white-space: nowrap;
  }
  .note {
    margin-top: 18px; padding: 14px 17px; border-left: 3px solid var(--about-accent);
    border-radius: 7px; background: var(--about-panel-2); color: var(--about-muted);
    font-size: 12px; line-height: 1.65;
  }
  .footer { margin-top: 28px; color: var(--about-subtle); font-size: 12px; text-align: center; line-height: 1.7; }
  .footer strong { color: var(--about-muted); }
  @media (max-width: 650px) {
    body { padding: 24px 14px 40px; }
    .hero { padding: 27px 22px; }
    .brand { font-size: 31px; }
    .grid { grid-template-columns: 1fr; }
    .wide { grid-column: auto; }
  }
</style>
</head>
<body>
<main class="page">
  <header class="hero">
    <h1 class="brand">SPARROW</h1>
    <p class="tagline">A focused browser that keeps your browsing yours.</p>
    <div class="version"><span class="dot"></span> Version ` + sparrowVersion + ` <span aria-hidden="true">&middot;</span> Built ` + sparrowBuildDate + `</div>
  </header>

  <p class="intro">
    Sparrow brings everyday browsing, useful built-in tools, and local-first controls together in one desktop app.
    Use the tab strip and address bar to browse, then tailor the browser with settings, themes, and privacy controls.
    This page describes the features currently available in Sparrow.
  </p>

  <h2>Browse your way</h2>
  <section class="grid" aria-label="Browsing features">
    <article class="card">
      <h3>Tabs and sessions</h3>
      <p>Keep pages in separate tabs, pin important tabs to the start of the tab strip, and use keyboard shortcuts to switch between them. Reopen the most recently closed tab, or restore open tabs after an unexpected shutdown when crash recovery is enabled.</p>
      <ul>
        <li><span class="key">Ctrl + T</span> new tab; <span class="key">Ctrl + W</span> close current tab</li>
        <li><span class="key">Ctrl + Shift + T</span> reopen last closed tab</li>
        <li><span class="key">Ctrl + Shift + W</span> close all tabs and open Sparrow home</li>
        <li><span class="key">Ctrl + Tab</span> / <span class="key">Ctrl + Shift + Tab</span> switch tabs</li>
      </ul>
    </article>
    <article class="card">
      <h3>Search and navigation</h3>
      <p>Search from the address bar using your selected search provider, enter a web address, or pick suggestions from browsing history and bookmarks. Back, forward, reload, and home controls are available in the toolbar.</p>
      <ul>
        <li><span class="key">Ctrl + L</span> focus the address bar</li>
        <li><span class="key">Alt + Left</span> / <span class="key">Alt + Right</span> back and forward</li>
        <li><span class="key">Ctrl + R</span> reload the current page</li>
      </ul>
    </article>
    <article class="card">
      <h3>Bookmarks, favorites, and saved links</h3>
      <p>Save pages to bookmarks or favorites, and keep a separate collection of shortcuts on Sparrow home. Search and manage saved items from their built-in pages. Bookmarks, Favorites, Saved Links, History, and Downloads each provide a Home button to return to Sparrow home.</p>
      <ul>
        <li><span class="key">Ctrl + D</span> save the current page as a favorite</li>
        <li><span class="key">Ctrl + Shift + O</span> open bookmarks</li>
      </ul>
    </article>
    <article class="card">
      <h3>Sparrow home</h3>
      <p>Start with a search box, saved shortcuts, and frequently visited sites. Customize the home page with themes and available wallpaper choices.</p>
    </article>
  </section>

  <h2>Privacy and protection</h2>
  <section class="grid" aria-label="Privacy and protection features">
    <article class="card">
      <h3>Private browsing controls</h3>
      <p>Open an incognito tab for a private browsing session. Sparrow keeps browsing history in memory for the current run rather than saving it as persistent browser history; history can also be cleared from its page or with <span class="key">Ctrl + Shift + Delete</span>.</p>
    </article>
    <article class="card">
      <h3>Site information and protection</h3>
      <p>Use the site information control to inspect connection details and protection activity. Sparrow includes HTTPS enforcement and a configurable shield that can block known advertising and tracking resources.</p>
    </article>
    <article class="card">
      <h3>Cookies and site permissions</h3>
      <p>Review cookie activity and manage site-specific permissions from the built-in Cookies and Site Settings pages.</p>
    </article>
    <article class="card">
      <h3>Passwords and autofill</h3>
      <p>Manage saved login records and autofill details locally. Saved password records are encrypted before persistence; review the password and autofill pages to manage what is stored.</p>
    </article>
  </section>

  <h2>Tools for everyday work</h2>
  <section class="grid" aria-label="Browser tools">
    <article class="card">
      <h3>Reading and page controls</h3>
      <p>Use Reader Mode for a simplified reading view, find text on a page, adjust page zoom, and print when needed.</p>
      <ul>
        <li><span class="key">Ctrl + F</span> find in page</li>
        <li><span class="key">Ctrl + Plus</span> / <span class="key">Ctrl + Minus</span> zoom</li>
        <li><span class="key">Ctrl + 0</span> reset zoom</li>
        <li><span class="key">Ctrl + Shift + R</span> toggle Reader Mode</li>
      </ul>
    </article>
    <article class="card">
      <h3>Downloads and documents</h3>
      <p>Sparrow's download manager queues transfers, runs up to three at once, and lets you pause, resume, cancel, and monitor progress. The Downloads page lists recent transfers and can clear finished items. PDF documents can be opened in Sparrow's built-in PDF viewer, including supported PDF files opened from the computer.</p>
    </article>
    <article class="card">
      <h3>Page and window utilities</h3>
      <p>Capture screenshots, use picture-in-picture where supported by the page, and switch the browser window to fullscreen with <span class="key">F11</span>.</p>
    </article>
    <article class="card">
      <h3>Backup and restore</h3>
      <p>Export or restore supported browser data through Backup &amp; Restore. Passwords are excluded from the general backup and have a separate vault export.</p>
    </article>
  </section>

  <h2>Personalize and manage Sparrow</h2>
  <section class="grid" aria-label="Settings and customization">
    <article class="card">
      <h3>Settings and appearance</h3>
      <p>Choose a light or dark theme, select a search provider and homepage, and configure available browsing options such as JavaScript, images, and download location. Customization pages provide additional appearance choices.</p>
      <p style="margin-top:9px"><span class="key">Ctrl + ,</span> open settings</p>
    </article>
    <article class="card">
      <h3>More built-in pages</h3>
      <p>Explore History, Downloads, Bookmarks, Favorites, Saved Links, Shortcuts, Site Settings, Cookies, Passwords, Autofill, Backup &amp; Restore, Updates, and About from the browser menu. Use the Home button on the History, Downloads, Bookmarks, Favorites, and Saved Links pages to return to Sparrow home.</p>
    </article>
    <article class="card wide">
      <h3>Local-first by design</h3>
      <p>Sparrow does not require an account and does not provide cross-device sync. Settings, bookmarks, favorites, saved links, and supported saved data are stored on this device. The embedded page engine is provided by the operating system: Chromium/WebView2 on Windows and the platform's native WebView on other supported systems.</p>
    </article>
  </section>

  <aside class="note">
    Some features depend on the operating system, the embedded WebView, or the website being visited. Website behavior and compatibility are controlled in part by those external components.
  </aside>
  <footer class="footer">Made by <strong>Kaif Alvi</strong>. Built as a learning project.<br>Thank you for using Sparrow.</footer>
</main>
</body>
</html>
`
