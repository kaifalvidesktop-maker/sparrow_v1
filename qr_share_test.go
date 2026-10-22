package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateQRCodeDataURL(t *testing.T) {
	dataURL, err := generateQRCodeDataURL("https://example.com/a?b=c")
	if err != nil {
		t.Fatalf("generate QR code: %v", err)
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("QR code data URL has unexpected prefix: %q", dataURL[:min(len(dataURL), len(prefix))])
	}
	png, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		t.Fatalf("decode QR code data URL: %v", err)
	}
	if len(png) < 8 || string(png[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatal("QR code data URL does not contain a PNG image")
	}
}

func TestGenerateQRCodeDataURLRejectsNonWebURLs(t *testing.T) {
	for _, rawURL := range []string{"", "sparrow://home", "javascript:alert(1)", "example.com"} {
		if _, err := generateQRCodeDataURL(rawURL); err == nil {
			t.Errorf("generateQRCodeDataURL(%q) should reject non-web URLs", rawURL)
		}
	}
}
