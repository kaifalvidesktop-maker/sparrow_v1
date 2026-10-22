package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------
// SAVE PAGE AS TEXT (exports visible page text to a .txt file
// in the downloads folder — a simple, dependency-free way to
// keep a copy of an article without needing a PDF engine)
// ---------------------------------------------------

// SavePageText writes extracted page text to disk and returns the file path
func SavePageText(pageTitle string, pageText string) (string, error) {
	if err := os.MkdirAll(settings.DownloadPath, 0755); err != nil {
		return "", err
	}

	safeTitle := sanitizeFileName(pageTitle)
	if safeTitle == "" {
		safeTitle = "page"
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := safeTitle + "_" + timestamp + ".txt"
	fullPath := filepath.Join(settings.DownloadPath, fileName)

	content := pageTitle + "\n" + strings.Repeat("=", len(pageTitle)) + "\n\n" + pageText

	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

// Injectable JS that extracts visible page text for export
const extractPageTextJS = `
(function() {
	var title = document.title || 'Untitled Page';
	var text = document.body ? document.body.innerText : '';
	return JSON.stringify({ title: title, text: text });
})();
`
