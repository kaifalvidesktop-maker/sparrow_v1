package main

import "testing"

func TestShortcutManagerAddEditAndRemoveLink(t *testing.T) {
	manager := &ShortcutManager{NextID: 1}
	link, err := manager.AddLink("  Reading list  ", "example.com/articles")
	if err != nil {
		t.Fatalf("AddLink returned error: %v", err)
	}
	if link.ID != 1 || link.Title != "Reading list" || link.URL != "https://example.com/articles" {
		t.Fatalf("AddLink returned %+v", link)
	}

	if err := manager.UpdateLink(link.ID, "Research", "https://example.org"); err != nil {
		t.Fatalf("UpdateLink returned error: %v", err)
	}
	links := manager.GetShortcuts()
	if len(links) != 1 || links[0].Title != "Research" || links[0].URL != "https://example.org" {
		t.Fatalf("updated links = %+v", links)
	}

	if manager.RemoveLink(link.ID + 100) {
		t.Fatal("RemoveLink returned true for an unknown ID")
	}
	if !manager.RemoveLink(link.ID) || len(manager.GetShortcuts()) != 0 {
		t.Fatal("RemoveLink did not remove the saved link")
	}
}

func TestShortcutManagerRejectsInvalidAndDuplicateLinks(t *testing.T) {
	manager := &ShortcutManager{NextID: 1}
	if _, err := manager.AddLink("Bad", "javascript:alert(1)"); err == nil {
		t.Fatal("AddLink accepted a non-HTTP URL")
	}
	if _, err := manager.AddLink("First", "https://example.com"); err != nil {
		t.Fatalf("AddLink returned error: %v", err)
	}
	if _, err := manager.AddLink("Duplicate", "https://example.com"); err == nil {
		t.Fatal("AddLink accepted a duplicate URL")
	}
	if err := manager.UpdateLink(1, "Bad", "file:///etc/passwd"); err == nil {
		t.Fatal("UpdateLink accepted a non-HTTP URL")
	}
}

func TestShortcutManagerRestoresLegacyLinks(t *testing.T) {
	manager := &ShortcutManager{}
	manager.ReplaceAll([]PinnedShortcut{{Title: "Example", URL: "example.com"}})
	links := manager.GetShortcuts()
	if len(links) != 1 || links[0].ID == 0 || links[0].URL != "https://example.com" {
		t.Fatalf("legacy link restore = %+v", links)
	}
	if _, err := manager.AddLink("Next", "https://another.example"); err != nil {
		t.Fatalf("AddLink after restore returned error: %v", err)
	}
	if got := manager.GetShortcuts()[1].ID; got == links[0].ID {
		t.Fatalf("restored IDs were not advanced: both links have ID %d", got)
	}
}

func TestLinkSaverInternalRoute(t *testing.T) {
	page, ok := internalPageHTML("sparrow://links")
	if !ok || page != linkSaverPageHTML {
		t.Fatal("expected sparrow://links to resolve to the saved links page")
	}
}
