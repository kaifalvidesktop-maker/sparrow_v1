package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------
// FULL DATA BACKUP / RESTORE
// Bundles bookmarks, settings, and shortcuts into one JSON
// file the user can save and restore later — this is how
// "no account, no sync" still lets users move between machines.
// Note: passwords are intentionally excluded from plain backup;
// see ExportEncryptedVault() below for those.
// ---------------------------------------------------

type BackupBundle struct {
	Version       string               `json:"version"`
	CreatedAt     string               `json:"createdAt"`
	Bookmarks     []ExportedBookmark   `json:"bookmarks"`
	Settings      map[string]any       `json:"settings"`
	Shortcuts     []PinnedShortcut     `json:"shortcuts"`
	Favorites     []FavoriteEntry      `json:"favorites,omitempty"`
	History       []BackupHistoryEntry `json:"history,omitempty"`
	Downloads     []BackupDownload     `json:"downloads,omitempty"`
	Autofill      *AutofillData        `json:"autofill,omitempty"`
	CustomTheme   *CustomThemeSnapshot `json:"customTheme,omitempty"`
	HomeAnimation string               `json:"homeAnimation,omitempty"`
	HomeWallpaper string               `json:"homeWallpaper,omitempty"`
}

type BackupHistoryEntry struct {
	URL   string    `json:"url"`
	Title string    `json:"title"`
	Time  time.Time `json:"time"`
}

type BackupDownload struct {
	ID         int       `json:"id"`
	FileName   string    `json:"fileName"`
	URL        string    `json:"url"`
	Progress   int       `json:"progress"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`
}

// BuildBackup assembles all exportable data into one bundle
func BuildBackup() BackupBundle {
	all := bookmarkManager.GetAll()
	exportList := make([]ExportedBookmark, 0, len(all))
	for _, b := range all {
		exportList = append(exportList, ExportedBookmark{
			Title: b.Title, URL: b.URL, Folder: b.Folder, CreatedAt: b.CreatedAt,
		})
	}

	historyEntries := history.GetAll()
	exportedHistory := make([]BackupHistoryEntry, 0, len(historyEntries))
	for _, entry := range historyEntries {
		exportedHistory = append(exportedHistory, BackupHistoryEntry{
			URL: entry.URL, Title: entry.Title, Time: entry.Time,
		})
	}

	downloadItems := downloadManager.GetAll()
	exportedDownloads := make([]BackupDownload, 0, len(downloadItems))
	for _, item := range downloadItems {
		if item == nil {
			continue
		}
		exportedDownloads = append(exportedDownloads, BackupDownload{
			ID: item.ID, FileName: item.FileName, URL: item.URL, Progress: item.Progress,
			Status: item.Status, StartedAt: item.StartedAt, FinishedAt: item.FinishedAt,
		})
	}

	customThemeSnapshot := customTheme.Snapshot()
	autofill := getAutofillProfile()
	return BackupBundle{
		Version:       sparrowVersion,
		CreatedAt:     time.Now().Format(time.RFC3339),
		Bookmarks:     exportList,
		Settings:      settings.GetAll(),
		Shortcuts:     shortcutManagerNewTab.GetShortcuts(),
		Favorites:     favoriteManager.GetAll(),
		History:       exportedHistory,
		Downloads:     exportedDownloads,
		Autofill:      &autofill,
		CustomTheme:   &customThemeSnapshot,
		HomeAnimation: homeAnimationManager.Get(),
		HomeWallpaper: wallpaperManager.GetSelectedID(),
	}
}

// ExportBackupJSON returns the full backup as a JSON string
func ExportBackupJSON() string {
	bundle := BuildBackup()
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// SaveBackupToFile writes the backup JSON to the downloads folder
func SaveBackupToFile() (string, error) {
	downloadsPath, err := userDownloadsDirectory()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(downloadsPath, 0755); err != nil {
		return "", err
	}

	fileName := "sparrow-backup_" + time.Now().Format("2006-01-02_15-04-05") + ".json"
	fullPath := filepath.Join(downloadsPath, fileName)

	err = os.WriteFile(fullPath, []byte(ExportBackupJSON()), 0644)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func userDownloadsDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("could not determine the current user's home directory")
	}
	return filepath.Join(home, "Downloads"), nil
}

