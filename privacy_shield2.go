package main

import (
	"encoding/base64"
	"net/url"
	"strings"
)

// ---------------------------------------------------
// PRIVACY SHIELD (round 2)
//  1. fingerprinting protection (canvas / audio / hardware farbling)
//  2. redirect-tracker "debouncing" + AMP-cache unwrapping
//  3. De-AMP (jump from AMP pages to the real page)
//  4. mixed-content auto-upgrade (http subresources on https pages)
//  5. hyperlink-auditing (ping="") blocking
// All follow the "Tracker Block" switch.
// ---------------------------------------------------

func validWebURL(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return "", false
	}
	return u.String(), true
}

// debounceOnce returns the real destination hidden behind a known tracking
// redirect or an AMP cache URL, or "" when raw is neither.
func debounceOnce(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	path := u.Path
	q := u.Query()

	pick := func(keys ...string) string {
		for _, k := range keys {
			if dest, ok := validWebURL(q.Get(k)); ok {
				return dest
			}
		}
		return ""
	}

	switch {
	case host == "google.com" && path == "/url":
		return pick("q", "url")
	case (host == "facebook.com" || host == "l.facebook.com" || host == "lm.facebook.com") && path == "/l.php":
		return pick("u")
	case host == "l.instagram.com":
		return pick("u")
	case host == "out.reddit.com":
		return pick("url")
	case (host == "youtube.com" || host == "m.youtube.com") && path == "/redirect":
		return pick("q")
	case host == "duckduckgo.com" && strings.HasPrefix(path, "/l/"):
		return pick("uddg")
	case host == "linkedin.com" && strings.HasPrefix(path, "/safety/go"):
		return pick("url")
	case host == "steamcommunity.com" && strings.HasPrefix(path, "/linkfilter"):
		return pick("url", "u")
	case host == "vk.com" && path == "/away.php":
		return pick("to")
	case host == "t.umblr.com" && path == "/redirect":
		return pick("z")
	case host == "bing.com" && strings.HasPrefix(path, "/ck/a"):
		v := q.Get("u")
		if strings.HasPrefix(v, "a1") {
			if b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(v[2:], "=")); err == nil {
				if dest, ok := validWebURL(string(b)); ok {
					return dest
				}
			}
		}
	case host == "google.com" && strings.HasPrefix(path, "/amp/s/"):
		return ampOrigin("https://", strings.TrimPrefix(path, "/amp/s/"), u.RawQuery)
	case strings.HasSuffix(host, ".cdn.ampproject.org") && strings.HasPrefix(path, "/c/s/"):
		return ampOrigin("https://", strings.TrimPrefix(path, "/c/s/"), u.RawQuery)
	case strings.HasSuffix(host, ".cdn.ampproject.org") && strings.HasPrefix(path, "/c/"):
		return ampOrigin("http://", strings.TrimPrefix(path, "/c/"), u.RawQuery)
	}
	return ""
}

func ampOrigin(scheme, rest, rawQuery string) string {
	if rest == "" {
		return ""
	}
	dest := scheme + rest
	if rawQuery != "" {
		dest += "?" + rawQuery
	}
	if v, ok := validWebURL(dest); ok {
		return v
	}
	return ""
}

// debounceURL follows up to 3 nested redirect wrappers.
func debounceURL(raw string) string {
	for i := 0; i < 3; i++ {
		next := debounceOnce(raw)
		if next == "" || next == raw {
			break
		}
		raw = next
	}
	return raw
}

func BuildPrivacy2JS() string {
	shield.mu.Lock()
	on := shield.TrackerBlockOn
	shield.mu.Unlock()
	if !on {
		return ""
	}
	return privacy2Script
}

