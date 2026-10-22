package main

import (
	"sync"
	"time"
)

// ---------------------------------------------------
// CRASH RECOVERY (tracks open tabs so they can be restored
// if SPARROW is force-closed unexpectedly — kept in RAM and
// mirrored to a small recovery file that's deleted on clean exit)
// ---------------------------------------------------

type RecoverySnapshot struct {
	mu        sync.Mutex
	TabURLs   []string
	UpdatedAt time.Time
	Enabled   bool
}

var recoverySnapshot = &RecoverySnapshot{
	Enabled: true,
}

// UpdateSnapshot refreshes the list of currently open tab URLs
func (rs *RecoverySnapshot) UpdateSnapshot(urls []string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.Enabled {
		return
	}
	rs.TabURLs = urls
	rs.UpdatedAt = time.Now()
}

// GetSnapshot returns the last known set of open tabs
func (rs *RecoverySnapshot) GetSnapshot() []string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.TabURLs
}

// HasRecoverableSession checks if there's a non-empty snapshot to offer
func (rs *RecoverySnapshot) HasRecoverableSession() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return len(rs.TabURLs) > 0
}

// ClearSnapshot wipes the recovery data (called on clean shutdown)
func (rs *RecoverySnapshot) ClearSnapshot() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.TabURLs = []string{}
}

// ToggleEnabled turns crash recovery tracking on/off
func (rs *RecoverySnapshot) ToggleEnabled() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Enabled = !rs.Enabled
	if !rs.Enabled {
		rs.TabURLs = []string{}
	}
	return rs.Enabled
}
