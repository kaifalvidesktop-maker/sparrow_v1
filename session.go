package main

import (
	"sync"
	"time"
)

// ---------------------------------------------------
// SESSION MANAGEMENT (RECENTLY CLOSED TABS - RAM ONLY)
// ---------------------------------------------------

type ClosedTab struct {
	Title    string
	URL      string
	ClosedAt time.Time
}

type SessionManager struct {
	mu       sync.Mutex
	History  []ClosedTab
	MaxLimit int
}

var sessionManager = &SessionManager{
	History:  []ClosedTab{},
	MaxLimit: 25, // it will remeber highest 25 closed tabs
}

// Record a closed tab
func (sm *SessionManager) RecordClosed(title string, url string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if url == "" || url == "home" || url == "sparrow://home" {
		return
	}

	item := ClosedTab{
		Title:    title,
		URL:      url,
		ClosedAt: time.Now(),
	}

	// new will be added in 1st
	sm.History = append([]ClosedTab{item}, sm.History...)

	// if limit finishes , we  will delete the last
	if len(sm.History) > sm.MaxLimit {
		sm.History = sm.History[:sm.MaxLimit]
	}
}

// Get all recently closed tabs
func (sm *SessionManager) GetAll() []ClosedTab {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.History
}

// Pop/Restore the last closed tab
func (sm *SessionManager) PopLastClosed() *ClosedTab {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.History) == 0 {
		return nil
	}

	last := sm.History[0]
	sm.History = sm.History[1:]
	return &last
}

// Clear all closed history
func (sm *SessionManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.History = []ClosedTab{}
}
