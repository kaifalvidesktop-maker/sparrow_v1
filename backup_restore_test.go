package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupRestorePageDoesNotShadowNativeRestoreBinding(t *testing.T) {
	if strings.Contains(backupPageHTML, "function restoreBackup(") {
		t.Fatal("restore page JavaScript must not shadow the native restoreBackup binding")
	}
	if !strings.Contains(backupPageHTML, `onchange="restoreBackupFileSelected()"`) {
		t.Fatal("backup file input should use its separate file-selection handler")
	}
	if strings.Contains(backupPageHTML, "event.target.files") {
		t.Fatal("restore handler must read the file from the checked input element, not an event target")
	}
	if !strings.Contains(backupPageHTML, "!input.files || input.files.length === 0") {
		t.Fatal("restore handler should safely handle missing or empty file selections")
	}
	if !strings.Contains(backupPageHTML, "window.restoreBackup(contents)") {
		t.Fatal("restore page should pass file contents to the native restore binding")
	}
}

func TestSaveBackupToUsersDownloadsFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	path, err := SaveBackupToFile()
	if err != nil {
		t.Fatalf("SaveBackupToFile returned an error: %v", err)
	}
	if filepath.Dir(path) != filepath.Join(home, "Downloads") {
		t.Fatalf("backup path = %q, want it under %q", path, filepath.Join(home, "Downloads"))
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("backup file was not created: %v", err)
	}
}

