package main

import "strings"

// ---------------------------------------------------
// PRIVACY SHIELD (round 3)
//  1. site permissions that really are enforced (camera, microphone,
//     location, notifications) - "Block" in Site Settings now blocks
//  2. ad-tech browser APIs switched off (Topics, Protected Audience,
//     Attribution Reporting, Private State Tokens)
//  3. fingerprinting APIs hidden (Battery, Network Information, UA hints)
//  4. Google One Tap / FedCM sign-in prompts blocked
//  5. social-media widgets / embeds blocked (like buttons, tweet embeds)
// Items 2-5 follow the "Tracker Block" switch; item 1 always runs.
// ---------------------------------------------------

// Lookup returns the stored permissions for a host WITHOUT creating an entry
// (GetFor adds one for every domain it is asked about).
func (pm *PermissionManager) Lookup(host string) map[string]string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	host = strings.ToLower(strings.TrimSpace(host))
	candidates := []string{host, strings.TrimPrefix(host, "www."), "www." + strings.TrimPrefix(host, "www.")}
	parts := strings.Split(strings.TrimPrefix(host, "www."), ".")
	for i := 1; i < len(parts)-1; i++ {
		candidates = append(candidates, strings.Join(parts[i:], "."))
	}
	for _, c := range candidates {
		if p, ok := pm.Sites[c]; ok {
			return pm.toMap(p)
		}
	}
	return map[string]string{
		"camera":        string(PermissionAsk),
		"microphone":    string(PermissionAsk),
		"location":      string(PermissionAsk),
		"notifications": string(PermissionAsk),
	}
}

func BuildPrivacy3JS() string {
	shield.mu.Lock()
	trackers := shield.TrackerBlockOn
	shield.mu.Unlock()
	flag := "false"
	if trackers {
		flag = "true"
	}
	return strings.Replace(privacy3Script, "__SHIELD_ON__", flag, 1)
}

