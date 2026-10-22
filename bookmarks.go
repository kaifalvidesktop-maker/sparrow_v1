package main

import (
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// BOOKMARK SYSTEM (RAM ONLY)
// ---------------------------------------------------

type Bookmark struct {
	ID        int
	Title     string
	URL       string
	Folder    string
	CreatedAt time.Time
}

type BookmarkManager struct {
	mu        sync.Mutex
	Bookmarks []*Bookmark
	NextID    int
}

var bookmarkManager = &BookmarkManager{
	Bookmarks: []*Bookmark{},
	NextID:    1,
}

// Add a new bookmark
func (bm *BookmarkManager) Add(title string, url string, folder string) *Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if folder == "" {
		folder = "Uncategorized"
	}

	b := &Bookmark{
		ID:        bm.NextID,
		Title:     title,
		URL:       url,
		Folder:    folder,
		CreatedAt: time.Now(),
	}

	bm.Bookmarks = append(bm.Bookmarks, b)
	bm.NextID++

	return b
}

// Remove a bookmark by ID
func (bm *BookmarkManager) Remove(id int) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	newList := []*Bookmark{}
	removed := false

	for _, b := range bm.Bookmarks {
		if b.ID == id {
			removed = true
			continue
		}
		newList = append(newList, b)
	}

	bm.Bookmarks = newList
	return removed
}

// Remove a bookmark by URL (used for quick toggle)
func (bm *BookmarkManager) RemoveByURL(url string) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	newList := []*Bookmark{}
	removed := false

	for _, b := range bm.Bookmarks {
		if b.URL == url {
			removed = true
			continue
		}
		newList = append(newList, b)
	}

	bm.Bookmarks = newList
	return removed
}

// Check if a URL is already bookmarked
func (bm *BookmarkManager) IsBookmarked(url string) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, b := range bm.Bookmarks {
		if b.URL == url {
			return true
		}
	}
	return false
}

// Toggle bookmark: add if not present, remove if present
func (bm *BookmarkManager) Toggle(title string, url string) bool {
	if bm.IsBookmarked(url) {
		bm.RemoveByURL(url)
		return false // now NOT bookmarked
	}
	bm.Add(title, url, "Uncategorized")
	return true // now bookmarked
}

// Get all bookmarks
func (bm *BookmarkManager) GetAll() []*Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.Bookmarks
}

// Search bookmarks by title or URL keyword
func (bm *BookmarkManager) Search(keyword string) []*Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	keyword = strings.ToLower(keyword)
	results := []*Bookmark{}

	for _, b := range bm.Bookmarks {
		if strings.Contains(strings.ToLower(b.Title), keyword) ||
			strings.Contains(strings.ToLower(b.URL), keyword) {
			results = append(results, b)
		}
	}
	return results
}

// Get bookmarks grouped by folder (for UI sidebar)
func (bm *BookmarkManager) GetFolders() map[string][]*Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	grouped := make(map[string][]*Bookmark)
	for _, b := range bm.Bookmarks {
		grouped[b.Folder] = append(grouped[b.Folder], b)
	}
	return grouped

}

// ---------------------------------------------------
// FAVORITES (local saved websites, persisted on disk)
// ---------------------------------------------------

type FavoriteEntry struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}

type FavoriteManager struct {
	mu        sync.Mutex
	Favorites []FavoriteEntry
	NextID    int
}

var favoriteManager = &FavoriteManager{
	Favorites: []FavoriteEntry{},
	NextID:    1,
}

// normFavURL makes "https://a.com" and "https://a.com/#x" count as the same page.
func normFavURL(u string) string {
	u = strings.TrimSpace(u)
	if i := strings.Index(u, "#"); i >= 0 {
		u = u[:i]
	}
	return strings.ToLower(strings.TrimRight(u, "/"))
}

func (fm *FavoriteManager) Add(title string, url string) *FavoriteEntry {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	title = strings.TrimSpace(title)
	url = strings.TrimSpace(url)
	if url == "" {
		return nil
	}
	if title == "" {
		title = extractDomain(url)
		if title == "" {
			title = url
		}
	}
	if fm.isFavoriteLocked(url) {
		return nil
	}

	entry := &FavoriteEntry{ID: fm.NextID, Title: title, URL: url, CreatedAt: time.Now()}
	fm.NextID++
	fm.Favorites = append(fm.Favorites, *entry)
	return entry
}

func (fm *FavoriteManager) RemoveByURL(url string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	filtered := []FavoriteEntry{}
	removed := false
	for _, item := range fm.Favorites {
		if normFavURL(item.URL) == normFavURL(url) {
			removed = true
			continue
		}
		filtered = append(filtered, item)
	}
	fm.Favorites = filtered
	return removed
}

func (fm *FavoriteManager) RemoveByID(id int) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	filtered := []FavoriteEntry{}
	removed := false
	for _, item := range fm.Favorites {
		if item.ID == id {
			removed = true
			continue
		}
		filtered = append(filtered, item)
	}
	fm.Favorites = filtered
	return removed
}

func (fm *FavoriteManager) IsFavorite(url string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.isFavoriteLocked(url)
}

func (fm *FavoriteManager) isFavoriteLocked(url string) bool {
	for _, item := range fm.Favorites {
		if normFavURL(item.URL) == normFavURL(url) {
			return true
		}
	}
	return false
}

func (fm *FavoriteManager) Toggle(title string, url string) bool {
	if fm.IsFavorite(url) {
		fm.RemoveByURL(url)
		return false
	}
	fm.Add(title, url)
	return true
}

func (fm *FavoriteManager) GetAll() []FavoriteEntry {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	items := make([]FavoriteEntry, len(fm.Favorites))
	copy(items, fm.Favorites)
	return items
}

func (fm *FavoriteManager) ReplaceAll(items []FavoriteEntry) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.Favorites = append([]FavoriteEntry{}, items...)
	if len(fm.Favorites) > 0 {
		maxID := 0
		for _, item := range fm.Favorites {
			if item.ID > maxID {
				maxID = item.ID
			}
		}
		fm.NextID = maxID + 1
	} else {
		fm.NextID = 1
	}
}