func TestRestoreBackupRestoresAllSupportedData(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())

	settings.mu.Lock()
	oldSettings := map[string]any{
		"theme": settings.Theme, "searchEngine": settings.DefaultSearchEngine,
		"homepage": settings.Homepage, "jsEnabled": settings.JSEnabled,
		"imagesEnabled": settings.ImagesEnabled, "downloadPath": settings.DownloadPath,
	}
	settings.mu.Unlock()
	defer func() {
		settings.mu.Lock()
		settings.Theme = oldSettings["theme"].(string)
		settings.DefaultSearchEngine = oldSettings["searchEngine"].(string)
		settings.Homepage = oldSettings["homepage"].(string)
		settings.JSEnabled = oldSettings["jsEnabled"].(bool)
		settings.ImagesEnabled = oldSettings["imagesEnabled"].(bool)
		settings.DownloadPath = oldSettings["downloadPath"].(string)
		settings.mu.Unlock()
	}()

	bookmarkManager.mu.Lock()
	oldBookmarks := append([]*Bookmark(nil), bookmarkManager.Bookmarks...)
	oldBookmarkNextID := bookmarkManager.NextID
	bookmarkManager.Bookmarks = nil
	bookmarkManager.NextID = 1
	bookmarkManager.mu.Unlock()
	defer func() {
		bookmarkManager.mu.Lock()
		bookmarkManager.Bookmarks = oldBookmarks
		bookmarkManager.NextID = oldBookmarkNextID
		bookmarkManager.mu.Unlock()
	}()

	oldFavorites := favoriteManager.GetAll()
	favoriteManager.ReplaceAll(nil)
	defer favoriteManager.ReplaceAll(oldFavorites)

	oldShortcuts := shortcutManagerNewTab.GetShortcuts()
	shortcutManagerNewTab.ReplaceAll(nil)
	defer shortcutManagerNewTab.ReplaceAll(oldShortcuts)

	history.mu.Lock()
	oldHistory := append([]HistoryEntry(nil), history.Entries...)
	history.Entries = nil
	history.mu.Unlock()
	defer func() {
		history.mu.Lock()
		history.Entries = oldHistory
		history.mu.Unlock()
	}()

	downloadManager.mu.Lock()
	oldDownloads := make([]*DownloadItem, len(downloadManager.Items))
	for i, item := range downloadManager.Items {
		if item != nil {
			copy := *item
			oldDownloads[i] = &copy
		}
	}
	oldNextDownloadID := downloadManager.NextID
	downloadManager.Items = nil
	downloadManager.NextID = 1
	downloadManager.mu.Unlock()
	defer func() {
		downloadManager.mu.Lock()
		downloadManager.Items = oldDownloads
		downloadManager.NextID = oldNextDownloadID
		downloadManager.mu.Unlock()
	}()

	oldAutofill := getAutofillProfile()
	defer saveAutofillProfile(oldAutofill)
	oldCustomTheme := customTheme.Snapshot()
	defer customTheme.Restore(oldCustomTheme)
	oldAnimation := homeAnimationManager.Get()
	defer homeAnimationManager.Restore(oldAnimation)
	oldWallpaper := wallpaperManager.GetSelectedID()
	defer wallpaperManager.Restore(oldWallpaper)

	visitedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	downloadStarted := visitedAt.Add(time.Hour)
	bundle := BackupBundle{
		Version: sparrowVersion,
		Bookmarks: []ExportedBookmark{{
			Title: "Docs", URL: "https://docs.example", Folder: "Work", CreatedAt: visitedAt,
		}},
		Settings: map[string]any{
			"theme": "light", "searchEngine": "bing", "homepage": "https://home.example",
			"jsEnabled": false, "imagesEnabled": false, "downloadPath": t.TempDir(),
		},
		Shortcuts: []PinnedShortcut{{ID: 1, Title: "Saved link", URL: "https://saved.example"}},
		Favorites: []FavoriteEntry{{
			Title: "Favorite", URL: "https://favorite.example", CreatedAt: visitedAt,
		}},
		History: []BackupHistoryEntry{{
			URL: "https://visited.example", Title: "Visited", Time: visitedAt,
		}},
		Downloads: []BackupDownload{{
			ID: 1, FileName: "download.zip", URL: "https://files.example/download.zip",
			Progress: 35, Status: "downloading", StartedAt: downloadStarted,
		}},
		Autofill:      &AutofillData{Name: "Jane Doe", Email: "jane@example.com"},
		CustomTheme:   &CustomThemeSnapshot{AccentColor: "#123456", TabRadius: 8, Compact: true},
		HomeAnimation: "bubbles",
		HomeWallpaper: "midnight",
	}
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}

	count, err := RestoreBackupJSON(string(data))
	if err != nil {
		t.Fatalf("RestoreBackupJSON returned an error: %v", err)
	}
	if count != 5 {
		t.Fatalf("restored data item count = %d, want 5", count)
	}

	if got := bookmarkManager.GetAll(); len(got) != 1 || got[0].CreatedAt != visitedAt {
		t.Fatalf("restored bookmarks = %+v", got)
	}
	if got := favoriteManager.GetAll(); len(got) != 1 || got[0].URL != "https://favorite.example" {
		t.Fatalf("restored favorites = %+v", got)
	}
	if got := shortcutManagerNewTab.GetShortcuts(); len(got) != 1 || got[0].URL != "https://saved.example" {
		t.Fatalf("restored shortcuts = %+v", got)
	}
	if got := history.GetAll(); len(got) != 1 || got[0].Time != visitedAt {
		t.Fatalf("restored history = %+v", got)
	}
	if got := downloadManager.GetAll(); len(got) != 1 || got[0].Status != "failed" || got[0].FinishedAt.IsZero() {
		t.Fatalf("restored downloads = %+v, expected an interrupted download marked failed", got)
	}
	if got := settings.GetAll(); got["theme"] != "light" || got["searchEngine"] != "bing" ||
		got["homepage"] != "https://home.example" || got["jsEnabled"] != false ||
		got["imagesEnabled"] != false || got["downloadPath"] != bundle.Settings["downloadPath"] {
		t.Fatalf("restored settings = %+v", got)
	}
	if got := getAutofillProfile(); got.Name != "Jane Doe" || got.Email != "jane@example.com" {
		t.Fatalf("restored autofill = %+v", got)
	}
	if got := customTheme.Snapshot(); got != *bundle.CustomTheme {
		t.Fatalf("restored custom theme = %+v", got)
	}
	if got := homeAnimationManager.Get(); got != "bubbles" {
		t.Fatalf("restored home animation = %q", got)
	}
	if got := wallpaperManager.GetSelectedID(); got != "midnight" {
		t.Fatalf("restored wallpaper = %q", got)
	}
}

func TestRestoreBackupRejectsUnrelatedJSON(t *testing.T) {
	if _, err := RestoreBackupJSON(`{"notABackup":true}`); err == nil {
		t.Fatal("expected unrelated JSON to be rejected")
	}
	if _, err := RestoreBackupJSON(`{`); err == nil {
		t.Fatal("expected invalid JSON to be rejected")
	}
}
