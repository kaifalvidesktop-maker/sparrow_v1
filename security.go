package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
)

// ---------------------------------------------------
// SECURITY SHIELD SYSTEM
// ---------------------------------------------------

type SecurityShield struct {
	mu             sync.Mutex
	AdBlockOn      bool
	TrackerBlockOn bool
	HTTPSOnlyOn    bool
	BlockedCount   int
}

var shield = &SecurityShield{
	AdBlockOn:      true,
	TrackerBlockOn: true,
	HTTPSOnlyOn:    true,
	BlockedCount:   0,
}

// Known Ad Domains (basic starter list, will grow later)
var adBlocklist = []string{
	"doubleclick.net",
	"googlesyndication.com",
	"googleadservices.com",
	"google-analytics.com",
	"googletagmanager.com",
	"adservice.google.com",
	"pagead2.googlesyndication.com",
	"partner.googleadservices.com",
	"securepubads.g.doubleclick.net",
	"stats.g.doubleclick.net",
	"adnxs.com",
	"adsrvr.org",
	"adform.net",
	"advertising.com",
	"adtech.com",
	"adroll.com",
	"amazon-adsystem.com",
	"amazonadservices.com",
	"criteo.com",
	"criteo.net",
	"outbrain.com",
	"taboola.com",
	"pubmatic.com",
	"rubiconproject.com",
	"openx.net",
	"indexexchange.com",
	"casalemedia.com",
	"33across.com",
	"appnexus.com",
	"sharethrough.com",
	"sovrn.com",
	"lijit.com",
	"gumgum.com",
	"triplelift.com",
	"smartadserver.com",
	"yieldmo.com",
	"media.net",
	"contextweb.com",
	"sonobi.com",
	"revcontent.com",
	"mgid.com",
	"zedo.com",
	"undertone.com",
	"districtm.io",
	"emxdigital.com",
	"smaato.net",
	"inmobi.com",
	"mopub.com",
	"unityads.unity3d.com",
	"ads-twitter.com",
	"ads.linkedin.com",
	"ads.pinterest.com",
	"ads.snapchat.com",
	"ads.tiktok.com",
	"ads.reddit.com",
	"ads.facebook.com",
	"an.facebook.com",
	"connect.facebook.net",
	"pixel.facebook.com",
	"analytics.facebook.com",
	"bat.bing.com",
	"clarity.ms",
	"bingads.microsoft.com",
	"ads.microsoft.com",
	"advertising.microsoft.com",
	"gemini.google.com",
	"scorecardresearch.com",
	"quantserve.com",
	"bluekai.com",
	"krxd.net",
	"demdex.net",
	"omtrdc.net",
	"2o7.net",
	"everesttech.net",
	"everestjs.net",
	"imrworldwide.com",
	"nielsen.com",
	"nielsen.net",
	"comscore.com",
	"doubleverify.com",
	"integralads.com",
	"adsafeprotected.com",
	"moatads.com",
	"measurementapi.com",
	"branch.io",
	"appsflyer.com",
	"adjust.com",
	"kochava.com",
	"singular.net",
	"app-measurement.com",
	"mixpanel.com",
	"segment.io",
	"segment.com",
	"amplitude.com",
	"hotjar.com",
	"mouseflow.com",
	"fullstory.com",
	"luckyorange.com",
	"inspectlet.com",
	"smartlook.com",
	"logrocket.com",
	"heap.io",
	"heapanalytics.com",
	"crazyegg.com",
	"clicky.com",
	"statcounter.com",
	"chartbeat.com",
	"parsely.com",
	"newrelic.com",
	"bugsnag.com",
	"sentry.io",
	"googletagservices.com",
	"adzerk.net",
	"adskeeper.co.uk",
	"adskeeper.com",
	"bidvertiser.com",
	"clickadu.com",
	"clickaine.com",
	"clicktripz.com",
	"exoclick.com",
	"juicyads.com",
	"popads.net",
	"popcash.net",
	"propellerads.com",
	"adcash.com",
	"adsterra.com",
	"hilltopads.net",
	"trafficjunky.net",
	"trafficfactory.biz",
	"plugrush.com",
	"ero-advertising.com",
	"onclickads.net",
	"ad-maven.com",
	"admaven.com",
	"pushground.com",
	"richpush.co",
	"evadav.com",
	"galaksion.com",
	"admixer.net",
	"smartyads.com",
	"bidmatic.io",
	"adkernel.com",
	"adtarget.com",
	"adcolony.com",
	"chartboost.com",
	"vungle.com",
	"ironsource.com",
	"applovin.com",
	"liftoff.io",
	"mintegral.com",
	"fyber.com",
	"tapjoy.com",
	"startapp.com",
	"inmobi.net",
	"mng-ads.com",
	"airpush.com",
	"leadbolt.com",
	"unity.com",
	"unityads.com",
	"tapdaq.com",
	"ogury.com",
	"hyprmx.com",
	"digitalturbine.com",
	"admob.com",
	"admob.google.com",
	"adsense.google.com",
	"ads.google.com",
	"ads.youtube.com",
	"analytics.google.com",
	"marketingplatform.google.com",
	"doubleclickbygoogle.com",
	"ad-delivery.net",
	"ad.gt",
	"adlooxtracking.com",
	"adentifi.com",
	"adbutler.com",
	"adgear.com",
	"adglare.com",
	"adition.com",
	"admanmedia.com",
	"admarvel.com",
	"admedo.com",
	"adnami.io",
	"adotmob.com",
	"adpushup.com",
	"adrecover.com",
	"adskeeper.net",
	"adsnative.com",
	"adspirit.de",
	"adswizz.com",
	"adthrive.com",
	"adzerk.com",
	"affinity.com",
	"afftrack.com",
	"alcmpn.com",
	"appier.com",
	"bidtellect.com",
	"conversantmedia.com",
	"districtm.net",
	"example.com",
}

