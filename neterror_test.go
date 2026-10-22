package main

import (
	"strings"
	"testing"
)

func TestClassifyNetError(t *testing.T) {
	cases := map[string]string{
		"":                                  "generic",
		"ERR_INTERNET_DISCONNECTED":         "offline",
		"DNS_PROBE_FINISHED_NO_INTERNET":    "offline",
		"ERR_NAME_NOT_RESOLVED":             "dns",
		"DNS_PROBE_FINISHED_NXDOMAIN":       "dns",
		"ERR_CONNECTION_TIMED_OUT":          "timeout",
		"ERR_CONNECTION_REFUSED":            "refused",
		"ERR_CERT_AUTHORITY_INVALID":        "ssl",
		"ERR_SOMETHING_ELSE":                "generic",
	}
	for code, want := range cases {
		if got := classifyNetError(code); got != want {
			t.Errorf("%q: got %s want %s", code, got, want)
		}
	}
}

func TestErrorPageEscapesTargetAndHidesEdge(t *testing.T) {
	page := buildErrorPageForCode(`https://x.com/"><script>alert(1)</script>`, "ERR_NAME_NOT_RESOLVED")
	if strings.Contains(page, "<script>alert(1)") {
		t.Fatal("target not escaped")
	}
	if !strings.Contains(page, "Site Can't Be Found") && !strings.Contains(page, "Site Can&#39;t Be Found") {
		t.Fatal("wrong page kind")
	}
	if extractErrCode("blah dns_probe_finished_nxdomain blah") != "DNS_PROBE_FINISHED_NXDOMAIN" {
		t.Fatal("code extraction failed")
	}
}
