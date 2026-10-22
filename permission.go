package main

import (
	"strings"
	"sync"
)

// ---------------------------------------------------
// PER-SITE PERMISSION MANAGER
// Controls camera, microphone, location, notifications
// ---------------------------------------------------

type PermissionState string

const (
	PermissionAsk   PermissionState = "ask"
	PermissionAllow PermissionState = "allow"
	PermissionDeny  PermissionState = "deny"
)

type SitePermissions struct {
	Camera        PermissionState
	Microphone    PermissionState
	Location      PermissionState
	Notifications PermissionState
}

type PermissionManager struct {
	mu    sync.Mutex
	Sites map[string]*SitePermissions
}

var permissionManager = &PermissionManager{
	Sites: make(map[string]*SitePermissions),
}

// Get (or create default) permissions for a domain
func (pm *PermissionManager) GetFor(domain string) *SitePermissions {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	domain = strings.ToLower(domain)

	if perms, ok := pm.Sites[domain]; ok {
		return perms
	}

	fresh := &SitePermissions{
		Camera:        PermissionAsk,
		Microphone:    PermissionAsk,
		Location:      PermissionAsk,
		Notifications: PermissionAsk,
	}
	pm.Sites[domain] = fresh
	return fresh
}

// Set a specific permission type for a domain
func (pm *PermissionManager) Set(domain string, permType string, state string) map[string]string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	domain = strings.ToLower(domain)
	perms, ok := pm.Sites[domain]
	if !ok {
		perms = &SitePermissions{
			Camera:        PermissionAsk,
			Microphone:    PermissionAsk,
			Location:      PermissionAsk,
			Notifications: PermissionAsk,
		}
		pm.Sites[domain] = perms
	}

	val := PermissionState(state)
	if val != PermissionAllow && val != PermissionDeny && val != PermissionAsk {
		val = PermissionAsk
	}

	switch permType {
	case "camera":
		perms.Camera = val
	case "microphone":
		perms.Microphone = val
	case "location":
		perms.Location = val
	case "notifications":
		perms.Notifications = val
	}

	return pm.toMap(perms)
}

func (pm *PermissionManager) toMap(p *SitePermissions) map[string]string {
	return map[string]string{
		"camera":        string(p.Camera),
		"microphone":    string(p.Microphone),
		"location":      string(p.Location),
		"notifications": string(p.Notifications),
	}
}

// Get all site permissions as a map (for a "Site Settings" UI page)
func (pm *PermissionManager) GetAllAsMap() map[string]map[string]string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	result := make(map[string]map[string]string)
	for domain, perms := range pm.Sites {
		result[domain] = pm.toMap(perms)
	}
	return result
}

// Reset permissions for one domain back to "ask"
func (pm *PermissionManager) ResetDomain(domain string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	domain = strings.ToLower(domain)
	delete(pm.Sites, domain)
}

// Reset ALL site permissions
func (pm *PermissionManager) ResetAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.Sites = make(map[string]*SitePermissions)
}
