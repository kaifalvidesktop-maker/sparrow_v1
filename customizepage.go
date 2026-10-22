package main

// ---------------------------------------------------
// CUSTOMIZE PAGE (browser chrome and Sparrow home)
// ---------------------------------------------------

const customizePageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Customize - SPARROW</title>
<style>
  * { box-sizing: border-box; }
  :root {
    --custom-bg: #0b0b0d; --custom-panel: #17181a; --custom-panel2: #1c1d20;
    --custom-border: #292a2e; --custom-text: #eee; --custom-muted: #a5a6ad;
    --custom-accent: #4d90fe;
  }
  html[data-theme="light"] {
    --custom-bg: #f5f7fa; --custom-panel: #fff; --custom-panel2: #f1f3f5;
    --custom-border: #dfe3e8; --custom-text: #202124; --custom-muted: #5f6368;
    --custom-accent: #1967d2;
  }
  body { margin: 0; background: var(--custom-bg); color: var(--custom-text); font: 13px 'Segoe UI',sans-serif; padding: 30px; }
  main { width: min(100%, 900px); margin: 0 auto; }
  h1 { font-size: 25px; margin: 0 0 8px; }
  .intro { color: var(--custom-muted); margin: 0 0 24px; line-height: 1.6; }
  .section { margin: 20px 0; padding: 22px 24px; background: var(--custom-panel); border: 1px solid var(--custom-border); border-radius: 14px; }
  h2 { font-size: 16px; margin: 0 0 7px; }
  .description { color: var(--custom-muted); margin: 0 0 16px; line-height: 1.55; }
  .row { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; margin: 16px 0; }
  label { min-width: 160px; color: var(--custom-muted); font-size: 13px; }
  input[type=color] { width: 54px; height: 34px; border: 1px solid var(--custom-border); border-radius: 7px; background: var(--custom-panel2); cursor: pointer; }
  input[type=range] { width: min(260px, 55vw); accent-color: var(--custom-accent); }
  .toggle-btn, .reset-btn {
    padding: 9px 14px; border: 1px solid var(--custom-border); border-radius: 8px;
    background: var(--custom-panel2); color: var(--custom-text); cursor: pointer;
  }
  .toggle-btn.on { background: var(--custom-accent); color: white; }
  .reset-btn { border-color: rgba(255,107,107,.4); color: #ff9999; }
  #radiusVal { color: var(--custom-accent); }
  .animation-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(175px, 1fr)); gap: 10px; }
  .animation-option {
    min-height: 52px; padding: 10px 12px; display: flex; align-items: center; gap: 10px;
    text-align: left; border: 1px solid var(--custom-border); border-radius: 9px;
    color: var(--custom-text); background: var(--custom-panel2); cursor: pointer;
  }
  .animation-option:hover { border-color: var(--custom-accent); }
  .animation-option[aria-pressed="true"] { border-color: var(--custom-accent); box-shadow: 0 0 0 2px color-mix(in srgb, var(--custom-accent) 25%, transparent); }
  .animation-option .check { width: 18px; color: var(--custom-accent); font-size: 15px; }
  .wallpaper-option { align-items: center; }
  .wallpaper-swatch {
    display: block; width: 48px; height: 32px; flex: 0 0 48px;
    border: 1px solid rgba(255,255,255,.28); border-radius: 6px;
    background-color: #111; background-size: cover; background-position: center;
  }
  .wallpaper-option[aria-pressed="true"] .wallpaper-swatch { outline: 2px solid var(--custom-accent); outline-offset: 1px; }
  .status { min-height: 18px; margin-top: 12px; color: var(--custom-muted); }
  @media (max-width: 600px) { body { padding: 22px 14px; } .section { padding: 18px; } }
</style>
</head>
<body>
<main>
  <h1>Customize SPARROW</h1>
  <p class="intro">Personalize the browser chrome and the Sparrow home page. Your animation selection is saved on this device and follows you back to home.</p>

  <section class="section">
    <h2>Browser appearance</h2>
    <div class="row">
      <label for="accentColor">Browser color</label>
      <input type="color" id="accentColor">
    </div>
    <div class="row">
      <label for="tabRadius">Tab corner radius</label>
      <input type="range" id="tabRadius" min="0" max="20" step="1">
      <span id="radiusVal"></span>
    </div>
    <div class="row">
      <label for="compactBtn">Compact toolbar</label>
      <button class="toggle-btn" id="compactBtn" type="button">Toggle</button>
    </div>
  </section>

  <section class="section">
    <h2>Home wallpaper</h2>
    <p class="description">Choose the background behind Sparrow home. Your selection is saved automatically and restored when you return to home.</p>
    <div class="animation-grid" id="wallpaperOptions" role="group" aria-label="Home wallpapers"></div>
    <div class="status" id="wallpaperStatus" role="status" aria-live="polite"></div>
  </section>

  <section class="section">
    <h2>Home screen animation</h2>
    <p class="description">Choose an animated background for Sparrow home. Your selection is saved automatically. Select “Default Sparrow home” whenever you want to restore the original home screen.</p>
    <div class="animation-grid" id="animationOptions" role="group" aria-label="Home screen animations"></div>
    <div class="status" id="animationStatus" role="status" aria-live="polite"></div>
  </section>
</main>

