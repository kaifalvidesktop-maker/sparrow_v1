package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalHTMLServerOnlyAllowsLocalHost(t *testing.T) {
	tempDir := t.TempDir()
	page := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(page, []byte("<html><body>hello</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	url, err := localHTMLServer.Start(page)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer localHTMLServer.Stop()

	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Fatalf("URL = %q, want localhost URL", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Host = "example.com"
	client := &http.Client{Timeout: 2 * time.Second}
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatalf("forbidden host request failed unexpectedly: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("host check status = %d, want %d", resp2.StatusCode, http.StatusForbidden)
	}
}

func TestLocalHTMLServerServesSiblingAssets(t *testing.T) {
	tempDir := t.TempDir()
	page := filepath.Join(tempDir, "index.html")
	asset := filepath.Join(tempDir, "app.js")
	if err := os.WriteFile(page, []byte(`<html><script src="/app.js"></script></html>`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(asset, []byte(`window.__sparrowAssetLoaded = true;`), 0644); err != nil {
		t.Fatal(err)
	}

	url, err := localHTMLServer.Start(page)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer localHTMLServer.Stop()

	resp, err := http.Get(url + "app.js")
	if err != nil {
		t.Fatalf("HTTP GET asset failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("asset status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
