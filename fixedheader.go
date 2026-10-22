package main

func BuildFixedHeaderJS() string {
	return `(function () {
  'use strict';
  if (window.top !== window || window.__sparrowFixedHeaderInstalled) return;
  if (window.location.protocol !== 'http:' &&
      window.location.protocol !== 'https:' &&
      window.location.protocol !== 'file:') return;
  var isYouTube = /(^|\.)youtube\.com$/i.test(window.location.hostname) ||
    /(^|\.)youtu\.be$/i.test(window.location.hostname);
  window.__sparrowFixedHeaderInstalled = true;

  var MARK = 'data-sparrow-header-shifted';
  var SAMPLE_X = [0.1, 0.5, 0.9];
  var SAMPLE_Y = [6, 50, 100, 150];
  var SAMPLE_PX = [12, 90];

  function chromeHeight() {
    try {
      var chrome = document.getElementById('__sparrow_chrome');
      if (chrome) {
        var rect = chrome.getBoundingClientRect();
        if (rect.bottom > 0) {
          window.__SPARROW_CHROME_HEIGHT__ = rect.bottom;
          return rect.bottom;
        }
      }
    } catch (e) {}
    return Number(window.__SPARROW_CHROME_HEIGHT__) || 86;
  }

  function number(value) {
    var parsed = parseFloat(value);
    return isFinite(parsed) ? parsed : null;
  }

  function insideSparrow(element) {
    return !!(element.closest &&
      element.closest('#__sparrow_chrome, dialog, [role="dialog"], [data-sparrow-fixed-header-ignore], .modal, .popup, .dropdown, video, audio, canvas, [data-sparrow-content-fullscreen]'));
  }

  function insideNavigation(element) {
    var from = isYouTube ? element : element.parentElement;
    return !!(from && from.closest &&
      from.closest('nav, aside, [role="navigation"], [role="complementary"]'));
  }

  function hasInnerScroller(element) {
    var parent = element.parentElement;
    var depth = 0;
    while (parent && parent !== document.body &&
           parent !== document.documentElement && depth++ < 40) {
      var overflow = window.getComputedStyle(parent).overflowY;
      if (overflow === 'auto' || overflow === 'scroll' ||
          overflow === 'hidden' || overflow === 'overlay') return true;
      parent = parent.parentElement;
    }
    return false;
  }

  // A side rail is a fixed element glued to the left/right edge that runs
  // down toward the bottom of the window (YouTube's left guide drawer and its
  // narrow mini-guide). It must move down together with the shifted header,
  // otherwise it sits behind the Sparrow toolbar and looks too high.
  function isSideRail(element, style, rect) {
    if (style.position !== 'fixed') return false;
    if (rect.width < 40 || rect.width > window.innerWidth * 0.5) return false;
    if (rect.left > 2 && rect.right < window.innerWidth - 2) return false;
    var tallEnough = rect.height > window.innerHeight * 0.5;
    var anchoredBottom = style.bottom !== 'auto' && number(style.bottom) !== null;
    return tallEnough || anchoredBottom;
  }

  // Remember what we overwrote so it can be put back if the site later
  // switches the element out of fixed/sticky (e.g. YouTube's chips bar).
  var shifted = [];
  var PROPS = ['top', 'bottom', 'height'];
  function remember(element) {
    var saved = {};
    PROPS.forEach(function (name) {
      saved[name] = [element.style.getPropertyValue(name), element.style.getPropertyPriority(name)];
    });
    shifted.push({ element: element, saved: saved });
  }

  function restoreStale() {
    for (var i = shifted.length - 1; i >= 0; i--) {
      var entry = shifted[i];
      var element = entry.element;
      if (!element.isConnected) { shifted.splice(i, 1); continue; }
      var position = window.getComputedStyle(element).position;
      if (position === 'fixed' || position === 'sticky') continue;
      PROPS.forEach(function (name) {
        var value = entry.saved[name];
        if (value[0]) element.style.setProperty(name, value[0], value[1]);
        else element.style.removeProperty(name);
      });
      element.removeAttribute(MARK);
      shifted.splice(i, 1);
    }
  }

  function shiftRail(element, style, top, offset) {
    remember(element);
    element.setAttribute(MARK, '1');
    element.style.setProperty('top', (top + offset) + 'px', 'important');
    // Keep the bottom edge on the window bottom instead of overflowing it.
    if (style.bottom === 'auto') {
      element.style.setProperty('bottom', '0px', 'important');
      element.style.setProperty('height', 'auto', 'important');
    }
  }

  function shift(element, top, offset) {
    remember(element);
    element.setAttribute(MARK, '1');
    element.style.setProperty('top', (top + offset) + 'px', 'important');
  }

  function ensureYouTubeCSS() {
    if (document.getElementById('__sparrow_yt_fix')) return;
    var st = document.createElement('style');
    st.id = '__sparrow_yt_fix';
    st.textContent = '#masthead-container, ytd-masthead { top: var(--sparrow-chrome-height, 86px) !important; }';
    (document.head || document.documentElement).appendChild(st);
  }

  function shiftYouTubeMasthead(offset) {
    ensureYouTubeCSS();
    var selectors = ['#masthead-container', '#masthead-positioner', 'ytd-masthead'];
    for (var i = 0; i < selectors.length; i++) {
      var element = document.querySelector(selectors[i]);
      if (!element) continue;
      var style = window.getComputedStyle(element);
      if (style.position !== 'fixed' && style.position !== 'sticky') continue;
      var top = number(style.top);
      var rect = element.getBoundingClientRect();
      if (top === null || rect.bottom <= 0 || rect.top >= offset) return;
      element.style.setProperty('top', (top + offset - rect.top) + 'px', 'important');
      return;
    }
  }

  // Adaptive: measure where #page-manager really starts and nudge its
  // margin until it sits right under the (shifted) masthead. Unlike a blind
  // "+offset" this never double-shifts when html padding already moved the
  // page, so the chips bar and the video grid stay directly under the header.
  function shiftYouTubeContent(offset) {
    var page = document.querySelector('#page-manager');
    if (!page || (window.scrollY || 0) > 0) return;
    var masthead = document.querySelector('ytd-masthead');
    var mh = masthead ? masthead.getBoundingClientRect().height : 56;
    if (!(mh > 20 && mh < 120)) mh = 56;
    var margin = number(window.getComputedStyle(page).marginTop) || 0;
    var delta = (offset + mh) - page.getBoundingClientRect().top;
    if (Math.abs(delta) < 2) return;
    page.style.setProperty('margin-top', Math.max(0, margin + delta) + 'px', 'important');
  }

  function shiftYouTubeSideRail(offset) {
    var selectors = ['#guide', 'ytd-mini-guide-renderer'];
    for (var i = 0; i < selectors.length; i++) {
      var element = document.querySelector(selectors[i]);
      if (!element || element.hasAttribute(MARK)) continue;
      var style = window.getComputedStyle(element);
      if (style.position !== 'fixed') continue;
      var top = number(style.top);
      var rect = element.getBoundingClientRect();
      if (top === null || rect.height <= 0 || rect.top >= offset) continue;
      shiftRail(element, style, top, offset);
    }
  }

  function fixYouTubeCategories(offset) {
    var categories = document.querySelector('#chips-wrapper') ||
      document.querySelector('ytd-feed-filter-chip-bar-renderer');
    if (!categories) return;

    var masthead = document.querySelector('#masthead-container') ||
      document.querySelector('ytd-masthead');
    var mastheadHeight = masthead ? masthead.getBoundingClientRect().height : 56;
    if (!(mastheadHeight > 20 && mastheadHeight < 120)) mastheadHeight = 56;

    categories.style.setProperty('position', 'sticky', 'important');
    categories.style.setProperty('top', (offset + mastheadHeight) + 'px', 'important');
    categories.style.setProperty('z-index', '2200', 'important');
  }

  var ignoredAt = new WeakMap();
  var scanNo = 0;

  function consider(element, offset) {
    if (!element || element.hasAttribute(MARK)) return;
    var last = ignoredAt.get(element);
    if (last !== undefined && scanNo - last < 6) return;
    var style = window.getComputedStyle(element);
    var position = style.position;
    if (position !== 'fixed' && position !== 'sticky') { ignoredAt.set(element, scanNo); return; }
    if (insideSparrow(element)) return;
    var inNav = insideNavigation(element);

    var top = number(style.top);
    if (top === null || top >= offset) return;

    var rect = element.getBoundingClientRect();
    if (rect.height <= 0 || rect.bottom <= 0 || rect.top >= offset) return;

    if (isSideRail(element, style, rect)) {
      shiftRail(element, style, top, offset);
      return;
    }
    if (inNav) return;

    if (rect.width < 180) return;
    if (rect.height > window.innerHeight * 0.6) return;

    if (position === 'fixed') {
      var expectedTop = top + (number(style.marginTop) || 0);
      if (Math.abs(rect.top - expectedTop) > 3) return;
    } else if (hasInnerScroller(element)) {
      return;
    }
    shift(element, top, offset);
  }

  function walk(stack, offset, visited) {
    for (var i = 0; i < stack.length; i++) {
      var element = stack[i];
      var depth = 0;
      while (element && element.nodeType === 1 &&
             element !== document.documentElement && depth++ < 30) {
        if (visited.has(element)) break;
        visited.add(element);
        consider(element, offset);
        element = element.parentElement;
      }
    }
  }

  var HEADER_SELECTOR = 'header, nav, [role="banner"], [class~="navbar"], [class~="header"], [id="header"], [id="navbar"]';
  var headerBudget = 10;

  function scanHeaders(offset) {
    var list;
    try { list = document.querySelectorAll(HEADER_SELECTOR); } catch (e) { return; }
    var max = Math.min(list.length, 60);
    for (var i = 0; i < max; i++) consider(list[i], offset);
  }

  function scan() {
    if (document.hidden || document.fullscreenElement) return;
    restoreStale();
    scanNo++;

    var offset = chromeHeight();
    if (isYouTube) {
      shiftYouTubeMasthead(offset);
      shiftYouTubeContent(offset);
      shiftYouTubeSideRail(offset);
      fixYouTubeCategories(offset);
      return;
    }
    if (!document.body || typeof document.elementsFromPoint !== 'function') return;
    if (!isYouTube && headerBudget > 0) { headerBudget--; scanHeaders(offset); }

    var width = window.innerWidth;
    var visited = new Set();
    for (var yi = 0; yi < SAMPLE_Y.length; yi++) {
      for (var xi = 0; xi < SAMPLE_X.length; xi++) {
        var stack = document.elementsFromPoint(
          Math.round(width * SAMPLE_X[xi]),
          SAMPLE_Y[yi]
        );
        walk(stack, offset, visited);
      }
      for (var pi = 0; pi < SAMPLE_PX.length; pi++) {
        walk(document.elementsFromPoint(SAMPLE_PX[pi], SAMPLE_Y[yi]), offset, visited);
      }
    }
  }

  var scanTimer = 0;
  function runScan() {
    scanTimer = 0;
    var go = function () { try { scan(); } catch (e) {} };
    if (typeof requestIdleCallback === 'function') requestIdleCallback(go, { timeout: 600 });
    else go();
  }
  function scheduleScan(delay) {
    if (scanTimer) return;
    scanTimer = setTimeout(runScan, typeof delay === 'number' ? delay : 0);
  }

  (isYouTube ? [300, 900, 2000, 4000, 8000] : [0, 150, 500, 1200, 2500, 5000]).forEach(function (delay) {
    setTimeout(function () { scheduleScan(0); }, delay);
  });
  window.addEventListener('load', function () { headerBudget = Math.max(headerBudget, 2); scheduleScan(0); });
  window.addEventListener('resize', function () { headerBudget = Math.max(headerBudget, 2); scheduleScan(150); }, { passive: true });

  // while scrolling only the cheap "put it back" check runs; the full scan waits until scrolling stops
  var restoreFrame = 0;
  window.addEventListener('scroll', function () {
    if (!restoreFrame) {
      restoreFrame = requestAnimationFrame(function () { restoreFrame = 0; try { restoreStale(); } catch (e) {} });
    }
    scheduleScan(220);
  }, { passive: true });
  ['fullscreenchange', 'webkitfullscreenchange'].forEach(function (name) {
    document.addEventListener(name, function () { scheduleScan(0); }, true);
  });
  window.setInterval(function () { if (!document.hidden) scheduleScan(0); }, 3000);
})();`
}
