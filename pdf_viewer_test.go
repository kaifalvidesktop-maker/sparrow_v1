package main

import (
	"html"
	"strings"
	"testing"
)

func TestPDFViewerHTMLForURL(t *testing.T) {
	rawURL := `https://example.com/report.pdf?token="quoted"&next=%2Fchapter`
	page := pdfViewerHTMLForURL(rawURL)

	if strings.Contains(page, "__SPARROW_PDF_URL__") {
		t.Fatal("PDF source placeholder was not replaced")
	}
	if want := `data-pdf-url="` + html.EscapeString(rawURL) + `"`; !strings.Contains(page, want) {
		t.Fatalf("PDF source was not safely embedded; expected %q", want)
	}
	if !strings.Contains(page, `config.getAttribute('data-pdf-url')`) {
		t.Fatal("PDF viewer does not read the embedded source URL")
	}
}

func TestIsPDFURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "direct pdf url", raw: "https://example.com/report.pdf", want: true},
		{name: "uppercase pdf url", raw: "https://example.com/FILE.PDF?download=1", want: true},
		{name: "html page", raw: "https://example.com/index.html", want: false},
		{name: "internal pdf route", raw: "sparrow://pdf?url=https%3A%2F%2Fexample.com%2Fdoc.pdf", want: true},
		{name: "empty", raw: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPDFURL(tt.raw); got != tt.want {
				t.Fatalf("isPDFURL(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
