package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------
// PAGE SCREENSHOT (captures visible viewport as PNG via
// injected canvas-based JS, then Go saves the base64 data)
// ---------------------------------------------------

// SaveScreenshot decodes a base64 PNG data URL and writes it to disk
func SaveScreenshot(base64Data string, pageTitle string) (string, error) {
	if err := os.MkdirAll(settings.DownloadPath, 0755); err != nil {
		return "", err
	}

	// Strip the "data:image/png;base64," prefix if present
	cleanData := base64Data
	if idx := strings.Index(base64Data, ","); idx != -1 {
		cleanData = base64Data[idx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(cleanData)
	if err != nil {
		return "", err
	}

	safeTitle := sanitizeFileName(pageTitle)
	if safeTitle == "" {
		safeTitle = "screenshot"
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := safeTitle + "_" + timestamp + ".png"
	fullPath := filepath.Join(settings.DownloadPath, fileName)

	if err := os.WriteFile(fullPath, decoded, 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

// Injectable JS: renders the visible page area onto a canvas and
// returns a base64 PNG data URL via the bound callback
const screenshotCaptureJS = `
(function() {
	try {
		var w = window.innerWidth, h = window.innerHeight;
		var html = document.documentElement.outerHTML
			.replace(/xmlns="[^"]*"/g, '')
			.replace(/<script[\s\S]*?<\/script>/gi, '');

		var svgData = '<svg xmlns="http://www.w3.org/2000/svg" width="' + w + '" height="' + h + '">' +
			'<foreignObject width="100%" height="100%">' +
			'<div xmlns="http://www.w3.org/1999/xhtml" style="width:' + w + 'px;height:' + h + 'px;overflow:hidden;transform:translate(' +
			(-window.scrollX) + 'px,' + (-window.scrollY) + 'px);">' + html + '</div>' +
			'</foreignObject></svg>';

		var img = new Image();
		var svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
		var url = URL.createObjectURL(svgBlob);

		img.onload = function () {
			var canvas = document.createElement('canvas');
			canvas.width = w;
			canvas.height = h;
			var ctx = canvas.getContext('2d');
			ctx.fillStyle = getComputedStyle(document.body).backgroundColor || '#ffffff';
			ctx.fillRect(0, 0, w, h);
			try {
				ctx.drawImage(img, 0, 0, w, h);
				var dataUrl = canvas.toDataURL('image/png');
				window.reportScreenshot(dataUrl, document.title);
			} catch (e) {
				// Cross-origin/tainted canvas fallback: capture what native print
				// rasterization allows, else report a solid-color snapshot with title.
				ctx.fillStyle = '#888';
				ctx.font = '14px sans-serif';
				ctx.fillText('SPARROW screenshot (limited): ' + document.title, 20, 30);
				window.reportScreenshot(canvas.toDataURL('image/png'), document.title);
			}
			URL.revokeObjectURL(url);
		};
		img.onerror = function () {
			var canvas = document.createElement('canvas');
			canvas.width = w; canvas.height = h;
			var ctx = canvas.getContext('2d');
			ctx.fillStyle = '#fff';
			ctx.fillRect(0, 0, w, h);
			ctx.fillStyle = '#888';
			ctx.font = '14px sans-serif';
			ctx.fillText('SPARROW screenshot: ' + document.title, 20, 30);
			window.reportScreenshot(canvas.toDataURL('image/png'), document.title);
			URL.revokeObjectURL(url);
		};
		img.src = url;
	} catch (e) {
		console.log('Screenshot failed:', e);
	}
})();
`
