package main

import (
	"strings"
	"testing"
)

func TestReaderModeManagerTogglesPerTab(t *testing.T) {
	manager := &ReaderModeManager{Active: make(map[int]bool)}

	if manager.Toggle(1) != true || !manager.IsActive(1) {
		t.Fatal("reader mode should turn on for the selected tab")
	}
	if manager.IsActive(2) {
		t.Fatal("reader mode state should not leak to another tab")
	}
	if manager.Toggle(1) != false || manager.IsActive(1) {
		t.Fatal("reader mode should turn off when toggled again")
	}
}

func TestBuildReaderModeJSIncludesEnableAndRestorePaths(t *testing.T) {
	enableJS := BuildReaderModeJS(true)
	disableJS := BuildReaderModeJS(false)

	if !strings.Contains(enableJS, "var enabled = true;") {
		t.Fatal("enabled script should activate reader mode")
	}
	if !strings.Contains(enableJS, "sparrow-reader-panel") ||
		!strings.Contains(enableJS, "sparrow-reader-font-size") {
		t.Fatal("enabled script should provide reading controls")
	}
	if !strings.Contains(enableJS, `{ name: "Dark", background: "#0b0b0d", color: "#e8e8e8", link: "#60a5fa" }`) ||
		!strings.Contains(enableJS, `background: rgba(21,22,27,.96)`) ||
		!strings.Contains(enableJS, `>Dark</button>`) {
		t.Fatal("reader mode should default to Sparrow Home's dark theme")
	}
	if !strings.Contains(disableJS, "var enabled = false;") ||
		!strings.Contains(disableJS, "restore();") {
		t.Fatal("disabled script should restore the original page")
	}
}
