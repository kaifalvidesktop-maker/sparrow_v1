package main

import (
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// TAB SYSTEM (RAM ONLY - MULTIPLE TAB TRACKING)
// ---------------------------------------------------

type Tab struct {
	ID         int
	Title      string
	Favicon    string
	URL        string
	CreatedAt  time.Time
	IsActive   bool
	History    []string `json:"-"`
	HistoryPos int      `json:"-"`
	// ExpectNav is true while a navigation that SPARROW started (typed URL,
	// back/forward) has not landed yet; the first location report then is the
	// same navigation (or its redirect), not a new history entry.
	ExpectNav bool `json:"-"`
}

type TabManager struct {
	mu       sync.Mutex
	Tabs     []*Tab
	NextID   int
	ActiveID int
}

var tabManager = &TabManager{
	Tabs:     []*Tab{},
	NextID:   1,
	ActiveID: 0,
}

// Create a new tab
func tabDisplayTitle(rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "New Tab"
	}
	if strings.HasPrefix(trimmed, "sparrow://") {
		route := strings.TrimPrefix(trimmed, "sparrow://")
		if query := strings.IndexByte(route, '?'); query >= 0 {
			route = route[:query]
		}
		internalTitles := map[string]string{
			"home":          "New Tab",
			"settings":      "Settings",
			"bookmarks":     "Bookmarks",
			"favorites":     "Favorites",
			"downloads":     "Downloads",
			"history":       "History",
			"shortcuts":     "Shortcuts",
			"links":         "Saved Links",
			"about":         "About Sparrow",
			"site-settings": "Site Settings",
			"cookies":       "Cookies",
			"welcome":       "Welcome",
			"autofill":      "Autofill",
			"feedback":      "Feedback",
			"updates":       "Updates",
			"backup":        "Backup",
			"pdf":           "PDF Viewer",
		}
		if title, ok := internalTitles[route]; ok {
			return title
		}
		return route
	}
	domain := extractDomain(trimmed)
	if domain != "" {
		return strings.TrimPrefix(domain, "www.")
	}
	return trimmed
}

func (tm *TabManager) NewTab(url string) *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Deactivate all existing tabs
	for _, t := range tm.Tabs {
		t.IsActive = false
	}

	tab := &Tab{
		ID:         tm.NextID,
		Title:      tabDisplayTitle(url),
		URL:        url,
		CreatedAt:  time.Now(),
		IsActive:   true,
		HistoryPos: -1,
	}

	tm.Tabs = append(tm.Tabs, tab)
	tm.ActiveID = tab.ID
	tm.NextID++

	return tab
}

// Close a tab by ID
func (tm *TabManager) CloseTab(id int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	newTabs := []*Tab{}
	wasActive := false

	for _, t := range tm.Tabs {
		if t.ID == id {
			if t.IsActive {
				wasActive = true
			}
			continue
		}
		newTabs = append(newTabs, t)
	}

	tm.Tabs = newTabs

	// If the closed tab was active, activate the last remaining tab
	if wasActive && len(tm.Tabs) > 0 {
		lastTab := tm.Tabs[len(tm.Tabs)-1]
		lastTab.IsActive = true
		tm.ActiveID = lastTab.ID
	}
}

// Switch to a specific tab
func (tm *TabManager) SwitchTab(id int) *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var found *Tab
	for _, t := range tm.Tabs {
		if t.ID == id {
			t.IsActive = true
			found = t
			tm.ActiveID = id
		} else {
			t.IsActive = false
		}
	}
	return found
}

// Update a tab's title and URL after navigation
func (tm *TabManager) UpdateTab(id int, title string, url string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.Tabs {
		if t.ID == id {
			t.Title = title
			t.URL = url
			break
		}
	}
}

// SetTitle updates just a tab's display title (used after a page loads
// and the chrome script reports back a short title for the tab strip).
func (tm *TabManager) SetTitle(id int, title string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			t.Title = title
			return
		}
	}
}

func (tm *TabManager) SetFavicon(id int, favicon string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			t.Favicon = favicon
			return
		}
	}
}

