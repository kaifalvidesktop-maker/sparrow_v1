package main

import "testing"

func TestRecordNavigationPreservesTabTitle(t *testing.T) {
	tm := &TabManager{Tabs: []*Tab{}, NextID: 1, ActiveID: 0}
	first := tm.NewTab("https://first.example")
	if first.Title != "first.example" {
		t.Fatalf("first tab title = %q, want %q", first.Title, "first.example")
	}

	tm.SetTitle(first.ID, "First page")
	tm.RecordNavigation(first.ID, "https://second.example")
	if got := tm.Tabs[0].Title; got != "First page" {
		t.Fatalf("updated tab title = %q, want %q", got, "First page")
	}
	if got := tm.GetActiveTab().Title; got != "First page" {
		t.Fatalf("active tab title = %q, want %q", got, "First page")
	}
}

func TestSetTabFavicon(t *testing.T) {
	tm := &TabManager{Tabs: []*Tab{}, NextID: 1, ActiveID: 0}
	tab := tm.NewTab("https://example.com")

	tm.SetFavicon(tab.ID, "https://example.com/favicon.ico")
	if got := tm.Tabs[0].Favicon; got != "https://example.com/favicon.ico" {
		t.Fatalf("tab favicon = %q, want %q", got, "https://example.com/favicon.ico")
	}
}

func TestTabDisplayTitleInternalRoutes(t *testing.T) {
	tests := map[string]string{
		"sparrow://home":      "New Tab",
		"sparrow://settings":  "Settings",
		"sparrow://bookmarks": "Bookmarks",
		"sparrow://links":     "Saved Links",
		"sparrow://pdf?url=x": "PDF Viewer",
	}
	for url, want := range tests {
		if got := tabDisplayTitle(url); got != want {
			t.Errorf("tabDisplayTitle(%q) = %q, want %q", url, got, want)
		}
	}
}

func TestBackForwardNavigationAcrossWebsites(t *testing.T) {
	tm := &TabManager{Tabs: []*Tab{}, NextID: 1, ActiveID: 0}
	tab := tm.NewTab("https://first.example")
	tm.RecordNavigation(tab.ID, "https://second.example")
	tm.RecordNavigation(tab.ID, "https://third.example")

	if got, ok := tm.GoBack(tab.ID); !ok || got != "https://second.example" {
		t.Fatalf("GoBack() = (%q, %t), want (%q, true)", got, ok, "https://second.example")
	}
	if !tm.CanGoForward(tab.ID) {
		t.Fatal("CanGoForward() = false after going back")
	}
	if got, ok := tm.GoForward(tab.ID); !ok || got != "https://third.example" {
		t.Fatalf("GoForward() = (%q, %t), want (%q, true)", got, ok, "https://third.example")
	}
	if tm.CanGoForward(tab.ID) {
		t.Fatal("CanGoForward() = true at the end of history")
	}
}
