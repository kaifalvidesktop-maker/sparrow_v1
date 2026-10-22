package main

import (
	"strings"
)

// ---------------------------------------------------
// REAL AD / TRACKER BLOCKING
// ---------------------------------------------------

func buildDomainMapJS(domains map[string]bool) string {
	names := make([]string, 0, len(domains))
	for d := range domains {
		names = append(names, `"`+d+`":true`)
	}
	if len(names) == 0 {
		return "Object.create(null)"
	}
	return "{" + strings.Join(names, ",") + "}"
}

// allAdDomains / allTrackerDomains merge the built-in lists with the extras.
func mergedDomainMap(base map[string]bool, extra []string) map[string]bool {
	out := make(map[string]bool, len(base)+len(extra))
	for d := range base {
		out[d] = true
	}
	for _, d := range extra {
		if n := normalizeHost(d); n != "" {
			out[n] = true
		}
	}
	return out
}

func BuildAdblockJS() string {
	shield.mu.Lock()
	adOn := shield.AdBlockOn
	trackerOn := shield.TrackerBlockOn
	shield.mu.Unlock()

	adList := buildDomainMapJS(mergedDomainMap(adDomains, extraAdTrackerDomains))
	trackerList := buildDomainMapJS(mergedDomainMap(trackerDomains, nil))

	r := strings.NewReplacer(
		"__AD_ON__", boolJS(adOn),
		"__TRACKER_ON__", boolJS(trackerOn),
		"__AD_LIST__", adList,
		"__TRACKER_LIST__", trackerList,
	)
	return r.Replace(adblockScriptTemplate)
}

