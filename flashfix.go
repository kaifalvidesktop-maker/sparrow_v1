package main

import (
	"net/url"
	"os"
	"strings"

	webview "github.com/webview/webview_go"
)

// ---------------------------------------------------
// BLACK FLASH FIX
// While WebView2 swaps one document for another it paints its default
// background. We make that background white (like Chrome) instead of black.
// Change the value to "FF0B0B0D" if you prefer a dark flash.
// Must run before webview.New(), so init() is the right place.
// ---------------------------------------------------

const webviewBackdropColor = "FFFFFFFF" // AARRGGBB

func init() {
	if os.Getenv("WEBVIEW2_DEFAULT_BACKGROUND_COLOR") == "" {
		os.Setenv("WEBVIEW2_DEFAULT_BACKGROUND_COLOR", webviewBackdropColor)
	}
	// Stops Chromium from treating the window as "hidden" and blanking it
	// for a moment while video/GPU surfaces are re-created (seek, fullscreen).
	const flags = "--disable-features=CalculateNativeWinOcclusion"
	cur := os.Getenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS")
	if !strings.Contains(cur, "CalculateNativeWinOcclusion") {
		os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", strings.TrimSpace(cur+" "+flags))
	}
}

func sameHTTPHost(a, b string) bool {
	ua, errA := url.Parse(strings.TrimSpace(a))
	ub, errB := url.Parse(strings.TrimSpace(b))
	if errA != nil || errB != nil {
		return false
	}
	if ua.Scheme != "http" && ua.Scheme != "https" {
		return false
	}
	return ua.Scheme == ub.Scheme && ua.Host != "" && strings.EqualFold(ua.Host, ub.Host)
}

// stepHistory goes back/forward. Inside the same site (e.g. YouTube video to
// video) it uses the browser's own history, which keeps the page cached and
// avoids a full reload (the black flash). Anything else loads normally.
func stepHistory(w webview.WebView, delta int, from, to string) {
	if delta != 0 && sameHTTPHost(from, to) {
		js := "history.go(" + itoa(delta) + ");"
		w.Dispatch(func() { w.Eval(js) })
		return
	}
	navigateTo(w, to)
}
