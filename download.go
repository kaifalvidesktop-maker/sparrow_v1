package main

import (
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type DownloadItem struct {
	ID         int
	FileName   string
	Path       string
	URL        string
	Progress   int
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
}

type DownloadManager struct {
	mu     sync.Mutex
	Items  []*DownloadItem
	NextID int
}

var downloadManager = &DownloadManager{Items: []*DownloadItem{}, NextID: 1}

func (dm *DownloadManager) Start(rawURL string) *DownloadItem {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dm.NextID < 1 {
		dm.NextID = 1
	}
	item := &DownloadItem{
		ID:        dm.NextID,
		FileName:  fileNameFromURL(rawURL),
		URL:       rawURL,
		Status:    "downloading",
		StartedAt: time.Now(),
	}
	dm.Items = append(dm.Items, item)
	dm.NextID++
	return item
}

func (dm *DownloadManager) SetFile(id int, name, fullPath string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	for _, item := range dm.Items {
		if item.ID == id {
			item.FileName = name
			item.Path = fullPath
			return
		}
	}
}

func (dm *DownloadManager) SetStatus(id int, status string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	for _, item := range dm.Items {
		if item.ID == id {
			item.Status = status
			if status == "completed" || status == "failed" || status == "cancelled" {
				item.FinishedAt = time.Now()
			}
			return
		}
	}
}

func (dm *DownloadManager) UpdateProgress(id, progress int) {
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	for _, item := range dm.Items {
		if item.ID == id {
			item.Progress = progress
			return
		}
	}
}

func (dm *DownloadManager) Complete(id int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	for _, item := range dm.Items {
		if item.ID == id && item.Status != "cancelled" {
			item.Status = "completed"
			item.Progress = 100
			item.FinishedAt = time.Now()
			return
		}
	}
}

func (dm *DownloadManager) Fail(id int) {
	dm.SetStatus(id, "failed")
}

func (dm *DownloadManager) Cancel(id int) {
	dm.SetStatus(id, "cancelled")
}

func (dm *DownloadManager) ClearFinished() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	active := dm.Items[:0]
	for _, item := range dm.Items {
		if item.Status == "downloading" || item.Status == "queued" || item.Status == "paused" {
			active = append(active, item)
		}
	}
	dm.Items = active
}

func (dm *DownloadManager) ClearAll() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.Items = nil
	dm.NextID = 1
}

func (dm *DownloadManager) GetAll() []*DownloadItem {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	items := make([]*DownloadItem, len(dm.Items))
	for i, item := range dm.Items {
		if item != nil {
			copy := *item
			items[i] = &copy
		}
	}
	return items
}

func (dm *DownloadManager) Import(items []*DownloadItem) int {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	existing := make(map[string]struct{}, len(dm.Items))
	usedIDs := make(map[int]struct{}, len(dm.Items))
	for _, item := range dm.Items {
		if item == nil {
			continue
		}
		existing[downloadImportKey(item)] = struct{}{}
		usedIDs[item.ID] = struct{}{}
	}
	if dm.NextID < 1 {
		dm.NextID = 1
	}

	imported := 0
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.URL) == "" {
			continue
		}
		key := downloadImportKey(item)
		if _, found := existing[key]; found {
			continue
		}
		copy := *item
		if copy.ID < 1 {
			copy.ID = dm.NextID
		}
		if _, found := usedIDs[copy.ID]; found {
			copy.ID = dm.NextID
			for {
				if _, found := usedIDs[copy.ID]; !found {
					break
				}
				copy.ID++
			}
		}
		if copy.ID >= dm.NextID {
			dm.NextID = copy.ID + 1
		}
		if copy.Progress < 0 {
			copy.Progress = 0
		} else if copy.Progress > 100 {
			copy.Progress = 100
		}
		switch copy.Status {
		case "completed", "failed", "cancelled", "downloading", "queued", "paused":
		default:
			copy.Status = "failed"
		}
		if copy.Status == "downloading" || copy.Status == "queued" || copy.Status == "paused" {
			copy.Status = "failed"
			copy.FinishedAt = time.Now()
		}
		dm.Items = append(dm.Items, &copy)
		existing[key] = struct{}{}
		usedIDs[copy.ID] = struct{}{}
		imported++
	}
	return imported
}

func fileNameFromURL(rawURL string) string {
	name := ""
	if parsed, err := url.Parse(strings.TrimSpace(rawURL)); err == nil {
		name = path.Base(parsed.Path)
		if decoded, err := url.PathUnescape(name); err == nil {
			name = decoded
		}
	}
	if name == "" || name == "." || name == "/" {
		return "unknown_file"
	}
	return name
}

func downloadImportKey(item *DownloadItem) string {
	return item.URL + "\x00" + item.FileName + "\x00" + item.StartedAt.UTC().Format(time.RFC3339Nano)
}