const adblockScriptTemplate = `(function () {
  "use strict";
  if (window.__sparrowAdblock) return;
  window.__sparrowAdblock = true;

  var AD_ON = __AD_ON__;
  var TRACKER_ON = __TRACKER_ON__;
  var AD_DOMAINS = __AD_LIST__;
  var TRACKER_DOMAINS = __TRACKER_LIST__;

  var hostCache = Object.create(null);
  var hostCacheSize = 0;
  var pendingBlocks = Object.create(null);
  var pendingTimer = null;
  var HOST_RE = /^(?:https?:)?\/\/(?:[^\/?#@]*@)?([a-z0-9.\-]+)(?::\d+)?(?:[\/?#]|$)/i;
  var pageHost = String(location.hostname || "").toLowerCase();

  function hostFromURL(u) {
    if (!u) return "";
    u = String(u);
    var m = HOST_RE.exec(u);
    if (m) return m[1].toLowerCase();
    if (/^(?:data|blob|about|javascript):/i.test(u)) return "";
    try { return new URL(u, location.href).hostname.toLowerCase(); } catch (e) { return ""; }
  }

  function matchesList(host, map) {
    var c = host;
    while (c) {
      if (map[c]) return true;
      var dot = c.indexOf(".");
      if (dot < 0) break;
      c = c.slice(dot + 1);
    }
    return false;
  }

  // never break an ad/tracker company's own website
  var SELF_EXEMPT = matchesList(pageHost, AD_DOMAINS) || matchesList(pageHost, TRACKER_DOMAINS);

  function isBlocked(url) {
    if (SELF_EXEMPT) return null;
    var host = hostFromURL(url);
    if (!host || host === pageHost) return null;
    var cached = hostCache[host];
    if (cached !== undefined) return cached || null;
    var kind = null;
    if (AD_ON && matchesList(host, AD_DOMAINS)) kind = "ad";
    else if (TRACKER_ON && matchesList(host, TRACKER_DOMAINS)) kind = "tracker";
    if (hostCacheSize > 2000) { hostCache = Object.create(null); hostCacheSize = 0; }
    hostCache[host] = kind || "";
    hostCacheSize++;
    return kind;
  }

  function flushReports() {
    pendingTimer = null;
    var batch = [];
    for (var k in pendingBlocks) batch.push(pendingBlocks[k]);
    pendingBlocks = Object.create(null);
    if (!batch.length || !window.reportBlockedBatch) return;
    for (var i = 0; i < batch.length; i += 200) {
      try { window.reportBlockedBatch(batch.slice(i, i + 200)); } catch (e) {}
    }
  }

  function report(kind, url) {
    if (!kind || !url) return;
    url = String(url);
    var key = kind + "|" + url;
    var entry = pendingBlocks[key];
    if (entry) entry.N++;
    else pendingBlocks[key] = { Kind: kind, URL: url.slice(0, 300), N: 1 };
    if (!pendingTimer) pendingTimer = setTimeout(flushReports, 1000);
  }

  // ---------------------------------------------------
  // NETWORK: fetch / XHR / beacon / WebSocket / EventSource
  // ---------------------------------------------------
  var nativeFetch = window.fetch;
  if (nativeFetch) {
    window.fetch = function (input, init) {
      var url = (typeof input === "string") ? input : (input && (input.url || input.href));
      var kind = isBlocked(url || "");
      if (kind) {
        report(kind, url);
        return Promise.reject(new TypeError("SPARROW Shield blocked: " + url));
      }
      return nativeFetch.apply(this, arguments);
    };
  }

  if (window.XMLHttpRequest) {
    var nativeOpen = XMLHttpRequest.prototype.open;
    XMLHttpRequest.prototype.open = function (method, url) {
      var kind = isBlocked(url || "");
      if (kind) {
        report(kind, url);
        arguments[1] = "data:text/plain,";
      }
      return nativeOpen.apply(this, arguments);
    };
  }

  if (navigator.sendBeacon) {
    var nativeBeacon = navigator.sendBeacon;
    navigator.sendBeacon = function (url) {
      var kind = isBlocked(url || "");
      if (kind) { report(kind, url); return true; }
      return nativeBeacon.apply(navigator, arguments);
    };
  }

  if (window.WebSocket) {
    var NativeWS = window.WebSocket;
    var WS = function (url, protocols) {
      var kind = isBlocked(url || "");
      if (kind) { report(kind, url); throw new DOMException("blocked", "SecurityError"); }
      return protocols === undefined ? new NativeWS(url) : new NativeWS(url, protocols);
    };
    WS.prototype = NativeWS.prototype;
    ["CONNECTING", "OPEN", "CLOSING", "CLOSED"].forEach(function (n) { WS[n] = NativeWS[n]; });
    window.WebSocket = WS;
  }

  if (window.open) {
    var nativeOpenWin = window.open;
    window.open = function (url) {
      var kind = url ? isBlocked(url) : null;
      var gesture = !navigator.userActivation || navigator.userActivation.isActive;
      if (kind || (!gesture && AD_ON)) { report(kind || "ad", url || "popup"); return null; }
      return nativeOpenWin.apply(window, arguments);
    };
  }

  // ---------------------------------------------------
  // DOM: block ad scripts / images / frames before they load
  // ---------------------------------------------------
  var TAGS = { SCRIPT: "src", IMG: "src", IFRAME: "src", LINK: "href", EMBED: "src", SOURCE: "src", VIDEO: "src", AUDIO: "src" };

  function hookProp(proto, prop) {
    try {
      var d = Object.getOwnPropertyDescriptor(proto, prop);
      if (!d || !d.set) return;
      Object.defineProperty(proto, prop, {
        configurable: true, enumerable: d.enumerable, get: d.get,
        set: function (v) {
          var kind = isBlocked(v);
          if (kind) { report(kind, v); return; }
          d.set.call(this, v);
        }
      });
    } catch (e) {}
  }
  hookProp(HTMLScriptElement.prototype, "src");
  hookProp(HTMLImageElement.prototype, "src");
  hookProp(HTMLIFrameElement.prototype, "src");
  hookProp(HTMLLinkElement.prototype, "href");
  if (window.HTMLEmbedElement) hookProp(HTMLEmbedElement.prototype, "src");
  if (window.HTMLSourceElement) hookProp(HTMLSourceElement.prototype, "src");

  var nativeSetAttr = Element.prototype.setAttribute;
  Element.prototype.setAttribute = function (name, value) {
    if (typeof name === "string") {
      var n = name.toLowerCase();
      if ((n === "src" || n === "href") && TAGS[this.tagName] === n) {
        var kind = isBlocked(value);
        if (kind) { report(kind, value); return; }
      }
    }
    return nativeSetAttr.apply(this, arguments);
  };

  function strip(el) {
    if (!el || el.nodeType !== 1) return;
    var attr = TAGS[el.tagName];
    if (!attr) return;
    var url = el.getAttribute(attr);
    if (!url) return;
    var kind = isBlocked(url);
    if (!kind) return;
    report(kind, url);
    el.removeAttribute(attr);
    if (el.parentNode && (el.tagName === "SCRIPT" || el.tagName === "IFRAME" || el.tagName === "IMG")) {
      el.parentNode.removeChild(el);
    }
  }

  var SELECTOR = "script[src], img[src], iframe[src], link[href], embed[src], source[src]";
  function scan(root) {
    var list = root.querySelectorAll(SELECTOR);
    for (var i = 0; i < list.length; i++) strip(list[i]);
  }

  if (!SELF_EXEMPT) {
    var observer = new MutationObserver(function (mutations) {
      for (var i = 0; i < mutations.length; i++) {
        var added = mutations[i].addedNodes;
        for (var j = 0; j < added.length; j++) {
          var node = added[j];
          if (node.nodeType !== 1) continue;
          strip(node);
          if (node.firstElementChild) scan(node);
        }
      }
    });
    var start = function () {
      scan(document);
      observer.observe(document.documentElement || document, { childList: true, subtree: true });
    };
    if (document.documentElement) start();
    else document.addEventListener("DOMContentLoaded", start);
  }

  // ---------------------------------------------------
  // COSMETIC: hide leftover ad containers
  // ---------------------------------------------------
  var COSMETIC = [
    "ins.adsbygoogle", ".adsbygoogle", "[id^='google_ads_iframe']", "[id^='google_ads_']", "[id^='div-gpt-ad']",
    "iframe[id^='google_ads']", "iframe[src*='doubleclick.net']", "iframe[src*='googlesyndication.com']",
    "[data-ad-slot]", "[data-ad-client]", "[data-adunit]", "[data-google-query-id]",
    ".ad-slot", ".ad-unit", ".ad-banner", ".ad-container", ".ad-wrapper", ".ad-box", ".adbox", ".adsbox",
    ".advert", ".advertisement", ".advertising", ".ads-container", ".ads-banner", ".banner-ads", ".top-ads", ".sidebar-ads",
    ".text-ads", ".textad", ".sponsored-ad", ".sponsored-links", "[aria-label='Advertisement']", "[aria-label='advertisement']",
    "#taboola-below-article", "[id^='taboola-']", ".taboola", ".trc_rbox_container", ".OUTBRAIN", ".ob-widget", "[data-widget-id^='AR_']",
    ".mgid_container", "[id^='mgid_']", ".revcontent", "[id^='rcjsload']", "#bottomads", "#tads", "#tadsb", "#taw .commercial-unit-desktop-top",
    ".b_ad", "li.b_ad", ".result--ad", "[data-testid='ad']", ".ads-ad", ".pla-unit", "#rhs-ads", ".cookie-ads"
  ];

  function injectCSS(css) {
    var s = document.createElement("style");
    s.setAttribute("data-sparrow-adblock", "1");
    s.textContent = css;
    (document.head || document.documentElement).appendChild(s);
  }

  function cosmetic() {
    var css = COSMETIC.join(",") + "{display:none!important;visibility:hidden!important;height:0!important;min-height:0!important;margin:0!important;padding:0!important;}";
    if (document.documentElement) injectCSS(css);
    else document.addEventListener("DOMContentLoaded", function () { injectCSS(css); });
  }
  if (AD_ON && !SELF_EXEMPT) cosmetic();

  // ---------------------------------------------------
  // YOUTUBE: strip ad payloads, hide ad UI, skip any ad that slips through
  // ---------------------------------------------------
  if (AD_ON && /(^|\.)youtube\.com$/.test(pageHost)) {
    var AD_KEYS = ["adPlacements", "playerAds", "adSlots", "adBreakHeartbeatParams"];
    var stripAds = function (obj) {
      if (!obj || typeof obj !== "object") return obj;
      for (var i = 0; i < AD_KEYS.length; i++) {
        if (AD_KEYS[i] in obj) { try { delete obj[AD_KEYS[i]]; } catch (e) { obj[AD_KEYS[i]] = undefined; } }
      }
      if (obj.playerResponse) stripAds(obj.playerResponse);
      return obj;
    };

    var nativeParse = JSON.parse;
    JSON.parse = function () { return stripAds(nativeParse.apply(this, arguments)); };

    if (window.Response && Response.prototype.json) {
      var nativeJson = Response.prototype.json;
      Response.prototype.json = function () {
        var url = this.url || "";
        var p = nativeJson.apply(this, arguments);
        if (url.indexOf("/youtubei/v1/") === -1) return p;
        return p.then(function (data) { return stripAds(data); });
      };
    }

    try {
      var initial;
      Object.defineProperty(window, "ytInitialPlayerResponse", {
        configurable: true,
        get: function () { return initial; },
        set: function (v) { initial = stripAds(v); }
      });
    } catch (e) {}

    var YT_CSS = [
      ".ytp-ad-module", ".ytp-ad-overlay-container", ".ytp-ad-overlay-slot", ".ytp-ad-image-overlay",
      "ytd-ad-slot-renderer", "ytd-in-feed-ad-layout-renderer", "ytd-display-ad-renderer", "ytd-promoted-sparkles-web-renderer",
      "ytd-promoted-video-renderer", "ytd-banner-promo-renderer", "ytd-statement-banner-renderer", "ytd-companion-slot-renderer",
      "ytd-action-companion-ad-renderer", "ytd-player-legacy-desktop-watch-ads-renderer", "#masthead-ad", "#player-ads",
      "ytd-engagement-panel-section-list-renderer[target-id='engagement-panel-ads']", ".ytp-ad-player-overlay",
      "ytd-rich-item-renderer:has(ytd-ad-slot-renderer)", "ytd-rich-section-renderer:has(ytd-statement-banner-renderer)"
    ].join(",") + "{display:none!important;}";
    if (document.documentElement) injectCSS(YT_CSS);
    else document.addEventListener("DOMContentLoaded", function () { injectCSS(YT_CSS); });

    var adWasPlaying = false, savedMuted = false;
    setInterval(function () {
      if (document.hidden) return;
      var player = document.querySelector(".html5-video-player");
      var video = document.querySelector("video.html5-main-video") || document.querySelector("video");
      var skip = document.querySelector(".ytp-ad-skip-button, .ytp-ad-skip-button-modern, .ytp-skip-ad-button, .ytp-ad-skip-button-container button");
      if (skip) { try { skip.click(); } catch (e) {} }
      var close = document.querySelector(".ytp-ad-overlay-close-button");
      if (close) { try { close.click(); } catch (e) {} }
      var adShowing = !!(player && player.classList.contains("ad-showing"));
      if (adShowing && video) {
        if (!adWasPlaying) savedMuted = video.muted;
        adWasPlaying = true;
        video.muted = true;
        video.playbackRate = 16;
        if (isFinite(video.duration) && video.duration > 0) video.currentTime = video.duration;
      } else if (adWasPlaying && video) {
        adWasPlaying = false;
        video.muted = savedMuted;
        video.playbackRate = 1;
      }
    }, 400);
  }
})();`

func boolJS(b bool) string {
	if b {
		return "true"
	}

	return "false"
}
