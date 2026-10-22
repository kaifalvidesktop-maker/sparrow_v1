package main

import "testing"

func TestPermissionLookupDoesNotCreateEntries(t *testing.T) {
	pm := &PermissionManager{Sites: map[string]*SitePermissions{}}
	got := pm.Lookup("news.example.com")
	if got["location"] != "ask" || len(pm.Sites) != 0 {
		t.Fatalf("lookup must default to ask and not store anything: %v %d", got, len(pm.Sites))
	}
}

func TestPermissionLookupMatchesWWWAndParentDomain(t *testing.T) {
	pm := &PermissionManager{Sites: map[string]*SitePermissions{}}
	pm.Set("example.com", "location", "deny")
	for _, host := range []string{"example.com", "www.example.com", "m.example.com", "a.b.example.com"} {
		if pm.Lookup(host)["location"] != "deny" {
			t.Errorf("%s should inherit example.com", host)
		}
	}
	if pm.Lookup("other.com")["location"] != "ask" {
		t.Error("unrelated host must stay at ask")
	}
}
