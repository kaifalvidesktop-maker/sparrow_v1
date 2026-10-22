package main

import (
	"strings"
	"sync"
)

// ---------------------------------------------------
// SPEED DIAL (user-customizable grid of shortcuts on
// the home page — separate from auto-detected Top Sites)
// ---------------------------------------------------

type SpeedDialTile struct {
	ID    int
	Title string
	URL   string
	Color string
}

type SpeedDialManager struct {
	mu     sync.Mutex
	Tiles  []*SpeedDialTile
	NextID int
}

var speedDialManager = &SpeedDialManager{
	Tiles: []*SpeedDialTile{
		{ID: 1, Title: "YouTube", URL: "https://youtube.com", Color: "#ff4d4d"},
		{ID: 2, Title: "GitHub", URL: "https://github.com", Color: "#6b6b6b"},
	},
	NextID: 3,
}

var speedDialColors = []string{"#4d90fe", "#ff4d4d", "#3ddc84", "#ffd93d", "#b58fd8", "#ff9d4d"}

// AddTile adds a new custom speed dial shortcut
func (sd *SpeedDialManager) AddTile(title string, url string) *SpeedDialTile {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" {
		title = extractDomain(url)
	}

	color := speedDialColors[sd.NextID%len(speedDialColors)]

	tile := &SpeedDialTile{
		ID:    sd.NextID,
		Title: title,
		URL:   url,
		Color: color,
	}
	sd.Tiles = append(sd.Tiles, tile)
	sd.NextID++

	// Cap at 12 tiles for a clean grid
	if len(sd.Tiles) > 12 {
		sd.Tiles = sd.Tiles[len(sd.Tiles)-12:]
	}
	return tile
}

// RemoveTile deletes a speed dial tile by ID
func (sd *SpeedDialManager) RemoveTile(id int) bool {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	newList := []*SpeedDialTile{}
	removed := false
	for _, t := range sd.Tiles {
		if t.ID == id {
			removed = true
			continue
		}
		newList = append(newList, t)
	}
	sd.Tiles = newList
	return removed
}

// Reorder moves a tile to a new position in the grid
func (sd *SpeedDialManager) Reorder(fromIndex int, toIndex int) bool {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if fromIndex < 0 || fromIndex >= len(sd.Tiles) || toIndex < 0 || toIndex >= len(sd.Tiles) {
		return false
	}

	tile := sd.Tiles[fromIndex]
	sd.Tiles = append(sd.Tiles[:fromIndex], sd.Tiles[fromIndex+1:]...)

	newList := make([]*SpeedDialTile, 0, len(sd.Tiles)+1)
	newList = append(newList, sd.Tiles[:toIndex]...)
	newList = append(newList, tile)
	newList = append(newList, sd.Tiles[toIndex:]...)
	sd.Tiles = newList

	return true
}

// GetAll returns all speed dial tiles in order
func (sd *SpeedDialManager) GetAll() []*SpeedDialTile {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.Tiles
}
