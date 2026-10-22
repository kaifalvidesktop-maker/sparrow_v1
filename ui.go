package main

// ---------------------------------------------------
// SPARROW BROWSER CHROME (toolbar + tab strip)
//
// This used to be a static "shell" page with an <iframe> where every
// other page (internal or external) was loaded. That's what broke
// search and most real browsing: sites like Google send
// X-Frame-Options / CSP headers that refuse to render inside someone
// else's iframe, and even when a site *did* allow it, JS here trying
// to reach into `frame.contentWindow` for a different-origin page is
// blocked by the browser's same-origin policy. Both together meant
// external navigation basically never worked.
//
// The fix: there is no more iframe. Every navigation (internal
// sparrow:// page or real external site) is a REAL top-level
// w.Navigate() call, exactly like a normal browser tab. This script
// is registered with w.Init() in main.go, so it re-runs and re-mounts
// the toolbar on top of whatever page has just loaded — internal or
// external, it doesn't matter, since it's injected directly into that
// page's own document rather than a sandboxed frame.
// ---------------------------------------------------

const chromeInjectionJS = `
(function () {
  if (window.top !== window) return;

  if (!window.__sparrowOfflineNavigationWrapped &&
      typeof window.realNavigate === 'function' &&
      typeof window.openLinkInNewTab === 'function' &&
      typeof window.navigateOffline === 'function') {
    var sparrowNativeNavigate = window.realNavigate;
    var sparrowNativeOpenLinkInNewTab = window.openLinkInNewTab;
    function shouldShowOfflinePage(input) {
      var value = String(input || '').trim();
      return navigator.onLine === false &&
        value !== '' &&
        !/^(sparrow:|file:)/i.test(value) &&
        !/^https?:\/\/(?:localhost|127\.0\.0\.1|\[::1\])(?:[:/]|$)/i.test(value);
    }
    window.realNavigate = function (input) {
      if (shouldShowOfflinePage(input)) return window.navigateOffline(String(input), false);
      return sparrowNativeNavigate.apply(window, arguments);
    };
    window.openLinkInNewTab = function (input) {
      if (shouldShowOfflinePage(input)) return window.navigateOffline(String(input), true);
      return sparrowNativeOpenLinkInNewTab.apply(window, arguments);
    };
    window.__sparrowOfflineNavigationWrapped = true;
  }

  var sparrowNativeOpen = window.open;
  function sparrowNavigatePopup(rawURL) {
    if (rawURL == null || String(rawURL).trim() === '') return false;
    try {
      var destination = new URL(String(rawURL), window.location.href);
      if (destination.protocol !== 'http:' && destination.protocol !== 'https:') return false;
      if (typeof window.openLinkInNewTab !== 'function') return false;
      Promise.resolve(window.openLinkInNewTab(destination.href)).catch(function (error) {
        console.error('SPARROW could not open link in a new tab:', error);
      });
      return true;
    } catch (error) {
      console.error('SPARROW could not open account link:', error);
      return false;
    }
  }

  function isSearchEnginePage() {
    var host = window.location.hostname.toLowerCase();
    return /(^|\.)google\.(com|[a-z]{2,3}(?:\.[a-z]{2})?)$/.test(host) ||
      /(^|\.)bing\.com$/.test(host) ||
      /(^|\.)duckduckgo\.com$/.test(host) ||
      /(^|\.)search\.yahoo\.com$/.test(host) ||
      /(^|\.)search\.brave\.com$/.test(host);
  }

  function isDownloadableURL(rawURL) {
    try {
      var pathname = new URL(rawURL, window.location.href).pathname.toLowerCase();
      return /\.(zip|rar|7z|exe|msi|msix|msixbundle|appx|appxbundle|dmg|pkg|deb|rpm|jar|bin|img|iso|apk|ipa|tar|gz|tgz|bz2|xz|zst|cab|whl|torrent|pdf|docx?|xlsx?|pptx?|csv|rtf|odt|ods|odp|epub|mp3|wav|flac|m4a|aac|mp4|mkv|avi|mov|wmv|flv)$/.test(pathname);
    } catch (error) {
      return false;
    }
  }

  // Route every download through Sparrow's own engine (links, programmatic
  // anchor.click(), blob:/data: downloads).
  function sparrowStartAnchorDownload(anchor) {
    try {
      if (!anchor || !anchor.href || (anchor.closest && anchor.closest('#__sparrow_chrome'))) return false;
      var href = anchor.href;
      if (/^(blob|data):/i.test(href)) {
        if (!anchor.hasAttribute('download') || typeof window.saveDownloadData !== 'function') return false;
        var fileName = anchor.getAttribute('download') || 'download';
        fetch(href).then(function (r) { return r.blob(); }).then(function (blob) {
          var reader = new FileReader();
          reader.onload = function () {
            var data = String(reader.result);
            window.saveDownloadData(fileName, data.slice(data.indexOf(',') + 1));
          };
          reader.readAsDataURL(blob);
        }).catch(function (error) {
          console.error('SPARROW could not save download:', error);
        });
        return true;
      }
      if (!/^https?:/i.test(href) || typeof window.downloadFile !== 'function') return false;
      if (!anchor.hasAttribute('download') && !isDownloadableURL(href)) return false;
      Promise.resolve(window.downloadFile(href)).catch(function (error) {
        console.error('SPARROW could not start download:', error);
      });
      return true;
    } catch (error) {
      return false;
    }
  }

  var sparrowNativeAnchorClick = HTMLAnchorElement.prototype.click;
  HTMLAnchorElement.prototype.click = function () {
    if (sparrowStartAnchorDownload(this)) return;
    return sparrowNativeAnchorClick.apply(this, arguments);
  };

  // Keep account/sign-in popups in Sparrow instead of allowing the embedded
  // webview to create a separate native window.
  window.open = function (url) {
    try {
      var popupURL = new URL(String(url || ''), window.location.href);
      var popupHost = popupURL.hostname.toLowerCase();
      var isYouTube = /(^|\.)youtube\.com$/.test(window.location.hostname.toLowerCase()) ||
        /(^|\.)youtu\.be$/.test(window.location.hostname.toLowerCase());
      var isGoogleSignIn = popupHost === 'accounts.google.com' ||
        popupHost === 'accounts.youtube.com' ||
        popupHost === 'consent.google.com';
      if (isYouTube && isGoogleSignIn) {
        return sparrowNativeOpen.apply(window, arguments);
      }
    } catch (error) {
      console.error('SPARROW could not inspect popup destination:', error);
    }
    if (sparrowNavigatePopup(url)) return null;
    return sparrowNativeOpen.apply(window, arguments);
  };

  var sparrowAutoPiPAttempted = false;
  function sparrowRequestAutoPiP() {
    if (sparrowAutoPiPAttempted || document.pictureInPictureElement ||
        !document.pictureInPictureEnabled) return;
    var videos = document.querySelectorAll('video');
    var video = null;
    var largestArea = 0;
    for (var i = 0; i < videos.length; i++) {
      var candidate = videos[i];
      if (candidate.paused || candidate.ended || candidate.readyState < 2 ||
          typeof candidate.requestPictureInPicture !== 'function') continue;
      var area = candidate.videoWidth * candidate.videoHeight;
      if (!video || area > largestArea) {
        video = candidate;
        largestArea = area;
      }
    }
    if (!video) return;
    sparrowAutoPiPAttempted = true;
    video.requestPictureInPicture().catch(function (error) {
      console.warn('SPARROW could not open automatic picture-in-picture:', error);
    });
  }
  window.addEventListener('blur', sparrowRequestAutoPiP);
  document.addEventListener('visibilitychange', function () {
    if (document.visibilityState === 'hidden') sparrowRequestAutoPiP();
  });
  window.addEventListener('focus', function () { sparrowAutoPiPAttempted = false; });

  document.addEventListener('click', function (event) {
    if (event.defaultPrevented || event.button !== 0) return;
    if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
    var element = event.target && event.target.nodeType === 1 ? event.target : event.target && event.target.parentElement;
    var anchor = element && element.closest ? element.closest('a[href]') : null;
    if (!anchor || anchor.closest('#__sparrow_chrome')) return;
    if (sparrowStartAnchorDownload(anchor)) {
      event.preventDefault();
      event.stopImmediatePropagation();
      return;
    }
    if (!/^https?:/i.test(anchor.href)) return;
    var target = (anchor.getAttribute('target') || '').trim().toLowerCase();
    if ((!target || target === '_self') && !isSearchEnginePage()) return;
    if (!sparrowNavigatePopup(anchor.href)) return;
    event.preventDefault();
    event.stopImmediatePropagation();
  }, true);

  // ---- Trusted Types-safe DOM helpers ----------------------------------
  // YouTube (and other Google sites) send "require-trusted-types-for 'script'",
  // which makes every plain-string innerHTML assignment throw. That threw
  // inside mountChrome(), so the title bar, tab bar and address bar never
  // got mounted. These helpers try innerHTML first and fall back to building
  // the nodes by hand, which Trusted Types does not restrict.
  var SPARROW_VOID = { input: 1, br: 1, img: 1, hr: 1, meta: 1, link: 1 };
  var SPARROW_ENTITIES = { amp: '&', lt: '<', gt: '>', quot: '"', apos: "'", nbsp: '\u00a0', times: '\u00d7' };

  function sparrowDecode(text) {
    return text.replace(/&(#x[0-9a-fA-F]+|#\d+|[a-zA-Z]+);/g, function (all, body) {
      if (body.charAt(0) === '#') {
        var code = body.charAt(1) === 'x' || body.charAt(1) === 'X'
          ? parseInt(body.slice(2), 16) : parseInt(body.slice(1), 10);
        try { return String.fromCodePoint(code); } catch (e) { return all; }
      }
      return Object.prototype.hasOwnProperty.call(SPARROW_ENTITIES, body) ? SPARROW_ENTITIES[body] : all;
    });
  }

  function sparrowBuild(parent, html) {
    var token = /<!--[\s\S]*?-->|<\/([a-zA-Z][\w-]*)\s*>|<([a-zA-Z][\w-]*)((?:\s+[^\s=>\/]+(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+))?)*)\s*(\/?)>|[^<]+|</g;
    var attrRe = /([^\s=>\/]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+)))?/g;
    var stack = [parent];
    var m;
    while ((m = token.exec(html)) !== null) {
      var current = stack[stack.length - 1];
      var raw = m[0];
      if (raw.indexOf('<!--') === 0) continue;
      if (m[1]) {
        var name = m[1].toLowerCase();
        for (var i = stack.length - 1; i > 0; i--) {
          if (stack[i].localName === name) { stack.length = i; break; }
        }
      } else if (m[2]) {
        var el = document.createElement(m[2].toLowerCase());
        var a;
        attrRe.lastIndex = 0;
        while ((a = attrRe.exec(m[3] || '')) !== null) {
          var attrName = a[1];
          if (/^on/i.test(attrName)) continue;
          var value = a[2] !== undefined ? a[2] : a[3] !== undefined ? a[3] : a[4] !== undefined ? a[4] : '';
          try { el.setAttribute(attrName, sparrowDecode(value)); } catch (e) {}
        }
        current.appendChild(el);
        if (!m[4] && !SPARROW_VOID[el.localName]) stack.push(el);
      } else {
        current.appendChild(document.createTextNode(sparrowDecode(raw)));
      }
    }
  }

  function sparrowClear(el) {
    while (el.firstChild) el.removeChild(el.firstChild);
  }

  function sparrowAppendHTML(el, html) {
    try {
      el.insertAdjacentHTML('beforeend', html);
    } catch (e) {
      sparrowBuild(el, html);
    }
  }

  function sparrowSetHTML(el, html) {
    if (html === '') { sparrowClear(el); return; }
    try {
      el.innerHTML = html;
    } catch (e) {
      sparrowClear(el);
      sparrowBuild(el, html);
    }
  }
  // ----------------------------------------------------------------------


  // This script is injected into every top-level page. Some websites replace
  // document.body/document.documentElement or otherwise remove nodes that we
  // inject. Never permanently lock ourselves out of remounting the browser UI.
  if (window.__sparrowChromeController) {
    try { window.__sparrowChromeController.ensure(); } catch (e) {}
    return;
  }

  var controller = {
    mounted: false,
    mounting: false,
    observer: null,
    interval: null,
    recoveryInterval: null,
    ensureTimer: null,
    handlers: {},
    ensure: function () {}
  };
  window.__sparrowChromeController = controller;

  var pageHistory = controller.pageHistory = {
    entries: [window.location.href],
    index: 0,
    pendingIndex: -1
  };
  var nativeHistory = {
    pushState: window.history.pushState,
    replaceState: window.history.replaceState,
    back: window.history.back,
    forward: window.history.forward,
    go: window.history.go
  };

  function updateHistoryButtons() {
    var back = document.getElementById('btn-back');
    var forward = document.getElementById('btn-forward');
    if (!back || !forward) return;
    var pageCanGoBack = controller.pageHistory.index > 0;
    var pageCanGoForward = controller.pageHistory.index < controller.pageHistory.entries.length - 1;
    var browserCanGoBack = !!(controller.navigationState && controller.navigationState.canBack);
    var browserCanGoForward = !!(controller.navigationState && controller.navigationState.canForward);
    back.className = pageCanGoBack || browserCanGoBack ? 'nav-btn' : 'nav-btn disabled';
    forward.className = pageCanGoForward || browserCanGoForward ? 'nav-btn' : 'nav-btn disabled';
  }

  function recordPageHistoryEntry() {
    pageHistory.entries = pageHistory.entries.slice(0, pageHistory.index + 1);
    pageHistory.entries.push(window.location.href);
    pageHistory.index = pageHistory.entries.length - 1;
    pageHistory.pendingIndex = -1;
    updateHistoryButtons();
  }

  function trackPageHistoryMove(delta) {
    var target = pageHistory.index + delta;
    if (target >= 0 && target < pageHistory.entries.length) {
      pageHistory.index = target;
      pageHistory.pendingIndex = target;
      updateHistoryButtons();
    }
  }

  var isImageDocument = /^image\//i.test(document.contentType || '');
  if (isImageDocument && document.documentElement) {
    document.documentElement.classList.add('sparrow-image-document');
  }

  try {
    window.history.pushState = function () {
      var result = nativeHistory.pushState.apply(this, arguments);
      recordPageHistoryEntry();
      return result;
    };
    window.history.replaceState = function () {
      var result = nativeHistory.replaceState.apply(this, arguments);
      pageHistory.entries[pageHistory.index] = window.location.href;
      updateHistoryButtons();
      return result;
    };
    window.history.back = function () {
      trackPageHistoryMove(-1);
      return nativeHistory.back.apply(this, arguments);
    };
    window.history.forward = function () {
      trackPageHistoryMove(1);
      return nativeHistory.forward.apply(this, arguments);
    };
    window.history.go = function (delta) {
      trackPageHistoryMove(Number(delta) || 0);
      return nativeHistory.go.apply(this, arguments);
    };
  } catch (e) {}

  document.addEventListener('popstate', function () {
    if (pageHistory.pendingIndex >= 0) {
      pageHistory.index = pageHistory.pendingIndex;
      pageHistory.pendingIndex = -1;
    } else {
      var currentIndex = pageHistory.entries.lastIndexOf(window.location.href);
      if (currentIndex >= 0) pageHistory.index = currentIndex;
    }
    updateHistoryButtons();
  });

  function whenBodyReady(cb) {
    if (document.body) { cb(); return; }
    document.addEventListener('DOMContentLoaded', cb, { once: true });
    var t = setInterval(function () {
      if (document.body) { clearInterval(t); cb(); }
    }, 10);
  }

  function mountChrome() {
    if (controller.mounting) return;
    if (document.getElementById('__sparrow_chrome')) {
      controller.mounted = true;
      return;
    }

    controller.mounting = true;
    setTimeout(function () {
      // Prevent a failed mount from permanently disabling future recovery.
      if (!document.getElementById('__sparrow_chrome')) controller.mounting = false;
    }, 1000);

    var oldStyle = document.getElementById('__sparrow_chrome_style');
    if (oldStyle) oldStyle.remove();

    var style = document.createElement('style');
    style.id = '__sparrow_chrome_style';
    style.setAttribute('data-sparrow-chrome', 'true');
    style.textContent =
      ':root {' +
      '  --shell-bg: #0b0b0d;' +
      '  --window-bar: #0b0b0d;' +
      '  --tabstrip-bg: #0f0f11;' +
      '  --toolbar-bg: #1e1f22;' +
      '  --chrome-panel: #17181a;' +
      '  --chrome-panel-2: #1c1d20;' +
      '  --chrome-accent: #4d90fe;' +
      '  --chrome-text: #eee;' +
      '  --chrome-muted: #aaa;' +
      '  --chrome-subtle: #666;' +
      '  --chip-bg: #1a1b1e;' +
      '  --chip-active: #1e1f22;' +
      '  --chip-text: #999;' +
      '  --chip-border: rgba(255,255,255,0.04);' +
      '  --menu-bg: #1c1d20;' +
      '  --menu-border: #2a2a2c;' +
      '  --menu-text: #ccc;' +
      '  --shadow: rgba(0,0,0,0.5);' +
      '}' +
      'html[data-theme="light"] {' +
      '  --shell-bg: #f5f7fa;' +
      '  --window-bar: #ffffff;' +
      '  --tabstrip-bg: #f0f2f4;' +
      '  --toolbar-bg: #ffffff;' +
      '  --chrome-panel: #ffffff;' +
      '  --chrome-panel-2: #ffffff;' +
      '  --chrome-accent: #4d90fe;' +
      '  --chrome-text: #1f2328;' +
      '  --chrome-muted: #4b5563;' +
      '  --chrome-subtle: #9ca3af;' +
      '  --chip-bg: #f1f3f4;' +
      '  --chip-active: #ffffff;' +
      '  --chip-text: #4b5563;' +
      '  --chip-border: #dfe3e8;' +
      '  --menu-bg: #ffffff;' +
      '  --menu-border: #e5e7eb;' +
      '  --menu-text: #374151;' +
      '  --shadow: rgba(17,24,39,0.12);' +
      '}' +
      '#__sparrow_chrome, #__sparrow_chrome * { box-sizing: border-box !important; }' +
      '#__sparrow_chrome {' +
      '  all: initial !important;' +
      '  position: fixed !important; top: 0 !important; left: 0 !important; right: 0 !important;' +
      '  width: 100vw !important; height: 86px !important; min-width: 0 !important; min-height: 86px !important;' +
      '  display: block !important; visibility: visible !important; opacity: 1 !important;' +
      '  overflow: visible !important; z-index: 2147483647 !important; pointer-events: auto !important;' +
      '  font-family: "Segoe UI", sans-serif !important; font-size: 13px !important; line-height: normal !important;' +
      '  text-align: left !important; direction: ltr !important; white-space: normal !important;' +
      '  color: #eee !important; background: var(--tabstrip-bg, #0b0b0d) !important; transform: none !important;' +
      '  filter: none !important; clip: auto !important; clip-path: none !important;' +
      '  contain: layout style !important; isolation: isolate !important;' +
      '}' +
      '#__sparrow_chrome * { visibility: visible !important; opacity: 1 !important; pointer-events: auto !important; }' +
      '#__sparrow_chrome button, #__sparrow_chrome input { all: unset; box-sizing: border-box !important; font-family: "Segoe UI", sans-serif !important; }' +
      '#__sparrow_chrome button { cursor: pointer !important; }' +
      '#__sparrow_chrome input { cursor: text !important; }' +
      '#__sparrow_chrome button, #__sparrow_chrome input, #__sparrow_chrome div, #__sparrow_chrome span {' +
      '  font-family: "Segoe UI", sans-serif !important;' +
      '}' +
      'html:not(.sparrow-image-document) { margin-top: 0 !important; padding-top: 86px !important; scroll-padding-top: 86px; }' +
      'html:not(.sparrow-image-document) { padding-left: 0 !important; }' +
      'html.sparrow-vertical-tabs:not(.sparrow-image-document) { padding-top: 78px !important; padding-left: 220px !important; scroll-padding-top: 78px; }' +
      'html:not(.sparrow-image-document) body { padding-top: 0 !important; }' +
      '#__sparrow_chrome {' +
      '  position: fixed; top: 0; left: 0; right: 0; z-index: 2147483647;' +
      '  font-family: "Segoe UI", sans-serif; background: var(--shell-bg);' +
      '  box-shadow: 0 1px 0 rgba(0,0,0,0.12);' +
      '}' +
      '#window-titlebar {' +
      '  position: fixed !important; top: 0 !important; left: 0 !important; right: 0 !important; width: 100vw !important; height: 30px !important;' +
      '  display: flex !important; align-items: center !important; justify-content: flex-end !important;' +
      '  padding: 0 8px 0 12px !important; background: transparent !important; color: var(--chrome-subtle) !important; font-size: 11px !important;' +
      '  user-select: none !important; cursor: default !important; z-index: 30 !important; visibility: visible !important; opacity: 1 !important; pointer-events: none !important;' +
      '}' +
      '#window-title { display: none; }' +
      '.vertical-tabs-toggle {' +
      '  position: relative !important; left: auto !important; top: auto !important; width: 30px !important; height: 30px !important; min-width: 30px !important; flex: 0 0 30px !important; padding: 0 8px !important;' +
      '  display: flex !important; align-items: center !important; justify-content: center !important; border: 0 !important; border-radius: 6px !important;' +
      '  color: var(--chrome-muted) !important; background: transparent !important; font-size: 16px !important; cursor: pointer !important;' +
      '}' +
      '.vertical-tabs-toggle:hover { background: rgba(255,255,255,0.08) !important; color: var(--chrome-text) !important; }' +
      '#__sparrow_chrome:not(.vertical-tabs) #tabstrip > .vertical-tabs-toggle { position: sticky !important; left: 6px !important; z-index: 5 !important; background: var(--tabstrip-bg) !important; box-shadow: 4px 0 6px var(--tabstrip-bg); }' +
      '#__sparrow_chrome:not(.vertical-tabs) > #new-tab-btn { position: fixed !important; top: 4px !important; right: 150px !important; z-index: 25 !important; margin: 0 !important; }' +
      '#window-controls { display: flex; height: 30px; pointer-events: auto !important; }' +
      '.window-control {' +
      '  width: 46px !important; height: 30px !important; min-width: 46px !important; min-height: 30px !important; max-width: 46px !important; max-height: 30px !important;' +
      '  border: 0 !important; background: transparent !important; color: var(--chrome-muted) !important;' +
      '  display: flex !important; align-items: center !important; justify-content: center !important; cursor: pointer !important; font-size: 13px !important;' +
      '}' +
      '.window-control:hover { background: rgba(255,255,255,0.06); color: var(--chrome-text); }' +
      '.window-control.close:hover { background: #c42b1c; color: #fff; }' +
      '#tabstrip {' +
      '  position: fixed !important; top: 0 !important; left: 0 !important; right: 180px !important; width: calc(100vw - 180px) !important; height: 38px !important;' +
      '  z-index: 20 !important; display: flex !important; align-items: center !important; background: var(--tabstrip-bg) !important;' +
      '  padding: 0 6px !important; gap: 4px !important; overflow-x: auto !important; overflow-y: hidden !important; white-space: nowrap !important;' +
      '  visibility: visible !important; opacity: 1 !important;' +
      '  scroll-behavior: smooth;' +
      '}' +
      '*::-webkit-scrollbar {' +
      '  width: 10px;' +
      '  height: 10px;' +
      '}' +
      '*::-webkit-scrollbar-track {' +
      '  background: #050505;' +
      '}' +
      '*::-webkit-scrollbar-thumb {' +
      '  background: linear-gradient(180deg, #333, #111, #292929);' +
      '  background-size: 100% 200%; background-position: 0 0;' +
      '  border: 3px solid transparent;' +
      '  background-clip: padding-box;' +
      '  border-radius: 999px;' +
      '  min-height: 36px;' +
      '  transition: background-position 0.35s ease, box-shadow 0.2s ease, transform 0.2s ease;' +
      '}' +
      '*::-webkit-scrollbar-thumb:hover {' +
      '  background-position: 0 100%; box-shadow: 0 0 8px color-mix(in srgb, var(--chrome-accent) 55%, transparent); transform: scaleX(1.08);' +
      '}' +
      '*::-webkit-scrollbar:vertical { width: 12px; }' +
      '*::-webkit-scrollbar:horizontal { height: 10px; }' +
      '* { scrollbar-width: thin; scrollbar-color: #242424 #050505; }' +
      '#tabstrip::-webkit-scrollbar { height: 8px; }' +
      '#tabstrip::-webkit-scrollbar-thumb { background: var(--chrome-accent); background-size: 100% 100%; border-width: 1px; }' +
      '#menu-dropdown::-webkit-scrollbar, #notif-panel::-webkit-scrollbar, #suggest-box::-webkit-scrollbar {' +
      '  width: 8px;' +
      '  height: 8px;' +
      '}' +
      '.tab-chip {' +
      '  display: flex; flex: 0 0 auto; align-items: center; gap: 8px; background: var(--chip-bg);' +
      '  color: var(--chip-text); padding: 7px 10px 7px 14px; border-radius: 10px 10px 0 0;' +
      '  font-size: 12px; max-width: 180px; min-width: 120px; cursor: pointer;' +
      '  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;' +
      '  transition: background 0.15s, color 0.15s; position: relative; top: 1px; border: 1px solid var(--chip-border); border-bottom: none;' +
      '}' +
      '.tab-chip.active { background: var(--chip-active); color: var(--chrome-text); box-shadow: 0 -3px 10px rgba(0,0,0,0.15); }' +
      '.tab-chip .tab-favicon { width: 16px; height: 16px; flex: 0 0 16px; object-fit: contain; border-radius: 2px; }' +
      '.tab-chip .tab-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1 1 auto; min-width: 0; }' +
      '.tab-chip .tab-close { width: 24px; height: 24px; display: flex; align-items: center; justify-content: center; color: var(--chrome-subtle); font: inherit; font-size: 18px; line-height: 1; padding: 0; border: 0; border-radius: 5px; background: transparent; cursor: pointer; flex: 0 0 24px; position: relative; z-index: 1; }' +
      '.tab-chip .tab-close:hover { background: #c42b1c !important; color: #fff !important; }' +
      '.tab-chip.incognito { background: #1a1420; border: 1px solid #3a2a4a; }' +
      '.tab-chip:hover { background: color-mix(in srgb, var(--chrome-accent) 20%, var(--chip-bg)); color: var(--chrome-text); }' +
      '#__sparrow_chrome.vertical-tabs #tabstrip {' +
      '  flex-direction: column !important; align-items: stretch !important; justify-content: flex-start !important;' +
      '  gap: 5px !important; border-right: 1px solid var(--chip-border) !important; box-shadow: 2px 0 12px rgba(0,0,0,0.12) !important;' +
      '}' +
      '#__sparrow_chrome.vertical-tabs .vertical-tabs-toggle {' +
      '  width: 100% !important; flex-basis: 30px !important; justify-content: flex-start !important; padding-left: 8px !important;' +
      '}' +
      '#tab-search-input-inline { display: none !important; }' +
      '#__sparrow_chrome.vertical-tabs #tab-search-input-inline {' +
      '  display: block !important; flex: 0 0 32px !important; width: 100% !important; height: 32px !important; min-height: 32px !important;' +
      '  padding: 0 9px !important; border: 1px solid var(--chip-border) !important; border-radius: 8px !important;' +
      '  background: var(--chip-bg) !important; color: var(--chrome-text) !important; font-size: 12px !important; outline: none !important;' +
      '}' +
      '#__sparrow_chrome.vertical-tabs #tab-search-input-inline:focus { border-color: var(--chrome-accent) !important; }' +
      '#__sparrow_chrome.vertical-tabs .tab-search-empty { padding: 10px 6px; color: var(--chrome-muted); font-size: 12px; text-align: center; }' +
      '#__sparrow_chrome.vertical-tabs .tab-chip {' +
      '  top: 0 !important; flex: 0 0 40px !important; width: 100% !important; max-width: none !important; min-width: 0 !important; height: 40px !important;' +
      '  border: 1px solid var(--chip-border) !important; border-radius: 8px !important; padding: 0 8px 0 10px !important; box-shadow: none !important;' +
      '}' +
      '#__sparrow_chrome.vertical-tabs .tab-chip .tab-title { display: block !important; flex: 1 1 auto !important; min-width: 0 !important; text-align: left !important; direction: ltr !important; }' +
      '#__sparrow_chrome.vertical-tabs .tab-chip.active { border-color: var(--chrome-accent) !important; box-shadow: inset 3px 0 var(--chrome-accent) !important; }' +
      '#__sparrow_chrome.vertical-tabs #new-tab-btn {' +
      '  position: relative !important; top: auto !important; right: auto !important; margin: 8px 0 0 2px !important; align-self: flex-start !important;' +
      '}' +
      '#new-tab-btn {' +
      '  width: 30px; height: 30px; border-radius: 8px; background: var(--chrome-accent); color: var(--new-tab-text, #fff);' +
      '  display: flex; align-items: center; justify-content: center; cursor: pointer;' +
      '  font-size: 18px; font-weight: 500; flex-shrink: 0; margin-bottom: 2px; border: 1px solid var(--chip-border);' +
      '  position: relative; right: auto; margin-left: 0; align-self: center; overflow: hidden;' +
      '  box-shadow: 0 0 0 1px rgba(77,144,254,0.08), 0 2px 6px rgba(0,0,0,0.14);' +
      '  transition: color 0.15s ease, background 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;' +
      '}' +
      '#new-tab-btn::before { content: ""; position: absolute; top: 0; left: 7px; right: 7px; height: 2px; background: var(--new-tab-text, #fff); opacity: 0.8; }' +
      '#new-tab-btn:hover { background: color-mix(in srgb, var(--chrome-accent) 86%, var(--chip-bg)); color: var(--new-tab-text, #fff); border-color: color-mix(in srgb, var(--chrome-accent) 55%, var(--chip-border)); box-shadow: 0 0 0 3px rgba(77,144,254,0.12), 0 4px 10px rgba(0,0,0,0.18); transform: translateY(-1px); }' +
      '#new-tab-btn:focus-visible { outline: 2px solid var(--chrome-accent); outline-offset: 2px; }' +
      '#toolbar {' +
      '  position: fixed !important; top: 38px !important; left: 0 !important; right: 0 !important; width: 100vw !important; height: 48px !important;' +
      '  display: flex !important; align-items: center !important; background: var(--toolbar-bg) !important; padding: 8px 12px !important;' +
      '  gap: 8px !important; box-shadow: 0 4px 12px var(--shadow) !important; z-index: 10 !important; visibility: visible !important; opacity: 1 !important;' +
      '}' +
      '.nav-btn {' +
      '  width: 32px !important; height: 32px !important; min-width: 32px !important; min-height: 32px !important; border-radius: 8px !important;' +
      '  background: transparent !important; color: var(--chrome-muted) !important; border: 0 !important;' +
      '  display: flex !important; align-items: center !important; justify-content: center !important; cursor: pointer !important;' +
      '  font-size: 15px !important; flex-shrink: 0 !important; transition: 0.15s;' +
      '}' +
      '.nav-btn:hover { background: rgba(255,255,255,0.06); color: var(--chrome-text); }' +
      '#btn-back:not(.disabled):hover, #btn-forward:not(.disabled):hover, #btn-reload:hover, #btn-home:hover { background: color-mix(in srgb, var(--chrome-accent) 24%, var(--toolbar-bg)) !important; color: var(--chrome-text) !important; }' +
      '#btn-back:not(.disabled):active, #btn-forward:not(.disabled):active, #btn-reload:active, #btn-home:active { background: color-mix(in srgb, var(--chrome-accent) 48%, var(--toolbar-bg)) !important; color: var(--chrome-text) !important; transform: scale(0.94); }' +
      '.nav-btn.active { background: color-mix(in srgb, var(--chrome-accent) 20%, transparent) !important; color: var(--chrome-accent) !important; }' +
      '.nav-btn.disabled { opacity: 0.3; pointer-events: none; }' +
      '#address-wrap {' +
      '  flex: 1 1 auto !important; min-width: 0 !important; display: flex !important; align-items: center !important;' +
      '  background: var(--chrome-panel) !important; border: 1px solid var(--menu-border) !important; border-radius: 20px !important;' +
      '  padding: 0 8px 0 16px !important; height: 34px !important; min-height: 34px !important; gap: 8px !important; transition: border-color 0.15s;' +
      '}' +
      '#address-wrap:focus-within { border-color: var(--chrome-accent); box-shadow: 0 0 0 3px rgba(77,144,254,0.15); }' +
      '#shield-icon { font-size: 12px; color: #3ddc84; flex-shrink: 0; cursor: pointer; }' +
      '#address-bar { flex: 1 1 auto !important; width: auto !important; min-width: 0 !important; height: 32px !important; background: transparent !important; border: none !important; outline: none !important; color: var(--chrome-text) !important; font-size: 13px !important; display: block !important; }' +
      '#address-bar::placeholder { color: var(--chrome-subtle); }' +
      '#__sparrow_chrome #btn-app-lock { width: 26px; height: 26px; flex: 0 0 26px; display: flex; align-items: center; justify-content: center; border: 0; border-radius: 6px; padding: 4px; background: transparent; color: #fff !important; cursor: pointer; }' +
      '#__sparrow_chrome #btn-app-lock:hover { background: rgba(255,255,255,0.08); color: var(--chrome-text); }' +
      '#__sparrow_chrome #btn-app-lock svg { width: 16px; height: 16px; fill: currentColor; }' +
      '#star-btn { width: 26px; height: 26px; border-radius: 6px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--chrome-subtle); font-size: 15px; flex-shrink: 0; }' +
      '#star-btn:hover { background: rgba(255,255,255,0.06); color: #ffd93d; }' +
      '#star-btn.saved { color: #ffd93d; }' +
      '#__sparrow_chrome #qr-share-btn { width: 26px; height: 26px; padding: 4px; border: 0; border-radius: 6px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: #fff !important; background: transparent; flex-shrink: 0; transition: background 0.16s ease, color 0.16s ease, box-shadow 0.16s ease; }' +
      '#qr-share-btn svg { width: 17px; height: 17px; fill: currentColor; }' +
      '#__sparrow_chrome #qr-share-btn:hover { background: rgba(77,144,254,0.2); color: #fff; box-shadow: 0 0 10px rgba(77,144,254,0.28); }' +
      '#menu-btn { width: 32px; height: 32px; border-radius: 8px; background: transparent; color: var(--chrome-muted); display: flex; align-items: center; justify-content: center; cursor: pointer; font-size: 16px; flex-shrink: 0; position: relative; }' +
      '#menu-btn:hover { background: rgba(255,255,255,0.06); color: var(--chrome-text); }' +
      '#menu-dropdown {' +
      '  position: absolute; top: 46px; right: 8px; background: var(--menu-bg); border: 1px solid var(--menu-border);' +
      '  border-radius: 12px; box-shadow: 0 10px 30px var(--shadow); width: 220px; padding: 8px; z-index: 100;' +
      '  max-height: 80vh; overflow-y: auto; display: none; opacity: 0; pointer-events: none; transform: translateY(-8px) scale(0.98);' +
      '  transition: opacity 0.18s ease, transform 0.18s ease, box-shadow 0.18s ease;' +
      '}' +
      '#menu-dropdown.open { display: block; opacity: 1; pointer-events: auto; transform: translateY(0) scale(1); }' +
      '.menu-item { padding: 10px 14px; border-radius: 8px; color: var(--menu-text); font-size: 13px; cursor: pointer; display: flex; justify-content: space-between; transition: background 0.15s ease, transform 0.15s ease; }' +
      '.menu-item:hover { background: rgba(0,0,0,0.05); color: var(--chrome-text); }' +
      '.menu-item .shortcut-hint { color: var(--chrome-subtle); font-size: 11px; }' +
      '.menu-divider { height: 1px; background: var(--menu-border); margin: 6px 4px; }' +
      '#incognito-badge {' +
      '  position: fixed; bottom: 14px; right: 14px; background: #1a1420; border: 1px solid #3a2a4a; color: #b58fd8;' +
      '  padding: 8px 14px; border-radius: 20px; font-size: 11px; box-shadow: 0 6px 16px rgba(0,0,0,0.4); display: none; z-index: 10000;' +
      '}' +
      '#incognito-badge.show { display: block; }' +
      '#suggest-box {' +
      '  position: absolute; top: 46px; left: 60px; right: 60px; background: var(--menu-bg); border: 1px solid var(--menu-border);' +
      '  border-radius: 12px; box-shadow: 0 12px 32px var(--shadow); padding: 6px; display: none; z-index: 90;' +
      '}' +
      '#suggest-box.open { display: block; }' +
      '.suggest-item { padding: 9px 14px; border-radius: 8px; color: var(--menu-text); font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 10px; }' +
      '.suggest-item:hover, .suggest-item.hovered { background: rgba(0,0,0,0.04); color: var(--chrome-text); }' +
      '.suggest-icon { font-size: 12px; color: var(--chrome-subtle); width: 16px; }' +
      '#find-bar {' +
      '  position: fixed; top: 92px; right: 20px; background: var(--menu-bg); border: 1px solid var(--menu-border); border-radius: 10px;' +
      '  box-shadow: 0 8px 24px var(--shadow); padding: 8px 10px; display: none; align-items: center; gap: 8px; z-index: 10000;' +
      '}' +
      '#find-bar.show { display: flex; }' +
      '#find-input { background: var(--chrome-panel); border: 1px solid var(--menu-border); border-radius: 6px; padding: 6px 10px; color: var(--chrome-text); font-size: 12px; outline: none; width: 160px; }' +
      '#find-count { font-size: 11px; color: var(--chrome-subtle); min-width: 50px; }' +
      '.find-btn { width: 24px; height: 24px; border-radius: 6px; display: flex; align-items: center; justify-content: center; color: var(--chrome-muted); cursor: pointer; font-size: 12px; }' +
      '.find-btn:hover { background: rgba(0,0,0,0.05); color: var(--chrome-text); }' +
      '#notif-badge { position: absolute; top: 3px; right: 3px; width: 8px; height: 8px; border-radius: 50%; background: #ff5c5c; display: none; }' +
      '#notif-badge.show { display: block; }' +
      '#notif-panel {' +
      '  position: absolute; top: 46px; right: 8px; width: 280px; max-height: 360px; overflow-y: auto; background: var(--menu-bg);' +
      '  border: 1px solid var(--menu-border); border-radius: 12px; box-shadow: 0 12px 32px var(--shadow); padding: 8px; display: none; z-index: 100;' +
      '}' +
      '#notif-panel.open { display: block; }' +
      '.notif-item { padding: 10px 12px; border-radius: 8px; margin-bottom: 4px; background: var(--chrome-panel); }' +
      '.notif-title { font-size: 12px; color: var(--chrome-text); margin-bottom: 3px; }' +
      '.notif-msg { font-size: 11px; color: var(--chrome-subtle); }' +
      '#qr-share-modal { position: fixed; inset: 0; display: none; align-items: center; justify-content: center; padding: 16px; background: rgba(0,0,0,0.76); z-index: 2147483647; backdrop-filter: blur(7px); }' +
      '#qr-share-modal.open { display: flex; }' +
      '.qr-share-dialog { width: min(360px, calc(100vw - 32px)); padding: 22px; border: 1px solid rgba(151,154,255,0.28); border-radius: 16px; background: #15161b; color: #e8e8e8; box-shadow: 0 20px 60px rgba(0,0,0,0.6), 0 0 30px rgba(85,96,255,0.16); text-align: center; }' +
      '.qr-share-dialog h3 { margin: 0 0 8px; font-size: 18px; }' +
      '#qr-share-url { margin: 0 0 16px; color: #a6a6d8; font-size: 11px; line-height: 1.45; overflow-wrap: anywhere; }' +
      '#qr-share-image { display: block; width: 240px; height: 240px; max-width: 100%; margin: 0 auto 16px; padding: 10px; border-radius: 10px; background: #fff; image-rendering: pixelated; }' +
      '#qr-share-image[hidden] { display: none; }' +
      '#qr-share-status { min-height: 20px; margin: 12px 0; color: #a6a6d8; font-size: 12px; }' +
      '.qr-share-actions { display: flex; justify-content: center; gap: 10px; margin-top: 16px; }' +
      '#qr-share-modal .qr-share-actions button { min-width: 92px; border: 1px solid rgba(77,144,254,0.6); border-radius: 999px; padding: 10px 18px; cursor: pointer; color: #fff; background: rgba(21,22,27,0.9); font-weight: 600; transition: transform 0.16s ease, box-shadow 0.16s ease, background 0.16s ease; }' +
      '#qr-share-modal .qr-share-actions button:hover { transform: translateY(-1px); box-shadow: 0 0 14px rgba(77,144,254,0.28); background: rgba(77,144,254,0.16); }' +
      '#qr-share-modal .qr-share-actions #qr-share-download { border-color: transparent; color: #fff; background: linear-gradient(110deg, #7c3aed, #4d90fe); box-shadow: 0 5px 16px rgba(77,144,254,0.24); }' +
      '#qr-share-modal .qr-share-actions #qr-share-download:hover { background: linear-gradient(110deg, #6d28d9, #2878f0); box-shadow: 0 0 18px rgba(124,58,237,0.38); }' +
      '#qr-share-modal .qr-share-actions button:disabled { opacity: 0.5; cursor: default; }' +
      '#favorite-toast { position: fixed; left: 50%; bottom: 26px; transform: translate(-50%, 12px); opacity: 0 !important; pointer-events: none; padding: 10px 14px; border-radius: 9px; background: var(--chrome-panel-2); border: 1px solid var(--chrome-accent); color: var(--chrome-text); font-size: 12px; z-index: 10000; transition: opacity 0.18s ease, transform 0.18s ease; }' +
      '#favorite-toast.show { opacity: 1 !important; transform: translate(-50%, 0); }' +
      '#__sparrow_download_bar { all: initial !important; display: block !important; position: fixed !important; right: 64px !important; top: 56px !important; z-index: 2147483647 !important; width: min(380px, calc(100vw - 36px)) !important; max-height: 70vh !important; overflow: hidden !important; border: 1px solid var(--menu-border, #353535) !important; border-radius: 12px !important; background: var(--menu-bg, #1c1d20) !important; color: var(--chrome-text, #eee) !important; box-shadow: 0 10px 30px rgba(0,0,0,0.4) !important; font: 13px "Segoe UI", sans-serif !important; }' +
      '#__sparrow_download_bar[hidden] { display: none !important; }' +
      '#__sparrow_download_bar *, #__sparrow_download_bar *::before { box-sizing: border-box; }' +
      '#__sparrow_download_bar .dl-head { display: flex; align-items: center; justify-content: space-between; padding: 12px 14px; border-bottom: 1px solid var(--menu-border, #353535); }' +
      '#__sparrow_download_bar .dl-title { font-size: 16px; font-weight: 600; }' +
      '#__sparrow_download_bar button { all: unset; box-sizing: border-box; cursor: pointer; color: inherit; font: 14px "Segoe UI", sans-serif; text-align: center; border-radius: 6px; }' +
      '#__sparrow_download_bar .dl-ic { width: 30px; height: 30px; line-height: 30px; margin-left: 4px; }' +
      '#__sparrow_download_bar .dl-ic:hover, #__sparrow_download_bar .dl-b:hover { background: rgba(128,128,128,0.25); }' +
      '#__sparrow_download_bar .dl-search { display: block; width: calc(100% - 28px); margin: 8px 14px 0; padding: 7px 10px; border: 1px solid var(--menu-border, #353535); border-radius: 8px; background: transparent; color: inherit; font: 13px "Segoe UI", sans-serif; outline: none; }' +
      '#__sparrow_download_bar .dl-search[hidden] { display: none; }' +
      '#__sparrow_download_bar .dl-list { max-height: 50vh; overflow-y: auto; padding: 10px 12px 4px; }' +
      '#__sparrow_download_bar .dl-row { display: flex; align-items: center; gap: 10px; padding: 10px 12px; margin-bottom: 8px; border-radius: 10px; background: var(--chrome-panel-2, #2b2d30); }' +
      '#__sparrow_download_bar .dl-row[hidden] { display: none; }' +
      '#__sparrow_download_bar .dl-main { flex: 1; min-width: 0; }' +
      '#__sparrow_download_bar .dl-name { font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }' +
      '#__sparrow_download_bar .dl-name.done { cursor: pointer; }' +
      '#__sparrow_download_bar .dl-status { font-size: 11px; color: var(--chrome-muted, #aaa); margin: 3px 0 6px; }' +
      '#__sparrow_download_bar .dl-bar { height: 4px; border-radius: 3px; background: rgba(128,128,128,0.3); overflow: hidden; }' +
      '#__sparrow_download_bar .dl-bar i { display: block; height: 100%; width: 0; background: #4d90fe; box-shadow: 0 0 8px rgba(77,144,254,0.75); transition: width 0.25s; }' +
      '#__sparrow_download_bar .dl-row.fin .dl-bar { display: none; }' +
      '#__sparrow_download_bar .dl-b { width: 30px; height: 30px; line-height: 30px; flex: 0 0 auto; }' +
      '#__sparrow_download_bar .dl-b[hidden] { display: none; }' +
      '#__sparrow_download_bar .dl-empty { padding: 20px; text-align: center; color: var(--chrome-muted, #aaa); }' +
      '#__sparrow_download_bar .dl-foot { padding: 6px 12px 12px; }' +
      '#__sparrow_download_bar .dl-more { padding: 7px 14px; border: 1px solid #4d90fe; border-radius: 7px; background: #4d90fe; color: #fff; font-weight: 600; transition: filter 0.15s ease, transform 0.15s ease; }' +
      '#__sparrow_download_bar .dl-more:hover { filter: brightness(1.12); transform: translateY(-1px); }' +
      '#btn-download:not(.show) { display: none !important; }' +
      '#btn-download { position: relative; color: #4d90fe !important; }' +
      '#btn-download svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }' +
      '#btn-download .dl-prog { position: absolute; left: 22%; bottom: 3px; height: 2px; width: 0; border-radius: 2px; background: #4d90fe; box-shadow: 0 0 6px rgba(77,144,254,0.9); transition: width 0.3s; }' +
      '#btn-download.active { color: #4d90fe !important; background: rgba(77,144,254,0.14) !important; animation: spDlPulse 1.4s infinite; }' +
      '#btn-download.failed { color: #ff5c5c; }' +
      '#btn-download.active svg { filter: drop-shadow(0 0 4px rgba(77,144,254,0.9)); animation: spDlBounce 0.9s ease-in-out infinite; }' +
      '@keyframes spDlBounce { 0%, 100% { transform: translateY(-2px); } 50% { transform: translateY(2px); } }' +
      '@keyframes spDlPulse { 0% { box-shadow: 0 0 0 0 rgba(77,144,254,0.55); } 100% { box-shadow: 0 0 0 10px rgba(77,144,254,0); } }' +
      '.tab-chip.pinned { min-width: 42px; max-width: 42px; padding: 7px; }' +
      '.tab-chip.pinned .tab-title { display: none; }' +
      '.pin-icon { font-size: 11px; margin-right: 4px; }';
    document.head.appendChild(style);

    var root = document.createElement('div');
    root.id = '__sparrow_chrome';
    var verticalTabsEnabled = false;
    try { verticalTabsEnabled = localStorage.getItem('sparrow-vertical-tabs') === '1'; } catch (e) {}
    root.classList.toggle('vertical-tabs', verticalTabsEnabled);
    if (document.documentElement) {
      document.documentElement.classList.toggle('sparrow-vertical-tabs', verticalTabsEnabled);
    }
    var chromeMarkup =
      '<div id="window-titlebar">' +
      '  <div id="window-controls">' +
      '    <button class="window-control" id="window-minimize" title="Minimize">&#8212;</button>' +
      '    <button class="window-control" id="window-fullscreen" title="Full screen">&#9633;</button>' +
      '    <button class="window-control close" id="window-close" title="Close">&#10005;</button>' +
      '  </div>' +
      '</div>' +
      '<div id="tabstrip"><button type="button" class="vertical-tabs-toggle" id="vertical-tabs-toggle" title="Use vertical tabs" aria-label="Use vertical tabs">&#9776;</button><input id="tab-search-input-inline" type="search" placeholder="Search tabs" aria-label="Search tabs" autocomplete="off"></div>' +
      '<div id="toolbar">' +
      '  <div class="nav-btn" id="btn-back" title="Undo" aria-label="Undo">&#8630;</div>' +
      '  <div class="nav-btn" id="btn-forward" title="Redo" aria-label="Redo">&#8631;</div>' +
      '  <div class="nav-btn" id="btn-reload">&#8635;</div>' +
      '  <div class="nav-btn" id="btn-home">&#8962;</div>' +
      '  <div id="address-wrap">' +
      '    <span id="shield-icon">&#128737;</span>' +
      '    <input id="address-bar" placeholder="Search Google or type a URL">' +
      '    <button type="button" id="btn-app-lock" title="Lock app" aria-label="Lock app">' +
      '      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M17 8h-1V6a4 4 0 0 0-8 0v2H7a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-9a2 2 0 0 0-2-2ZM10 6a2 2 0 1 1 4 0v2h-4Zm2 11a2 2 0 1 1 0-4 2 2 0 0 1 0 4Z"/></svg>' +
      '    </button>' +
      '    <button type="button" id="qr-share-btn" title="Share link as QR code" aria-label="Share link as QR code">' +
      '      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 3h8v8H3V3zm2 2v4h4V5H5zm8-2h8v8h-8V3zm2 2v4h4V5h-4zM3 13h8v8H3v-8zm2 2v4h4v-4H5zm9-2h2v2h-2zm3 0h4v2h-2v2h-2zm-3 4h2v4h-4v-2h2zm4 0h2v2h2v2h-4zm-2 3h2v2h-2z"/></svg>' +
      '    </button>' +
      '    <div id="star-btn">&#9733;</div>' +
      '  </div>' +
      '  <div id="suggest-box"></div>' +
      '  <div class="nav-btn" id="btn-reader" title="Reading mode" role="button" tabindex="0" aria-label="Reading mode" aria-pressed="false">&#128214;</div>' +
      '  <div class="nav-btn" id="btn-download" title="Downloads" aria-label="Downloads"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3v12m0 0l-4-4m4 4l4-4M5 21h14"/></svg><span class="dl-prog"></span></div>' +
      '  <div class="nav-btn" id="btn-notif" style="position:relative;">' +
      '    &#128276;<div id="notif-badge"></div><div id="notif-panel"></div>' +
      '  </div>' +
      '  <div id="menu-btn">' +
      '    &#8942;' +
      '    <div id="menu-dropdown">' +
      '      <div class="menu-item" data-go="sparrow://bookmarks"><span>Bookmarks</span></div>' +
      '      <div class="menu-item" data-go="sparrow://favorites"><span>Favorites</span></div>' +
      '      <div class="menu-item" data-go="sparrow://links"><span>Link Saver</span></div>' +
      '      <div class="menu-item" data-go="sparrow://history"><span>History</span></div>' +
      '      <div class="menu-item" data-go="sparrow://downloads"><span>Downloads</span></div>' +
      '      <div class="menu-divider"></div>' +
      '      <div class="menu-item" id="menu-open-local-html"><span>Open Local HTML</span></div>' +
      '      <div class="menu-item" id="menu-incognito"><span>New Incognito Tab</span></div>' +
      '      <div class="menu-item" data-go="sparrow://lock"><span>App Lock</span></div>' +
      '      <div class="menu-item" data-go="sparrow://site-settings"><span>Site Settings</span></div>' +
      '      <div class="menu-item" data-go="sparrow://shortcuts"><span>Shortcuts</span></div>' +
      '      <div class="menu-item" data-go="sparrow://about"><span>About SPARROW</span></div>' +
      '      <div class="menu-item" data-go="sparrow://cookies"><span>Cookies</span></div>' +
      '      <div class="menu-item" data-go="sparrow://updates"><span>About &amp; Updates</span></div>' +
      '      <div class="menu-item" data-go="sparrow://backup"><span>Backup &amp; Restore</span></div>' +
      '      <div class="menu-item" data-go="sparrow://customize"><span>Customize</span></div>' +
      '      <div class="menu-item" id="menu-print"><span>Print</span></div>' +
      '      <div class="menu-divider"></div>' +
      '      <div class="menu-item" data-go="sparrow://settings"><span>Settings</span></div>' +
      '    </div>' +
      '  </div>' +
      '</div>' +
      '<div id="find-bar">' +
      '  <input id="find-input" placeholder="Find in page">' +
      '  <span id="find-count">0/0</span>' +
      '  <div class="find-btn" id="find-prev">&#8593;</div>' +
      '  <div class="find-btn" id="find-next">&#8595;</div>' +
      '  <div class="find-btn" id="find-close">&times;</div>' +
      '</div>' +
      '<div id="incognito-badge">Incognito Mode Active</div>';
    sparrowSetHTML(root, chromeMarkup);
    var initialLayoutToggle = root.querySelector('#vertical-tabs-toggle');
    initialLayoutToggle.title = verticalTabsEnabled ? 'Use horizontal tabs' : 'Use vertical tabs';
    initialLayoutToggle.setAttribute('aria-label', initialLayoutToggle.title);
    initialLayoutToggle.setAttribute('aria-pressed', verticalTabsEnabled ? 'true' : 'false');
    var chromeMarkup2 = '<div id="favorite-toast" role="status">Saved to Favorites</div>';
    chromeMarkup2 +=
      '<div id="qr-share-modal" role="dialog" aria-modal="true" aria-labelledby="qr-share-heading">' +
      '  <div class="qr-share-dialog">' +
      '    <h3 id="qr-share-heading">Share this link</h3>' +
      '    <p id="qr-share-url"></p>' +
      '    <img id="qr-share-image" alt="QR code for this page link" hidden>' +
      '    <p id="qr-share-status" role="status">Creating QR code...</p>' +
      '    <div class="qr-share-actions">' +
      '      <button type="button" id="qr-share-download" disabled>Download</button>' +
      '      <button type="button" id="qr-share-close">Close</button>' +
      '    </div>' +
      '  </div>' +
      '</div>';
    sparrowAppendHTML(root, chromeMarkup2);
    document.documentElement.appendChild(root);
    var downloadBar = document.createElement('div');
    downloadBar.id = '__sparrow_download_bar';
    downloadBar.setAttribute('role', 'region');
    downloadBar.setAttribute('aria-label', 'Downloads');
    downloadBar.hidden = true;
    sparrowSetHTML(downloadBar,
      '<div class="dl-head"><span class="dl-title">Downloads</span><span>' +
      '<button type="button" class="dl-ic" data-act="folder" title="Open downloads folder">&#128193;</button>' +
      '<button type="button" class="dl-ic" data-act="search" title="Search downloads">&#128269;</button>' +
      '<button type="button" class="dl-ic" data-act="clear" title="Clear finished">&#8943;</button>' +
      '<button type="button" class="dl-ic" data-act="close" title="Close">&#10005;</button></span></div>' +
      '<input type="text" class="dl-search" placeholder="Search downloads" hidden>' +
      '<div class="dl-list"></div>' +
      '<div class="dl-foot"><button type="button" class="dl-more" data-act="more">See more</button></div>');
    document.documentElement.appendChild(downloadBar);
    var dlSearch = downloadBar.querySelector('.dl-search');
    function dlFmt(n) {
      n = Number(n) || 0;
      var u = ['B', 'KB', 'MB', 'GB', 'TB'], i = 0;
      while (n >= 1024 && i < 4) { n /= 1024; i++; }
      return (i === 0 ? n : n.toFixed(1)) + ' ' + u[i];
    }
    function dlApplyFilter() {
      var q = (dlSearch.value || '').toLowerCase();
      downloadBar.querySelectorAll('.dl-row').forEach(function (r) {
        r.hidden = !!q && (r.getAttribute('data-name') || '').toLowerCase().indexOf(q) === -1;
      });
    }
    dlSearch.oninput = dlApplyFilter;
    controller.renderDownloads = function (downloads, stats) {
      var list = downloadBar.querySelector('.dl-list');
      var st = {};
      stats.forEach(function (x) { st[x.id] = x; });
      var seen = {};
      downloads.slice().sort(function (a, b) { return a.ID - b.ID; }).forEach(function (d) {
        seen[d.ID] = true;
        var row = list.querySelector('[data-id="' + d.ID + '"]');
        if (!row) {
          row = document.createElement('div');
          row.className = 'dl-row';
          row.setAttribute('data-id', d.ID);
          var main = document.createElement('div'); main.className = 'dl-main';
          var nm = document.createElement('div'); nm.className = 'dl-name';
          var stx = document.createElement('div'); stx.className = 'dl-status';
          var bar = document.createElement('div'); bar.className = 'dl-bar';
          bar.appendChild(document.createElement('i'));
          main.appendChild(nm); main.appendChild(stx); main.appendChild(bar);
          var b1 = document.createElement('button'); b1.type = 'button'; b1.className = 'dl-b b1';
          var b2 = document.createElement('button'); b2.type = 'button'; b2.className = 'dl-b b2';
          row.appendChild(main); row.appendChild(b1); row.appendChild(b2);
          list.insertBefore(row, list.firstChild);
        }
        var s = st[d.ID] || {};
        var name = d.FileName || 'Download';
        var nameEl = row.querySelector('.dl-name');
        var statusEl = row.querySelector('.dl-status');
        var b1 = row.querySelector('.b1'), b2 = row.querySelector('.b2');
        row.setAttribute('data-name', name);
        nameEl.textContent = name; nameEl.title = name;
        row.querySelector('.dl-bar i').style.width = Math.max(0, Math.min(100, Number(d.Progress) || 0)) + '%';
        var active = d.Status === 'downloading' || d.Status === 'queued' || d.Status === 'paused';
        var fin = !active;
        row.classList.toggle('fin', fin);
        nameEl.classList.toggle('done', d.Status === 'completed');
        nameEl.setAttribute('data-act', d.Status === 'completed' ? 'open' : '');
        if (active) {
          var sizeTxt = s.total ? dlFmt(s.received) + ' of ' + dlFmt(s.total) : dlFmt(s.received || 0);
          statusEl.textContent = d.Status === 'queued' ? 'Queued' :
            (d.Status === 'paused' || s.paused ? sizeTxt + ', Paused' : dlFmt(s.speed || 0) + '/s - ' + sizeTxt);
          b1.hidden = false; b1.textContent = s.paused ? '\u25B6' : '\u23F8'; b1.title = s.paused ? 'Resume' : 'Pause'; b1.setAttribute('data-act', 'toggle');
          b2.hidden = false; b2.textContent = '\u2715'; b2.title = 'Cancel'; b2.setAttribute('data-act', 'cancel');
        } else if (d.Status === 'completed') {
          statusEl.textContent = (s.total ? dlFmt(s.total) + ' - ' : '') + 'Completed';
          b1.hidden = false; b1.textContent = '\u2197'; b1.title = 'Open file'; b1.setAttribute('data-act', 'open');
          b2.hidden = false; b2.textContent = '\uD83D\uDCC1'; b2.title = 'Show in folder'; b2.setAttribute('data-act', 'reveal');
        } else {
          statusEl.textContent = d.Status === 'failed' ? 'Failed' : 'Cancelled';
          b1.hidden = true; b2.hidden = true;
        }
      });
      list.querySelectorAll('.dl-row').forEach(function (r) {
        if (!seen[r.getAttribute('data-id')]) r.remove();
      });
      dlApplyFilter();
    };
    downloadBar.addEventListener('click', function (event) {
      event.stopPropagation();
      var t = event.target.closest ? event.target.closest('[data-act]') : null;
      if (!t) return;
      var act = t.getAttribute('data-act');
      var rowEl = t.closest('[data-id]');
      var id = rowEl ? Number(rowEl.getAttribute('data-id')) : 0;
      if (act === 'close') downloadBar.hidden = true;
      else if (act === 'folder' && window.openDownloadsFolder) window.openDownloadsFolder();
      else if (act === 'search') { dlSearch.hidden = !dlSearch.hidden; if (!dlSearch.hidden) dlSearch.focus(); else { dlSearch.value = ''; dlApplyFilter(); } }
      else if (act === 'clear' && window.clearFinishedDownloads) window.clearFinishedDownloads();
      else if (act === 'more') {
        downloadBar.hidden = true;
        if (typeof window.go === 'function') window.go('sparrow://downloads');
        else if (typeof window.realNavigate === 'function') window.realNavigate('sparrow://downloads');
      }
      else if (act === 'toggle' && window.toggleDownloadPause) window.toggleDownloadPause(id);
      else if (act === 'cancel' && window.cancelDownload) window.cancelDownload(id);
      else if (act === 'open' && window.openDownloadedFile) window.openDownloadedFile(id);
      else if (act === 'reveal' && window.showDownloadInFolder) window.showDownloadInFolder(id);
    });
    var qrShareModal = document.getElementById('qr-share-modal');
    if (qrShareModal) {
      document.documentElement.appendChild(qrShareModal);
    }

    var activeTabId = null;
    var findDebounce = null;

    window.go = async function (input) {
      if (!input || String(input).trim() === '') return;
      if (typeof window.realNavigate !== 'function') return;
      try {
        return await window.realNavigate(String(input).trim());
      } catch (e) {
        console.error('SPARROW navigation failed:', e);
        return null;
      }
    };

    // webview_go does not expose native SetZoom in this version, so apply
    // page zoom with Chromium's CSS zoom while explicitly excluding SPARROW's
    // own browser chrome.
    window.__sparrowSetZoom = function (level) {
      var value = Number(level);
      if (!isFinite(value)) value = 1;
      value = Math.max(0.5, Math.min(3, value));

      var style = document.getElementById('__sparrow_zoom_style');
      if (!style) {
        style = document.createElement('style');
        style.id = '__sparrow_zoom_style';
        style.textContent = 'html:not(.sparrow-image-document) body > *:not(#__sparrow_chrome) { zoom: var(--sparrow-page-zoom) !important; }';
        if (document.head) document.head.appendChild(style);
      }
      document.documentElement.style.setProperty('--sparrow-page-zoom', String(value));
      window.__sparrowZoom = value;
      return value;
    };

    window.goBack = async function () {
      if (traversePageHistory(-1)) return;
      var state = await window.goBackNav();
      if (state) updateNavButtons(state);
    };
    window.goForward = async function () {
      if (traversePageHistory(1)) return;
      var state = await window.goForwardNav();
      if (state) updateNavButtons(state);
    };
    window.reload = function () { location.reload(); };
    window.goHome = function () { window.go('sparrow://home'); };

    // A page can remove our chrome without destroying the document. If that
    // happens, remove handlers from the previous chrome instance before the
    // new instance installs its own handlers.
    async function navigateFromBar() {
      var val = document.getElementById('address-bar').value;
      if (val.trim() === '') return;
      await window.go(val);
    }

    async function createNewTab() {
      await window.openNewTab();
    }

    async function switchToTab(id) {
      await window.switchTab(id);
    }

    function filterTabsByName() {
      var input = document.getElementById('tab-search-input-inline');
      var root = document.getElementById('__sparrow_chrome');
      var query = input && root && root.classList.contains('vertical-tabs') ? input.value.trim().toLowerCase() : '';
      var tabs = document.querySelectorAll('#tabstrip .tab-chip');
      var visibleCount = 0;
      Array.prototype.forEach.call(tabs, function (chip) {
        var title = chip.querySelector('.tab-title');
        var matches = !query || (title && title.textContent.toLocaleLowerCase().indexOf(query) !== -1);
        chip.style.display = matches ? '' : 'none';
        if (matches) visibleCount++;
      });

      var empty = document.getElementById('tab-search-empty');
      if (query && visibleCount === 0) {
        if (!empty) {
          empty = document.createElement('div');
          empty.id = 'tab-search-empty';
          empty.className = 'tab-search-empty';
          empty.textContent = 'No matching tabs';
          document.getElementById('tabstrip').appendChild(empty);
        }
      } else if (empty) {
        empty.remove();
      }
    }

    async function closeThisTab(id) {
      try {
        await window.closeTab(id);
      } catch (error) {
        console.error('SPARROW could not close tab:', error);
      }
    }

    function closeAllTabs() {
      window.closeAllTabs();
    }

    async function renderTabs(currentTabs) {
      var tabs = currentTabs || await window.getAllTabsUI();
      var strip = document.getElementById('tabstrip');
      var root = document.getElementById('__sparrow_chrome');
      var layoutToggle = document.getElementById('vertical-tabs-toggle');
      var searchInput = document.getElementById('tab-search-input-inline');
      var existingNewBtn = document.getElementById('new-tab-btn');
      if (existingNewBtn) existingNewBtn.remove();
      sparrowClear(strip);
      if (layoutToggle) strip.appendChild(layoutToggle);
      if (searchInput) strip.appendChild(searchInput);

      tabs.forEach(function (t) {
        var chip = document.createElement('div');
        chip.className = 'tab-chip' + (t.ID === activeTabId ? ' active' : '');
        chip.setAttribute('data-tab-id', String(t.ID));
        chip.onclick = function () { switchToTab(t.ID); };

        if (t.Favicon) {
          var favicon = document.createElement('img');
          favicon.className = 'tab-favicon';
          favicon.alt = '';
          favicon.draggable = false;
          favicon.onerror = function () { this.remove(); };
          favicon.src = t.Favicon;
          chip.appendChild(favicon);
        }

        var title = document.createElement('div');
        title.className = 'tab-title';
        title.innerText = t.Title || 'New Tab';
        chip.appendChild(title);

        var closeBtn = document.createElement('button');
        closeBtn.type = 'button';
        closeBtn.className = 'tab-close';
        closeBtn.textContent = '\u00d7';
        closeBtn.setAttribute('aria-label', 'Close ' + (t.Title || 'New Tab'));
        closeBtn.onclick = function (e) {
          e.preventDefault();
          e.stopPropagation();
          closeThisTab(t.ID);
        };
        chip.appendChild(closeBtn);

        strip.appendChild(chip);
      });

      var newBtn = document.createElement('div');
      newBtn.id = 'new-tab-btn';
      newBtn.textContent = '+';
      newBtn.onclick = createNewTab;
      (root.classList.contains('vertical-tabs') ? strip : root).appendChild(newBtn);
      filterTabsByName();

      var activeChip = strip.querySelector('.tab-chip.active');
      if (activeChip) {
        var stripBounds = strip.getBoundingClientRect();
        var chipBounds = activeChip.getBoundingClientRect();
        if (document.getElementById('__sparrow_chrome').classList.contains('vertical-tabs')) {
          if (chipBounds.top < stripBounds.top) strip.scrollTop -= stripBounds.top - chipBounds.top;
          else if (chipBounds.bottom > stripBounds.bottom) strip.scrollTop += chipBounds.bottom - stripBounds.bottom;
        } else {
          var tabChips = strip.querySelectorAll('.tab-chip');
          var lastTabChip = tabChips[tabChips.length - 1];
          if (activeChip === lastTabChip) strip.scrollLeft = strip.scrollWidth;
          else if (chipBounds.left < stripBounds.left) strip.scrollLeft -= stripBounds.left - chipBounds.left;
          else if (chipBounds.right > stripBounds.right) strip.scrollLeft += chipBounds.right - stripBounds.right;
        }
      }
    }

    function updateNavButtons(state) {
      controller.navigationState = state;
      updateHistoryButtons();
    }

    function traversePageHistory(delta) {
      var target = controller.pageHistory.index + delta;
      if (target < 0 || target >= controller.pageHistory.entries.length) return false;
      window.history.go(delta);
      return true;
    }

    function applyShellTheme(theme) {
      var t = theme === 'light' ? 'light' : 'dark';
      try { localStorage.setItem('sparrow-shell-theme', t); } catch (e) {}
      document.documentElement.setAttribute('data-theme', t);
    }

    window.applyCustomShellColor = function (color) {
      if (!/^#[0-9a-fA-F]{6}$/.test(color || '')) return;
      var light = document.documentElement.getAttribute('data-theme') === 'light';
      var colorMix = function (amount, neutral) {
        return 'color-mix(in srgb, ' + color + ' ' + amount + '%, ' + neutral + ')';
      };
      var channel = function (offset) {
        var value = parseInt(color.slice(offset, offset + 2), 16) / 255;
        return value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
      };
      var luminance = 0.2126 * channel(1) + 0.7152 * channel(3) + 0.0722 * channel(5);
      var surfaces = light ? {
        '--shell-bg': colorMix(25, '#ffffff'),
        '--window-bar': colorMix(20, '#ffffff'),
        '--tabstrip-bg': colorMix(20, '#f0f2f4'),
        '--toolbar-bg': colorMix(16, '#ffffff'),
        '--chrome-panel': colorMix(16, '#ffffff'),
        '--chrome-panel-2': colorMix(18, '#ffffff'),
        '--chip-bg': colorMix(18, '#f1f3f4'),
        '--chip-active': colorMix(15, '#ffffff'),
        '--menu-bg': colorMix(16, '#ffffff')
      } : {
        '--shell-bg': colorMix(45, '#0b0b0d'),
        '--window-bar': colorMix(38, '#0b0b0d'),
        '--tabstrip-bg': colorMix(32, '#0f0f11'),
        '--toolbar-bg': colorMix(24, '#1e1f22'),
        '--chrome-panel': colorMix(22, '#17181a'),
        '--chrome-panel-2': colorMix(24, '#1c1d20'),
        '--chip-bg': colorMix(22, '#1a1b1e'),
        '--chip-active': colorMix(24, '#1e1f22'),
        '--menu-bg': colorMix(22, '#1c1d20')
      };
      surfaces['--chrome-accent'] = color;
      surfaces['--new-tab-text'] = luminance > 0.179 ? '#111111' : '#ffffff';
      Object.keys(surfaces).forEach(function (name) {
        document.documentElement.style.setProperty(name, surfaces[name]);
      });
    };

    function applySearchPlaceholder(engine) {
      var name = (engine || 'google').toLowerCase();
      if (name === 'bing') name = 'Bing';
      else if (name === 'brave') name = 'Brave Search';
      else if (name === 'duckduckgo') name = 'DuckDuckGo';
      else if (name === 'yahoo') name = 'Yahoo';
      else name = 'Google';
      var label = 'Search ' + name + ' or type a URL';
      var addressBar = document.getElementById('address-bar');
      if (addressBar) addressBar.placeholder = label;
      var homeBar = document.getElementById('searchBar');
      if (homeBar) homeBar.placeholder = label;
    }

    window.updateSearchPlaceholder = applySearchPlaceholder;

    async function refreshStar(url, isSaved) {
      var starBtn = document.getElementById('star-btn');
      if (!url || url.indexOf('sparrow://') === 0) {
        starBtn.className = '';
        return;
      }
      if (typeof isSaved !== 'boolean') isSaved = await window.isFavorite(url);
      starBtn.className = isSaved ? 'saved' : '';
    }

    function currentPageURL(state) {
      var u = (state && state.url) || '';
      if (/^https?:/i.test(window.location.href)) u = window.location.href;
      return u;
    }

    async function openQRShareModal() {
      var state = await window.getNavState();
      var url = currentPageURL(state);
      if (!/^https?:\/\/\S+/i.test(url)) {
        showFavoriteSaved('Only web links can be shared as QR codes');
        return;
      }

      var modal = document.getElementById('qr-share-modal');
      var image = document.getElementById('qr-share-image');
      var download = document.getElementById('qr-share-download');
      var status = document.getElementById('qr-share-status');
      document.getElementById('qr-share-url').textContent = url;
      image.hidden = true;
      image.removeAttribute('src');
      download.disabled = true;
      status.textContent = 'Creating QR code...';
      modal.classList.add('open');
      document.getElementById('qr-share-close').focus();

      try {
        var imageData = await window.generateQRCode(url);
        if (typeof imageData !== 'string' || imageData.indexOf('data:image/png;base64,') !== 0) {
          throw new Error('The QR code generator returned invalid image data.');
        }
        image.src = imageData;
        image.hidden = false;
        status.textContent = 'Scan to open this link.';
        download.disabled = false;
      } catch (error) {
        console.error('SPARROW could not create a QR code:', error);
        status.textContent = 'Could not create the QR code. Close and try again.';
      }
    }

    function closeQRShareModal() {
      document.getElementById('qr-share-modal').classList.remove('open');
      document.getElementById('qr-share-image').removeAttribute('src');
      document.getElementById('qr-share-btn').focus();
    }

    function downloadQRShare() {
      var image = document.getElementById('qr-share-image');
      if (image.hidden || !image.src) return;
      var link = document.createElement('a');
      link.href = image.src;
      link.download = 'sparrow-link-qr.png';
      document.body.appendChild(link);
      link.click();
      link.remove();
    }

    async function toggleCurrentFavorite() {
      var state = await window.getNavState();
      var url = currentPageURL(state);
      if (!/^https?:\/\/\S+/i.test(url)) {
        showFavoriteSaved('Only web links can be saved to Favorites');
        return;
      }
      var starBtn = document.getElementById('star-btn');
      try {
        if (await window.isFavorite(url)) {
          await window.removeFavoriteByURL(url);
          starBtn.className = '';
          showFavoriteSaved('Removed from Favorites');
        } else {
          var saved = await window.addFavorite(document.title || url, url);
          if (!saved) {
            showFavoriteSaved('Could not save favorite');
            return;
          }
          starBtn.className = 'saved';
          showFavoriteSaved('Saved to Favorites');
        }
        var favorites = await window.getFavorites();
        try { localStorage.setItem('sparrow-favorites', JSON.stringify(favorites)); } catch (error) {}
      } catch (error) {
        console.error('SPARROW could not update favorites:', error);
        showFavoriteSaved('Could not update Favorites');
      }
    }

    function showFavoriteSaved(msg) {
      var toast = document.getElementById('favorite-toast');
      toast.textContent = msg || 'Saved to Favorites';
      if (controller.favoriteToastTimer) clearTimeout(controller.favoriteToastTimer);
      toast.classList.add('show');
      controller.favoriteToastTimer = setTimeout(function () {
        toast.classList.remove('show');
        controller.favoriteToastTimer = null;
      }, 2000);
    }

    function pageFaviconURL() {
      if (window.location.protocol !== 'http:' && window.location.protocol !== 'https:') return '';
      var icon = document.querySelector('link[rel~="icon"]');
      return icon && icon.href ? icon.href : new URL('/favicon.ico', window.location.href).href;
    }

    var lastTitleSig = '';
    async function updateTabTitleFromPage() {
      if (activeTabId === null) return;
      var title = (document.title || '').trim();
      var favicon = pageFaviconURL();
      var signature = activeTabId + '\u0001' + title + '\u0001' + window.location.href + '\u0001' + favicon;
      if (signature === lastTitleSig) return;
      lastTitleSig = signature;
      await window.updateTabTitle(activeTabId, title, window.location.href, favicon);
      var chip = document.querySelector('.tab-chip[data-tab-id="' + activeTabId + '"]');
      if (!chip) return;
      var titleElement = chip.querySelector('.tab-title');
      if (titleElement) titleElement.textContent = title || 'New Tab';
      var faviconElement = chip.querySelector('.tab-favicon');
      if (favicon && faviconElement) {
        faviconElement.src = favicon;
      } else if (favicon && !faviconElement) {
        faviconElement = document.createElement('img');
        faviconElement.className = 'tab-favicon';
        faviconElement.alt = '';
        faviconElement.draggable = false;
        faviconElement.onerror = function () { this.remove(); };
        faviconElement.src = favicon;
        chip.insertBefore(faviconElement, chip.firstChild);
      } else if (!favicon && faviconElement) {
        faviconElement.remove();
      }
    }

    var titleUpdateTimer = null;
    function scheduleTitleUpdate() {
      if (titleUpdateTimer) return;
      titleUpdateTimer = setTimeout(function () {
        titleUpdateTimer = null;
        updateTabTitleFromPage();
      }, 400);
    }

    function registerPageTitleListener() {
      if (controller.titleListenerRegistered) return;
      controller.titleListenerRegistered = true;

      if (document.head && typeof MutationObserver !== 'undefined') {
        if (controller.titleObserver) { try { controller.titleObserver.disconnect(); } catch (e) {} }
        controller.titleObserver = new MutationObserver(scheduleTitleUpdate);
        controller.titleObserver.observe(document.head, {
          childList: true,
          characterData: true,
          subtree: true,
          attributes: true,
          attributeFilter: ['href', 'rel']
        });
      }

      window.addEventListener('load', function () {
        updateTabTitleFromPage();
      }, { once: true });
      if (document.readyState === 'complete') updateTabTitleFromPage();
    }

    async function syncPageLocation() {
      if (activeTabId === null) return;
      var location = window.location.href;
      if (location.indexOf('http://') !== 0 && location.indexOf('https://') !== 0) return;
      if (controller.lastLocation === location && controller.lastLocationTabId === activeTabId) return;
      controller.lastLocation = location;
      controller.lastLocationTabId = activeTabId;
      await window.syncTabLocation(activeTabId, location, controller.pageHistory.entries.length > 1);
      var addressBar = document.getElementById('address-bar');
      if (addressBar && document.activeElement !== addressBar) {
        addressBar.value = location;
      }
    }

    async function refreshChrome() {
      var initial = await Promise.all([
        window.getNavState(),
        window.getSettings(),
        window.getCustomTheme()
      ]);
      var state = initial[0];
      var settings = initial[1];
      var customTheme = initial[2];
      activeTabId = state.tabId;
      updateNavButtons(state);
      var details = await Promise.all([
        window.getZoom(activeTabId).catch(function (error) {
          console.error('SPARROW could not load page zoom:', error);
          return 1;
        }),
        window.getAllTabsUI(),
        state.url && state.url.indexOf('sparrow://') !== 0
          ? window.isFavorite(state.url) : Promise.resolve(false),
        window.isIncognitoActive(),
        window.getUnreadNotifCount()
      ]);
      window.__sparrowSetZoom(details[0]);
      await renderTabs(details[1]);
      setReaderButtonState(await window.isReaderModeActive(activeTabId));

      var theme = (settings && settings.theme) || 'dark';
      applyShellTheme(theme);
      window.applyCustomShellColor(customTheme.accentColor);

      var name = (settings && settings.searchEngine) || 'google';
      applySearchPlaceholder(name);
      var addressBar = document.getElementById('address-bar');
      if (document.activeElement !== addressBar) {
        addressBar.value = state.url || '';
      }
      await refreshStar(state.url, details[2]);
      registerPageTitleListener();
      if (document.readyState === 'complete') await updateTabTitleFromPage();

      var incognitoOn = details[3];
      document.getElementById('incognito-badge').className = incognitoOn ? 'show' : '';
      var menuIncognito = document.getElementById('menu-incognito');
      if (menuIncognito) {
        sparrowSetHTML(menuIncognito, incognitoOn ? '<span>Disable Incognito</span>' : '<span>New Incognito Tab</span>');
      }

      refreshNotifBadge(details[4]);
      window.injectContextMenu();
      window.injectFindInPage();
    }

    window.__sparrowRefreshChrome = refreshChrome;

    syncPageLocation();
    if (!controller.locationInterval) {
      controller.locationInterval = setInterval(syncPageLocation, 350);
    }

    async function refreshNotifBadge(count) {
      if (typeof count !== 'number') count = await window.getUnreadNotifCount();
      document.getElementById('notif-badge').className = count > 0 ? 'show' : '';
    }

    async function toggleNotifPanel(e) {
      e.stopPropagation();
      var panel = document.getElementById('notif-panel');
      if (panel.classList.contains('open')) {
        panel.classList.remove('open');
        return;
      }
      var items = await window.getNotifications();
      sparrowSetHTML(panel, items.length === 0
        ? '<div style="padding:16px;color:#666;font-size:12px;text-align:center;">No notifications</div>'
        : items.map(function (n) {
            return '<div class="notif-item"><div class="notif-title">' + n.Title +
              '</div><div class="notif-msg">' + n.Message + '</div></div>';
          }).join(''));
      panel.classList.add('open');
      await window.markAllNotifsRead();
      refreshNotifBadge();
    }

    var suggestSelected = -1;
    var currentSuggestions = [];

    async function onAddressInput() {
      var val = document.getElementById('address-bar').value;
      var box = document.getElementById('suggest-box');
      if (val.trim() === '') {
        box.classList.remove('open');
        return;
      }
      currentSuggestions = await window.getSuggestions(val) || [];
      suggestSelected = -1;
      renderSuggestions();
    }

    function renderSuggestions() {
      var box = document.getElementById('suggest-box');
      if (currentSuggestions.length === 0) {
        box.classList.remove('open');
        return;
      }
      sparrowSetHTML(box, currentSuggestions.map(function (s, i) {
        var icon = s.source === 'bookmark' ? '&#9733;' : s.source === 'history' ? '&#8635;' : '&#128269;';
        return '<div class="suggest-item' + (i === suggestSelected ? ' hovered' : '') + '" data-idx="' + i + '">' +
          '<span class="suggest-icon">' + icon + '</span><span>' + s.text + '</span></div>';
      }).join(''));
      Array.prototype.forEach.call(box.querySelectorAll('.suggest-item'), function (el) {
        el.onclick = function () { selectSuggestion(parseInt(el.getAttribute('data-idx'), 10)); };
      });
      box.classList.add('open');
    }

    function selectSuggestion(i) {
      var s = currentSuggestions[i];
      if (!s || !s.url) return;
      document.getElementById('address-bar').value = s.url;
      document.getElementById('suggest-box').classList.remove('open');
      window.go(s.url);
    }

    function closeSuggestions() {
      document.getElementById('suggest-box').classList.remove('open');
    }

    async function showSiteInfo(e) {
      e.stopPropagation();
      var url = document.getElementById('address-bar').value;
      if (!url || url.indexOf('sparrow://') === 0) return;
      var info = await window.getSiteInfo(url);
      alert(
        'Domain: ' + info.domain + '\\n' +
        'Secure (HTTPS): ' + (info.isSecure ? 'Yes' : 'No') + '\\n' +
        'Ads Blocked: ' + (info.adsBlocked ? 'On' : 'Off') + '\\n' +
        'Trackers Blocked: ' + (info.trackersBlocked ? 'On' : 'Off') + '\\n' +
        'Total Threats Blocked This Session: ' + info.threatsBlockedTotal
      );
    }

    function onFindInput() {
      clearTimeout(findDebounce);
      findDebounce = setTimeout(function () {
        var query = document.getElementById('find-input').value;
        try {
          var count = window.__sparrowFind.search(query);
          document.getElementById('find-count').innerText = count > 0 ? '1/' + count : '0/0';
        } catch (e) {}
      }, 200);
    }

    function findNext() {
      try {
        var idx = window.__sparrowFind.next();
        var total = window.__sparrowFind.matches.length;
        document.getElementById('find-count').innerText = (idx + 1) + '/' + total;
      } catch (e) {}
    }

    function findPrev() {
      try {
        var idx = window.__sparrowFind.prev();
        var total = window.__sparrowFind.matches.length;
        document.getElementById('find-count').innerText = (idx + 1) + '/' + total;
      } catch (e) {}
    }

    function openFindBar() {
      document.getElementById('find-bar').classList.add('show');
      document.getElementById('find-input').focus();
    }

    function closeFindBar() {
      try { window.__sparrowFind.clear(); } catch (e) {}
      document.getElementById('find-bar').classList.remove('show');
      document.getElementById('find-input').value = '';
      document.getElementById('find-count').innerText = '0/0';
    }

    function printCurrentPage() {
      try { window.print(); } catch (e) { alert('Cannot print this page'); }
    }

    function minimizeWithVideoPiP() {
      var activePiP = document.pictureInPictureElement;
      var video = Array.prototype.find.call(document.querySelectorAll('video'), function (candidate) {
        return !candidate.paused && !candidate.ended && candidate.readyState >= 2 &&
          !candidate.disablePictureInPicture;
      });
      if (!video || activePiP) {
        window.minimizeWindow();
        return;
      }

      try {
        Promise.resolve(video.requestPictureInPicture()).then(function () {
          window.minimizeWindow();
        }).catch(function (error) {
          console.error('SPARROW could not open video picture-in-picture:', error);
          window.minimizeWindow();
        });
      } catch (error) {
        console.error('SPARROW could not request video picture-in-picture:', error);
        window.minimizeWindow();
      }
    }

    async function switchToRelativeTab(delta) {
      var tabs = window.getAllTabsUI ? await window.getAllTabsUI() : [];
      if (!tabs || tabs.length === 0) return;
      var idx = -1;
      for (var i = 0; i < tabs.length; i++) {
        if (tabs[i].ID === activeTabId) { idx = i; break; }
      }
      if (idx === -1) idx = 0;
      var nextIdx = (idx + delta + tabs.length) % tabs.length;
      await window.switchTab(tabs[nextIdx].ID);
      setTimeout(function () { refreshChrome(); }, 50);
    }

    async function reopenClosedTab() {
      var item = await window.reopenLastClosed();
      if (!item || !item.URL) return;
      await window.openNewTab();
      await window.go(item.URL);
    }

    function toggleMenu(e) {
      e.stopPropagation();
      var panel = document.getElementById('menu-dropdown');
      var willOpen = !panel.classList.contains('open');
      panel.classList.toggle('open', willOpen);
    }

    function setVerticalTabs(enabled) {
      var root = document.getElementById('__sparrow_chrome');
      if (!root) return;
      root.classList.toggle('vertical-tabs', enabled);
      document.documentElement.classList.toggle('sparrow-vertical-tabs', enabled);
      var newTabButton = document.getElementById('new-tab-btn');
      var strip = document.getElementById('tabstrip');
      if (newTabButton && strip) (enabled ? strip : root).appendChild(newTabButton);
      filterTabsByName();
      var toggle = document.getElementById('vertical-tabs-toggle');
      if (toggle) {
        toggle.title = enabled ? 'Use horizontal tabs' : 'Use vertical tabs';
        toggle.setAttribute('aria-label', toggle.title);
        toggle.setAttribute('aria-pressed', enabled ? 'true' : 'false');
      }
      try {
        localStorage.setItem('sparrow-vertical-tabs', enabled ? '1' : '0');
      } catch (error) {
        console.error('SPARROW could not save tab layout:', error);
      }
      repairChromeStyle();
    }

    function setReaderButtonState(enabled) {
      var button = document.getElementById('btn-reader');
      if (button) {
        button.classList.toggle('active', enabled);
        button.title = enabled ? 'Exit reading mode' : 'Reading mode';
        button.setAttribute('aria-label', button.title);
        button.setAttribute('aria-pressed', enabled ? 'true' : 'false');
      }
    }

    async function toggleReader() {
      var enabled = await window.toggleReaderMode(activeTabId);
      setReaderButtonState(enabled);
    }

    async function requestPiPForTab() {
      await window.requestPiP(activeTabId);
    }

    async function openIncognito() {
      await window.enableIncognito();
      await createNewTab();
    }

    // ---- wire up events ----
    document.getElementById('btn-back').onclick = window.goBack;
    document.getElementById('btn-forward').onclick = window.goForward;
    document.getElementById('btn-reload').onclick = window.reload;
    document.getElementById('btn-home').onclick = window.goHome;
    document.getElementById('shield-icon').onclick = showSiteInfo;
    document.getElementById('btn-app-lock').onclick = async function (e) {
      e.stopPropagation();
      try {
        var status = await window.getAppLockStatus();
        if (!status.configured) {
          await window.go('sparrow://lock');
          return;
        }
        await window.lockApp();
      } catch (error) {
        console.error('SPARROW could not lock the app:', error);
      }
    };
    document.getElementById('qr-share-btn').onclick = openQRShareModal;
    document.getElementById('qr-share-close').onclick = closeQRShareModal;
    document.getElementById('qr-share-download').onclick = downloadQRShare;
    document.getElementById('qr-share-modal').onclick = function (e) {
      if (e.target.id === 'qr-share-modal') closeQRShareModal();
    };
    document.getElementById('tab-search-input-inline').oninput = filterTabsByName;
    document.getElementById('star-btn').onclick = toggleCurrentFavorite;
    document.getElementById('btn-reader').onclick = toggleReader;
    document.getElementById('btn-notif').onclick = toggleNotifPanel;
    document.getElementById('btn-download').onclick = function (e) {
      e.stopPropagation();
      downloadBar.hidden = !downloadBar.hidden;
    };
    document.getElementById('window-minimize').onclick = function (e) { e.stopPropagation(); minimizeWithVideoPiP(); };
    document.getElementById('window-fullscreen').onclick = function (e) { e.stopPropagation(); window.toggleFullscreenWindow(); };
    document.getElementById('window-close').onclick = function (e) { e.stopPropagation(); window.closeWindow(); };
    document.getElementById('vertical-tabs-toggle').onclick = function (e) {
      e.stopPropagation();
      setVerticalTabs(!document.getElementById('__sparrow_chrome').classList.contains('vertical-tabs'));
    };
    document.getElementById('window-titlebar').onmousedown = function (e) {
      if (e.target.id === 'window-titlebar' || e.target.id === 'window-title') {
        window.beginWindowDrag(e.clientX, e.clientY);
      }
    };
    document.getElementById('tabstrip').onmousedown = function (e) {
      if (e.target.id === 'tabstrip') window.beginWindowDrag(e.clientX, e.clientY);
    };
    document.getElementById('tabstrip').onwheel = function (e) {
      var strip = e.currentTarget;
      if (document.getElementById('__sparrow_chrome').classList.contains('vertical-tabs') ||
          strip.scrollWidth <= strip.clientWidth) return;
      var delta = e.deltaX || e.deltaY;
      if (delta) {
        e.preventDefault();
        strip.scrollLeft += delta;
      }
    };
    document.getElementById('menu-btn').onclick = toggleMenu;
    document.getElementById('menu-incognito').onclick = async function (e) {
      e.stopPropagation();
      var active = await window.isIncognitoActive();
      if (active) {
        await window.disableIncognito();
      } else {
        await window.enableIncognito();
        await createNewTab();
      }
      refreshChrome();
    };
    document.getElementById('menu-open-local-html').onclick = async function (e) {
      e.stopPropagation();
      var filePath = window.prompt('Enter full path to a local HTML file:\nExample: C:\\temp\\page.html');
      if (!filePath || !filePath.trim()) return;
      var url = await window.startLocalHTMLServer(filePath.trim());
      if (!url) {
        alert('Unable to open that file as a localhost HTML page.\nUse a .html or .htm file path.');
        return;
      }
      window.go(url);
    };
    document.getElementById('menu-print').onclick = printCurrentPage;
    document.getElementById('find-prev').onclick = findPrev;
    document.getElementById('find-next').onclick = findNext;
    document.getElementById('find-close').onclick = closeFindBar;
    document.getElementById('find-input').oninput = onFindInput;
    document.getElementById('find-input').onkeydown = function (e) {
      if (e.key === 'Enter') { e.shiftKey ? findPrev() : findNext(); }
      if (e.key === 'Escape') { closeFindBar(); }
    };

    Array.prototype.forEach.call(document.querySelectorAll('[data-go]'), function (el) {
      el.onclick = function () { window.go(el.getAttribute('data-go')); };
    });

    var addressBar = document.getElementById('address-bar');
    addressBar.oninput = onAddressInput;
    addressBar.onfocus = onAddressInput;
    addressBar.onkeydown = function (e) {
      if (e.key === 'Enter') { navigateFromBar(); closeSuggestions(); }
    };

    if (controller.handlers.documentClick) {
      try { document.removeEventListener('click', controller.handlers.documentClick); } catch (e) {}
    }
    controller.handlers.documentClick = function (e) {
      var menu = document.getElementById('menu-dropdown');
      var suggestions = document.getElementById('suggest-box');
      var notifications = document.getElementById('notif-panel');
      if (menu && !e.target.closest('#menu-btn')) menu.classList.remove('open');
      if (suggestions) suggestions.classList.remove('open');
      if (notifications) notifications.classList.remove('open');
      var dlPanel = document.getElementById('__sparrow_download_bar');
      if (dlPanel && !e.target.closest('#__sparrow_download_bar') && !e.target.closest('#btn-download')) dlPanel.hidden = true;
    };
    document.addEventListener('click', controller.handlers.documentClick);

    if (controller.handlers.documentKeydown) {
      try { document.removeEventListener('keydown', controller.handlers.documentKeydown, true); } catch (e) {}
    }
    controller.handlers.documentKeydown = function (e) {
      var key = String(e.key || '').toLowerCase();
      var ctrl = e.ctrlKey && !e.altKey && !e.metaKey;
      var shift = e.shiftKey;
      var handled = true;

      if (key === 'escape' && document.getElementById('qr-share-modal').classList.contains('open')) {
        closeQRShareModal();
      } else if (ctrl && key === 'tab') {
        switchToRelativeTab(shift ? -1 : 1);
      } else if (ctrl && shift && key === 'w') {
        closeAllTabs();
      } else if (ctrl && shift && key === 't') {
        reopenClosedTab();
      } else if (ctrl && shift && key === 'o') {
        window.go('sparrow://bookmarks');
      } else if (ctrl && shift && key === 'n') {
        openIncognito();
      } else if (ctrl && shift && key === 'delete') {
        window.clearHistory('all');
      } else if (ctrl && shift && key === 'r') {
        toggleReader();
      } else if (ctrl && (key === '+' || key === '=' || e.code === 'NumpadAdd')) {
        window.zoomIn(activeTabId);
      } else if (ctrl && key === '-') {
        window.zoomOut(activeTabId);
      } else if (ctrl && key === '0') {
        window.zoomReset(activeTabId);
      } else if (ctrl && !shift && key === 't') {
        createNewTab();
      } else if (ctrl && !shift && key === 'w') {
        closeThisTab(activeTabId);
      } else if (ctrl && !shift && key === 'l') {
        addressBar.focus();
        addressBar.select();
      } else if (ctrl && !shift && key === 'r') {
        window.reload();
      } else if (ctrl && !shift && key === 'd') {
        toggleCurrentFavorite();
      } else if (ctrl && !shift && key === 'h') {
        window.go('sparrow://history');
      } else if (ctrl && !shift && key === 'j') {
        window.go('sparrow://downloads');
      } else if (ctrl && !shift && key === ',') {
        window.go('sparrow://settings');
      } else if (ctrl && !shift && key === 'f') {
        openFindBar();
      } else if (ctrl && !shift && key === 'p') {
        printCurrentPage();
      } else if (e.altKey && !e.ctrlKey && !e.metaKey && !shift && (key === 'arrowleft' || key === 'left')) {
        window.goBack();
      } else if (e.altKey && !e.ctrlKey && !e.metaKey && !shift && (key === 'arrowright' || key === 'right')) {
        window.goForward();
      } else if (!e.ctrlKey && !e.altKey && !e.metaKey && !shift && key === 'f11') {
        window.toggleFullscreenWindow();
      } else {
        handled = false;
      }

      if (handled) {
        e.preventDefault();
        e.stopPropagation();
      }
    };
    document.addEventListener('keydown', controller.handlers.documentKeydown, true);

    // Keep only one recovery timer even if the page destroys and remounts the
    // browser chrome.
    if (!controller.recoveryInterval) {
      controller.recoveryInterval = setInterval(async function () {
        try {
          var tabs = await window.getAllTabsUI();
          var urls = tabs.map(function (t) { return t.URL; }).filter(function (u) { return u; });
          await window.updateRecoverySnapshot(urls);
        } catch (e) {}
      }, 10000);
    }

    if (!controller.downloadPollInterval) {
      controller.lastDownloadID = null;
      controller.downloadPollInterval = setInterval(async function () {
        if (controller.downloadPollPending || typeof window.getDownloads !== 'function') return;
        controller.downloadPollPending = true;
        try {
          var downloads = await window.getDownloads();
          var stats = [];
          try { if (typeof window.getDownloadStats === 'function') stats = (await window.getDownloadStats()) || []; } catch (e) {}
          var pausedIds = {};
          stats.forEach(function (x) { if (x.paused) pausedIds[x.id] = true; });
          var dlBtn = document.getElementById('btn-download');
          if (!Array.isArray(downloads) || downloads.length === 0) {
            if (dlBtn) dlBtn.classList.remove('show', 'active');
            downloadBar.hidden = true;
            return;
          }
          var running = downloads.filter(function (d) { return d.Status === 'downloading' && !pausedIds[d.ID]; });
          if (dlBtn) {
            dlBtn.classList.add('show');
            dlBtn.classList.toggle('active', running.length > 0);
            var avg = running.length
              ? running.reduce(function (sum, d) { return sum + (Number(d.Progress) || 0); }, 0) / running.length
              : 0;
            dlBtn.querySelector('.dl-prog').style.width = Math.round(avg * 0.56) + '%';
          }
          var latest = downloads.reduce(function (current, item) {
            return !current || item.ID > current.ID ? item : current;
          }, null);
          if (!latest) return;
          if (dlBtn) dlBtn.classList.toggle('failed', latest.Status === 'failed');
          if (!downloadBar.hidden && controller.renderDownloads) controller.renderDownloads(downloads, stats);
        } catch (error) {
          console.error('SPARROW could not refresh download status:', error);
        } finally {
          controller.downloadPollPending = false;
        }
      }, 500);
    }

    refreshChrome();

    controller.mounted = true;
    controller.mounting = false;
  }

  function repairChromeStyle() {
    try {
      var style = document.getElementById('__sparrow_chrome_style');
      var root = document.getElementById('__sparrow_chrome');
      if (!style || !root) {
        ensureChrome();
        return;
      }

      // Page fullscreen (video etc.): keep the chrome hidden, don't re-enforce it.
      if (document.documentElement && document.documentElement.hasAttribute('data-sparrow-fs-on')) {
        root.style.setProperty('display', 'none', 'important');
        return;
      }

      // Re-append our stylesheet so it stays at the end of <head>.
      // This matters for pages that inject their own CSS after our Init().
      if (document.head && style.parentNode !== document.head) {
        document.head.appendChild(style);
      }

      var titlebar = document.getElementById('window-titlebar');
      var tabs = document.getElementById('tabstrip');
      var toolbar = document.getElementById('toolbar');
      var addressWrap = document.getElementById('address-wrap');
      var addressBar = document.getElementById('address-bar');
      var appLockButton = document.getElementById('btn-app-lock');
      var qrShareButton = document.getElementById('qr-share-btn');
      var verticalTabs = root.classList.contains('vertical-tabs');
      var contentHeight = verticalTabs ? 78 : 86;
      var toolbarTop = verticalTabs ? '30px' : '38px';
      var toolbarLeft = verticalTabs ? '220px' : '0px';
      var toolbarWidth = verticalTabs ? 'calc(100vw - 220px)' : '100vw';
      var tabsWidth = verticalTabs ? '220px' : 'calc(100vw - 180px)';
      var tabsTop = verticalTabs ? '30px' : '0px';
      var tabsHeight = verticalTabs ? 'calc(100vh - 30px)' : '38px';
      var tabsRight = verticalTabs ? 'auto' : '180px';
      function intact(element, property, value) {
        return !!element &&
          element.style.getPropertyValue(property) === value &&
          element.style.getPropertyPriority(property) === 'important';
      }
      function addressButtonIntact(button) {
        var icon = button && button.querySelector('svg');
        return intact(button, 'display', 'flex') &&
          intact(button, 'visibility', 'visible') &&
          intact(button, 'opacity', '1') &&
          intact(button, 'position', 'static') &&
          intact(button, 'width', '26px') &&
          intact(button, 'height', '26px') &&
          intact(button, 'min-width', '26px') &&
          intact(button, 'flex', '0 0 26px') &&
          intact(button, 'color', '#fff') &&
          intact(icon, 'fill', '#fff');
      }
      if (intact(root, 'height', contentHeight + 'px') &&
          intact(root, 'z-index', '2147483647') &&
          intact(root, 'display', 'block') &&
          intact(root, 'visibility', 'visible') &&
          intact(titlebar, 'z-index', '30') &&
          intact(tabs, 'top', tabsTop) &&
          intact(tabs, 'right', tabsRight) &&
          intact(tabs, 'width', tabsWidth) &&
          intact(tabs, 'height', tabsHeight) &&
          intact(toolbar, 'top', toolbarTop) &&
          intact(toolbar, 'left', toolbarLeft) &&
          intact(toolbar, 'width', toolbarWidth) &&
          intact(addressWrap, 'position', 'relative') &&
          intact(addressBar, 'position', 'static') &&
          addressButtonIntact(appLockButton) &&
          addressButtonIntact(qrShareButton)) {
        return;
      }
      root.style.setProperty('display', 'block', 'important');
      root.style.setProperty('visibility', 'visible', 'important');
      root.style.setProperty('opacity', '1', 'important');
      root.style.setProperty('position', 'fixed', 'important');
      root.style.setProperty('top', '0', 'important');
      root.style.setProperty('left', '0', 'important');
      root.style.setProperty('right', '0', 'important');
      root.style.setProperty('width', '100vw', 'important');
      root.style.setProperty('min-width', '0', 'important');
      root.style.setProperty('height', contentHeight + 'px', 'important');
      root.style.setProperty('min-height', contentHeight + 'px', 'important');
      root.style.setProperty('z-index', '2147483647', 'important');
      root.style.setProperty('transform', 'none', 'important');
      root.style.setProperty('translate', 'none', 'important');
      root.style.setProperty('zoom', '1', 'important');
      window.__SPARROW_CHROME_HEIGHT__ = contentHeight;
      if (document.documentElement) {
        document.documentElement.style.setProperty('--sparrow-chrome-height', contentHeight + 'px');
      }

      if (titlebar) {
        titlebar.style.setProperty('display', 'flex', 'important');
        titlebar.style.setProperty('visibility', 'visible', 'important');
        titlebar.style.setProperty('opacity', '1', 'important');
        titlebar.style.setProperty('position', 'fixed', 'important');
        titlebar.style.setProperty('top', '0', 'important');
        titlebar.style.setProperty('height', '30px', 'important');
        titlebar.style.setProperty('z-index', '30', 'important');
      }
      if (tabs) {
        tabs.style.setProperty('display', 'flex', 'important');
        tabs.style.setProperty('visibility', 'visible', 'important');
        tabs.style.setProperty('opacity', '1', 'important');
        tabs.style.setProperty('position', 'fixed', 'important');
        tabs.style.setProperty('top', tabsTop, 'important');
        tabs.style.setProperty('left', '0px', 'important');
        tabs.style.setProperty('right', tabsRight, 'important');
        tabs.style.setProperty('width', tabsWidth, 'important');
        tabs.style.setProperty('height', tabsHeight, 'important');
        tabs.style.setProperty('flex-direction', verticalTabs ? 'column' : 'row', 'important');
        tabs.style.setProperty('align-items', verticalTabs ? 'stretch' : 'flex-end', 'important');
        tabs.style.setProperty('overflow-x', verticalTabs ? 'hidden' : 'auto', 'important');
        tabs.style.setProperty('overflow-y', verticalTabs ? 'auto' : 'hidden', 'important');
        tabs.style.setProperty('padding', verticalTabs ? '10px 8px' : '0px 6px', 'important');
        tabs.style.setProperty('z-index', '20', 'important');
      }
      if (toolbar) {
        toolbar.style.setProperty('display', 'flex', 'important');
        toolbar.style.setProperty('visibility', 'visible', 'important');
        toolbar.style.setProperty('opacity', '1', 'important');
        toolbar.style.setProperty('position', 'fixed', 'important');
        toolbar.style.setProperty('top', toolbarTop, 'important');
        toolbar.style.setProperty('left', toolbarLeft, 'important');
        toolbar.style.setProperty('right', '0px', 'important');
        toolbar.style.setProperty('width', toolbarWidth, 'important');
        toolbar.style.setProperty('height', '48px', 'important');
        toolbar.style.setProperty('z-index', '10', 'important');
      }

      if (addressWrap) {
        addressWrap.style.setProperty('position', 'relative', 'important');
        addressWrap.style.setProperty('top', 'auto', 'important');
        addressWrap.style.setProperty('left', 'auto', 'important');
        addressWrap.style.setProperty('right', 'auto', 'important');
        addressWrap.style.setProperty('transform', 'none', 'important');
        addressWrap.style.setProperty('flex', '1 1 auto', 'important');
        addressWrap.style.setProperty('min-width', '0', 'important');
      }
      if (addressBar) {
        addressBar.style.setProperty('position', 'static', 'important');
        addressBar.style.setProperty('top', 'auto', 'important');
        addressBar.style.setProperty('left', 'auto', 'important');
        addressBar.style.setProperty('right', 'auto', 'important');
        addressBar.style.setProperty('transform', 'none', 'important');
        addressBar.style.setProperty('flex', '1 1 auto', 'important');
        addressBar.style.setProperty('min-width', '0', 'important');
      }
      [appLockButton, qrShareButton].forEach(function (button) {
        if (!button) return;
        button.style.setProperty('display', 'flex', 'important');
        button.style.setProperty('visibility', 'visible', 'important');
        button.style.setProperty('opacity', '1', 'important');
        button.style.setProperty('position', 'static', 'important');
        button.style.setProperty('width', '26px', 'important');
        button.style.setProperty('height', '26px', 'important');
        button.style.setProperty('min-width', '26px', 'important');
        button.style.setProperty('flex', '0 0 26px', 'important');
        button.style.setProperty('color', '#fff', 'important');
        var icon = button.querySelector('svg');
        if (icon) icon.style.setProperty('fill', '#fff', 'important');
      });
    } catch (e) {}
  }

  function ensureChrome() {
    if (controller.ensureTimer) return;
    controller.ensureTimer = setTimeout(function () {
      controller.ensureTimer = null;

      var root = document.getElementById('__sparrow_chrome');
      if (root && document.documentElement && document.documentElement.contains(root)) {
        controller.mounted = true;
        return;
      }

      controller.mounted = false;
      whenBodyReady(mountChrome);
    }, 0);
  }

  controller.ensure = ensureChrome;
  controller.repair = repairChromeStyle;

  function startChromeWatchdog() {
    ensureChrome();

    if (controller.observer) {
      try { controller.observer.disconnect(); } catch (e) {}
    }
    controller.observedBody = null;

    function observeBody() {
      if (!controller.observer || !document.body || controller.observedBody === document.body) return;
      controller.observedBody = document.body;
      try { controller.observer.observe(document.body, { childList: true }); } catch (e) {}
    }

    if (typeof MutationObserver !== 'undefined') {
      controller.observer = new MutationObserver(function () {
        if (!document.getElementById('__sparrow_chrome')) {
          ensureChrome();
        }
        observeBody();
      });

      try {
        controller.observer.observe(document.documentElement, { childList: true });
      } catch (e) {}
      observeBody();
    }

    if (!controller.interval) {
      controller.interval = setInterval(function () {
        observeBody();
        if (!document.getElementById('__sparrow_chrome')) {
          ensureChrome();
        } else {
          repairChromeStyle();
        }
      }, 1000);
    }
  }

  whenBodyReady(function () {
    startChromeWatchdog();
  });
})();
`
