package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/skip2/go-qrcode"
)

func generateQRCodeDataURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("QR sharing requires a valid HTTP or HTTPS URL")
	}

	png, err := qrcode.Encode(rawURL, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("generate QR code: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}
