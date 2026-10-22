package main

import "sync"

// ---------------------------------------------------
// PERFORMANCE / BATTERY SAVER MODE
// Tracks which tabs are "background" (inactive) so the UI
// can visually indicate they're throttled/suspended.
// ---------------------------------------------------

type PerformanceManager struct {
	mu            sync.Mutex
	SaverEnabled  bool
	SuspendedTabs map[int]bool
}

var perfManager = &PerformanceManager{
	SuspendedTabs: make(map[int]bool),
}

// ToggleSaver turns battery/performance saver mode on or off
func (pm *PerformanceManager) ToggleSaver() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.SaverEnabled = !pm.SaverEnabled
	if !pm.SaverEnabled {
		pm.SuspendedTabs = make(map[int]bool)
	}
	return pm.SaverEnabled
}

// IsSaverOn checks current saver state
func (pm *PerformanceManager) IsSaverOn() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.SaverEnabled
}

// SuspendTab marks a background tab as suspended (saver mode only)
func (pm *PerformanceManager) SuspendTab(tabID int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.SaverEnabled {
		pm.SuspendedTabs[tabID] = true
	}
}

// WakeTab marks a tab as active again (e.g. user switched to it)
func (pm *PerformanceManager) WakeTab(tabID int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.SuspendedTabs, tabID)
}

// IsSuspended checks if a tab is currently suspended
func (pm *PerformanceManager) IsSuspended(tabID int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.SuspendedTabs[tabID]
}

// SuspendedCount returns how many tabs are currently suspended
func (pm *PerformanceManager) SuspendedCount() int {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return len(pm.SuspendedTabs)
}
