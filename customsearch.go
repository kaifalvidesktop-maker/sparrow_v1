package main

import (
	"net/url"
	"strings"
	"sync"
)

// ---------------------------------------------------
// CUSTOM SEARCH ENGINES (user-added, beyond the built-in three)
// A search engine is a name + a URL template with %s as the query placeholder
// ---------------------------------------------------

type CustomSearchEngine struct {
	ID       int
	Name     string
	Template string // e.g. "https://example.com/search?q=%s"
	IsActive bool
}

type CustomSearchManager struct {
	mu      sync.Mutex
	Engines []*CustomSearchEngine
	NextID  int
}

var customSearchManager = &CustomSearchManager{
	Engines: []*CustomSearchEngine{},
	NextID:  1,
}

// AddEngine registers a new custom search engine
func (csm *CustomSearchManager) AddEngine(name string, template string) (*CustomSearchEngine, bool) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	name = strings.TrimSpace(name)
	template = strings.TrimSpace(template)

	if name == "" || !strings.Contains(template, "%s") {
		return nil, false
	}

	engine := &CustomSearchEngine{
		ID:       csm.NextID,
		Name:     name,
		Template: template,
	}
	csm.Engines = append(csm.Engines, engine)
	csm.NextID++
	return engine, true
}

// RemoveEngine deletes a custom search engine
func (csm *CustomSearchManager) RemoveEngine(id int) bool {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	newList := []*CustomSearchEngine{}
	removed := false
	for _, e := range csm.Engines {
		if e.ID == id {
			removed = true
			continue
		}
		newList = append(newList, e)
	}
	csm.Engines = newList
	return removed
}

// SetActive marks one engine as the active default (deactivates others)
func (csm *CustomSearchManager) SetActive(id int) bool {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	found := false
	for _, e := range csm.Engines {
		if e.ID == id {
			e.IsActive = true
			found = true
		} else {
			e.IsActive = false
		}
	}
	return found
}

// GetActiveTemplate returns the URL template of the active custom engine, if any
func (csm *CustomSearchManager) GetActiveTemplate() string {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	for _, e := range csm.Engines {
		if e.IsActive {
			return e.Template
		}
	}
	return ""
}

// BuildSearchURL replaces %s in a template with the URL-encoded query
func BuildCustomSearchURL(template string, query string) string {
	encoded := url.QueryEscape(query)
	return strings.Replace(template, "%s", encoded, 1)
}

// GetAll returns all custom search engines
func (csm *CustomSearchManager) GetAll() []*CustomSearchEngine {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	return csm.Engines
}