const privacy3Script = `(function () {
  "use strict";
  if (window.__sparrowPrivacy3) return;
  window.__sparrowPrivacy3 = true;
  if (!/^https?:$/.test(location.protocol)) return;

  var SHIELD_ON = __SHIELD_ON__;
  var host = String(location.hostname || "").toLowerCase();
  function hostIs(h) { return host === h || host.slice(-(h.length + 1)) === "." + h; }
  function stub(obj, name, fn) {
    try { if (obj && name in obj) Object.defineProperty(obj, name, { value: fn, configurable: true, writable: true }); } catch (e) {}
  }
  function define(obj, name, value) {
    try { Object.defineProperty(obj, name, { get: function () { return value; }, configurable: true }); } catch (e) {}
  }
  function denied(msg) { return new DOMException(msg || "Permission denied", "NotAllowedError"); }

  // ---------------------------------------------------
  // 1. SITE PERMISSIONS
  // ---------------------------------------------------
  var permsCache = null;
  function perms() {
    if (permsCache) return permsCache;
    var p;
    try {
      p = (typeof window.getPagePermissions === "function") ? window.getPagePermissions(host) : null;
    } catch (e) { p = null; }
    permsCache = Promise.resolve(p).then(function (v) { return v || {}; }, function () { return {}; });
    return permsCache;
  }

  // location
  try {
    var geo = navigator.geolocation;
    if (geo) {
      var nGet = geo.getCurrentPosition.bind(geo);
      var nWatch = geo.watchPosition.bind(geo);
      var nClear = geo.clearWatch.bind(geo);
      var errObj = function () { return { code: 1, message: "User denied Geolocation", PERMISSION_DENIED: 1, POSITION_UNAVAILABLE: 2, TIMEOUT: 3 }; };
      var watchMap = {}, watchId = 0;
      geo.getCurrentPosition = function (ok, fail, opts) {
        perms().then(function (p) {
          if (p.location === "deny") { if (fail) try { fail(errObj()); } catch (e) {} }
          else nGet(ok, fail, opts);
        });
      };
      geo.watchPosition = function (ok, fail, opts) {
        var id = ++watchId;
        perms().then(function (p) {
          if (p.location === "deny") { if (fail) try { fail(errObj()); } catch (e) {} }
          else watchMap[id] = nWatch(ok, fail, opts);
        });
        return id;
      };
      geo.clearWatch = function (id) {
        if (watchMap[id] !== undefined) { nClear(watchMap[id]); delete watchMap[id]; }
      };
    }
  } catch (e) {}

  // camera / microphone
  try {
    var md = navigator.mediaDevices;
    if (md && md.getUserMedia) {
      var nGum = md.getUserMedia.bind(md);
      md.getUserMedia = function (c) {
        return perms().then(function (p) {
          if ((c && c.video && p.camera === "deny") || (c && c.audio && p.microphone === "deny")) throw denied();
          return nGum(c);
        });
      };
    }
  } catch (e) {}

  // notifications: only sites explicitly allowed may ask; everyone else is denied quietly
  try {
    if (window.Notification) {
      var NN = window.Notification;
      var nReq = NN.requestPermission;
      NN.requestPermission = function (cb) {
        return perms().then(function (p) {
          if (p.notifications === "allow") return nReq.call(NN, cb);
          if (cb) try { cb("denied"); } catch (e) {}
          return "denied";
        });
      };
      perms().then(function (p) {
        if (p.notifications !== "allow") define(NN, "permission", "denied");
      });
    }
    if (window.PushManager && PushManager.prototype.subscribe) {
      var nSub = PushManager.prototype.subscribe;
      PushManager.prototype.subscribe = function () {
        var self = this, args = arguments;
        return perms().then(function (p) {
          if (p.notifications !== "allow") throw denied();
          return nSub.apply(self, args);
        });
      };
    }
  } catch (e) {}

  if (!SHIELD_ON) return;

  // ---------------------------------------------------
  // 2. AD-TECH APIs (Privacy Sandbox)
  // ---------------------------------------------------
  stub(Document.prototype, "browsingTopics", function () { return Promise.resolve([]); });
  stub(Document.prototype, "interestCohort", function () { return Promise.reject(denied("blocked")); });
  stub(Document.prototype, "hasPrivateToken", function () { return Promise.resolve(false); });
  stub(Document.prototype, "hasRedemptionRecord", function () { return Promise.resolve(false); });
  ["joinAdInterestGroup", "leaveAdInterestGroup", "updateAdInterestGroups", "clearOriginJoinedAdInterestGroups"].forEach(function (n) {
    stub(Navigator.prototype, n, function () { return Promise.resolve(); });
  });
  stub(Navigator.prototype, "runAdAuction", function () { return Promise.resolve(null); });
  stub(Navigator.prototype, "canLoadAdAuctionFencedFrame", function () { return false; });

  var nFetch = window.fetch;
  if (nFetch) {
    window.fetch = function (input, init) {
      if (init && typeof init === "object" &&
          (init.browsingTopics || init.attributionReporting || init.privateToken || init.adAuctionHeaders)) {
        init = Object.assign({}, init);
        delete init.browsingTopics; delete init.attributionReporting; delete init.privateToken; delete init.adAuctionHeaders;
        arguments[1] = init;
      }
      return nFetch.apply(this, arguments);
    };
  }
  var nSet = Element.prototype.setAttribute;
  Element.prototype.setAttribute = function (name) {
    if (typeof name === "string" && name.toLowerCase() === "attributionsrc") return;
    return nSet.apply(this, arguments);
  };
  document.addEventListener("click", function (e) {
    var a = e.target && e.target.closest ? e.target.closest("[attributionsrc]") : null;
    if (a) a.removeAttribute("attributionsrc");
  }, true);

  // ---------------------------------------------------
  // 3. FINGERPRINTING APIs
  // ---------------------------------------------------
  stub(Navigator.prototype, "getBattery", function () { return Promise.reject(denied("Battery status is not available")); });
  if ("connection" in Navigator.prototype) define(Navigator.prototype, "connection", undefined);
  try {
    if (window.NavigatorUAData && NavigatorUAData.prototype.getHighEntropyValues) {
      var nHigh = NavigatorUAData.prototype.getHighEntropyValues;
      NavigatorUAData.prototype.getHighEntropyValues = function () {
        return nHigh.call(this, []); // low-entropy brands / mobile / platform only
      };
    }
  } catch (e) {}

  // ---------------------------------------------------
  // 4. GOOGLE ONE TAP / FedCM SIGN-IN PROMPTS
  // ---------------------------------------------------
  if (!hostIs("google.com") && !hostIs("youtube.com") && !hostIs("gstatic.com")) {
    try {
      if (window.CredentialsContainer && CredentialsContainer.prototype.get) {
        var nCred = CredentialsContainer.prototype.get;
        CredentialsContainer.prototype.get = function (opts) {
          if (opts && opts.identity) return Promise.reject(denied("Sign-in prompt blocked"));
          return nCred.apply(this, arguments);
        };
      }
    } catch (e) {}
  }

  // ---------------------------------------------------
  // 5. SOCIAL WIDGETS / EMBEDS + One Tap containers
  // ---------------------------------------------------
  var SOCIAL_HOSTS = {
    "platform.twitter.com": 1, "syndication.twitter.com": 1, "platform.linkedin.com": 1, "platform.instagram.com": 1,
    "widgets.pinterest.com": 1, "assets.pinterest.com": 1, "platform.tumblr.com": 1, "embed.redditmedia.com": 1,
    "badge.facebook.com": 1, "tweetdeck.twitter.com": 1
  };
  var OWN = ["facebook.com", "twitter.com", "x.com", "linkedin.com", "instagram.com", "pinterest.com", "tumblr.com", "reddit.com", "fb.com"];
  var ownSite = OWN.some(hostIs);

  function socialBlocked(u) {
    if (ownSite || !u) return false;
    var url;
    try { url = new URL(String(u), location.href); } catch (e) { return false; }
    var h = url.hostname.toLowerCase();
    var c = h;
    while (c) {
      if (SOCIAL_HOSTS[c]) return true;
      var dot = c.indexOf(".");
      if (dot < 0) break;
      c = c.slice(dot + 1);
    }
    if ((h === "facebook.com" || h.slice(-13) === ".facebook.com") && /(^|\/)(v\d+\.\d+\/)?plugins\//.test(url.pathname)) return true;
    if (h === "accounts.google.com" && url.pathname.indexOf("/gsi/iframe/select") === 0) return true;
    return false;
  }

  function hookProp(proto, prop) {
    try {
      var d = Object.getOwnPropertyDescriptor(proto, prop);
      if (!d || !d.set) return;
      Object.defineProperty(proto, prop, {
        configurable: true, enumerable: d.enumerable, get: d.get,
        set: function (v) { if (socialBlocked(v)) return; d.set.call(this, v); }
      });
    } catch (e) {}
  }
  hookProp(HTMLIFrameElement.prototype, "src");
  hookProp(HTMLScriptElement.prototype, "src");

  var nSet2 = Element.prototype.setAttribute;
  Element.prototype.setAttribute = function (name, value) {
    if (typeof name === "string" && name.toLowerCase() === "src" &&
        (this.tagName === "IFRAME" || this.tagName === "SCRIPT") && socialBlocked(value)) return;
    return nSet2.apply(this, arguments);
  };

  function sweep(root) {
    var list = root.querySelectorAll("iframe[src], script[src]");
    for (var i = 0; i < list.length; i++) {
      if (socialBlocked(list[i].getAttribute("src")) && list[i].parentNode) list[i].parentNode.removeChild(list[i]);
    }
  }
  var mo = new MutationObserver(function (muts) {
    for (var i = 0; i < muts.length; i++) {
      var added = muts[i].addedNodes;
      for (var j = 0; j < added.length; j++) {
        var n = added[j];
        if (n.nodeType !== 1) continue;
        if ((n.tagName === "IFRAME" || n.tagName === "SCRIPT") && socialBlocked(n.getAttribute("src")) && n.parentNode) n.parentNode.removeChild(n);
        else if (n.firstElementChild) sweep(n);
      }
    }
  });
  var start = function () {
    sweep(document);
    mo.observe(document.documentElement, { childList: true, subtree: true });
  };
  if (document.documentElement) start(); else document.addEventListener("DOMContentLoaded", start);

  var css = "#credential_picker_container,#credential_picker_iframe,iframe[src*='accounts.google.com/gsi/iframe/select']," +
    "iframe[src*='facebook.com/plugins/'],.fb-like,.fb-share-button,.fb-comments,.fb-page,.fb-post,.twitter-follow-button,.twitter-share-button," +
    ".IN-widget,.addthis_toolbox .at-share-btn[class*='facebook']{display:none!important;}";
  var style = document.createElement("style");
  style.setAttribute("data-sparrow-privacy3", "1");
  style.textContent = css;
  var attach = function () { (document.head || document.documentElement).appendChild(style); };
  if (document.documentElement) attach(); else document.addEventListener("DOMContentLoaded", attach);
})();`
