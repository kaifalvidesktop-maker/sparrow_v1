package main

import (
	"net/url"
	"strings"
)

// ---------------------------------------------------
// PRIVACY SHIELD (round 1)
//  1. strip tracking parameters from URLs (utm_*, fbclid, gclid ...)
//  2. send Global Privacy Control / Do Not Track signals
//  3. hide cookie-consent banners
// All three follow the "Tracker Block" switch.
// ---------------------------------------------------

var trackingParamNames = map[string]bool{
	"fbclid": true, "gclid": true, "gclsrc": true, "dclid": true, "gbraid": true, "wbraid": true,
	"msclkid": true, "yclid": true, "twclid": true, "ttclid": true, "li_fat_id": true,
	"mc_eid": true, "mc_cid": true, "igshid": true, "igsh": true,
	"_hsenc": true, "_hsmi": true, "__hssc": true, "__hstc": true, "__hsfp": true, "hsctatracking": true,
	"mkt_tok": true, "vero_id": true, "vero_conv": true, "oly_enc_id": true, "oly_anon_id": true,
	"s_kwcid": true, "ef_id": true, "_openstat": true, "rb_clickid": true, "ml_subscriber": true,
	"ml_subscriber_hash": true, "wickedid": true, "_ga": true, "_gl": true, "ref_src": true, "ref_url": true,
	"srsltid": true,
}

// parameters that only count as tracking on specific hosts
var hostTrackingParams = map[string][]string{
	"youtube.com":      {"si", "pp"},
	"youtu.be":         {"si"},
	"open.spotify.com": {"si"},
	"twitter.com":      {"s", "t"},
	"x.com":            {"s", "t"},
	"facebook.com":     {"mibextid"},
	"pinterest.com":    {"epik"},
	"pin.it":           {"epik"},
	"snapchat.com":     {"scid"},
}

func isTrackingParam(name, host string) bool {
	n := strings.ToLower(name)
	if strings.HasPrefix(n, "utm_") || trackingParamNames[n] {
		return true
	}
	for h, params := range hostTrackingParams {
		if host == h || strings.HasSuffix(host, "."+h) {
			for _, p := range params {
				if n == p {
					return true
				}
			}
		}
	}
	return false
}

// cleanTrackingURL removes tracking parameters from http(s) URLs.
// Other URLs and URLs without tracking parameters are returned unchanged.
func cleanTrackingURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.RawQuery == "" {
		return raw
	}
	host := strings.ToLower(u.Hostname())
	kept := make([]string, 0, 4)
	removed := false
	for _, pair := range strings.Split(u.RawQuery, "&") {
		if pair == "" {
			continue
		}
		name := pair
		if i := strings.Index(pair, "="); i >= 0 {
			name = pair[:i]
		}
		if decoded, err := url.QueryUnescape(name); err == nil {
			name = decoded
		}
		if isTrackingParam(name, host) {
			removed = true
			continue
		}
		kept = append(kept, pair)
	}
	if !removed {
		return raw
	}
	u.RawQuery = strings.Join(kept, "&")
	return u.String()
}

func BuildPrivacyJS() string {
	shield.mu.Lock()
	on := shield.TrackerBlockOn
	shield.mu.Unlock()
	if !on {
		return ""
	}
	return privacyScript
}