var adDomains = map[string]bool{}

func init() {
	for _, domain := range adBlocklist {
		d := normalizeHost(domain)
		if d == "" {
			continue
		}
		adDomains[d] = true
	}
}

// Known Tracker Domains
var trackerDomains = map[string]bool{
	"google-analytics.com":         true,
	"analytics.google.com":         true,
	"googletagmanager.com":         true,
	"googletagservices.com":        true,
	"googleadservices.com":         true,
	"googlesyndication.com":        true,
	"doubleclick.net":              true,
	"2mdn.net":                     true,
	"adnxs.com":                    true,
	"appnexus.com":                 true,
	"adsrvr.org":                   true,
	"adform.net":                   true,
	"advertising.com":              true,
	"adroll.com":                   true,
	"amazon-adsystem.com":          true,
	"bluekai.com":                  true,
	"krxd.net":                     true,
	"demdex.net":                   true,
	"omtrdc.net":                   true,
	"everesttech.net":              true,
	"everestjs.net":                true,
	"scorecardresearch.com":        true,
	"quantserve.com":               true,
	"comscore.com":                 true,
	"nielsen.com":                  true,
	"nielsen.net":                  true,
	"imrworldwide.com":             true,
	"moatads.com":                  true,
	"doubleverify.com":             true,
	"integralads.com":              true,
	"adsafeprotected.com":          true,
	"criteo.com":                   true,
	"criteo.net":                   true,
	"outbrain.com":                 true,
	"taboola.com":                  true,
	"pubmatic.com":                 true,
	"rubiconproject.com":           true,
	"openx.net":                    true,
	"indexexchange.com":            true,
	"casalemedia.com":              true,
	"33across.com":                 true,
	"sharethrough.com":             true,
	"sovrn.com":                    true,
	"lijit.com":                    true,
	"gumgum.com":                   true,
	"triplelift.com":               true,
	"smartadserver.com":            true,
	"yieldmo.com":                  true,
	"media.net":                    true,
	"contextweb.com":               true,
	"sonobi.com":                   true,
	"revcontent.com":               true,
	"mgid.com":                     true,
	"zedo.com":                     true,
	"undertone.com":                true,
	"emxdigital.com":               true,
	"smaato.net":                   true,
	"inmobi.com":                   true,
	"mopub.com":                    true,
	"adcolony.com":                 true,
	"chartboost.com":               true,
	"vungle.com":                   true,
	"ironsource.com":               true,
	"applovin.com":                 true,
	"liftoff.io":                   true,
	"mintegral.com":                true,
	"fyber.com":                    true,
	"tapjoy.com":                   true,
	"startapp.com":                 true,
	"airpush.com":                  true,
	"leadbolt.com":                 true,
	"ogury.com":                    true,
	"hyprmx.com":                   true,
	"branch.io":                    true,
	"appsflyer.com":                true,
	"adjust.com":                   true,
	"kochava.com":                  true,
	"singular.net":                 true,
	"app-measurement.com":          true,
	"mixpanel.com":                 true,
	"segment.com":                  true,
	"segment.io":                   true,
	"amplitude.com":                true,
	"heap.io":                      true,
	"heapanalytics.com":            true,
	"hotjar.com":                   true,
	"mouseflow.com":                true,
	"fullstory.com":                true,
	"luckyorange.com":              true,
	"inspectlet.com":               true,
	"smartlook.com":                true,
	"logrocket.com":                true,
	"crazyegg.com":                 true,
	"clicky.com":                   true,
	"statcounter.com":              true,
	"chartbeat.com":                true,
	"parsely.com":                  true,
	"clarity.ms":                   true,
	"bat.bing.com":                 true,
	"facebook.net":                 true,
	"analytics.twitter.com":        true,
	"connect.facebook.net":         true,
	"pixel.facebook.com":           true,
	"analytics.facebook.com":       true,
	"an.facebook.com":              true,
	"ads.linkedin.com":             true,
	"analytics.linkedin.com":       true,
	"snap.licdn.com":               true,
	"ads.pinterest.com":            true,
	"ads.snapchat.com":             true,
	"tr.snapchat.com":              true,
	"ads.tiktok.com":               true,
	"analytics.tiktok.com":         true,
	"ads.reddit.com":               true,
	"events.reddit.com":            true,
	"pixel.reddit.com":             true,
	"ads.yahoo.com":                true,
	"analytics.yahoo.com":          true,
	"beap.gemini.yahoo.com":        true,
	"geo.yahoo.com":                true,
	"sp.analytics.yahoo.com":       true,
	"advertising.yahoo.com":        true,
	"api.segment.io":               true,
	"cdn.mxpnl.com":                true,
	"api.mixpanel.com":             true,
	"api.amplitude.com":            true,
	"cdn.amplitude.com":            true,
	"statsig.com":                  true,
	"events.statsig.com":           true,
	"datadog.com":                  true,
	"browser-intake-datadoghq.com": true,
	"rum.datadoghq.com":            true,
	"static.hotjar.com":            true,
	"insights.hotjar.com":          true,
	"cdn.mouseflow.com":            true,
	"hn.inspectlet.com":            true,
	"web-sdk.smartlook.com":        true,
	"r.lr-ingest.io":               true,
	"app.logrocket.com":            true,
	"cdn.heapanalytics.com":        true,
	"api.heapanalytics.com":        true,
	"script.crazyegg.com":          true,
	"w1.luckyorange.com":           true,
	"w2.luckyorange.com":           true,
	"pixel.wp.com":                 true,
	"stats.wp.com":                 true,
	"matomo.org":                   true,
	"piwik.org":                    true,
	"plausible.io":                 true,
	"snowplow.io":                  true,
	"snowplowanalytics.com":        true,
	"rudderstack.com":              true,
	"rudderlabs.com":               true,
	"posthog.com":                  true,
	"app.posthog.com":              true,
	"telemetrydeck.com":            true,
	"openpanel.dev":                true,
	"goatcounter.com":              true,
	"simpleanalytics.com":          true,
	"fathomanalytics.com":          true,
	"countly.com":                  true,
	"kissmetrics.com":              true,
	"kissmetrics.io":               true,
	"woopra.com":                   true,
	"sessioncam.com":               true,
	"decibelinsight.net":           true,
	"tealeaf.com":                  true,
	"quantummetric.com":            true,
	"glassbox.com":                 true,
	"contentsquare.com":            true,
	"contentsquare.net":            true,
	"userzoom.com":                 true,
	"permutive.com":                true,
	"neustar.biz":                  true,
	"onetrust.com":                 true,
	"trustarc.com":                 true,
	"cookiebot.com":                true,
	"cookielaw.org":                true,
	"didomi.io":                    true,
	"sourcepoint.com":              true,
	"appier.com":                   true,
	"bidtellect.com":               true,
	"conversantmedia.com":          true,
	"districtm.io":                 true,
	"districtm.net":                true,
	"imonomy.com":                  true,
	"justpremium.com":              true,
	"keywee.co":                    true,
	"kargo.com":                    true,
	"mediavine.com":                true,
	"nativo.com":                   true,
	"nextmillennium.io":            true,
	"optinmonster.com":             true,
	"playwire.com":                 true,
	"quantcast.com":                true,
	"remerge.io":                   true,
	"skimlinks.com":                true,
	"teads.tv":                     true,
	"zemanta.com":                  true,
	"chartbeat.net":                true,
	"crwdcntrl.net":                true,
	"cookiebot-analytics.com":      true,
	"creativecdn.com":              true,
	"clicktale.net":                true,
	"id5-sync.com":                 true,
	"insightexpressai.com":         true,
	"liveramp.com":                 true,
	"mathtag.com":                  true,
	"mookie1.com":                  true,
	"nexac.com":                    true,
	"pippio.com":                   true,
	"rlcdn.com":                    true,
	"sail-horizon.com":             true,
	"semasio.net":                  true,
	"smetrics.com":                 true,
	"turn.com":                     true,
	"tynt.com":                     true,
	"zeotap.com":                   true,
}