<script>
  (async function () {
    var storageKey = 'sparrow-custom-theme';
    var storedTheme = null;
    try {
      storedTheme = JSON.parse(localStorage.getItem(storageKey) || 'null');
    } catch (e) {}

    if (storedTheme && /^#[0-9a-fA-F]{6}$/.test(storedTheme.accentColor || '')) {
      await window.setAccentColor(storedTheme.accentColor);
      if (Number.isInteger(storedTheme.tabRadius)) await window.setTabRadius(storedTheme.tabRadius);
      if (typeof storedTheme.compact === 'boolean') {
        var currentTheme = await window.getCustomTheme();
        if (currentTheme.compact !== storedTheme.compact) await window.toggleCompactToolbar();
      }
      await window.saveCustomTheme();
    }

    var theme = await window.getCustomTheme();
    var accent = document.getElementById('accentColor');
    var radius = document.getElementById('tabRadius');
    var radiusVal = document.getElementById('radiusVal');
    var compactBtn = document.getElementById('compactBtn');

    async function saveCustomThemeNow() {
      var current = await window.getCustomTheme();
      try { localStorage.setItem(storageKey, JSON.stringify(current)); } catch (e) {}
      if (!await window.saveCustomTheme()) console.error('Unable to save custom theme settings');
    }

    accent.value = theme.accentColor;
    radius.value = theme.tabRadius;
    radiusVal.innerText = theme.tabRadius + 'px';
    compactBtn.className = 'toggle-btn' + (theme.compact ? ' on' : '');
    if (window.applyCustomShellColor) window.applyCustomShellColor(theme.accentColor);

    accent.oninput = async function () {
      var color = await window.setAccentColor(accent.value);
      accent.value = color;
      if (window.applyCustomShellColor) window.applyCustomShellColor(color);
      await saveCustomThemeNow();
    };
    radius.oninput = async function () {
      radiusVal.innerText = radius.value + 'px';
      await window.setTabRadius(parseInt(radius.value, 10));
      await saveCustomThemeNow();
    };
    compactBtn.onclick = async function () {
      var on = await window.toggleCompactToolbar();
      compactBtn.className = 'toggle-btn' + (on ? ' on' : '');
      await saveCustomThemeNow();
    };

    var animationGrid = document.getElementById('animationOptions');
    var animationStatus = document.getElementById('animationStatus');
    var wallpaperGrid = document.getElementById('wallpaperOptions');
    var wallpaperStatus = document.getElementById('wallpaperStatus');
    async function loadWallpapers() {
      var options = await window.getWallpaperOptions();
      var selected = await window.getSelectedWallpaper();
      wallpaperGrid.replaceChildren();
      options.forEach(function (option) {
        var button = document.createElement('button');
        button.type = 'button';
        button.className = 'animation-option wallpaper-option';
        button.setAttribute('aria-pressed', option.ID === selected ? 'true' : 'false');
        var check = document.createElement('span');
        check.className = 'check';
        check.textContent = option.ID === selected ? '\u2713' : '';
        var swatch = document.createElement('span');
        swatch.className = 'wallpaper-swatch';
        swatch.setAttribute('aria-hidden', 'true');
        swatch.style.background = option.CSSValue;
        var name = document.createElement('span');
        name.textContent = option.Name;
        button.append(check, swatch, name);
        button.addEventListener('click', async function () {
          button.disabled = true;
          wallpaperStatus.textContent = 'Saving wallpaper...';
          try {
            selected = await window.setWallpaper(option.ID);
            if (selected !== option.ID) throw new Error('The wallpaper was not saved.');
            wallpaperStatus.textContent = option.Name + ' saved for Sparrow home.';
            await loadWallpapers();
          } catch (error) {
            console.error('Unable to save home wallpaper:', error);
            wallpaperStatus.textContent = 'Could not save the wallpaper. Please try again.';
            button.disabled = false;
          }
        });
        wallpaperGrid.appendChild(button);
      });
    }

    async function loadAnimations() {
      var options = await window.getHomeAnimationOptions();
      var selected = await window.getHomeAnimation();
      animationGrid.replaceChildren();
      options.forEach(function (option) {
        var button = document.createElement('button');
        button.type = 'button';
        button.className = 'animation-option';
        button.setAttribute('aria-pressed', option.ID === selected ? 'true' : 'false');
        var check = document.createElement('span');
        check.className = 'check';
        check.textContent = option.ID === selected ? '\u2713' : '';
        var name = document.createElement('span');
        name.textContent = option.Name;
        button.append(check, name);
        button.addEventListener('click', async function () {
          button.disabled = true;
          animationStatus.textContent = 'Saving animation...';
          try {
            selected = await window.setHomeAnimation(option.ID);
            if (selected !== option.ID) throw new Error('The animation was not saved.');
            animationStatus.textContent = option.ID === 'default' ? 'Default Sparrow home restored.' : option.Name + ' saved for Sparrow home.';
            await loadAnimations();
          } catch (error) {
            console.error('Unable to save home animation:', error);
            animationStatus.textContent = 'Could not save the animation. Please try again.';
            button.disabled = false;
          }
        });
        animationGrid.appendChild(button);
      });
    }
    await loadWallpapers();
    await loadAnimations();
  })().catch(function (error) {
    console.error('Unable to load customization settings:', error);
  });
</script>
</body>
</html>
`
