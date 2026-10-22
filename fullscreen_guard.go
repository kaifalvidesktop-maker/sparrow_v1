package main

func BuildFullscreenGuardJS() string {
	return `(function () {
  'use strict';
  if (window.__sparrowFsInstalled) return;
  window.__sparrowFsInstalled = true;

  var IS_TOP = window.top === window;
  var ATTR = 'data-sparrow-fs';
  var fsElement = null;
  var childFrame = null;
  var nativeExit = null;
  var nativeFullscreenGetter = null;
  var swallowNativeExit = false;

  var NEUTRAL = {
'transform': 'none', 'translate': 'none', 'rotate': 'none', 'scale': 'none',
'filter': 'none', 'backdrop-filter': 'none', 'perspective': 'none',
'contain': 'none', 'clip-path': 'none', 'zoom': '1',
'will-change': 'auto', 'content-visibility': 'visible'
};

var saved = [];

var css =

'[' + ATTR + '] {' +

' position: fixed !important;' +

' top: 0 !important; left: 0 !important; right: auto !important; bottom: auto !important;' +

' width: 100vw !important; height: 100vh !important;' +

' min-width: 0 !important; min-height: 0 !important;' +

' max-width: none !important; max-height: none !important;' +

' margin: 0 !important; transform: none !important; zoom: 1 !important;' +

' z-index: 2147483000 !important; background-color: #000 !important;' +

' box-sizing: border-box !important; object-fit: contain !important;' +

' visibility: visible !important; opacity: 1 !important;' +

'}' +

'html[data-sparrow-fs-on][data-sparrow-fs-on][data-sparrow-fs-on] {' +

' overflow: hidden !important; padding: 0 !important; margin: 0 !important; scroll-padding: 0 !important;' +

'}' +

'html[data-sparrow-fs-on][data-sparrow-fs-on] body { overflow: hidden !important; }' +

'html[data-sparrow-fs-on][data-sparrow-fs-on] body [data-sparrow-fs]:not(#__sf_a):not(#__sf_b) { zoom: 1 !important; }' +

'html[data-sparrow-fs-on][data-sparrow-fs-on] #__sparrow_chrome { display: none !important; }';

function setChromeVisible(visible) {

if (!IS_TOP) return;

try {

var chrome = document.getElementById('__sparrow_chrome');

if (chrome) chrome.style.setProperty('display', visible ? 'block' : 'none', 'important');

} catch (e) {}

}

function restoreAncestors() {

while (saved.length) {

var item = saved.pop();

try {

Object.keys(item.prev).forEach(function (prop) {

var old = item.prev[prop];

if (old[0]) item.el.style.setProperty(prop, old[0], old[1]);

else item.el.style.removeProperty(prop);

});

} catch (e) {}

}

}

function neutralizeAncestors(element) {

restoreAncestors();

var node = element;

while (node) {

if (node.nodeType === 11) { node = node.host; continue; }

if (node.nodeType !== 1) break;

if (node !== document.documentElement) {

var prev = {};

try {

Object.keys(NEUTRAL).forEach(function (prop) {

prev[prop] = [node.style.getPropertyValue(prop), node.style.getPropertyPriority(prop)];

node.style.setProperty(prop, NEUTRAL[prop], 'important');

});

saved.push({ el: node, prev: prev });

} catch (e) {}

}

node = node.parentNode;

}

}

function ensureStyle() {
    if (window.__sparrowFsSheet || document.getElementById('__sparrow_fs_style')) return;
    try {
      var sheet = new CSSStyleSheet();
      sheet.replaceSync(css);
      document.adoptedStyleSheets = Array.prototype.concat.call(
        document.adoptedStyleSheets || [],
        [sheet]
      );
      window.__sparrowFsSheet = sheet;
      return;
    } catch (e) {}
    try {
      var style = document.createElement('style');
      style.id = '__sparrow_fs_style';
      style.textContent = css;
      (document.head || document.documentElement).appendChild(style);
    } catch (e) {}
  }

  function setRootState(active) {

var root = document.documentElement;

if (!root) return;

if (active) {

root.setAttribute('data-sparrow-fs-on', '1');

setChromeVisible(false);

} else if (!fsElement && !childFrame) {

root.removeAttribute('data-sparrow-fs-on');

restoreAncestors();

setChromeVisible(true);

}

}

var nativeTimer = 0;

function syncNative() {

if (!IS_TOP) return;

clearTimeout(nativeTimer);

nativeTimer = setTimeout(function () {

try {

if (typeof window.setPageFullscreen === 'function') {

var r = window.setPageFullscreen(!!currentElement());

if (r && r.catch) r.catch(function () {});

}

} catch (e) {}

[120, 400].forEach(function (ms) {

setTimeout(function () {

try { window.dispatchEvent(new Event('resize')); } catch (e) {}

}, ms);

});

}, 30);

}

function notifyParent(state) {
    if (IS_TOP) return;
    try {
      window.parent.postMessage({ __sparrowFullscreenState: state }, '*');
    } catch (e) {}
  }

  function fireFullscreenChange(element) {
    var target = element && element.isConnected ? element : document;
    ['fullscreenchange', 'webkitfullscreenchange', 'mozfullscreenchange', 'MSFullscreenChange'].forEach(function (type) {
      try { target.dispatchEvent(new Event(type, { bubbles: true })); } catch (e) {}
    });
  }

  function nudgeLayout() {
    setTimeout(function () {
      try { window.dispatchEvent(new Event('resize')); } catch (e) {}
    }, 0);
  }

  function isDocumentRoot(element) {
    return element === document.documentElement || element === document.body;
  }

  function currentElement() {
    return fsElement || childFrame || null;
  }

  function leave(silent) {
    if (!fsElement) return;
    var element = fsElement;
    fsElement = null;
    try { element.removeAttribute(ATTR); } catch (e) {}
    setRootState(false);
    if (!silent && !childFrame) notifyParent('exit');

syncNative();
    if (!silent) fireFullscreenChange(element);
    nudgeLayout();
  }

  function leaveChild(frame, silent) {
    if (childFrame !== frame) return;
    childFrame = null;
    try { frame.removeAttribute(ATTR); } catch (e) {}
    setRootState(false);
    if (!silent && !fsElement) notifyParent('exit');

syncNative();
    if (!silent) fireFullscreenChange(frame);
    nudgeLayout();
  }

  function enter(element) {
    if (!element || element.nodeType !== 1 || fsElement === element) return;
    if (childFrame) {
      try {
        childFrame.contentWindow.postMessage(
          { __sparrowFullscreenState: 'exit-request' },
          '*'
        );
      } catch (e) {}
      leaveChild(childFrame, true);
    }
    if (fsElement) leave(true);
    ensureStyle();
    fsElement = element;
    setRootState(true);
    if (!isDocumentRoot(element)) {
      element.setAttribute(ATTR, '1');
      neutralizeAncestors(element);
    }
    notifyParent('enter');
    syncNative();
    fireFullscreenChange(element);
    nudgeLayout();
  }

  function enterChild(frame) {
    if (childFrame === frame) return;
    if (fsElement) leave(true);
    if (childFrame) leaveChild(childFrame, true);
    ensureStyle();
    childFrame = frame;
    setRootState(true);
    frame.setAttribute(ATTR, '1');
    neutralizeAncestors(frame);
    notifyParent('enter');
    syncNative();
    fireFullscreenChange(frame);
    nudgeLayout();
  }

  function findChildFrame(source) {
    var frames = document.getElementsByTagName('iframe');
    for (var i = 0; i < frames.length; i++) {
      try {
        if (frames[i].contentWindow === source) return frames[i];
      } catch (e) {}
    }
    return null;
  }

  function nativeFullscreenElement() {
    try {
      return nativeFullscreenGetter ? nativeFullscreenGetter.call(document) : null;
    } catch (e) {
      return null;
    }
  }

  function requestFullscreen() {
    enter(this);
    return Promise.resolve();
  }

  function exitFullscreen() {
    if (fsElement) {
      leave(false);
    } else if (childFrame) {
      try {
        childFrame.contentWindow.postMessage(
          { __sparrowFullscreenState: 'exit-request' },
          '*'
        );
      } catch (e) {}
      leaveChild(childFrame, false);
    } else if (nativeFullscreenElement() && nativeExit) {
      try {
        var result = nativeExit.call(document);
        if (result && result.catch) result.catch(function () {});
      } catch (e) {}
    }
    return Promise.resolve();
  }

  function defineGetter(prototype, name, getter) {
    try {
      Object.defineProperty(prototype, name, {
        configurable: true,
        enumerable: true,
        get: getter
      });
    } catch (e) {}
  }

  if (typeof Document !== 'undefined') {
    try {
      var fullscreenDescriptor =
        Object.getOwnPropertyDescriptor(Document.prototype, 'fullscreenElement');
      nativeFullscreenGetter = fullscreenDescriptor && fullscreenDescriptor.get;
      nativeExit = Document.prototype.exitFullscreen ||
        Document.prototype.webkitExitFullscreen ||
        Document.prototype.mozCancelFullScreen ||
        Document.prototype.msExitFullscreen;
    } catch (e) {}
  }

  if (typeof Element !== 'undefined') {
    [
      'requestFullscreen',
      'webkitRequestFullscreen',
      'webkitRequestFullScreen',
      'mozRequestFullScreen',
      'msRequestFullscreen'
    ].forEach(function (name) {
      try {
        if (name === 'requestFullscreen' || typeof Element.prototype[name] === 'function') {
          Element.prototype[name] = requestFullscreen;
        }
      } catch (e) {}
    });
  }

  if (typeof Document !== 'undefined') {
    [
      'exitFullscreen',
      'webkitExitFullscreen',
      'webkitCancelFullScreen',
      'mozCancelFullScreen',
      'msExitFullscreen'
    ].forEach(function (name) {
      try {
        if (name === 'exitFullscreen' || typeof Document.prototype[name] === 'function') {
          Document.prototype[name] = exitFullscreen;
        }
      } catch (e) {}
    });
    ['fullscreenElement', 'webkitFullscreenElement', 'webkitCurrentFullScreenElement'].forEach(function (name) {
      defineGetter(Document.prototype, name, currentElement);
    });
    ['fullscreen', 'webkitIsFullScreen'].forEach(function (name) {
      defineGetter(Document.prototype, name, function () { return !!currentElement(); });
    });
  }

  if (typeof HTMLVideoElement !== 'undefined') {
    ['webkitEnterFullscreen', 'webkitEnterFullScreen'].forEach(function (name) {
      try { HTMLVideoElement.prototype[name] = function () { enter(this); }; } catch (e) {}
    });
    ['webkitExitFullscreen', 'webkitExitFullScreen'].forEach(function (name) {
      try {
        HTMLVideoElement.prototype[name] = function () {
          if (fsElement === this) leave(false);
        };
      } catch (e) {}
    });
    defineGetter(HTMLVideoElement.prototype, 'webkitDisplayingFullscreen', function () {
      return fsElement === this;
    });
  }

  function onNativeFullscreenChange(event) {
    if (!event.isTrusted) return;
    var element = nativeFullscreenElement();
    if (element) {
      event.stopImmediatePropagation();
      swallowNativeExit = true;
      try {
        var result = nativeExit && nativeExit.call(document);
        if (result && result.catch) result.catch(function () {});
      } catch (e) {}
      enter(element);
    } else if (swallowNativeExit) {
      swallowNativeExit = false;
      event.stopImmediatePropagation();
    }
  }

  ['fullscreenchange', 'webkitfullscreenchange', 'mozfullscreenchange', 'MSFullscreenChange'].forEach(function (name) {
    try { document.addEventListener(name, onNativeFullscreenChange, true); } catch (e) {}
  });

  window.addEventListener('pagehide', function () {

if (!IS_TOP || !currentElement()) return;

try { if (typeof window.setPageFullscreen === 'function') window.setPageFullscreen(false); } catch (e) {}

});

setInterval(function () {

if (fsElement && !fsElement.isConnected) leave(false);

if (childFrame && !childFrame.isConnected) leaveChild(childFrame, false);

}, 1000);

window.addEventListener('keydown', function (event) {
    if (event.key !== 'Escape' || !currentElement()) return;
    event.stopImmediatePropagation();
    event.preventDefault();
    exitFullscreen();
  }, true);

  window.addEventListener('message', function (event) {
    var state = event.data && event.data.__sparrowFullscreenState;
    if (!state) return;
    if (state === 'exit-request') {
      if (fsElement) leave(false);
      return;
    }
    var frame = findChildFrame(event.source);
    if (!frame) return;
    if (state === 'enter') enterChild(frame);
    else if (state === 'exit') leaveChild(frame, false);
  }, true);
})();`
}
