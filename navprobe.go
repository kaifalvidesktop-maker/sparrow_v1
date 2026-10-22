package main

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	webview "github.com/webview/webview_go"
)

// ---------------------------------------------------
// NETWORK PROBE
// WebView2 shows Edge's own "can't reach this site" page when a load fails.
// Before an http(s) page is opened we check, off the UI thread, whether the
// host can be resolved. If the machine is offline or the name does not exist,
// the custom SPARROW error page is shown instead of Edge's page.
// ---------------------------------------------------

var navSeq int64

// probeHost returns the host to check, or "" when the target must not be probed.
func probeHost(target string) string {
	u, err := url.Parse(strings.TrimSpace(target))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".localhost") || net.ParseIP(host) != nil {
		return ""
	}
	return host
}

// internetReachable reports whether any well-known public endpoint answers.
func internetReachable() bool {
	endpoints := []string{"1.1.1.1:443", "8.8.8.8:443", "9.9.9.9:443", "208.67.222.222:443"}
	ok := make(chan bool, len(endpoints))
	for _, ep := range endpoints {
		go func(addr string) {
			c, err := net.DialTimeout("tcp", addr, 1500*time.Millisecond)
			if err == nil {
				c.Close()
			}
			ok <- err == nil
		}(ep)
	}
	for range endpoints {
		if <-ok {
			return true
		}
	}
	return false
}

// probeNetwork returns an error-page kind ("offline" or "dns") plus the Chrome
// style error code, or "" when the page should simply be loaded.
func probeNetwork(host string) (kind, code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := net.DefaultResolver.LookupHost(ctx, host)
	if err == nil {
		return "", ""
	}
	if !internetReachable() {
		return "offline", "ERR_INTERNET_DISCONNECTED"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return "dns", "ERR_NAME_NOT_RESOLVED"
	}
	return "", "" // online but resolver hiccup: let the browser try
}

// navigateExternal opens an http(s)/other external URL, replacing Edge's
// default error pages with SPARROW's own.
func navigateExternal(w webview.WebView, target string) {
	shield.mu.Lock()
	cleanLinks := shield.TrackerBlockOn
	shield.mu.Unlock()
	if cleanLinks {
		target = cleanTrackingURL(debounceURL(target))
	}
	host := probeHost(target)
	if host == "" {
		w.Dispatch(func() { w.Navigate(target) })
		return
	}
	seq := atomic.AddInt64(&navSeq, 1)
	go func() {
		kind, code := probeNetwork(host)
		if atomic.LoadInt64(&navSeq) != seq {
			return // a newer navigation started meanwhile
		}
		if kind != "" {
			page := buildErrorPageHTML(kind, target, code)
			w.Dispatch(func() { w.Navigate(dataURI(page)) })
			return
		}
		w.Dispatch(func() { w.Navigate(target) })
	}()
}
