package main

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------
// SMART TOP SITES (based on visit frequency from history,
// falls back to defaults when history is empty)
// ---------------------------------------------------

type TopSite struct {
	Domain string
	Visits int
	URL    string
}

type TopSitesEngine struct {
	mu sync.Mutex
}

var topSitesEngine = &TopSitesEngine{}

var defaultTopSites = []TopSite{
	{Domain: "youtube.com", URL: "https://youtube.com"},
	{Domain: "facebook.com", URL: "https://facebook.com"},
	{Domain: "github.com", URL: "https://github.com"},
}

// GetTopSites returns most-visited domains from history, or defaults
func (te *TopSitesEngine) GetTopSites(limit int) []TopSite {
	te.mu.Lock()
	defer te.mu.Unlock()

	entries := history.GetAll()
	if len(entries) == 0 {
		return capSites(defaultTopSites, limit)
	}

	counts := make(map[string]int)
	firstURL := make(map[string]string)

	for _, e := range entries {
		domain := extractDomain(e.URL)
		if domain == "" {
			continue
		}
		counts[domain]++
		if _, exists := firstURL[domain]; !exists {
			firstURL[domain] = e.URL
		}
	}

	sites := make([]TopSite, 0, len(counts))
	for domain, count := range counts {
		sites = append(sites, TopSite{
			Domain: domain,
			Visits: count,
			URL:    firstURL[domain],
		})
	}

	sort.SliceStable(sites, func(i, j int) bool {
		return sites[i].Visits > sites[j].Visits
	})

	if len(sites) == 0 {
		return capSites(defaultTopSites, limit)
	}

	return capSites(sites, limit)
}

func capSites(sites []TopSite, limit int) []TopSite {
	if len(sites) > limit {
		return sites[:limit]
	}
	return sites
}

// PinnedTopSite lets a user manually pin a shortcut on the new tab page
type PinnedShortcut struct {
	ID    int
	Title string
	URL   string
}

type ShortcutManager struct {
	mu        sync.Mutex
	Shortcuts []PinnedShortcut
	NextID    int
}

var shortcutManagerNewTab = &ShortcutManager{
	Shortcuts: []PinnedShortcut{},
	NextID:    1,
}

// AddShortcut pins a custom shortcut to the new tab page
func (sm *ShortcutManager) AddShortcut(title string, url string) []PinnedShortcut {
	_, _ = sm.AddLink(title, url)
	return sm.GetShortcuts()
}

func normalizeSavedLinkURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("website URL is required")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return "", fmt.Errorf("enter a valid HTTP or HTTPS website URL")
	}
	return parsed.String(), nil
}

func (sm *ShortcutManager) AddLink(title string, rawURL string) (*PinnedShortcut, error) {
	normalizedURL, err := normalizeSavedLinkURL(rawURL)
	if err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = strings.TrimPrefix(extractDomain(normalizedURL), "www.")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, s := range sm.Shortcuts {
		if strings.EqualFold(s.URL, normalizedURL) {
			return nil, fmt.Errorf("this website is already saved")
		}
	}

	if sm.NextID < 1 {
		sm.NextID = 1
	}
	link := PinnedShortcut{ID: sm.NextID, Title: title, URL: normalizedURL}
	sm.NextID++
	sm.Shortcuts = append(sm.Shortcuts, link)
	return &link, nil
}

func (sm *ShortcutManager) UpdateLink(id int, title string, rawURL string) error {
	normalizedURL, err := normalizeSavedLinkURL(rawURL)
	if err != nil {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = strings.TrimPrefix(extractDomain(normalizedURL), "www.")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	for i := range sm.Shortcuts {
		if sm.Shortcuts[i].ID != id {
			if strings.EqualFold(sm.Shortcuts[i].URL, normalizedURL) {
				return fmt.Errorf("this website is already saved")
			}
			continue
		}
		sm.Shortcuts[i].Title = title
		sm.Shortcuts[i].URL = normalizedURL
		return nil
	}
	return fmt.Errorf("saved link not found")
}

// RemoveShortcut removes a pinned shortcut by index
func (sm *ShortcutManager) RemoveShortcut(index int) []PinnedShortcut {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if index < 0 || index >= len(sm.Shortcuts) {
		return sm.Shortcuts
	}

	newList := []PinnedShortcut{}
	for i, s := range sm.Shortcuts {
		if i != index {
			newList = append(newList, s)
		}
	}
	sm.Shortcuts = newList
	return sm.Shortcuts
}

func (sm *ShortcutManager) RemoveLink(id int) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for i := range sm.Shortcuts {
		if sm.Shortcuts[i].ID == id {
			sm.Shortcuts = append(sm.Shortcuts[:i], sm.Shortcuts[i+1:]...)
			return true
		}
	}
	return false
}

// GetShortcuts returns all pinned shortcuts
func (sm *ShortcutManager) GetShortcuts() []PinnedShortcut {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	shortcuts := make([]PinnedShortcut, len(sm.Shortcuts))
	copy(shortcuts, sm.Shortcuts)
	return shortcuts
}

func (sm *ShortcutManager) ReplaceAll(shortcuts []PinnedShortcut) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.Shortcuts = make([]PinnedShortcut, 0, len(shortcuts))
	usedIDs := make(map[int]bool, len(shortcuts))
	nextID := 1
	for _, shortcut := range shortcuts {
		normalizedURL, err := normalizeSavedLinkURL(shortcut.URL)
		if err != nil {
			continue
		}
		if shortcut.ID < 1 || usedIDs[shortcut.ID] {
			for usedIDs[nextID] {
				nextID++
			}
			shortcut.ID = nextID
		}
		usedIDs[shortcut.ID] = true
		if shortcut.ID >= nextID {
			nextID = shortcut.ID + 1
		}
		shortcut.URL = normalizedURL
		shortcut.Title = strings.TrimSpace(shortcut.Title)
		if shortcut.Title == "" {
			shortcut.Title = strings.TrimPrefix(extractDomain(normalizedURL), "www.")
		}
		sm.Shortcuts = append(sm.Shortcuts, shortcut)
	}
	sm.NextID = nextID
}