// Extract domain from a URL string
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	return strings.ToLower(host)
}

func normalizeHost(host string) string {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "" {
		return ""
	}
	h = strings.Trim(h, ".")
	h = strings.TrimPrefix(h, "[")
	h = strings.TrimSuffix(h, "]")
	if idx := strings.Index(h, "/"); idx >= 0 {
		h = h[:idx]
	}
	if parsedIP := net.ParseIP(h); parsedIP != nil {
		if parsedIP.IsLoopback() || parsedIP.IsPrivate() || parsedIP.IsUnspecified() {
			return ""
		}
		return h
	}
	if h == "localhost" || strings.HasSuffix(h, ".localhost") || h == "127.0.0.1" || h == "::1" {
		return ""
	}
	return h
}

func isLocalHost(host string) bool {
	n := normalizeHost(host)
	return n == "" || n == "localhost" || strings.HasSuffix(n, ".localhost")
}

func IsBlocked(host string) bool {
	h := normalizeHost(host)
	if h == "" || isLocalHost(h) {
		return false
	}
	if _, ok := adDomains[h]; ok {
		return true
	}
	parts := strings.Split(h, ".")
	for i := 1; i < len(parts); i++ {
		candidate := strings.Join(parts[i:], ".")
		if _, ok := adDomains[candidate]; ok {
			return true
		}
	}
	return false
}

