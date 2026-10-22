package main

import (
	"errors"
	"slices"
	"sync"
)

// Color validation map
var validGroupColors = map[string]bool{
	"blue":   true,
	"red":    true,
	"green":  true,
	"yellow": true,
	"purple": true,
	"orange": true,
	"pink":   true,
	"gray":   true,
}

type TabGroup struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	TabIDs []int  `json:"tab_ids"`
}

type TabGroupManager struct {
	mu     sync.RWMutex
	Groups []*TabGroup
	NextID int
}

// NewTabGroupManager creates a new manager
func NewTabGroupManager() *TabGroupManager {
	return &TabGroupManager{
		Groups: make([]*TabGroup, 0),
		NextID: 1,
	}
}

// CreateGroup creates a new tab group
func (tgm *TabGroupManager) CreateGroup(name, color string, tabID int) (*TabGroup, error) {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	if name == "" {
		name = "Untitled Group"
	}

	if !validGroupColors[color] {
		color = "blue"
	}

	// Check if tab already in any group
	if existing := tgm.findGroupByTabIDLocked(tabID); existing != nil {
		return nil, errors.New("tab already in a group")
	}

	group := &TabGroup{
		ID:     tgm.NextID,
		Name:   name,
		Color:  color,
		TabIDs: []int{tabID},
	}

	tgm.Groups = append(tgm.Groups, group)
	tgm.NextID++
	return group, nil
}

// AddTabToGroup adds a tab to existing group
func (tgm *TabGroupManager) AddTabToGroup(groupID, tabID int) error {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	// Check if tab already in any group
	if existing := tgm.findGroupByTabIDLocked(tabID); existing != nil {
		return errors.New("tab already in a group")
	}

	for _, group := range tgm.Groups {
		if group.ID == groupID {
			// Check duplicate
			if slices.Contains(group.TabIDs, tabID) {
				return errors.New("tab already in this group")
			}
			group.TabIDs = append(group.TabIDs, tabID)
			return nil
		}
	}
	return errors.New("group not found")
}

// RemoveTabFromGroups removes a tab from all groups
func (tgm *TabGroupManager) RemoveTabFromGroups(tabID int) {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	for _, group := range tgm.Groups {
		for i, id := range group.TabIDs {
			if id == tabID {
				group.TabIDs = append(group.TabIDs[:i], group.TabIDs[i+1:]...)
				break
			}
		}
	}
	tgm.cleanupEmptyGroupsLocked()
}

// cleanupEmptyGroups removes groups with no tabs
func (tgm *TabGroupManager) cleanupEmptyGroupsLocked() {
	// Filter out empty groups
	filtered := make([]*TabGroup, 0, len(tgm.Groups))
	for _, group := range tgm.Groups {
		if len(group.TabIDs) > 0 {
			filtered = append(filtered, group)
		}
	}
	tgm.Groups = filtered
}

// RenameGroup renames a group
func (tgm *TabGroupManager) RenameGroup(groupID int, newName string) error {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	if newName == "" {
		return errors.New("name cannot be empty")
	}

	for _, group := range tgm.Groups {
		if group.ID == groupID {
			group.Name = newName
			return nil
		}
	}
	return errors.New("group not found")
}

// SetGroupColor changes group color
func (tgm *TabGroupManager) SetGroupColor(groupID int, color string) error {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	if !validGroupColors[color] {
		return errors.New("invalid color")
	}

	for _, group := range tgm.Groups {
		if group.ID == groupID {
			group.Color = color
			return nil
		}
	}
	return errors.New("group not found")
}

// GetGroupForTab returns group containing a tab
func (tgm *TabGroupManager) GetGroupForTab(tabID int) *TabGroup {
	tgm.mu.RLock()
	defer tgm.mu.RUnlock()
	return tgm.findGroupByTabIDLocked(tabID)
}

// findGroupByTabIDLocked finds group containing tab (caller must hold lock)
func (tgm *TabGroupManager) findGroupByTabIDLocked(tabID int) *TabGroup {
	for _, group := range tgm.Groups {
		if slices.Contains(group.TabIDs, tabID) {
			return group
		}
	}
	return nil
}

// GetAllGroups returns all groups
func (tgm *TabGroupManager) GetAllGroups() []*TabGroup {
	tgm.mu.RLock()
	defer tgm.mu.RUnlock()

	result := make([]*TabGroup, len(tgm.Groups))
	copy(result, tgm.Groups)
	return result
}

// GetGroupByID returns group by ID
func (tgm *TabGroupManager) GetGroupByID(groupID int) *TabGroup {
	tgm.mu.RLock()
	defer tgm.mu.RUnlock()

	for _, group := range tgm.Groups {
		if group.ID == groupID {
			return group
		}
	}
	return nil
}

// DeleteGroup deletes a group
func (tgm *TabGroupManager) DeleteGroup(groupID int) error {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	for i, group := range tgm.Groups {
		if group.ID == groupID {
			tgm.Groups = append(tgm.Groups[:i], tgm.Groups[i+1:]...)
			return nil
		}
	}
	return errors.New("group not found")
}

// MoveTabToGroup moves tab from one group to another
func (tgm *TabGroupManager) MoveTabToGroup(tabID, fromGroupID, toGroupID int) error {
	tgm.mu.Lock()
	defer tgm.mu.Unlock()

	// Find source group
	var sourceGroup *TabGroup
	for _, group := range tgm.Groups {
		if group.ID == fromGroupID {
			sourceGroup = group
			break
		}
	}
	if sourceGroup == nil {
		return errors.New("source group not found")
	}

	// Remove from source
	found := false
	for i, id := range sourceGroup.TabIDs {
		if id == tabID {
			sourceGroup.TabIDs = append(sourceGroup.TabIDs[:i], sourceGroup.TabIDs[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("tab not found in source group")
	}

	// Add to destination
	for _, group := range tgm.Groups {
		if group.ID == toGroupID {
			group.TabIDs = append(group.TabIDs, tabID)
			return nil
		}
	}

	// If destination not found, create new group
	newGroup := &TabGroup{
		ID:     tgm.NextID,
		Name:   "Group",
		Color:  "blue",
		TabIDs: []int{tabID},
	}
	tgm.Groups = append(tgm.Groups, newGroup)
	tgm.NextID++

	tgm.cleanupEmptyGroupsLocked()
	return nil
}
