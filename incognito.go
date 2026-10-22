package main

import (
	"sync"
	"time"
)

// ---------------------------------------------------
// INCOGNITO / PRIVATE MODE
// Even stricter than normal mode: NOTHING is recorded,
// not even in the in-memory history list.
// ---------------------------------------------------

type IncognitoSession struct {
	mu        sync.Mutex
	Active    bool
	StartedAt time.Time
	TabIDs    map[int]bool // which tab IDs are incognito
}

var incognito = &IncognitoSession{
	TabIDs: make(map[int]bool),
}

// Turn incognito mode ON
func (i *IncognitoSession) Enable() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Active = true
	i.StartedAt = time.Now()
}

// Turn incognito mode OFF
func (i *IncognitoSession) Disable() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Active = false
	i.TabIDs = make(map[int]bool)
}

// Check if incognito mode is currently active globally
func (i *IncognitoSession) IsActive() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.Active
}

// Mark a specific tab as incognito
func (i *IncognitoSession) MarkTab(tabID int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.TabIDs[tabID] = true
}

// Unmark a tab (when it's closed)
func (i *IncognitoSession) UnmarkTab(tabID int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.TabIDs, tabID)
}

// Check if a specific tab is incognito
func (i *IncognitoSession) IsTabIncognito(tabID int) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.TabIDs[tabID]
}

// Count how many incognito tabs are open
func (i *IncognitoSession) ActiveTabCount() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return len(i.TabIDs)
}

// Get session duration so far
func (i *IncognitoSession) Duration() string {
	i.mu.Lock()
	defer i.mu.Unlock()
	if !i.Active {
		return "0m"
	}
	d := time.Since(i.StartedAt)
	minutes := int(d.Minutes())
	return itoa(minutes) + "m"
}

// Small helper: int to string without importing strconv everywhere
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
