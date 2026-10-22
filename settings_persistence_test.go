package main

import (
	"os"
	"strings"
	"testing"
)

func TestSettingsPersistImmediatelyToDisk(t *testing.T) {
	tempDir := t.TempDir()
	oldAppData, hadAppData := os.LookupEnv("APPDATA")
	oldXDG, hadXDG := os.LookupEnv("XDG_CONFIG_HOME")
	if err := os.Setenv("APPDATA", tempDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("XDG_CONFIG_HOME", tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if hadAppData {
			_ = os.Setenv("APPDATA", oldAppData)
		} else {
			_ = os.Unsetenv("APPDATA")
		}
		if hadXDG {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	settings.mu.Lock()
	settings.Theme = "dark"
	settings.DefaultSearchEngine = "google"
	settings.Homepage = "sparrow://home"
	settings.JSEnabled = true
	settings.ImagesEnabled = true
	settings.DownloadPath = "Downloads"
	settings.mu.Unlock()
	oldCustomTheme := customTheme.Snapshot()
	defer customTheme.Restore(oldCustomTheme)
	oldHomeAnimation := homeAnimationManager.Get()
	defer homeAnimationManager.Restore(oldHomeAnimation)
	oldWallpaper := wallpaperManager.GetSelectedID()
	defer wallpaperManager.Restore(oldWallpaper)
	oldSavedLinks := shortcutManagerNewTab.GetShortcuts()
	defer shortcutManagerNewTab.ReplaceAll(oldSavedLinks)
	wantCustomTheme := CustomThemeSnapshot{AccentColor: "#123456", TabRadius: 7, Compact: true}
	customTheme.Restore(wantCustomTheme)
	homeAnimationManager.Restore("bubbles")
	if _, ok := wallpaperManager.SetWallpaper("midnight"); !ok {
		t.Fatal("failed to set test wallpaper")
	}
	shortcutManagerNewTab.ReplaceAll([]PinnedShortcut{{ID: 7, Title: "Docs", URL: "https://example.com/docs"}})

	saveAutofillProfile(AutofillData{Name: "Jane Doe", Email: "jane@example.com"})
	settings.SetTheme("light")

	path, err := stateFilePath()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected persisted state file to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "\"theme\": \"light\"") {
		t.Fatalf("expected theme to be saved to disk, got: %s", content)
	}
	if !strings.Contains(content, "\"name\": \"Jane Doe\"") {
		t.Fatalf("expected autofill profile to be saved to disk, got: %s", content)
	}
	if !strings.Contains(content, "\"homeAnimation\": \"bubbles\"") {
		t.Fatalf("expected home animation to be saved to disk, got: %s", content)
	}
	if !strings.Contains(content, "\"homeWallpaper\": \"midnight\"") {
		t.Fatalf("expected home wallpaper to be saved to disk, got: %s", content)
	}
	state := loadPersistedState()
	if state == nil || state.CustomTheme == nil {
		t.Fatal("expected custom theme settings to be saved to disk")
	}
	if got := *state.CustomTheme; got != wantCustomTheme {
		t.Fatalf("persisted custom theme = %+v, want %+v", got, wantCustomTheme)
	}
	if state.HomeAnimation != "bubbles" {
		t.Fatalf("persisted home animation = %q, want %q", state.HomeAnimation, "bubbles")
	}
	if state.HomeWallpaper != "midnight" {
		t.Fatalf("persisted home wallpaper = %q, want %q", state.HomeWallpaper, "midnight")
	}
	if len(state.SavedLinks) != 1 || state.SavedLinks[0].ID != 7 || state.SavedLinks[0].Title != "Docs" {
		t.Fatalf("persisted saved links = %+v, want the Docs link", state.SavedLinks)
	}
	if strings.Contains(content, "bookmarks") || strings.Contains(content, "savedLogins") || strings.Contains(content, "closedTabs") {
		t.Fatalf("expected only settings and autofill to be persisted, got: %s", content)
	}
}

func TestResetToDefaultRestoresDefaults(t *testing.T) {
	settings.SetTheme("light")
	settings.SetSearchEngine("yahoo")
	settings.SetHomepage("https://example.com")
	settings.ToggleJS()
	settings.ToggleImages()

	result := settings.ResetToDefault()
	if result["theme"] != "dark" {
		t.Fatalf("expected default theme to be dark, got %v", result["theme"])
	}
	if result["searchEngine"] != "google" {
		t.Fatalf("expected default search engine to be google, got %v", result["searchEngine"])
	}
	if result["homepage"] != "sparrow://home" {
		t.Fatalf("expected homepage to reset to sparrow://home, got %v", result["homepage"])
	}
	if result["jsEnabled"] != true {
		t.Fatalf("expected JS to be enabled by default, got %v", result["jsEnabled"])
	}
	if result["imagesEnabled"] != true {
		t.Fatalf("expected images to be enabled by default, got %v", result["imagesEnabled"])
	}
}

func TestResolveInputRespectsDefaultSearchEngine(t *testing.T) {
	old := settings.DefaultSearchEngine
	defer func() { settings.DefaultSearchEngine = old }()

	settings.SetSearchEngine("bing")
	if got := resolveInput("cats and dogs"); got != "https://www.bing.com/search?q=cats+and+dogs" {
		t.Fatalf("expected Bing search URL, got %s", got)
	}

	settings.SetSearchEngine("duckduckgo")
	if got := resolveInput("cats and dogs"); got != "https://duckduckgo.com/?q=cats+and+dogs" {
		t.Fatalf("expected DuckDuckGo search URL, got %s", got)
	}

	settings.SetSearchEngine("brave")
	if got := resolveInput("cats and dogs"); got != "https://search.brave.com/search?q=cats+and+dogs" {
		t.Fatalf("expected Brave Search URL, got %s", got)
	}
}

func TestShouldOpenExternalURL(t *testing.T) {
	if !shouldOpenExternalURL("www.facebokkk.com") {
		t.Fatal("expected a mistyped domain like www.facebokkk.com to open in the default browser")
	}
	if shouldOpenExternalURL("www.google.com") {
		t.Fatal("expected a real domain to stay in-app")
	}
	if shouldOpenExternalURL("cats and dogs") {
		t.Fatal("expected a normal search query to remain a search")
	}
}

func TestSaveLoginPageRoutesAreRemoved(t *testing.T) {
	for _, route := range []string{"sparrow://save-login", "sparrow://passwords"} {
		if _, ok := internalPageHTML(route); ok {
			t.Fatalf("expected %s to no longer resolve to a page", route)
		}
	}
}
