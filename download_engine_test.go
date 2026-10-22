package main

import "testing"

func TestIsDownloadableURLIgnoresQueryAndFragment(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "query", url: "https://example.com/file.zip?token=123", want: true},
		{name: "fragment", url: "https://example.com/file.pdf#download", want: true},
		{name: "case insensitive", url: "https://example.com/FILE.MSI", want: true},
		{name: "ordinary page", url: "https://example.com/download?id=file.zip", want: false},
		{name: "invalid URL", url: "http://[::1", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsDownloadableURL(test.url); got != test.want {
				t.Errorf("IsDownloadableURL(%q) = %t, want %t", test.url, got, test.want)
			}
		})
	}
}