// RestoreBackupJSON parses a backup JSON string and applies it
func RestoreBackupJSON(jsonStr string) (int, error) {
	var bundle BackupBundle
	if err := json.Unmarshal([]byte(jsonStr), &bundle); err != nil {
		return 0, err
	}
	if bundle.Version == "" && bundle.Settings == nil && bundle.Bookmarks == nil &&
		bundle.Shortcuts == nil && bundle.Favorites == nil && bundle.History == nil &&
		bundle.Downloads == nil && bundle.Autofill == nil && bundle.CustomTheme == nil {
		return 0, fmt.Errorf("file does not contain a Sparrow backup")
	}

	restoredCount := 0

	// Restore bookmarks (skip duplicates)
	for _, b := range bundle.Bookmarks {
		if b.URL == "" {
			continue
		}
		if !bookmarkManager.IsBookmarked(b.URL) {
			bookmark := bookmarkManager.Add(b.Title, b.URL, b.Folder)
			if !b.CreatedAt.IsZero() {
				bookmark.CreatedAt = b.CreatedAt
			}
			restoredCount++
		}
	}

	// Restore settings, accepting only supported values from older and newer backups.
	settings.mu.Lock()
	if theme, ok := bundle.Settings["theme"].(string); ok {
		switch theme {
		case "dark", "light":
			settings.Theme = theme
		}
	}
	if engine, ok := bundle.Settings["searchEngine"].(string); ok {
		switch engine {
		case "google", "bing", "brave", "duckduckgo", "yahoo":
			settings.DefaultSearchEngine = engine
		}
	}
	if homepage, ok := bundle.Settings["homepage"].(string); ok {
		if homepage = strings.TrimSpace(homepage); homepage != "" {
			settings.Homepage = homepage
		}
	}
	if enabled, ok := bundle.Settings["jsEnabled"].(bool); ok {
		settings.JSEnabled = enabled
	}
	if enabled, ok := bundle.Settings["imagesEnabled"].(bool); ok {
		settings.ImagesEnabled = enabled
	}
	if downloadPath, ok := bundle.Settings["downloadPath"].(string); ok && downloadPath != "" {
		settings.DownloadPath = downloadPath
	}
	settings.mu.Unlock()

	// Restore shortcuts
	for _, s := range bundle.Shortcuts {
		if _, err := shortcutManagerNewTab.AddLink(s.Title, s.URL); err == nil {
			restoredCount++
		}
	}

	for _, favorite := range bundle.Favorites {
		if favorite.URL != "" && favoriteManager.Add(favorite.Title, favorite.URL) != nil {
			restoredCount++
		}
	}

	importedHistory := make([]HistoryEntry, 0, len(bundle.History))
	for _, entry := range bundle.History {
		importedHistory = append(importedHistory, HistoryEntry{
			URL: entry.URL, Title: entry.Title, Time: entry.Time,
		})
	}
	restoredCount += history.Import(importedHistory)

	importedDownloads := make([]*DownloadItem, 0, len(bundle.Downloads))
	for _, item := range bundle.Downloads {
		importedDownloads = append(importedDownloads, &DownloadItem{
			ID: item.ID, FileName: item.FileName, URL: item.URL, Progress: item.Progress,
			Status: item.Status, StartedAt: item.StartedAt, FinishedAt: item.FinishedAt,
		})
	}
	restoredCount += downloadManager.Import(importedDownloads)

	if bundle.Autofill != nil {
		saveAutofillProfile(*bundle.Autofill)
	}
	if bundle.CustomTheme != nil {
		customTheme.Restore(*bundle.CustomTheme)
	}
	if bundle.HomeAnimation != "" {
		homeAnimationManager.Restore(bundle.HomeAnimation)
	}
	if bundle.HomeWallpaper != "" {
		wallpaperManager.Restore(bundle.HomeWallpaper)
	}

	if err := savePersistedState(); err != nil {
		return restoredCount, err
	}

	return restoredCount, nil
}
