package main

import (
	"html"
	"regexp"
	"strings"
)

// ---------------------------------------------------
// CUSTOM NETWORK ERROR PAGES (replaces Edge/WebView2 default)
// ---------------------------------------------------

type netErrInfo struct {
	Kind    string
	Title   string
	Message string
	Tips    []string
	Icon    string
}

const (
	iconOffline = `<svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2 8.8a15 15 0 0 1 4.2-2.7M22 8.8a15 15 0 0 0-9.4-3.7M5 12.9a10 10 0 0 1 3.1-2M19 12.9a10 10 0 0 0-4.6-2.6M8.5 16.4a5 5 0 0 1 7 0"/><circle cx="12" cy="20" r="1"/><path d="M3 3l18 18"/></svg>`
	iconSearch  = `<svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3M8.5 8.5l5 5M13.5 8.5l-5 5"/></svg>`
	iconServer  = `<svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="6" rx="1.5"/><rect x="3" y="14" width="18" height="6" rx="1.5"/><path d="M7 7h.01M7 17h.01M13 12v2"/></svg>`
	iconLock    = `<svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="10" width="16" height="11" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3M12 14v3"/></svg>`
	iconWarn    = `<svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3l10 18H2L12 3z"/><path d="M12 10v5M12 18h.01"/></svg>`
)

var chromeErrCodeRe = regexp.MustCompile(`(?i)\b(ERR_[A-Z0-9_]+|DNS_PROBE_[A-Z_]+)\b`)

// classifyNetError maps a Chromium/Edge error code to a page kind.
func classifyNetError(code string) string {
	c := strings.ToUpper(strings.TrimSpace(code))
	switch {
	case c == "":
		return "generic"
	case strings.Contains(c, "INTERNET_DISCONNECTED"),
		strings.Contains(c, "NETWORK_CHANGED"),
		strings.Contains(c, "NO_INTERNET"),
		strings.Contains(c, "NETWORK_ACCESS_DENIED"):
		return "offline"
	case strings.Contains(c, "NAME_NOT_RESOLVED"),
		strings.Contains(c, "NXDOMAIN"),
		strings.Contains(c, "DNS_"):
		return "dns"
	case strings.Contains(c, "TIMED_OUT"):
		return "timeout"
	case strings.Contains(c, "CONNECTION_"),
		strings.Contains(c, "ADDRESS_UNREACHABLE"),
		strings.Contains(c, "EMPTY_RESPONSE"),
		strings.Contains(c, "PROXY_"),
		strings.Contains(c, "TUNNEL_"):
		return "refused"
	case strings.Contains(c, "CERT_"),
		strings.Contains(c, "SSL_"),
		strings.Contains(c, "BAD_SSL"):
		return "ssl"
	}
	return "generic"
}

func netErrorInfo(kind string) netErrInfo {
	switch kind {
	case "offline":
		return netErrInfo{"offline", "No Internet Connection",
			"SPARROW can't reach the network. Check your connection and try again.",
			[]string{"Check your Wi-Fi or Ethernet cable", "Restart your router or modem", "Turn off VPN or proxy and retry"}, iconOffline}
	case "dns":
		return netErrInfo{"dns", "Site Can't Be Found",
			"The server address could not be found. It may be mistyped or the site no longer exists.",
			[]string{"Check the address for typing mistakes", "Try searching for the site instead", "Check your DNS settings"}, iconSearch}
	case "timeout":
		return netErrInfo{"timeout", "Connection Timed Out",
			"The site took too long to respond. It may be down or your connection may be slow.",
			[]string{"Wait a moment and try again", "Check that your connection is stable", "Try again later if the site is busy"}, iconServer}
	case "refused":
		return netErrInfo{"refused", "Can't Connect to Site",
			"The site refused the connection or closed it unexpectedly.",
			[]string{"The site may be temporarily down", "Check firewall, VPN or proxy settings", "Try again in a few minutes"}, iconServer}
	case "ssl":
		return netErrInfo{"ssl", "Connection Not Secure",
			"SPARROW couldn't establish a secure connection to this site, so the page was not loaded.",
			[]string{"Check your computer's date and time", "The site's certificate may be invalid or expired", "Avoid entering passwords on this site"}, iconLock}
	}
	return netErrInfo{"generic", "This Page Can't Be Loaded",
		"Something went wrong while loading this page.",
		[]string{"Try reloading the page", "Check your internet connection", "Try again later"}, iconWarn}
}

const errorPageTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>__TITLE__</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; min-height: 100vh; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', system-ui, sans-serif; display: flex;
    align-items: center; justify-content: center; padding: 24px;
  }
  .card { max-width: 460px; width: 100%; text-align: center; animation: in .35s ease; }
  @keyframes in { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: none; } }
  .icon { color: #6b6b72; margin-bottom: 18px; }
  h1 { font-size: 22px; font-weight: 600; margin: 0 0 10px; }
  p.msg { color: #8a8a90; font-size: 13.5px; line-height: 1.6; margin: 0 0 6px; }
  .url { color: #9a9aa2; font-family: Consolas, monospace; font-size: 12px; word-break: break-all; margin: 10px 0 0; }
  ul { list-style: none; padding: 0; margin: 20px auto 0; text-align: left; display: inline-block; }
  li { color: #8a8a90; font-size: 12.5px; padding: 4px 0 4px 18px; position: relative; }
  li:before { content: ''; position: absolute; left: 4px; top: 11px; width: 4px; height: 4px; border-radius: 50%; background: #3a3a40; }
  .btns { margin-top: 24px; display: flex; gap: 10px; justify-content: center; }
  button {
    background: #232323; color: #ccc; border: 1px solid #2c2c2e; padding: 10px 24px;
    border-radius: 8px; cursor: pointer; font-size: 13px; font-family: inherit;
  }
  button:hover { background: #2a2a2c; color: #fff; }
  button.primary { background: #e8e8e8; color: #0b0b0d; border-color: #e8e8e8; font-weight: 600; }
  button.primary:hover { background: #fff; }
  .status { margin-top: 18px; font-size: 12px; color: #6b6b72; display: __STATUS_DISPLAY__; align-items: center; justify-content: center; gap: 8px; }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: #c0392b; }
  .dot.on { background: #2ecc71; }
  .code { margin-top: 22px; font-family: Consolas, monospace; font-size: 11px; color: #4a4a50; }
</style>
</head>
<body data-target="__TARGET__" data-kind="__KIND__">
  <div class="card">
    <div class="icon">__ICON__</div>
    <h1>__TITLE__</h1>
    <p class="msg">__MESSAGE__</p>
    <div class="url">__TARGET_TEXT__</div>
    <ul>__TIPS__</ul>
    <div class="btns">
      <button class="primary" onclick="window.realNavigate(document.body.getAttribute('data-target'))">Try Again</button>
      <button onclick="(window.goHome ? window.goHome() : window.realNavigate('sparrow://home'))">Go Home</button>
    </div>
    <div class="status" id="st"><span class="dot" id="dot"></span><span id="stt">Waiting for connection&hellip;</span></div>
    <div class="code">__CODE__</div>
  </div>
<script>
(function () {
  var b = document.body;
  if (b.getAttribute('data-kind') !== 'offline') return;
  var dot = document.getElementById('dot'), stt = document.getElementById('stt');
  function upd() {
    var on = navigator.onLine;
    dot.className = 'dot' + (on ? ' on' : '');
    stt.textContent = on ? 'Back online \u2014 reconnecting\u2026' : 'Waiting for connection\u2026';
  }
  window.addEventListener('online', function () {
    upd();
    var t = b.getAttribute('data-target');
    if (t && window.realNavigate) setTimeout(function () { window.realNavigate(t); }, 600);
  });
  window.addEventListener('offline', upd);
  upd();
})();
</script>
</body>
</html>
`

// buildErrorPageHTML renders the custom error page for a kind
// (offline, dns, timeout, refused, ssl, generic).
func buildErrorPageHTML(kind, target, code string) string {
	info := netErrorInfo(kind)
	var tips strings.Builder
	for _, t := range info.Tips {
		tips.WriteString("<li>" + html.EscapeString(t) + "</li>")
	}
	display := "none"
	if info.Kind == "offline" {
		display = "flex"
	}
	codeText := ""
	if code != "" {
		codeText = html.EscapeString(code)
	}
	r := strings.NewReplacer(
		"__TITLE__", html.EscapeString(info.Title),
		"__MESSAGE__", html.EscapeString(info.Message),
		"__ICON__", info.Icon,
		"__TIPS__", tips.String(),
		"__KIND__", info.Kind,
		"__STATUS_DISPLAY__", display,
		"__CODE__", codeText,
		"__TARGET_TEXT__", html.EscapeString(target),
		"__TARGET__", html.EscapeString(target),
	)
	return r.Replace(errorPageTemplate)
}

// buildErrorPageForCode picks the page kind from a raw Edge error code.
func buildErrorPageForCode(target, code string) string {
	return buildErrorPageHTML(classifyNetError(code), target, code)
}

// extractErrCode pulls an ERR_* / DNS_PROBE_* code from arbitrary text.
func extractErrCode(text string) string {
	return strings.ToUpper(chromeErrCodeRe.FindString(text))
}

// errorPageHookJS runs on every document (w.Init). It only acts when
// WebView2 shows its built-in chrome-error:// page: it hides it, reads
// the error code and asks Go to replace it with the SPARROW error page.
const errorPageHookJS = `
(function () {
  if (window.top !== window) return;
  if (String(location.href).indexOf('chrome-error://') !== 0) return;

  try {
    var s = document.createElement('style');
    s.textContent = 'html,body{background:#0b0b0d!important}body *{visibility:hidden!important}';
    (document.head || document.documentElement || document).appendChild(s);
  } catch (e) {}

  var done = false;
  function run() {
    if (done) return;
    var text = '';
    try { text = (document.documentElement.innerText || '') + ' ' + (document.documentElement.textContent || ''); } catch (e) {}
    var m = text.match(/\b(ERR_[A-Z0-9_]+|DNS_PROBE_[A-Z_]+)\b/i);
    var code = m ? m[1].toUpperCase() : '';
    if (typeof window.showNetworkError === 'function') {
      done = true;
      window.showNetworkError(code);
    }
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', run);
  } else {
    run();
  }
  window.addEventListener('load', run);
  setTimeout(run, 300);
})();
`
