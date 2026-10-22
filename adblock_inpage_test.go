package main

import (
	"strings"
	"testing"
)

func TestBuildAdblockJSIncludesExtrasAndHooks(t *testing.T) {
	js := BuildAdblockJS()
	for _, want := range []string{`"doubleclick.net":true`, `"hotjar.com":true`, `"zeotap.com":true`, "sendBeacon", "ytInitialPlayerResponse", "SELF_EXEMPT"} {
		if !strings.Contains(js, want) {
			t.Errorf("adblock script missing %q", want)
		}
	}
	if strings.Contains(js, "__AD_LIST__") || strings.Contains(js, "__AD_ON__") {
		t.Error("template placeholders left in script")
	}
}

func TestExtraDomainsAreNormalizedAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range extraAdTrackerDomains {
		if d != normalizeHost(d) {
			t.Errorf("%q is not normalized", d)
		}
		if seen[d] {
			t.Errorf("duplicate %q", d)
		}
		seen[d] = true
	}
}
