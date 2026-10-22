package main

import (
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------
// TAB SEARCH / QUICK SWITCHER (Ctrl+Shift+A style)
// Fuzzy-ish matching across open tab titles and URLs
// ---------------------------------------------------

type TabSearchResult struct {
	TabID int
	Title string
	URL   string
	Score int
}

type TabSearchEngine struct {
	mu sync.Mutex
}

var tabSearchEngine = &TabSearchEngine{}

// Search ranks open tabs by how well they match the query
func (tse *TabSearchEngine) Search(query string) []TabSearchResult {
	tse.mu.Lock()
	defer tse.mu.Unlock()

	query = strings.ToLower(strings.TrimSpace(query))
	allTabs := tabManager.GetAllTabs()

	if query == "" {
		results := make([]TabSearchResult, 0, len(allTabs))
		for _, t := range allTabs {
			results = append(results, TabSearchResult{TabID: t.ID, Title: t.Title, URL: t.URL, Score: 0})
		}
		return results
	}

	results := []TabSearchResult{}
	for _, t := range allTabs {
		score := scoreMatch(strings.ToLower(t.Title), query) + scoreMatch(strings.ToLower(t.URL), query)
		if score > 0 {
			results = append(results, TabSearchResult{TabID: t.ID, Title: t.Title, URL: t.URL, Score: score})
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// scoreMatch gives higher scores for prefix matches, lower for substring matches
func scoreMatch(text string, query string) int {
	if text == "" || query == "" {
		return 0
	}
	if strings.HasPrefix(text, query) {
		return 100
	}
	if strings.Contains(text, query) {
		return 50
	}

	// Fuzzy: check if all query characters appear in order
	qi := 0
	for i := 0; i < len(text) && qi < len(query); i++ {
		if text[i] == query[qi] {
			qi++
		}
	}
	if qi == len(query) {
		return 10
	}
	return 0
}

const tabSearchOverlayJS = `
window.__sparrowTabSearchOpen = true;
`