const privacyScript = `(function () {
  "use strict";
  if (window.__sparrowPrivacy) return;
  window.__sparrowPrivacy = true;
  if (!/^https?:$/.test(location.protocol)) return;

  // 2. Global Privacy Control + Do Not Track
  function define(obj, name, value) {
    try { Object.defineProperty(obj, name, { get: function () { return value; }, configurable: true }); } catch (e) {}
  }
  define(Navigator.prototype, "globalPrivacyControl", true);
  define(Navigator.prototype, "doNotTrack", "1");

  // 1. clean tracking parameters out of the address
  var EXACT = {};
  "fbclid gclid gclsrc dclid gbraid wbraid msclkid yclid twclid ttclid li_fat_id mc_eid mc_cid igshid igsh _hsenc _hsmi mkt_tok vero_id oly_enc_id oly_anon_id s_kwcid ef_id _openstat rb_clickid _ga _gl srsltid".split(" ")
    .forEach(function (n) { EXACT[n] = true; });
  var HOST_PARAMS = { "youtube.com": ["si", "pp"], "youtu.be": ["si"], "open.spotify.com": ["si"], "facebook.com": ["mibextid"], "pinterest.com": ["epik"], "pin.it": ["epik"], "snapchat.com": ["scid"] };

  function isTracking(name) {
    var n = String(name).toLowerCase();
    if (n.indexOf("utm_") === 0 || EXACT[n]) return true;
    var host = location.hostname.toLowerCase();
    for (var h in HOST_PARAMS) {
      if (host === h || host.slice(-(h.length + 1)) === "." + h) {
        if (HOST_PARAMS[h].indexOf(n) !== -1) return true;
      }
    }
    return false;
  }

  function cleanAddress() {
    try {
      if (!location.search) return;
      var params = new URLSearchParams(location.search);
      var drop = [];
      params.forEach(function (v, k) { if (isTracking(k)) drop.push(k); });
      if (!drop.length) return;
      drop.forEach(function (k) { params.delete(k); });
      var qs = params.toString();
      history.replaceState(history.state, "", location.pathname + (qs ? "?" + qs : "") + location.hash);
    } catch (e) {}
  }
  cleanAddress();

  // 3. cookie-consent banners
  var BANNERS = [
    "#onetrust-consent-sdk", "#onetrust-banner-sdk", ".onetrust-pc-dark-filter",
    "#CybotCookiebotDialog", "#CybotCookiebotDialogBodyUnderlay", "#CookiebotWidget",
    ".qc-cmp2-container", "#qc-cmp2-container", "#truste-consent-track", ".truste_overlay", ".truste_box_overlay",
    "#didomi-host", ".didomi-popup-backdrop", "#usercentrics-root", "#usercentrics-cmp-ui",
    ".fc-consent-root", "[id^='sp_message_container']", "iframe[id^='sp_message_iframe']",
    "#cmpbox", ".cmpwrapper", "#cmpwrapper", "#cookie-law-info-bar", "#cookie-notice", ".cookie-notice-container",
    "#cookieconsent", ".cc-window", ".cc-banner", "#CookieConsent", "#cookiescript_injected", "#cookie-banner",
    ".cookie-banner", "#cookieBanner", ".cookie-consent-banner", "#cookie-consent", "#gdpr-cookie-notice",
    ".gdpr-banner", "#gdpr-consent", ".evidon-banner", "#_evidon_banner", "#hs-eu-cookie-confirmation",
    "#iubenda-cs-banner", ".iubenda-cs-container", ".osano-cm-window", ".cky-consent-container", "#cmplz-cookiebanner-container"
  ];
  var css = BANNERS.join(",") + "{display:none!important;visibility:hidden!important;pointer-events:none!important;}";
  var style = document.createElement("style");
  style.setAttribute("data-sparrow-privacy", "1");
  style.textContent = css;
  function attach() { (document.head || document.documentElement).appendChild(style); }
  if (document.documentElement) attach(); else document.addEventListener("DOMContentLoaded", attach);

  // banners often lock page scrolling; undo that once, only if a banner was present
  var LOCK_CLASSES = ["modal-open", "no-scroll", "noscroll", "sp-message-open", "didomi-popup-open", "onetrust-no-scroll", "cookie-modal-open", "overflow-hidden"];
  var tries = 0;
  var timer = setInterval(function () {
    tries++;
    if (tries > 14 || !document.body) { if (tries > 14) clearInterval(timer); return; }
    var found = false;
    try { found = !!document.querySelector(BANNERS.join(",")); } catch (e) {}
    if (!found) return;
    LOCK_CLASSES.forEach(function (c) {
      document.body.classList.remove(c);
      document.documentElement.classList.remove(c);
    });
    document.body.style.setProperty("overflow", "auto", "important");
    document.documentElement.style.setProperty("overflow", "auto", "important");
  }, 500);
})();`
