package main

import "testing"

func TestCleanTrackingURL(t *testing.T) {
	cases := map[string]string{
		"https://a.com/p?utm_source=x&id=5&fbclid=abc":     "https://a.com/p?id=5",
		"https://a.com/p?utm_medium=m":                     "https://a.com/p",
		"https://a.com/p?id=5&q=go":                        "https://a.com/p?id=5&q=go",
		"https://a.com/p?gclid=1&msclkid=2#frag":           "https://a.com/p#frag",
		"https://youtu.be/abc?si=zzz&t=30":                 "https://youtu.be/abc?t=30",
		"https://www.youtube.com/watch?v=abc&si=zz&pp=yy":  "https://www.youtube.com/watch?v=abc",
		"https://a.com/p?si=keep":                          "https://a.com/p?si=keep",
		"sparrow://home?utm_source=x":                      "sparrow://home?utm_source=x",
		"https://a.com/?q=a%20b&UTM_Campaign=c":            "https://a.com/?q=a%20b",
		"https://shop.example.com/?srsltid=abc&item=1":     "https://shop.example.com/?item=1",
		"https://www.facebook.com/share/?mibextid=abc&x=1": "https://www.facebook.com/share/?x=1",
		"https://www.pinterest.com/pin/123/?epik=abc":      "https://www.pinterest.com/pin/123/",
		"https://www.snapchat.com/add/user?scid=abc&x=1":   "https://www.snapchat.com/add/user?x=1",
		"https://a.com/":                                   "https://a.com/",
	}
	for in, want := range cases {
		if got := cleanTrackingURL(in); got != want {
			t.Errorf("%s\n  got  %s\n  want %s", in, got, want)
		}
	}
}

func TestTrackerDomainListCoversAdditionalTrackers(t *testing.T) {
	for _, host := range []string{
		"collect.chartbeat.net",
		"events.liveramp.com",
		"pixel.smetrics.com",
		"collector.zeotap.com",
	} {
		if !isDomainBlocked(host, trackerDomains) {
			t.Errorf("%s should match the tracker blocklist", host)
		}
	}
	if isDomainBlocked("notchartbeat.net", trackerDomains) {
		t.Error("tracker matching must not use substring matches")
	}
}