// SetURL updates just a tab's current URL without changing its display title.
func (tm *TabManager) SetURL(id int, url string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			t.URL = url
			return
		}
	}
}

// Get the currently active tab
func (tm *TabManager) GetActiveTab() *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.Tabs {
		if t.ID == tm.ActiveID {
			return t
		}
	}
	return nil
}

// Get all tabs (for rendering tab bar UI)
func (tm *TabManager) GetAllTabs() []*Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.Tabs
}

// Close all tabs (used on full reset / all-delete)
func (tm *TabManager) CloseAllTabs() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Tabs = []*Tab{}
	tm.ActiveID = 0
}

// ---------------------------------------------------
// PER-TAB BACK/FORWARD HISTORY
// (Needed because SPARROW now does real top-level navigation
// instead of loading pages inside an iframe, so nothing in
// the page's own JS survives between navigations. The Go
// side is the only place that can remember where a tab has
// been.)
// ---------------------------------------------------

// RecordNavigation pushes a newly-visited URL onto a tab's history,
// discarding any "forward" entries beyond the current position.
func (tm *TabManager) RecordNavigation(id int, url string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.Tabs {
		if t.ID == id {
			if t.HistoryPos < -1 || t.HistoryPos >= len(t.History) {
				t.HistoryPos = len(t.History) - 1
			}
			if t.HistoryPos < 0 {
				t.History = append([]string{}, url)
				t.HistoryPos = 0
			} else {
				t.History = append(t.History[:t.HistoryPos+1], url)
				t.HistoryPos = len(t.History) - 1
			}
			t.URL = url
			t.ExpectNav = true
			return
		}
	}
}

// CanGoBack reports whether the tab has an earlier entry to go to.
func (tm *TabManager) CanGoBack(id int) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			return t.HistoryPos > 0
		}
	}
	return false
}

// CanGoForward reports whether the tab has a later entry to go to.
func (tm *TabManager) CanGoForward(id int) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			return t.HistoryPos < len(t.History)-1
		}
	}
	return false
}

// GoBack moves the tab's history pointer back one step and returns
// the URL that should now be loaded.
func (tm *TabManager) GoBack(id int) (string, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			if t.HistoryPos <= 0 {
				return "", false
			}
			t.HistoryPos--
			t.URL = t.History[t.HistoryPos]
			t.ExpectNav = true
			return t.URL, true
		}
	}
	return "", false
}

// GoForward moves the tab's history pointer forward one step and
// returns the URL that should now be loaded.
func (tm *TabManager) GoForward(id int) (string, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for _, t := range tm.Tabs {
		if t.ID == id {
			if t.HistoryPos >= len(t.History)-1 {
				return "", false
			}
			t.HistoryPos++
			t.URL = t.History[t.HistoryPos]
			t.ExpectNav = true
			return t.URL, true
		}
	}
	return "", false
}

func sameLocation(a, b string) bool {
	trim := func(u string) string { return strings.TrimRight(strings.TrimSpace(u), "/") }
	return trim(a) == trim(b)
}

// SyncLocation is called with the page's real address once it has loaded.
// It keeps back/forward working like Chrome for navigations SPARROW did not
// start itself (link clicks, JS redirects, form posts):
//   - same address as recorded      -> the navigation landed, nothing to do
//   - change inside the same page   -> (pushState / #hash) update the address only
//   - redirect of our own navigation -> replace the current entry
//   - anything else                 -> a new page: push a history entry
func (tm *TabManager) SyncLocation(id int, location string, sameDocument bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.Tabs {
		if t.ID != id {
			continue
		}
		if sameLocation(t.URL, location) {
			t.ExpectNav = false
			return
		}
		validPos := t.HistoryPos >= 0 && t.HistoryPos < len(t.History)
		if sameDocument || t.ExpectNav {
			t.ExpectNav = false
			t.URL = location
			if validPos {
				t.History[t.HistoryPos] = location
			}
			return
		}
		if !validPos {
			t.History = append([]string{}, location)
			t.HistoryPos = 0
		} else {
			t.History = append(t.History[:t.HistoryPos+1], location)
			t.HistoryPos = len(t.History) - 1
		}
		t.URL = location
		return
	}
}