// Check if a domain (or its parent domain) is in a block list
func isDomainBlocked(domain string, list map[string]bool) bool {
	domain = normalizeHost(domain)
	if domain == "" {
		return false
	}
	if list[domain] {
		return true
	}
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		sub := strings.Join(parts[i:], ".")
		if list[sub] {
			return true
		}
	}
	return false
}

// Main function: should this request be blocked?
func (s *SecurityShield) ShouldBlock(rawURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	domain := extractDomain(rawURL)
	domain = normalizeHost(domain)

	if s.AdBlockOn && IsBlocked(domain) {
		s.BlockedCount++
		fmt.Println("[AD BLOCKED]", domain)
		return true
	}

	if s.TrackerBlockOn && isDomainBlocked(domain, trackerDomains) {
		s.BlockedCount++
		fmt.Println("[TRACKER BLOCKED]", domain)
		return true
	}

	return false
}

// Force HTTPS on a URL if HTTPS-Only mode is on
func (s *SecurityShield) EnforceHTTPS(rawURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.HTTPSOnlyOn {
		return rawURL
	}

	if strings.HasPrefix(rawURL, "http://") {
		return strings.Replace(rawURL, "http://", "https://", 1)
	}

	return rawURL
}

// Toggle functions for UI switches
func (s *SecurityShield) ToggleAdBlock() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdBlockOn = !s.AdBlockOn
	return s.AdBlockOn
}

