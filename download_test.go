package main

import "testing"

func TestDownloadManagerClearAll(t *testing.T) {
	manager := &DownloadManager{NextID: 1}
	manager.Start("https://example.com/one.zip")
	manager.Complete(1)
	manager.Start("https://example.com/two.zip")

	manager.ClearAll()

	if downloads := manager.GetAll(); len(downloads) != 0 {
		t.Fatalf("ClearAll left %d downloads", len(downloads))
	}

	next := manager.Start("https://example.com/next.zip")
	if next.ID != 1 {
		t.Errorf("first ID after ClearAll = %d, want 1", next.ID)
	}
}
