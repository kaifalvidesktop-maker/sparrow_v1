package main

// sparrow://home page (kept out of main.go).

const homePageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW</title>
<style>
  * { box-sizing: border-box; }

  :root {
		--home-bg: var(--shell-bg, #0b0b0d);
		--home-bg-2: color-mix(in srgb, var(--shell-bg, #0b0b0d) 70%, #121214);
    --home-panel: rgba(21, 22, 27, 0.84);
    --home-panel-border: rgba(151, 154, 255, 0.28);
    --home-text: #e8e8e8;
    --home-muted: #a6a6d8;
    --home-btn-bg: rgba(25, 26, 35, 0.9);
    --home-btn-text: #eae9ff;
    --home-btn-hover: rgba(120, 96, 255, 0.16);
    --home-shadow: rgba(85, 96, 255, 0.32);
    --home-glow-1: #7c3aed;
    --home-glow-2: #60a5fa;
  }

  html[data-theme="light"] {
		--home-bg: var(--shell-bg, #f3f6fb);
		--home-bg-2: color-mix(in srgb, var(--shell-bg, #f3f6fb) 70%, #eaf0f7);
    --home-panel: rgba(255, 255, 255, 0.9);
    --home-panel-border: rgba(124, 58, 237, 0.18);
    --home-text: #1f2328;
    --home-muted: #5d6675;
    --home-btn-bg: rgba(255, 255, 255, 0.9);
    --home-btn-text: #2a2f36;
    --home-btn-hover: rgba(96, 165, 250, 0.08);
    --home-shadow: rgba(96, 165, 250, 0.12);
    --home-glow-1: #6d52ff;
    --home-glow-2: #3b82f6;
  }

  body {
    margin: 0;
    min-height: 100vh;
	background: radial-gradient(circle at top, color-mix(in srgb, var(--chrome-accent, #4d90fe) 18%, transparent), transparent 30%),
                var(--home-wallpaper, linear-gradient(180deg, var(--home-bg) 0%, var(--home-bg-2) 100%));
    color: var(--home-text);
    font-family: 'Segoe UI', sans-serif;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0 0 70px;
    box-sizing: border-box;
    position: relative;
    overflow-x: hidden;
    transition: background 0.25s ease, color 0.25s ease;
  }

  .starfield {
    position: fixed;
    inset: 0;
    overflow: hidden;
    pointer-events: none;
    opacity: 0.9;
  }

  .star {
    position: absolute;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: rgba(255,255,255,0.9);
    box-shadow: 0 0 10px rgba(255,255,255,0.8);
    animation: drift var(--speed, 12s) linear infinite;
    opacity: 0.7;
  }

  #home-animation-layer {
    position: fixed;
    z-index: 0;
    inset: 116px 0 0;
    overflow: hidden;
    pointer-events: none;
    opacity: 0;
    transition: opacity .4s ease;
    background: radial-gradient(ellipse at 50% 110%, rgba(77,144,254,.13), transparent 65%);
  }

  body[data-home-animation]:not([data-home-animation="default"]) #home-animation-layer { opacity: 1; }
  body[data-home-animation]:not([data-home-animation="default"]) .starfield { opacity: 0; }
  .home-particle {
    position: absolute;
    left: var(--x);
    top: var(--y);
    color: var(--particle-color, rgba(145,190,255,.7));
    font-size: var(--size, 18px);
    opacity: var(--opacity, .65);
    text-shadow: 0 0 12px currentColor;
    animation: var(--particle-motion, home-float) var(--duration, 12s) var(--delay, 0s) ease-in-out infinite;
    will-change: transform, opacity;
  }
  body[data-home-animation="bubbles"] .home-particle, body[data-home-animation="ripples"] .home-particle { width: var(--size); height: var(--size); }
  body[data-home-animation="bubbles"] .home-particle { border: 1px solid rgba(120,190,255,.5); border-radius: 50%; background: radial-gradient(circle at 35% 30%, rgba(255,255,255,.45), rgba(96,165,250,.05) 65%); }
  body[data-home-animation="rain"] .home-particle,
  body[data-home-animation="matrix"] .home-particle { width: 2px; height: var(--size); background: linear-gradient(transparent, currentColor); text-shadow: none; }
  body[data-home-animation="rainbow"] #home-animation-layer { background: linear-gradient(120deg, rgba(255,70,100,.10), rgba(255,210,70,.11), rgba(70,220,150,.10), rgba(70,140,255,.13), rgba(185,90,255,.10)); background-size: 300% 300%; animation: home-rainbow 18s ease infinite; }
  body[data-home-animation="aurora"] #home-animation-layer { background: radial-gradient(ellipse at 20% 15%, rgba(25,220,160,.24), transparent 45%), radial-gradient(ellipse at 75% 10%, rgba(105,95,255,.24), transparent 48%), radial-gradient(ellipse at 50% 30%, rgba(50,200,220,.13), transparent 55%); background-size: 180% 180%; animation: home-aurora 16s ease-in-out infinite alternate; }
  body[data-home-animation="waves"] #home-animation-layer { background: repeating-radial-gradient(ellipse at 50% 115%, transparent 0 28px, rgba(70,170,255,.13) 30px 32px, transparent 34px 58px); background-size: 140% 140%; animation: home-wave 10s ease-in-out infinite; }
  body[data-home-animation="neon"] #home-animation-layer { background: radial-gradient(ellipse at 20% 20%, rgba(255,35,190,.14), transparent 45%), radial-gradient(ellipse at 80% 70%, rgba(45,210,255,.15), transparent 48%); animation: home-neon 6s ease-in-out infinite alternate; }
  body[data-home-animation="bubbles"] .home-particle { animation-name: home-bubble; }
  body[data-home-animation="rain"] .home-particle { animation-name: home-rain; }
  body[data-home-animation="rainbow"] .home-particle { animation-name: home-drift; }
  body[data-home-animation="snowfall"] .home-particle { animation-name: home-snow; }
  body[data-home-animation="fireflies"] .home-particle { animation-name: home-twinkle; }
  body[data-home-animation="aurora"] .home-particle { animation-name: home-drift; }
  body[data-home-animation="shooting-stars"] .home-particle,
  body[data-home-animation="meteors"] .home-particle { animation-name: home-meteor; }
  body[data-home-animation="confetti"] .home-particle { animation-name: home-confetti; }
  body[data-home-animation="waves"] .home-particle { animation-name: home-drift; }
  body[data-home-animation="petals"] .home-particle { animation-name: home-petal; }
  body[data-home-animation="matrix"] .home-particle { color: #6dff9b; animation-name: home-rain; }
  body[data-home-animation="sparkles"] .home-particle { animation-name: home-twinkle; }
  body[data-home-animation="clouds"] .home-particle { animation-name: home-cloud; }
  body[data-home-animation="hearts"] .home-particle { color: #ff91b8; animation-name: home-heart; }
  body[data-home-animation="galaxy"] .home-particle { animation-name: home-orbit; }
  body[data-home-animation="ripples"] .home-particle { border: 1px solid rgba(95,195,255,.45); border-radius: 50%; animation-name: home-ripple; }
  body[data-home-animation="leaves"] .home-particle { color: #b7df78; animation-name: home-leaf; }
  body[data-home-animation="embers"] .home-particle { color: #ff9a51; animation-name: home-ember; }
  body[data-home-animation="neon"] .home-particle { animation-name: home-twinkle; }
  @keyframes home-float { 0%,100% { transform: translate3d(0,0,0); } 50% { transform: translate3d(25px,-28px,0); } }
  @keyframes home-bubble { from { transform: translateY(18vh) scale(.65); opacity: 0; } 15% { opacity: .8; } to { transform: translateY(-110vh) scale(1.2); opacity: 0; } }
  @keyframes home-rain { from { transform: translateY(-15vh); } to { transform: translateY(110vh); } }
  @keyframes home-rainbow { 0% { background-position: 0 50%; } 100% { background-position: 100% 50%; } }
  @keyframes home-snow { from { transform: translate3d(0,-8vh,0) rotate(0); } to { transform: translate3d(55px,105vh,0) rotate(300deg); } }
  @keyframes home-twinkle { 0%,100% { opacity: .15; transform: scale(.65); } 50% { opacity: 1; transform: scale(1.25); } }
  @keyframes home-drift { 0%,100% { transform: translate3d(-12px,8px,0); } 50% { transform: translate3d(22px,-24px,0); } }
  @keyframes home-meteor { 0%,70% { transform: translate3d(0,0,0); opacity: 0; } 74% { opacity: .9; } 100% { transform: translate3d(-180px,130px,0); opacity: 0; } }
  @keyframes home-confetti { from { transform: translateY(-12vh) rotate(0); } to { transform: translateY(110vh) rotate(540deg); } }
  @keyframes home-petal { from { transform: translate3d(0,-10vh,0) rotate(0); } to { transform: translate3d(70px,108vh,0) rotate(420deg); } }
  @keyframes home-aurora { from { background-position: 0 0; filter: hue-rotate(0); } to { background-position: 100% 100%; filter: hue-rotate(35deg); } }
  @keyframes home-wave { 0%,100% { background-position: 0 0; } 50% { background-position: 70px 20px; } }
  @keyframes home-cloud { from { transform: translateX(-20vw) scale(1); } to { transform: translateX(120vw) scale(1.15); } }
  @keyframes home-heart { from { transform: translateY(15vh) scale(.7) rotate(-12deg); opacity: 0; } 20% { opacity: .8; } to { transform: translateY(-100vh) scale(1.2) rotate(12deg); opacity: 0; } }
  @keyframes home-orbit { 0% { transform: rotate(0) translateX(25px) rotate(0); } 100% { transform: rotate(360deg) translateX(25px) rotate(-360deg); } }
  @keyframes home-ripple { from { transform: scale(.1); opacity: .8; } to { transform: scale(5); opacity: 0; } }
  @keyframes home-leaf { from { transform: translate3d(0,-10vh,0) rotate(0); } to { transform: translate3d(90px,110vh,0) rotate(620deg); } }
  @keyframes home-ember { from { transform: translateY(15vh) scale(.4); opacity: 0; } 20% { opacity: .9; } to { transform: translateY(-85vh) scale(1.2); opacity: 0; } }
  @keyframes home-neon { from { filter: hue-rotate(0) blur(0); opacity: .55; } to { filter: hue-rotate(70deg) blur(8px); opacity: 1; } }

  @keyframes drift {
    0% { transform: translate3d(0, 0, 0) scale(0.6); opacity: 0; }
    12% { opacity: 1; }
    50% { transform: translate3d(20px, -36px, 0) scale(1); }
    100% { transform: translate3d(-18px, -80px, 0) scale(0.7); opacity: 0; }
  }

  .logo-wrap {
    position: relative;
    z-index: 1;
    text-align: center;
    margin-bottom: 8px;
  }

  .logo {
    font-size: 54px;
    font-weight: 800;
    letter-spacing: 5px;
    margin-bottom: 2px;
    background: linear-gradient(90deg, var(--home-glow-1), var(--home-glow-2), var(--home-glow-1));
    background-size: 200% 100%;
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
    animation: glowPulse 4s ease-in-out infinite;
    text-shadow: 0 0 26px rgba(124,58,237,0.28);
  }

  @keyframes wavify {
    0%, 100% { background-position: 0% 50%; transform: translateY(0px); }
    25% { transform: translateY(-2px); }
    50% { background-position: 100% 50%; transform: translateY(2px); }
    75% { transform: translateY(-1px); }
  }

  @keyframes glowPulse {
    0%, 100% { filter: drop-shadow(0 0 0 rgba(124,58,237,0.2)); }
    50% { filter: drop-shadow(0 0 18px rgba(96,165,250,0.45)); }
  }

  .tagline {
    color: var(--home-muted);
    margin-bottom: 34px;
    font-size: 13px;
    letter-spacing: 1.4px;
    text-transform: uppercase;
    opacity: 0.8;
  }

  .search-bar {
    flex: 1;
    min-width: 0;
    padding: 12px 4px 12px 20px;
    border: 0;
    background: transparent;
    color: var(--home-text);
    font-size: 15px;
    outline: none;
  }

  .search-form {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    width: 560px;
    max-width: 82%;
    padding: 4px 6px;
    border: 1px solid var(--home-panel-border);
    border-radius: 30px;
    background: var(--home-panel);
    box-shadow: 0 6px 30px var(--home-shadow);
    backdrop-filter: blur(6px);
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .search-form:focus-within {
    border-color: var(--home-glow-2);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--home-glow-2) 22%, transparent), 0 6px 30px var(--home-shadow);
  }

  .search-submit {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    flex: 0 0 auto;
    border: 0;
    border-radius: 50%;
    background: linear-gradient(100deg, var(--home-glow-1), var(--home-glow-2), var(--home-glow-1));
    background-size: 200% 100%;
    color: #fff;
    cursor: pointer;
    box-shadow: 0 3px 12px var(--home-shadow);
    transition: background-position 0.3s ease, transform 0.15s ease, box-shadow 0.15s ease;
  }

  .search-submit:hover {
    background-position: 100% 0;
    transform: translateY(-1px);
    box-shadow: 0 5px 18px var(--home-shadow);
  }

  .search-submit:focus-visible { outline: 2px solid var(--home-glow-2); outline-offset: 3px; }
  .search-submit svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; }


  .sites-row {
    position: relative;
    z-index: 1;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 22px;
    width: min(760px, 90%);
    margin-top: 42px;
    flex-shrink: 0;
  }
  .saved-links-section, .saved-links, .top-sites, .top-site-item { display: contents; }

  .saved-link-form {
    display: grid;
    grid-template-columns: 1fr;
    gap: 10px;
    width: min(380px, 92vw);
    margin: 0;
    padding: 20px;
    border: 1px solid var(--home-panel-border);
    border-radius: 14px;
    background: var(--home-panel);
    box-shadow: 0 8px 18px rgba(0,0,0,0.12);
    backdrop-filter: blur(4px);
  }

  .saved-link-form input,
  .saved-link-form button {
    min-width: 0;
    padding: 10px 12px;
    border: 1px solid var(--home-panel-border);
    border-radius: 9px;
    background: var(--home-panel);
    color: var(--home-text);
    font: inherit;
  }

  .saved-link-form input { outline: none; }
  .saved-link-form input:focus { border-color: rgba(96,165,250,0.8); }
  .saved-link-form button {
    background: rgba(96,165,250,0.2);
    cursor: pointer;
    white-space: nowrap;
  }
  .saved-link-form button:hover { background: rgba(96,165,250,0.35); }
  .saved-link-form button:disabled { cursor: wait; opacity: 0.65; }
  .saved-link-form h3 { margin: 0 0 4px; font-size: 16px; color: var(--home-text); }
  .saved-link-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 4px; }
  .saved-link-actions .cancel { background: transparent; }
  .saved-link-overlay {
    position: fixed; inset: 0; z-index: 50; display: none; align-items: center; justify-content: center;
    background: rgba(0,0,0,0.5); backdrop-filter: blur(4px);
  }
  .saved-link-overlay.open { display: flex; animation: slFade .15s ease; }
  @keyframes slFade { from { opacity: 0; } to { opacity: 1; } }
  .saved-link-add {
    flex: 0 0 82px; width: 82px; height: 82px; border-radius: 18px; cursor: pointer; font-size: 28px; line-height: 1;
    border: 1px dashed var(--home-panel-border); background: var(--home-panel); color: var(--home-muted);
  }
  .saved-link-add:hover { background: var(--home-btn-hover); color: var(--home-text); }
  .add-saved-link {
    display: block; margin: 8px auto 0; padding: 8px 16px; border-radius: 20px; cursor: pointer;
    border: 1px dashed var(--home-panel-border); background: transparent; color: var(--home-muted); font-size: 12px;
  }
  .add-saved-link:hover { background: var(--home-btn-hover); color: var(--home-text); }
  .home-toast {
    position: fixed; left: 50%; bottom: 74px; z-index: 60; transform: translate(-50%, 12px); opacity: 0;
    pointer-events: none; padding: 10px 16px; border-radius: 10px; font-size: 12px;
    background: var(--home-panel); color: var(--home-text); border: 1px solid rgba(96,165,250,0.65);
    box-shadow: 0 8px 24px rgba(0,0,0,0.3); transition: opacity .2s ease, transform .2s ease;
  }
  .home-toast.show { opacity: 1; transform: translate(-50%, 0); }
  .saved-link-error {
    grid-column: 1 / -1;
    color: #ff8a80;
    font-size: 12px;
  }


  .saved-link-card {
    position: relative;
    width: 82px;
    height: 82px;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 10px;
    border: 1px solid var(--home-panel-border);
    border-radius: 18px;
    background: var(--home-panel);
    box-shadow: 0 8px 18px rgba(0,0,0,0.12);
    backdrop-filter: blur(4px);
  }

  .saved-link-title {
    width: 100%;
    overflow: hidden;
    color: var(--home-text);
    cursor: pointer;
    font-size: 12px;
    font-weight: 600;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .saved-link-more {
    position: absolute;
    top: 3px;
    right: 3px;
    width: 24px;
    height: 24px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--home-muted);
    cursor: pointer;
    font-size: 17px;
  }

  .saved-link-menu {
    position: absolute;
    z-index: 4;
    top: 27px;
    right: 3px;
    display: none;
    min-width: 90px;
    padding: 4px;
    border: 1px solid var(--home-panel-border);
    border-radius: 8px;
    background: var(--home-panel);
    box-shadow: 0 8px 20px rgba(0,0,0,0.3);
  }

  .saved-link-menu.open { display: grid; }
  .saved-link-menu button {
    padding: 7px 8px;
    border: 0;
    border-radius: 5px;
    background: transparent;
    color: var(--home-text);
    cursor: pointer;
    font-size: 11px;
    text-align: left;
  }
  .saved-link-menu button:hover { background: var(--home-btn-hover); }

  .site {
    width: 82px;
    height: 82px;
    background: var(--home-panel);
    border: 1px solid var(--home-panel-border);
    border-radius: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    color: var(--home-muted);
    cursor: pointer;
    transition: 0.2s ease;
    box-shadow: 0 8px 18px rgba(0,0,0,0.12);
    backdrop-filter: blur(4px);
  }

  .site:hover {
    background: var(--home-btn-hover);
    color: var(--home-text);
    border-color: rgba(96,165,250,0.65);
  }


  .site-save-link {
    width: 82px;
    height: 82px;
    flex: 0 0 82px;
    padding: 0;
    border: 1px solid var(--home-panel-border);
    border-radius: 18px;
    background: var(--home-panel);
    color: var(--home-text);
    cursor: pointer;
    font-size: 28px;
    line-height: 1;
  }

  .site-save-link:hover {
    background: var(--home-btn-hover);
    border-color: rgba(96,165,250,0.65);
  }

  .site-save-link.saved {
    color: #3ddc84;
    cursor: default;
  }

  .history-controls {
    position: fixed;
    bottom: 22px;
    display: flex;
    gap: 12px;
    z-index: 2;
  }

  .history-controls button {
    background: var(--home-btn-bg);
    color: var(--home-btn-text);
    border: 1px solid var(--home-panel-border);
    padding: 10px 16px;
    border-radius: 10px;
    cursor: pointer;
    font-size: 12px;
    letter-spacing: 0.02em;
    box-shadow: 0 6px 20px rgba(0,0,0,0.1);
    transition: transform 0.15s ease, background 0.2s ease, box-shadow 0.2s ease;
  }

  .history-controls button:nth-child(2) { animation-delay: 0.7s; }
  .history-controls button:nth-child(3) { animation-delay: 1.4s; }

  @keyframes pulseFloat {
    0%, 100% { transform: translateY(0); box-shadow: 0 6px 18px rgba(0,0,0,0.1); }
    50% { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(124,58,237,0.14); }
  }

  .history-controls button:hover {
    background: var(--home-btn-hover);
    color: var(--home-text);
  }

  @media (max-width: 620px) {
    .saved-link-form { grid-template-columns: 1fr; }
    .saved-link-error { grid-column: 1; }
  }
</style>
</head>
<body>
  <div id="home-animation-layer" aria-hidden="true"></div>
  <div class="starfield" id="starfield"></div>

  <div class="logo-wrap">
    <div class="logo">SPARROW</div>
    <div class="tagline">Fast. Private. Yours.</div>
  </div>

  <form class="search-form" role="search" onsubmit="handleSearch(event)">
    <input class="search-bar" id="searchBar" type="search" aria-label="Search or enter a web address"
           placeholder="Search or type a URL" autocomplete="off">
    <button class="search-submit" type="submit" aria-label="Search">
      <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="11" cy="11" r="7"></circle><path d="m20 20-4-4"></path></svg>
    </button>
  </form>

  <div class="sites-row">
    <div class="saved-links-section"><div class="saved-links" id="savedLinksContainer"></div></div>
    <div class="top-sites" id="topSitesContainer"></div>
    <button class="saved-link-add" id="addSavedLinkBtn" type="button" title="Add website" aria-label="Add website">+</button>
  </div>

  <div class="saved-link-overlay" id="savedLinkOverlay">
    <form class="saved-link-form" id="savedLinkForm" autocomplete="off" novalidate>
      <h3 id="savedLinkHeading">Save website</h3>
      <input id="savedLinkURL" type="text" inputmode="url" placeholder="Paste website link" aria-label="Website link" required>
      <input id="savedLinkTitle" type="text" maxlength="120" placeholder="Title" aria-label="Website title" required>
      <div class="saved-link-error" id="savedLinkError" role="status" aria-live="polite"></div>
      <div class="saved-link-actions">
        <button id="cancelSavedLink" class="cancel" type="button">Cancel</button>
        <button id="saveWebsiteButton" type="submit">Save</button>
      </div>
    </form>
  </div>
  <div class="home-toast" id="homeToast" role="status"></div>

  <div class="history-controls">
    <button onclick="go('sparrow://history')">History</button>
    <button onclick="go('sparrow://bookmarks')">Bookmarks</button>
    <button onclick="go('sparrow://settings')">Settings</button>
  </div>

<script>
  function applyHomeTheme(theme) {
    const t = theme === 'light' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', t);
    try { localStorage.setItem('sparrow-home-theme', t); } catch (e) {}
  }

  async function applyHomeAnimation(animation, save) {
    const layer = document.getElementById('home-animation-layer');
    const validIds = ['default','bubbles','rain','rainbow','snowfall','fireflies','aurora','shooting-stars','confetti','waves','petals','matrix','sparkles','clouds','meteors','hearts','galaxy','ripples','leaves','embers','neon'];
    const selected = validIds.indexOf(animation) >= 0 ? animation : 'default';
    document.body.setAttribute('data-home-animation', selected);
    if (!layer) return;
    layer.replaceChildren();
    if (selected === 'default') return;
    const shapes = {
      bubbles: '', rain: '', rainbow: '\u2726', snowfall: '\u2744', fireflies: '\u2022',
      aurora: '\u2726', 'shooting-stars': '\u2726', confetti: '\u25a0', waves: '\u2022',
      petals: '\u273f', matrix: '01', sparkles: '\u2726', clouds: '\u25cf',
      meteors: '\u2726', hearts: '\u2665', galaxy: '\u22c6', ripples: '',
      leaves: '\u274b', embers: '\u2022', neon: '\u25cf'
    };
    const count = selected === 'rain' || selected === 'matrix' ? 70 : selected === 'ripples' ? 18 : 34;
    for (let i = 0; i < count; i++) {
      const particle = document.createElement('span');
      particle.className = 'home-particle';
      particle.textContent = shapes[selected] || '';
      particle.style.setProperty('--x', (Math.random() * 100).toFixed(2) + '%');
      particle.style.setProperty('--y', (Math.random() * 100).toFixed(2) + '%');
      particle.style.setProperty('--size', (selected === 'bubbles' || selected === 'ripples' ? 8 + Math.random() * 30 : 7 + Math.random() * 15).toFixed(1) + 'px');
      particle.style.setProperty('--duration', (7 + Math.random() * 15).toFixed(2) + 's');
      particle.style.setProperty('--delay', (-Math.random() * 20).toFixed(2) + 's');
      particle.style.setProperty('--opacity', (0.25 + Math.random() * 0.65).toFixed(2));
      if (selected === 'rainbow') {
        const hues = [0, 45, 120, 190, 240, 290];
        particle.style.setProperty('--particle-color', 'hsla(' + hues[i % hues.length] + ', 90%, 68%, .8)');
      } else if (selected === 'confetti') {
        particle.style.setProperty('--particle-color', 'hsl(' + Math.floor(Math.random() * 360) + ', 90%, 68%)');
      }
      layer.appendChild(particle);
    }
    if (save && window.setHomeAnimation) {
      try {
        const persisted = await window.setHomeAnimation(selected);
        if (persisted !== selected) throw new Error('The selected home animation was not saved.');
      } catch (error) {
        console.error('Could not save home animation:', error);
      }
    }
  }

  function createStarField(preset) {
    const field = document.getElementById('starfield');
    if (!field) return;
    while (field.firstChild) field.removeChild(field.firstChild);

    const patterns = [
      { count: 18, spread: 1.0, speed: 8 },
      { count: 24, spread: 1.15, speed: 10 },
      { count: 16, spread: 1.4, speed: 12 },
      { count: 30, spread: 1.1, speed: 11 },
      { count: 20, spread: 1.25, speed: 9 },
      { count: 26, spread: 1.2, speed: 13 },
      { count: 22, spread: 1.35, speed: 10 },
      { count: 18, spread: 1.5, speed: 14 },
      { count: 28, spread: 1.05, speed: 7 },
      { count: 32, spread: 1.25, speed: 16 }
    ];

    const chosen = patterns[Math.min(Math.max(preset || 0, 0), patterns.length - 1)];
    for (let i = 0; i < chosen.count; i++) {
      const star = document.createElement('span');
      star.className = 'star';
      const left = (Math.random() * 100) + '%';
      const top = (Math.random() * 76) + '%';
      const size = (Math.random() * 4 + 1).toFixed(1) + 'px';
      const delay = (Math.random() * 8).toFixed(2) + 's';
      const speed = (chosen.speed + Math.random() * 6).toFixed(2) + 's';
      star.style.left = left;
      star.style.top = top;
      star.style.width = size;
      star.style.height = size;
      star.style.animationDelay = delay;
      star.style.setProperty('--speed', speed);
      star.style.opacity = (0.3 + Math.random() * 0.7).toFixed(2);
      field.appendChild(star);
    }
  }

  function handleSearch(event) {
    if (event) event.preventDefault();
    const input = document.getElementById('searchBar').value;
    if (input.trim() === '') return;
    go(input);
  }

  function normalizeSiteURL(raw) {
    const value = (raw || '').trim();
    if (!value) return '';
    return /^https?:\/\//i.test(value) ? value : 'https://' + value;
  }

  function domainLabel(url) {
    try {
      const parsed = new URL(url);
      return parsed.hostname.replace(/^www\./, '');
    } catch (e) {
      return 'Website';
    }
  }

  async function loadTopSites() {
    const container = document.getElementById('topSitesContainer');
    if (!container) return;

    container.replaceChildren();
    if (window.getTopSites) {
      const sites = await window.getTopSites(4);
      sites.forEach(function (site) {
        const item = document.createElement('div');
        item.className = 'top-site-item';

        const tile = document.createElement('div');
        tile.className = 'site';
        tile.title = site.URL;
        tile.textContent = site.Domain;
        tile.addEventListener('click', function () { go(site.URL); });
        item.appendChild(tile);

        container.appendChild(item);
      });
    }
  }

  async function loadSavedLinks() {
    const container = document.getElementById('savedLinksContainer');
    if (!container || !window.getShortcutsNewTab) return;
    const links = await window.getShortcutsNewTab();
    container.replaceChildren();
    const section = container.closest('.saved-links-section');
    if (section) section.style.display = '';
    links.forEach(function (link) {
      const card = document.createElement('div');
      card.className = 'saved-link-card';
      const title = document.createElement('span');
      title.className = 'saved-link-title';
      title.textContent = link.Title;
      title.title = link.Title;
      title.addEventListener('click', function () { go(link.URL); });

      const more = document.createElement('button');
      more.className = 'saved-link-more';
      more.type = 'button';
      more.textContent = '⋮';
      more.setAttribute('aria-label', 'Link options');
      const menu = document.createElement('div');
      menu.className = 'saved-link-menu';
      const edit = document.createElement('button');
      edit.type = 'button';
      edit.textContent = 'Edit';
      edit.addEventListener('click', function () {
        menu.classList.remove('open');
        openSavedLinkModal(link, true);
      });
      const remove = document.createElement('button');
      remove.type = 'button';
      remove.textContent = 'Delete';
      remove.addEventListener('click', async function () {
        menu.classList.remove('open');
        try {
          await window.deleteLink(link.ID);
          await loadSavedLinks();
          await loadTopSites();
          showHomeToast('Deleted "' + link.Title + '"');
        } catch (error) {
          showHomeToast('Could not delete: ' + error);
        }
      });
      more.addEventListener('click', function (event) {
        event.stopPropagation();
        menu.classList.toggle('open');
      });
      menu.append(edit, remove);
      card.append(title, more, menu);
      container.appendChild(card);
    });

  }

  let editingLinkID = 0;
  let toastTimer = null;

  function showHomeToast(message) {
    const toast = document.getElementById('homeToast');
    toast.textContent = message;
    toast.classList.add('show');
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { toast.classList.remove('show'); toastTimer = null; }, 2200);
  }

  function openSavedLinkModal(link, editing) {
    editingLinkID = editing && link ? link.ID : 0;
    document.getElementById('savedLinkHeading').textContent = editingLinkID ? 'Edit website' : 'Save website';
    document.getElementById('saveWebsiteButton').textContent = editingLinkID ? 'Update' : 'Save';
    document.getElementById('savedLinkURL').value = (link && link.URL) || '';
    document.getElementById('savedLinkTitle').value = (link && link.Title) || '';
    document.getElementById('savedLinkError').textContent = '';
    document.getElementById('savedLinkOverlay').classList.add('open');
    document.getElementById('savedLinkURL').focus();
  }

  function closeSavedLinkModal() {
    document.getElementById('savedLinkOverlay').classList.remove('open');
    editingLinkID = 0;
  }

  document.getElementById('addSavedLinkBtn').addEventListener('click', function () { openSavedLinkModal(null); });
  document.getElementById('cancelSavedLink').addEventListener('click', closeSavedLinkModal);
  document.getElementById('savedLinkOverlay').addEventListener('click', function (event) {
    if (event.target.id === 'savedLinkOverlay') closeSavedLinkModal();
  });
  document.addEventListener('keydown', function (event) {
    if (event.key === 'Escape') closeSavedLinkModal();
  });
  document.getElementById('savedLinkURL').addEventListener('input', function () {
    const titleInput = document.getElementById('savedLinkTitle');
    if (titleInput.dataset.touched === '1' && titleInput.value.trim()) return;
    try {
      titleInput.value = new URL(normalizeSiteURL(this.value)).hostname.replace(/^www\./, '');
    } catch (e) {}
  });
  document.getElementById('savedLinkTitle').addEventListener('input', function () { this.dataset.touched = '1'; });

  async function saveSavedLink(event) {
    event.preventDefault();
    const urlInput = document.getElementById('savedLinkURL');
    const titleInput = document.getElementById('savedLinkTitle');
    const error = document.getElementById('savedLinkError');
    const button = document.getElementById('saveWebsiteButton');
    const url = normalizeSiteURL(urlInput.value);
    const title = titleInput.value.trim();
    if (!url || !title) {
      error.textContent = 'Enter both a website link and a title.';
      return;
    }

    button.disabled = true;
    error.textContent = '';
    const wasEditing = editingLinkID !== 0;
    try {
      if (wasEditing) {
        await window.updateLink(editingLinkID, title, url);
      } else {
        await window.saveLink(title, url);
      }
    } catch (saveError) {
      console.error('SPARROW could not save the home-page link:', saveError);
      error.textContent = 'Could not save link: ' + (saveError && saveError.message ? saveError.message : String(saveError));
      button.disabled = false;
      return;
    }

    urlInput.value = '';
    titleInput.value = '';
    titleInput.dataset.touched = '';
    closeSavedLinkModal();
    button.disabled = false;
    try {
      await loadSavedLinks();
      await loadTopSites();
    } catch (refreshError) {
      console.error('SPARROW saved the link but could not refresh the home page:', refreshError);
      showHomeToast('Saved, but could not refresh the page.');
      return;
    }
    showHomeToast(wasEditing ? 'Updated "' + title + '"' : 'Saved "' + title + '"');
  }

  document.getElementById('savedLinkForm').addEventListener('submit', saveSavedLink);
  document.getElementById('saveWebsiteButton').addEventListener('click', saveSavedLink);

  document.addEventListener('click', function (event) {
    if (!event.target.closest('.saved-link-card')) {
      document.querySelectorAll('.saved-link-menu.open').forEach(function (menu) {
        menu.classList.remove('open');
      });
    }
  });

  async function applyHomeSearchPlaceholder() {
    try {
      const settings = window.getSettings ? await window.getSettings() : null;
      const engine = (settings && settings.searchEngine) || 'google';
      const name = engine === 'bing' ? 'Bing' : engine === 'brave' ? 'Brave Search' : engine === 'duckduckgo' ? 'DuckDuckGo' : engine === 'yahoo' ? 'Yahoo' : 'Google';
      const input = document.getElementById('searchBar');
      if (input) input.placeholder = 'Search ' + name + ' or type a URL';
    } catch (e) {}
  }

  const savedHomeTheme = (() => {
    try {
      return localStorage.getItem('sparrow-home-theme') || 'dark';
    } catch (e) {
      return 'dark';
    }
  })();

  applyHomeTheme(savedHomeTheme);
  (async function () {
    if (!window.getWallpaperCSS) return;
    try {
      const background = await window.getWallpaperCSS();
      if (background) document.body.style.setProperty('--home-wallpaper', background);
    } catch (error) {
      console.error('Could not load saved home wallpaper:', error);
    }
  })();
  (async function () {
    if (!window.getHomeAnimation) return;
    try {
      await applyHomeAnimation(await window.getHomeAnimation(), false);
    } catch (error) {
      console.error('Could not load saved home animation:', error);
    }
  })();
  applyHomeSearchPlaceholder();
  createStarField(2);
  loadSavedLinks();
  loadTopSites();
</script>

</body>
</html>
`
