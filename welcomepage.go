package main

// ---------------------------------------------------
// FIRST-RUN WELCOME / ABOUT PAGE
// ---------------------------------------------------

const welcomePageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Welcome to SPARROW</title>
<style>
  * { box-sizing: border-box; }
  :root {
    color-scheme: dark;
    --page-bg: #0b0b0d; --panel: #17181a; --panel-2: #1c1d20;
    --border: #292a2e; --text: #eee; --muted: #a5a6ad; --accent: #4d90fe;
  }
  html[data-theme="light"] {
    color-scheme: light;
    --page-bg: #f5f7fa; --panel: #fff; --panel-2: #f1f3f5;
    --border: #dfe3e8; --text: #202124; --muted: #5f6368; --accent: #1967d2;
  }
  body {
    margin: 0; min-height: 100vh; background: var(--page-bg); color: var(--text);
    font-family: 'Segoe UI', sans-serif; padding: 32px 20px 48px;
  }
  main { width: min(100%, 900px); margin: 0 auto; }
  .hero {
    padding: 30px 34px; border: 1px solid var(--border); border-radius: 18px;
    background: linear-gradient(125deg, var(--panel), var(--panel-2));
    box-shadow: 0 10px 32px rgba(0,0,0,.16);
  }
  h1 { margin: 0 0 6px; font-size: 38px; letter-spacing: 5px; }
  .tagline { margin: 0; color: var(--muted); font-size: 15px; }
  .version {
    display: inline-flex; align-items: center; gap: 8px; margin-top: 18px; padding: 6px 10px;
    border: 1px solid var(--border); border-radius: 20px; color: var(--muted); font-size: 12px;
  }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--accent); }
  .intro { max-width: 780px; margin: 22px 0; color: var(--muted); font-size: 14px; line-height: 1.7; }
  h2 { margin: 24px 0 12px; font-size: 18px; }
  .features { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
  .feature-card {
    padding: 18px 20px; border: 1px solid var(--border); border-radius: 13px;
    background: var(--panel); box-shadow: 0 3px 12px rgba(0,0,0,.10);
  }
  .feature-title { margin-bottom: 6px; font-size: 14px; font-weight: 600; }
  .feature-desc { color: var(--muted); font-size: 12px; line-height: 1.65; }
  .footer { margin-top: 24px; color: var(--muted); font-size: 12px; line-height: 1.7; text-align: center; }
  .start-btn {
    margin-top: 22px; padding: 13px 30px; border: 0; border-radius: 30px;
    background: linear-gradient(90deg, #4d90fe, #6bb0ff); color: #000;
    font-size: 14px; font-weight: 600; cursor: pointer;
    box-shadow: 0 8px 24px rgba(77,144,254,.30);
  }
  .start-btn:hover { transform: translateY(-1px); }
  @media (max-width: 600px) {
    body { padding: 20px 14px 32px; }
    .hero { padding: 24px 20px; }
    h1 { font-size: 31px; }
    .features { grid-template-columns: 1fr; }
  }
</style>
</head>
<body>
<main>
  <header class="hero">
    <h1>SPARROW</h1>
    <p class="tagline">A focused browser that keeps your browsing yours.</p>
    <div class="version"><span class="dot"></span> Version ` + sparrowVersion + ` <span aria-hidden="true">&middot;</span> Built ` + sparrowBuildDate + `</div>
  </header>

  <p class="intro">
    Sparrow brings everyday browsing, useful built-in tools, and local-first controls together in one desktop app.
    No account is required. Your settings and saved items stay on this device, and you can tailor the browser with
    themes, privacy controls, and your preferred search provider.
  </p>

  <h2>Explore Sparrow</h2>
  <section class="features" aria-label="Sparrow features">
    <article class="feature-card">
      <div class="feature-title">Tabs and navigation</div>
      <div class="feature-desc">Browse in separate tabs, pin important pages, and use keyboard shortcuts to switch tabs or reopen a recently closed one.</div>
    </article>
    <article class="feature-card">
      <div class="feature-title">Search and Sparrow home</div>
      <div class="feature-desc">Search from the address bar with your selected provider. Sparrow home includes saved shortcuts, frequently visited sites, and appearance options.</div>
    </article>
    <article class="feature-card">
      <div class="feature-title">Privacy and protection</div>
      <div class="feature-desc">Use incognito tabs, inspect site information, manage site permissions, and configure protection against known advertising and tracking resources.</div>
    </article>
    <article class="feature-card">
      <div class="feature-title">Saved items and local controls</div>
      <div class="feature-desc">Manage bookmarks, favorites, passwords, and autofill locally. Saved passwords are encrypted before they are persisted.</div>
    </article>
    <article class="feature-card">
      <div class="feature-title">Tools for everyday browsing</div>
      <div class="feature-desc">Use Reader Mode, find text on a page, adjust zoom, capture screenshots, print, and open PDFs in the built-in viewer.</div>
    </article>
    <article class="feature-card">
      <div class="feature-title">Your settings, your device</div>
      <div class="feature-desc">Choose a theme, search provider, homepage, and other browsing options. Sparrow does not require an account or provide cross-device sync.</div>
    </article>
  </section>

  <div class="footer">
    Some features depend on your operating system, embedded WebView, or the website being visited.<br>
    Made by <strong>Kaif Alvi</strong>. Built as a learning project.
  </div>
  <button class="start-btn" id="startBrowsing" type="button">Start Browsing</button>
</main>
<script>
  (function () {
    var button = document.getElementById('startBrowsing');
    if (!button) return;
    button.addEventListener('click', async function () {
      if (typeof window.go === 'function') {
        await window.go('sparrow://home');
      }
    });
  })();
</script>
</body>
</html>
`