const privacy2Script = `(function () {
  "use strict";
  if (window.__sparrowPrivacy2) return;
  window.__sparrowPrivacy2 = true;
  if (!/^https?:$/.test(location.protocol)) return;

  var root = document;

  // ---- 5. hyperlink auditing: drop ping="" before the browser can send it
  function dropPing(e) {
    var t = e.target;
    var a = t && t.closest ? t.closest("a[ping], area[ping]") : null;
    if (a) a.removeAttribute("ping");
  }
  ["mousedown", "click", "auxclick", "keydown", "touchstart"].forEach(function (n) {
    root.addEventListener(n, dropPing, true);
  });

  // ---- 4. mixed content: upgrade http:// subresources on https pages
  if (location.protocol === "https:") {
    var meta = document.createElement("meta");
    meta.httpEquiv = "Content-Security-Policy";
    meta.content = "upgrade-insecure-requests";
    var addMeta = function () {
      if (!document.head) return false;
      document.head.insertBefore(meta, document.head.firstChild);
      return true;
    };
    if (!addMeta()) {
      var mo = new MutationObserver(function () { if (addMeta()) mo.disconnect(); });
      mo.observe(document, { childList: true, subtree: true });
    }
  }

  // ---- 3. De-AMP
  function deAmp() {
    var html = document.documentElement;
    if (!html || !(html.hasAttribute("amp") || html.hasAttribute("\u26A1"))) return;
    var link = document.querySelector('link[rel="canonical"]');
    if (!link || !link.href) return;
    var target;
    try { target = new URL(link.href, location.href); } catch (e) { return; }
    if (!/^https?:$/.test(target.protocol) || target.href === location.href) return;
    var key = "__sparrow_deamp:" + target.href;
    try {
      if (sessionStorage.getItem(key)) return; // loop guard
      sessionStorage.setItem(key, "1");
    } catch (e) {}
    location.replace(target.href);
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", deAmp);
  else deAmp();

  // ---- 1. fingerprinting protection ("farbling"): tiny, stable-per-page noise
  var seed = (Math.random() * 4294967295) >>> 0 || 1;
  function rng(s) {
    return function () {
      s |= 0; s = (s + 0x6D2B79F5) | 0;
      var t = Math.imul(s ^ (s >>> 15), 1 | s);
      t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
      return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
  }

  function farbleBytes(data) {
    var pixels = data.length >> 2;
    if (pixels < 64) return;
    var r = rng(seed ^ data.length);
    var n = Math.min(24, Math.max(2, pixels >> 8));
    for (var i = 0; i < n; i++) {
      var idx = (Math.floor(r() * pixels) << 2) + Math.floor(r() * 3);
      data[idx] ^= 1;
    }
  }

  function define(obj, name, value) {
    try { Object.defineProperty(obj, name, { get: function () { return value; }, configurable: true }); } catch (e) {}
  }
  define(Navigator.prototype, "hardwareConcurrency", 4);
  define(Navigator.prototype, "deviceMemory", 8);

  if (window.HTMLCanvasElement && window.CanvasRenderingContext2D) {
    var nativeGetImageData = CanvasRenderingContext2D.prototype.getImageData;
    var nativeToDataURL = HTMLCanvasElement.prototype.toDataURL;
    var nativeToBlob = HTMLCanvasElement.prototype.toBlob;

    CanvasRenderingContext2D.prototype.getImageData = function () {
      var img = nativeGetImageData.apply(this, arguments);
      try { farbleBytes(img.data); } catch (e) {}
      return img;
    };

    var farbledCopy = function (canvas) {
      var w = canvas.width, h = canvas.height;
      if (!w || !h || w * h < 64 || w * h > 8000000) return null;
      try {
        var copy = document.createElement("canvas");
        copy.width = w; copy.height = h;
        var ctx = copy.getContext("2d");
        ctx.drawImage(canvas, 0, 0);
        var img = nativeGetImageData.call(ctx, 0, 0, w, h);
        farbleBytes(img.data);
        ctx.putImageData(img, 0, 0);
        return copy;
      } catch (e) { return null; }
    };

    HTMLCanvasElement.prototype.toDataURL = function () {
      var copy = farbledCopy(this);
      return nativeToDataURL.apply(copy || this, arguments);
    };
    HTMLCanvasElement.prototype.toBlob = function () {
      var copy = farbledCopy(this);
      return nativeToBlob.apply(copy || this, arguments);
    };
  }

  if (window.AudioBuffer) {
    var nativeChannel = AudioBuffer.prototype.getChannelData;
    var touched = new WeakSet();
    AudioBuffer.prototype.getChannelData = function () {
      var arr = nativeChannel.apply(this, arguments);
      if (!touched.has(arr)) {
        touched.add(arr);
        var r = rng(seed ^ arr.length);
        for (var i = 0; i < arr.length; i += 97) arr[i] += (r() - 0.5) * 1e-7;
      }
      return arr;
    };
  }
  if (window.AnalyserNode) {
    var nativeFloat = AnalyserNode.prototype.getFloatFrequencyData;
    AnalyserNode.prototype.getFloatFrequencyData = function (arr) {
      nativeFloat.apply(this, arguments);
      var r = rng(seed ^ arr.length);
      for (var i = 0; i < arr.length; i += 7) arr[i] += (r() - 0.5) * 0.01;
    };
  }
})();`
