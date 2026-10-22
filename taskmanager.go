package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type TaskEntry struct {
	TabID      int       `json:"tab_id"`
	Title      string    `json:"title"`
	MemoryMB   int       `json:"memory_mb"`
	CPUPercent float64   `json:"cpu_percent"`
	NetworkKB  int       `json:"network_kb"`
	LastActive time.Time `json:"last_active"`
	CreatedAt  time.Time `json:"created_at"`
	IsActive   bool      `json:"is_active"`
}

type TaskManager struct {
	mu        sync.RWMutex
	Stats     map[int]*TaskEntry
	MaxMemory int
}

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		Stats:     make(map[int]*TaskEntry),
		MaxMemory: 400, // Max simulated memory in MB
	}
}

// Track updates or creates a task entry
func (tm *TaskManager) Track(tabID int, title string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry, exists := tm.Stats[tabID]
	if !exists {
		entry = &TaskEntry{
			TabID:      tabID,
			Title:      title,
			MemoryMB:   30 + tabID%50, // Base memory varies by tab
			CPUPercent: 0.2 + float64(tabID%10)/10,
			NetworkKB:  0,
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			IsActive:   true,
		}
		tm.Stats[tabID] = entry
		return
	}

	// Update existing entry
	if title != "" {
		entry.Title = title
	}

	// Simulate memory growth (with some randomness)
	if entry.MemoryMB < tm.MaxMemory {
		increase := 1 + tabID%3
		entry.MemoryMB += increase
		if entry.MemoryMB > tm.MaxMemory {
			entry.MemoryMB = tm.MaxMemory
		}
	}

	// Simulate CPU usage (varies)
	if time.Since(entry.LastActive).Seconds() < 5 {
		entry.CPUPercent = 0.5 + float64(tabID%5)/5
	} else {
		entry.CPUPercent = 0.1 // CPU drops when inactive
	}

	// Simulate network activity
	if entry.NetworkKB > 0 {
		entry.NetworkKB = max(
			// Decreases over time
			entry.NetworkKB-10, 0)
	}

	entry.LastActive = time.Now()
	entry.IsActive = true
}

// TrackNetworkActivity simulates network usage
func (tm *TaskManager) TrackNetworkActivity(tabID int, kb int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if entry, exists := tm.Stats[tabID]; exists {
		entry.NetworkKB += kb
		if entry.NetworkKB > 9999 {
			entry.NetworkKB = 9999 // Cap at 9999 KB
		}
	}
}

// RemoveTab removes a tab from task manager
func (tm *TaskManager) RemoveTab(tabID int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.Stats, tabID)
}

// GetTask returns a single task entry
func (tm *TaskManager) GetTask(tabID int) *TaskEntry {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.Stats[tabID]
}

// GetAllTasks returns all task entries sorted by memory usage
func (tm *TaskManager) GetAllTasks() []*TaskEntry {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*TaskEntry, 0, len(tm.Stats))
	for _, entry := range tm.Stats {
		tasks = append(tasks, entry)
	}

	// Sort by memory usage (highest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].MemoryMB > tasks[j].MemoryMB
	})

	return tasks
}

// UpdateTitle updates a tab's title
func (tm *TaskManager) UpdateTitle(tabID int, title string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if entry, exists := tm.Stats[tabID]; exists {
		entry.Title = title
	}
}

// GetTotalMemory returns total memory usage
func (tm *TaskManager) GetTotalMemory() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	total := 0
	for _, entry := range tm.Stats {
		total += entry.MemoryMB
	}
	return total
}

// GetActiveTabCount returns number of active tabs
func (tm *TaskManager) GetActiveTabCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.Stats)
}

// CleanupInactive removes inactive tabs (optional)
func (tm *TaskManager) CleanupInactive(maxAge time.Duration) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	removed := 0
	now := time.Now()
	for tabID, entry := range tm.Stats {
		if now.Sub(entry.LastActive) > maxAge {
			delete(tm.Stats, tabID)
			removed++
		}
	}
	return removed
}

// GenerateTaskManagerHTML generates HTML for task manager page
func GenerateTaskManagerHTML(tm *TaskManager) string {
	tasks := tm.GetAllTasks()

	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Task Manager - Sparrow Browser</title>
    <style>
        body { 
            background: #1a1a2e; 
            color: #eee; 
            font-family: 'Segoe UI', Arial, sans-serif;
            padding: 20px;
        }
        h1 { color: #e94560; border-bottom: 2px solid #e94560; padding-bottom: 10px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th { background: #16213e; padding: 12px; text-align: left; }
        td { padding: 10px; border-bottom: 1px solid #333; }
        tr:hover { background: #16213e; }
        .memory-bar { 
            background: #e94560; 
            height: 20px; 
            border-radius: 10px;
            transition: width 0.3s;
        }
        .memory-bg { 
            background: #333; 
            border-radius: 10px; 
            width: 100%;
            height: 20px;
        }
        .cpu-text { color: #ffd93d; }
        .active { color: #6bcb77; }
        .inactive { color: #888; }
        .stats { 
            background: #16213e; 
            padding: 15px; 
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .stat-item { display: inline-block; margin-right: 30px; }
        .stat-value { color: #e94560; font-weight: bold; font-size: 1.2em; }
        .refresh-btn {
            background: #e94560;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
        }
        .refresh-btn:hover { background: #c73652; }
    </style>
</head>
<body>
    <h1>⚡ Task Manager</h1>`)

	// Stats
	totalMem := tm.GetTotalMemory()
	activeTabs := tm.GetActiveTabCount()
	html.WriteString(fmt.Sprintf(`
    <div class="stats">
        <div class="stat-item">📊 Active Tabs: <span class="stat-value">%d</span></div>
        <div class="stat-item">💾 Total Memory: <span class="stat-value">%d MB</span></div>
        <div class="stat-item">⏱️ Last Updated: <span class="stat-value">%s</span></div>
        <button class="refresh-btn" onclick="location.reload()">🔄 Refresh</button>
    </div>`, activeTabs, totalMem, time.Now().Format("15:04:05")))

	if len(tasks) == 0 {
		html.WriteString(`<p>No active tabs.</p>`)
	} else {
		html.WriteString(`<table>
    <tr>
        <th>Tab ID</th>
        <th>Title</th>
        <th>Memory</th>
        <th>CPU</th>
        <th>Network</th>
        <th>Status</th>
        <th>Active Since</th>
    </tr>`)

		for _, task := range tasks {
			memPercent := float64(task.MemoryMB) / float64(tm.MaxMemory) * 100
			status := "🟢 Active"
			statusClass := "active"
			if !task.IsActive {
				status = "⚪ Inactive"
				statusClass = "inactive"
			}

			html.WriteString(fmt.Sprintf(`
    <tr>
        <td>#%d</td>
        <td>%s</td>
        <td>
            <div class="memory-bg">
                <div class="memory-bar" style="width: %.1f%%"></div>
            </div>
            <small>%d MB</small>
        </td>
        <td class="cpu-text">%.1f%%</td>
        <td>%d KB</td>
        <td class="%s">%s</td>
        <td>%s</td>
    </tr>`, task.TabID, task.Title, memPercent, task.MemoryMB,
				task.CPUPercent, task.NetworkKB, statusClass, status,
				task.CreatedAt.Format("15:04:05")))
		}
		html.WriteString(`</table>`)
	}

	html.WriteString(`
</body>
</html>`)
	return html.String()
}
