package main

import (
	"encoding/base64"
	"testing"
)

func TestDebounceURL(t *testing.T) {
	bing := "https://www.bing.com/ck/a?u=a1" + base64.RawURLEncoding.EncodeToString([]byte("https://real.com/x?y=1"))
	cases := map[string]string{
		"https://www.google.com/url?q=https%3A%2F%2Freal.com%2Fa&sa=D":     "https://real.com/a",
		"https://l.facebook.com/l.php?u=https%3A%2F%2Freal.com%2Fb&h=zzz":  "https://real.com/b",
		"https://l.instagram.com/?u=https%3A%2F%2Freal.com%2Fc&e=1":        "https://real.com/c",
		"https://out.reddit.com/t3_x?url=https%3A%2F%2Freal.com%2Fd":       "https://real.com/d",
		"https://www.youtube.com/redirect?q=https%3A%2F%2Freal.com%2Fe":    "https://real.com/e",
		"https://duckduckgo.com/l/?uddg=https%3A%2F%2Freal.com%2Ff":        "https://real.com/f",
		bing: "https://real.com/x?y=1",
		"https://www.google.com/amp/s/news.com/story":                      "https://news.com/story",
		"https://news-com.cdn.ampproject.org/c/s/news.com/story?a=1":       "https://news.com/story?a=1",
		"https://www.google.com/url?q=javascript:alert(1)":                 "https://www.google.com/url?q=javascript:alert(1)",
		"https://example.com/page":                                         "https://example.com/page",
		"sparrow://home":                                                   "sparrow://home",
	}
	for in, want := range cases {
		if got := debounceURL(in); got != want {
			t.Errorf("%s\n  got  %s\n  want %s", in, got, want)
		}
	}
}

func TestPrivacy2ScriptRespectsSwitch(t *testing.T) {
	if BuildPrivacy2JS() == "" {
		t.Skip("tracker block default off")
	}
}
