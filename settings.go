package main

import (
	"log"
	"net/url"
	"strings"
	"sync"
)

// ---------------------------------------------------
// SETTINGS SYSTEM (RAM ONLY - RESETS ON RESTART)
// ---------------------------------------------------

type BrowserSettings struct {
	mu                  sync.Mutex
	Theme               string // "dark" or "light"
	DefaultSearchEngine string // "google", "bing", "brave", "duckduckgo", "yahoo"
	Homepage            string
	JSEnabled           bool
	ImagesEnabled       bool
	DownloadPath        string
}

var settings = &BrowserSettings{
	Theme:               "dark",
	DefaultSearchEngine: "google",
	Homepage:            "sparrow://home",
	JSEnabled:           true,
	ImagesEnabled:       true,
	DownloadPath:        "Downloads",
}

// Get all settings as a map (for sending to UI)
func (s *BrowserSettings) GetAll() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"theme":         s.Theme,
		"searchEngine":  s.DefaultSearchEngine,
		"homepage":      s.Homepage,
		"jsEnabled":     s.JSEnabled,
		"imagesEnabled": s.ImagesEnabled,
		"downloadPath":  s.DownloadPath,
	}
}

// Change theme (dark/light)
func (s *BrowserSettings) SetTheme(theme string) string {
	s.mu.Lock()
	theme = strings.ToLower(strings.TrimSpace(theme))
	if theme != "dark" && theme != "light" {
		theme = "dark"
	}
	s.Theme = theme
	value := s.Theme
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save theme change:", err)
	}
	return value
}

// Change default search engine
func (s *BrowserSettings) SetSearchEngine(engine string) string {
	s.mu.Lock()
	engine = strings.ToLower(strings.TrimSpace(engine))
	valid := map[string]bool{"google": true, "bing": true, "brave": true, "duckduckgo": true, "yahoo": true}

	if !valid[engine] {
		engine = "google"
	}
	s.DefaultSearchEngine = engine
	value := s.DefaultSearchEngine
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save search engine change:", err)
	}
	return value
}

// Change homepage URL
func (s *BrowserSettings) SetHomepage(url string) string {
	s.mu.Lock()
	url = strings.TrimSpace(url)
	if url == "" {
		url = "sparrow://home"
	}
	s.Homepage = url
	value := s.Homepage
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save homepage change:", err)
	}
	return value
}

// Toggle JavaScript on/off
func (s *BrowserSettings) ToggleJS() bool {
	s.mu.Lock()
	s.JSEnabled = !s.JSEnabled
	value := s.JSEnabled
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save JS toggle:", err)
	}
	return value
}

// Toggle Image loading on/off
func (s *BrowserSettings) ToggleImages() bool {
	s.mu.Lock()
	s.ImagesEnabled = !s.ImagesEnabled
	value := s.ImagesEnabled
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save image toggle:", err)
	}
	return value
}

// Change download folder path
func (s *BrowserSettings) SetDownloadPath(path string) string {
	s.mu.Lock()
	path = strings.TrimSpace(path)
	if path == "" {
		path = "Downloads"
	}
	s.DownloadPath = path
	value := s.DownloadPath
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save download path:", err)
	}
	return value
}

// Reset everything back to default values
func (s *BrowserSettings) ResetToDefault() map[string]any {
	s.mu.Lock()
	s.Theme = "dark"
	s.DefaultSearchEngine = "google"
	s.Homepage = "sparrow://home"
	s.JSEnabled = true
	s.ImagesEnabled = true
	s.DownloadPath = "Downloads"
	result := map[string]any{
		"theme":         s.Theme,
		"searchEngine":  s.DefaultSearchEngine,
		"homepage":      s.Homepage,
		"jsEnabled":     s.JSEnabled,
		"imagesEnabled": s.ImagesEnabled,
		"downloadPath":  s.DownloadPath,
	}
	s.mu.Unlock()

	if err := savePersistedState(); err != nil {
		log.Println("[settings] failed to save reset defaults:", err)
	}
	return result
}

func searchEngineName(engine string) string {
	switch strings.ToLower(strings.TrimSpace(engine)) {
	case "bing":
		return "Bing"
	case "brave":
		return "Brave Search"
	case "duckduckgo":
		return "DuckDuckGo"
	case "yahoo":
		return "Yahoo"
	default:
		return "Google"
	}
}

// Build a search URL based on the currently selected search engine
func buildSearchURL(query string) string {
	query = url.QueryEscape(strings.TrimSpace(query))

	switch strings.ToLower(strings.TrimSpace(settings.DefaultSearchEngine)) {
	case "bing":
		return "https://www.bing.com/search?q=" + query
	case "brave":
		return "https://search.brave.com/search?q=" + query
	case "duckduckgo":
		return "https://duckduckgo.com/?q=" + query
	case "yahoo":
		return "https://search.yahoo.com/search?p=" + query
	default:
		return "https://www.google.com/search?q=" + query
	}
}
