package main

import "testing"

func TestHomeAnimationManagerOptionsAndSelection(t *testing.T) {
	manager := &HomeAnimationManager{Selected: "default"}
	options := manager.Options()
	if len(options) != 21 {
		t.Fatalf("animation options = %d, want 21 including default", len(options))
	}

	selected, err := manager.Set("bubbles")
	if err != nil || selected != "bubbles" {
		t.Fatalf("Set(bubbles) = (%q, %v), want (bubbles, nil)", selected, err)
	}
	if got := manager.Get(); got != "bubbles" {
		t.Fatalf("Get() = %q, want bubbles", got)
	}

	selected, err = manager.Set("default")
	if err != nil || selected != "default" {
		t.Fatalf("Set(default) = (%q, %v), want (default, nil)", selected, err)
	}
}

func TestHomeAnimationManagerRejectsUnknownSelection(t *testing.T) {
	manager := &HomeAnimationManager{Selected: "default"}
	if _, err := manager.Set("not-an-animation"); err == nil {
		t.Fatal("Set() accepted an unknown animation")
	}
	if got := manager.Get(); got != "default" {
		t.Fatalf("unknown animation changed selection to %q", got)
	}
}

func TestHomeAnimationManagerRestoresPersistedSelection(t *testing.T) {
	manager := &HomeAnimationManager{Selected: "default"}
	manager.Restore("aurora")
	if got := manager.Get(); got != "aurora" {
		t.Fatalf("Restore(aurora) produced %q", got)
	}
	manager.Restore("invalid")
	if got := manager.Get(); got != "default" {
		t.Fatalf("invalid restored selection = %q, want default", got)
	}
}

func TestWallpaperManagerSelectionAndRestore(t *testing.T) {
	manager := &WallpaperManager{Selected: "default"}
	selected, ok := manager.SetWallpaper("midnight")
	if !ok || selected != "midnight" {
		t.Fatalf("SetWallpaper(midnight) = (%q, %t), want (midnight, true)", selected, ok)
	}
	if manager.GetSelectedID() != "midnight" {
		t.Fatalf("GetSelectedID() = %q, want midnight", manager.GetSelectedID())
	}
	manager.Restore("unknown")
	if manager.GetSelectedID() != "default" {
		t.Fatalf("invalid wallpaper restore = %q, want default", manager.GetSelectedID())
	}
}

func TestWallpaperPresetsHaveDistinctBackgrounds(t *testing.T) {
	if len(wallpaperOptions) < 12 {
		t.Fatalf("wallpaper presets = %d, want at least 12", len(wallpaperOptions))
	}
	ids := make(map[string]bool, len(wallpaperOptions))
	backgrounds := make(map[string]bool, len(wallpaperOptions))
	for _, option := range wallpaperOptions {
		if option.ID == "" || option.Name == "" || option.CSSValue == "" {
			t.Fatalf("wallpaper preset is incomplete: %+v", option)
		}
		if ids[option.ID] {
			t.Errorf("duplicate wallpaper ID %q", option.ID)
		}
		if backgrounds[option.CSSValue] {
			t.Errorf("wallpaper %q shares a background with another preset", option.Name)
		}
		ids[option.ID] = true
		backgrounds[option.CSSValue] = true
	}
}