func (s *SecurityShield) ToggleTrackerBlock() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TrackerBlockOn = !s.TrackerBlockOn
	return s.TrackerBlockOn
}

func (s *SecurityShield) ToggleHTTPSOnly() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HTTPSOnlyOn = !s.HTTPSOnlyOn
	return s.HTTPSOnlyOn
}

func (s *SecurityShield) ReportBlocked(kind string, rawURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(kind) == "" || strings.TrimSpace(rawURL) == "" {
		return
	}
	s.BlockedCount++
	fmt.Println("[SHIELD BLOCKED]", kind, rawURL)
}

type BlockedReport struct {
	Kind string
	URL  string
	N    int
}

func (s *SecurityShield) ReportBlockedBatch(entries []BlockedReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	groups := 0
	for _, entry := range entries {
		if strings.TrimSpace(entry.Kind) == "" || strings.TrimSpace(entry.URL) == "" {
			continue
		}
		count := entry.N
		if count < 1 {
			count = 1
		}
		if count > 10000 {
			count = 10000
		}
		total += count
		groups++
	}
	if total == 0 {
		return
	}
	s.BlockedCount += total
	fmt.Println("[SHIELD BLOCKED]", total, "request(s) in", groups, "group(s)")
}

// Get current shield stats (for UI display)
func (s *SecurityShield) GetStats() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"adBlock":      s.AdBlockOn,
		"trackerBlock": s.TrackerBlockOn,
		"httpsOnly":    s.HTTPSOnlyOn,
		"blockedCount": s.BlockedCount,
	}
}

// ---------------------------------------------------
// URL VALIDATION / SEARCH LOGIC
// ---------------------------------------------------

// Decide if user typed a URL or a search query
func isPDFURL(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return false
	}

	if strings.HasPrefix(rawURL, "sparrow://pdf") {
		if parsed, err := url.Parse(rawURL); err == nil {
			if pdfURL := parsed.Query().Get("url"); pdfURL != "" {
				return isPDFURL(pdfURL)
			}
		}
	}

	lower := strings.ToLower(rawURL)
	if sharpIdx := strings.IndexAny(lower, "#?"); sharpIdx >= 0 {
		lower = lower[:sharpIdx]
	}
	if !strings.Contains(lower, ".pdf") {
		return false
	}

	if strings.Contains(lower, ".pdf/") || strings.HasSuffix(lower, ".pdf") {
		return true
	}
	return strings.Contains(lower, ".pdf?") || strings.Contains(lower, ".pdf#")
}

func shouldOpenExternalURL(input string) bool {
	input = strings.TrimSpace(input)
	if input == "" || strings.Contains(input, " ") || strings.HasPrefix(input, "sparrow://") {
		return false
	}

	candidate := input
	if strings.HasPrefix(candidate, "http://") {
		candidate = strings.TrimPrefix(candidate, "http://")
	} else if strings.HasPrefix(candidate, "https://") {
		candidate = strings.TrimPrefix(candidate, "https://")
	}

	if candidate == "" || !strings.Contains(candidate, ".") {
		return false
	}

	if idx := strings.IndexAny(candidate, "/?#"); idx >= 0 {
		candidate = candidate[:idx]
	}
	if strings.Contains(candidate, ":") {
		candidate = strings.Split(candidate, ":")[0]
	}
	if candidate == "" || candidate == "localhost" {
		return false
	}

	_, err := net.LookupHost(candidate)
	return err != nil
}

func openExternalURL(rawURL string) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		trimmed = "https://" + trimmed
	}

	if strings.Contains(strings.ToLower(trimmed), "windows") {
		return nil
	}

	cmd := openURLCmd(trimmed)
	return cmd.Start()
}

func resolveInput(input string) string {
	input = strings.TrimSpace(input)

	if input == "" {
		return ""
	}

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return input
	}

	if strings.Contains(input, ".") && !strings.Contains(input, " ") {
		return "https://" + input
	}

	return buildSearchURL(input)
}
