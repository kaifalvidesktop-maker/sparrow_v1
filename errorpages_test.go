package main

import (
	"html"
	"strings"
	"testing"
)

func TestBuildOfflinePageHTMLSafelyEmbedsRetryTarget(t *testing.T) {
	target := `https://example.com/?query=" onmouseover="alert(1)&next=page`
	page := buildOfflinePageHTML(target)

	if !strings.Contains(page, `data-target="`+html.EscapeString(target)+`"`) {
		t.Fatal("offline page does not preserve the escaped retry target")
	}
	if strings.Contains(page, `data-target="`+target+`"`) {
		t.Fatal("offline page inserted the retry target without escaping it")
	}
	if !strings.Contains(page, "window.realNavigate(document.body.getAttribute('data-target'))") {
		t.Fatal("offline page retry button does not retry the requested URL")
	}
}
