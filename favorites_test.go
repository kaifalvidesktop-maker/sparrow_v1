package main

import "testing"

func TestFavoritePersistenceRoundTrip(t *testing.T) {
	favoriteManager.Add("Example", "https://example.com")
	state := buildPersistedState()
	if len(state.Favorites) == 0 {
		t.Fatal("expected favorite to be included in persisted state")
	}

	favoriteManager.RemoveByURL("https://example.com")
	if favoriteManager.IsFavorite("https://example.com") {
		t.Fatal("favorite should have been removed before restore")
	}

	applyPersistedState(state)
	if !favoriteManager.IsFavorite("https://example.com") {
		t.Fatal("favorite was not restored from persisted state")
	}
}
