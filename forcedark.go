package main

import "sync"

// ---------------------------------------------------
// FORCE DARK MODE (injects a CSS filter into any website
// that doesn't have its own dark theme)
// ---------------------------------------------------

type ForceDarkManager struct {
	mu     sync.Mutex
	Global bool
	PerTab map[int]bool
}

var forceDark = &ForceDarkManager{
	PerTab: make(map[int]bool),
}

// ToggleGlobal turns force-dark on/off for all new pages
func (fd *ForceDarkManager) ToggleGlobal() bool {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.Global = !fd.Global
	return fd.Global
}

// IsGlobalOn checks the global setting
func (fd *ForceDarkManager) IsGlobalOn() bool {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	return fd.Global
}

// ToggleForTab turns force-dark on/off for a specific tab only
func (fd *ForceDarkManager) ToggleForTab(tabID int) bool {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.PerTab[tabID] = !fd.PerTab[tabID]
	return fd.PerTab[tabID]
}

// ShouldApply decides whether force-dark should run for this tab
func (fd *ForceDarkManager) ShouldApply(tabID int) bool {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	if fd.Global {
		return true
	}
	return fd.PerTab[tabID]
}

// Injectable CSS-based dark mode filter (safe fallback for any site)
const forceDarkJS = `
(function() {
	var existing = document.getElementById('sparrow-force-dark');
	if (existing) { existing.remove(); return; }

	var style = document.createElement('style');
	style.id = 'sparrow-force-dark';
	style.textContent =
		'html { filter: invert(1) hue-rotate(180deg) !important; background: #fff !important; }' +
		'img, video, picture, iframe, canvas, svg, [style*="background-image"] { filter: invert(1) hue-rotate(180deg) !important; }';
	document.head.appendChild(style);
})();
`
