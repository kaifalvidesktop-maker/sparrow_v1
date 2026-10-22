package main

import "testing"

func newTestTab(tm *TabManager, url string) *Tab {
	tab := tm.NewTab(url)
	tm.RecordNavigation(tab.ID, url)
	return tab
}

func TestBackForwardAfterLinkClicks(t *testing.T) {
	tm := &TabManager{NextID: 1}
	tab := newTestTab(tm, "https://a.com/")
	tm.SyncLocation(tab.ID, "https://a.com", false) // landed
	tm.SyncLocation(tab.ID, "https://a.com/page2", false) // link click
	tm.SyncLocation(tab.ID, "https://a.com/page3", false) // link click

	if !tm.CanGoBack(tab.ID) || tm.CanGoForward(tab.ID) {
		t.Fatal("expected back only")
	}
	if u, _ := tm.GoBack(tab.ID); u != "https://a.com/page2" {
		t.Fatalf("back1 got %s", u)
	}
	tm.SyncLocation(tab.ID, "https://a.com/page2", false)
	if u, _ := tm.GoBack(tab.ID); u != "https://a.com/" {
		t.Fatalf("back2 got %s", u)
	}
	tm.SyncLocation(tab.ID, "https://a.com/", false)
	if !tm.CanGoForward(tab.ID) {
		t.Fatal("forward should be available")
	}
	if u, _ := tm.GoForward(tab.ID); u != "https://a.com/page2" {
		t.Fatalf("forward got %s", u)
	}
}

func TestRedirectReplacesEntry(t *testing.T) {
	tm := &TabManager{NextID: 1}
	tab := newTestTab(tm, "http://a.com")
	tm.SyncLocation(tab.ID, "https://www.a.com/", false) // redirect of our own navigation
	if tm.CanGoBack(tab.ID) {
		t.Fatal("redirect must not add an entry")
	}
	tm.SyncLocation(tab.ID, "https://www.a.com/x", false)
	if !tm.CanGoBack(tab.ID) {
		t.Fatal("link click must add an entry")
	}
}

func TestSameDocumentDoesNotAddEntry(t *testing.T) {
	tm := &TabManager{NextID: 1}
	tab := newTestTab(tm, "https://a.com/")
	tm.SyncLocation(tab.ID, "https://a.com/", false)
	tm.SyncLocation(tab.ID, "https://a.com/#/route", true)
	if tm.CanGoBack(tab.ID) {
		t.Fatal("in-page route change must not add a Go history entry")
	}
}
