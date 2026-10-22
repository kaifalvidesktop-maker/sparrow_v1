package main

import (
	"sync"
	"time"
)

// ---------------------------------------------------
// NOTIFICATION CENTER (in-browser toast notifications)
// ---------------------------------------------------

type NotificationItem struct {
	ID        int
	Title     string
	Message   string
	Kind      string // "info", "success", "warning", "danger"
	CreatedAt time.Time
	Read      bool
}

type NotificationCenter struct {
	mu     sync.Mutex
	Items  []*NotificationItem
	NextID int
}

var notifCenter = &NotificationCenter{
	Items:  []*NotificationItem{},
	NextID: 1,
}

// Push adds a new notification
func (nc *NotificationCenter) Push(title string, message string, kind string) *NotificationItem {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	validKinds := map[string]bool{"info": true, "success": true, "warning": true, "danger": true}
	if !validKinds[kind] {
		kind = "info"
	}

	item := &NotificationItem{
		ID:        nc.NextID,
		Title:     title,
		Message:   message,
		Kind:      kind,
		CreatedAt: time.Now(),
		Read:      false,
	}

	nc.Items = append(nc.Items, item)
	nc.NextID++

	// Cap to last 50 notifications
	if len(nc.Items) > 50 {
		nc.Items = nc.Items[len(nc.Items)-50:]
	}

	return item
}

// MarkRead marks one notification as read
func (nc *NotificationCenter) MarkRead(id int) {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for _, item := range nc.Items {
		if item.ID == id {
			item.Read = true
			break
		}
	}
}

// MarkAllRead marks every notification as read
func (nc *NotificationCenter) MarkAllRead() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for _, item := range nc.Items {
		item.Read = true
	}
}

// UnreadCount returns how many notifications are unread
func (nc *NotificationCenter) UnreadCount() int {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	count := 0
	for _, item := range nc.Items {
		if !item.Read {
			count++
		}
	}
	return count
}

// GetAll returns all notifications, most recent first
func (nc *NotificationCenter) GetAll() []*NotificationItem {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	reversed := make([]*NotificationItem, len(nc.Items))
	for i, item := range nc.Items {
		reversed[len(nc.Items)-1-i] = item
	}
	return reversed
}

// ClearAll wipes the notification list
func (nc *NotificationCenter) ClearAll() {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.Items = []*NotificationItem{}
}
