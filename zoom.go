package main

import (
	"math"
	"sync"
)

// ---------------------------------------------------
// ZOOM MANAGEMENT SYSTEM (RAM ONLY)
// ---------------------------------------------------

type ZoomManager struct {
	mu    sync.Mutex
	Zooms map[int]float64 // TabID -> Zoom Level (e.g. 1.0 = 100%)
}

var zoomManager = &ZoomManager{
	Zooms: make(map[int]float64),
}

// Zoom In (+10%)
func (zm *ZoomManager) ZoomIn(tabID int) float64 {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	current, exists := zm.Zooms[tabID]
	if !exists {
		current = 1.0
	}
	current = math.Round((current+0.1)*10) / 10
	if current > 3.0 {
		current = 3.0 // Maximum zoom limit (300%)
	}
	zm.Zooms[tabID] = current
	return current
}

// Zoom Out (-10%)
func (zm *ZoomManager) ZoomOut(tabID int) float64 {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	current, exists := zm.Zooms[tabID]
	if !exists {
		current = 1.0
	}
	current = math.Round((current-0.1)*10) / 10
	if current < 0.5 {
		current = 0.5 // Minimum zoom limit (50%)
	}
	zm.Zooms[tabID] = current
	return current
}

// Reset Zoom to default (100%)
func (zm *ZoomManager) Reset(tabID int) float64 {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	zm.Zooms[tabID] = 1.0
	return 1.0
}

// Get current zoom level for a tab
func (zm *ZoomManager) Get(tabID int) float64 {
	zm.mu.Lock()
	defer zm.mu.Unlock()

	current, exists := zm.Zooms[tabID]
	if !exists {
		return 1.0
	}
	return current
}
