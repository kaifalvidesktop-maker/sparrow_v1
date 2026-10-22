package main

import "testing"

func TestProbeHost(t *testing.T) {
	cases := map[string]string{
		"https://www.example.com/path": "www.example.com",
		"http://Example.COM":           "example.com",
		"https://localhost:8080":       "",
		"http://127.0.0.1/x":           "",
		"https://[::1]/":               "",
		"https://printer.local/":       "",
		"sparrow://home":               "",
		"file:///C:/a.pdf":             "",
		"data:text/html,hi":            "",
		"not a url":                    "",
	}
	for in, want := range cases {
		if got := probeHost(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}
