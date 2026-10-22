package main

import (
	"encoding/json"
	"strings"
	"time"
)

// ---------------------------------------------------
// BOOKMARK EXPORT / IMPORT (JSON-based, no account needed)
// ---------------------------------------------------

type ExportedBookmark struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Folder    string    `json:"folder"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// ExportBookmarksJSON converts all bookmarks into a JSON string
func ExportBookmarksJSON() string {
	all := bookmarkManager.GetAll()
	exportList := make([]ExportedBookmark, 0, len(all))

	for _, b := range all {
		exportList = append(exportList, ExportedBookmark{
			Title:  b.Title,
			URL:    b.URL,
			Folder: b.Folder,
		})
	}

	data, err := json.MarshalIndent(exportList, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(data)
}

// ImportBookmarksJSON parses a JSON string and adds bookmarks from it
func ImportBookmarksJSON(jsonStr string) int {
	var imported []ExportedBookmark

	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return 0
	}

	err := json.Unmarshal([]byte(jsonStr), &imported)
	if err != nil {
		return 0
	}

	count := 0
	for _, item := range imported {
		if item.URL == "" {
			continue
		}
		if !bookmarkManager.IsBookmarked(item.URL) {
			bookmarkManager.Add(item.Title, item.URL, item.Folder)
			count++
		}
	}
	return count
}
