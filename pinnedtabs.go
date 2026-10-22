package main

import "sync"

// ---------------------------------------------------
// TAB PINNING (pinned tabs stay small and always at the start)
// ---------------------------------------------------

type PinManager struct {
	mu     sync.Mutex
	Pinned map[int]bool
}

var pinManager = &PinManager{
	Pinned: make(map[int]bool),
}

// Toggle pin state for a tab
func (pm *PinManager) Toggle(tabID int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.Pinned[tabID] = !pm.Pinned[tabID]
	return pm.Pinned[tabID]
}

// Check if a tab is pinned
func (pm *PinManager) IsPinned(tabID int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.Pinned[tabID]
}

// Remove pin record when a tab closes
func (pm *PinManager) Clear(tabID int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.Pinned, tabID)
}

// Get all pinned tab IDs
func (pm *PinManager) GetAllPinned() []int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	ids := []int{}
	for id, pinned := range pm.Pinned {
		if pinned {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetOrderedTabs returns tabs with pinned ones first, preserving
// relative order within each group
func GetOrderedTabs() []*Tab {
	all := tabManager.GetAllTabs()
	pinned := []*Tab{}
	unpinned := []*Tab{}

	for _, t := range all {
		if pinManager.IsPinned(t.ID) {
			pinned = append(pinned, t)
		} else {
			unpinned = append(unpinned, t)
		}
	}

	return append(pinned, unpinned...)
}
