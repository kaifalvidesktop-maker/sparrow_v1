package main

import (
	"fmt"
	"sync"
)

// ---------------------------------------------------
// HOME PAGE WALLPAPER / BACKGROUND CUSTOMIZATION
// ---------------------------------------------------

type WallpaperOption struct {
	ID       string
	Name     string
	CSSValue string
}

type HomeAnimationOption struct {
	ID   string
	Name string
}

var homeAnimationOptions = []HomeAnimationOption{
	{ID: "default", Name: "Default Sparrow home"},
	{ID: "bubbles", Name: "Floating bubbles"},
	{ID: "rain", Name: "Soft rain"},
	{ID: "rainbow", Name: "Rainbow ribbons"},
	{ID: "snowfall", Name: "Gentle snowfall"},
	{ID: "fireflies", Name: "Fireflies"},
	{ID: "aurora", Name: "Northern lights"},
	{ID: "shooting-stars", Name: "Shooting stars"},
	{ID: "confetti", Name: "Confetti"},
	{ID: "waves", Name: "Ocean waves"},
	{ID: "petals", Name: "Falling petals"},
	{ID: "matrix", Name: "Digital rain"},
	{ID: "sparkles", Name: "Sparkles"},
	{ID: "clouds", Name: "Drifting clouds"},
	{ID: "meteors", Name: "Meteor shower"},
	{ID: "hearts", Name: "Floating hearts"},
	{ID: "galaxy", Name: "Galaxy swirl"},
	{ID: "ripples", Name: "Water ripples"},
	{ID: "leaves", Name: "Autumn leaves"},
	{ID: "embers", Name: "Glowing embers"},
	{ID: "neon", Name: "Neon pulses"},
}

type HomeAnimationManager struct {
	mu       sync.Mutex
	Selected string
}

var homeAnimationManager = &HomeAnimationManager{Selected: "default"}

func (am *HomeAnimationManager) Set(id string) (string, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, option := range homeAnimationOptions {
		if option.ID == id {
			am.Selected = id
			return am.Selected, nil
		}
	}
	return am.Selected, fmt.Errorf("unknown home animation %q", id)
}

func (am *HomeAnimationManager) Get() string {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.isValidLocked(am.Selected) {
		return am.Selected
	}
	return "default"
}

func (am *HomeAnimationManager) Restore(id string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if am.isValidLocked(id) {
		am.Selected = id
	} else {
		am.Selected = "default"
	}
}

func (am *HomeAnimationManager) isValidLocked(id string) bool {
	for _, option := range homeAnimationOptions {
		if option.ID == id {
			return true
		}
	}
	return false
}

func (am *HomeAnimationManager) Options() []HomeAnimationOption {
	options := make([]HomeAnimationOption, len(homeAnimationOptions))
	copy(options, homeAnimationOptions)
	return options
}

var wallpaperOptions = []WallpaperOption{
	{ID: "default", Name: "SPARROW Dark", CSSValue: "linear-gradient(180deg, #0b0b0d 0%, #121214 100%)"},
	{ID: "midnight", Name: "Midnight Stars", CSSValue: "radial-gradient(1px 1px at 12% 24%, #fff 99%, transparent 100%), radial-gradient(2px 2px at 78% 36%, #b9d8ff 99%, transparent 100%), radial-gradient(1px 1px at 58% 72%, #fff 99%, transparent 100%), linear-gradient(145deg, #020617 0%, #101b46 52%, #24104a 100%)"},
	{ID: "forest", Name: "Enchanted Forest", CSSValue: "radial-gradient(ellipse at 18% 95%, rgba(52,211,153,.58), transparent 42%), radial-gradient(ellipse at 82% 8%, rgba(132,204,22,.3), transparent 38%), linear-gradient(150deg, #041b16 0%, #075143 52%, #092b24 100%)"},
	{ID: "wine", Name: "Velvet Wine", CSSValue: "radial-gradient(ellipse at 82% 15%, rgba(244,63,94,.52), transparent 38%), radial-gradient(ellipse at 12% 88%, rgba(168,85,247,.4), transparent 42%), linear-gradient(135deg, #190719 0%, #51152e 48%, #21102f 100%)"},
	{ID: "sunset", Name: "Tropical Sunset", CSSValue: "radial-gradient(ellipse at 50% 110%, #ffd166 0%, #fb7185 24%, transparent 58%), linear-gradient(180deg, #27104b 0%, #8b2871 48%, #f97345 100%)"},
	{ID: "solid", Name: "Pure Black", CSSValue: "#000000"},
	{ID: "ocean", Name: "Ocean Depths", CSSValue: "radial-gradient(ellipse at 18% 22%, rgba(34,211,238,.55), transparent 40%), radial-gradient(ellipse at 82% 80%, rgba(37,99,235,.65), transparent 48%), linear-gradient(150deg, #031b3b 0%, #075985 52%, #082f49 100%)"},
	{ID: "aurora", Name: "Aurora Sky", CSSValue: "radial-gradient(ellipse at 18% 18%, rgba(45,212,191,.8), transparent 42%), radial-gradient(ellipse at 70% 14%, rgba(167,139,250,.78), transparent 42%), radial-gradient(ellipse at 82% 86%, rgba(34,197,94,.48), transparent 46%), linear-gradient(150deg, #07162b 0%, #10294c 54%, #192044 100%)"},
	{ID: "rose", Name: "Rose Quartz", CSSValue: "radial-gradient(ellipse at 20% 15%, rgba(255,255,255,.7), transparent 35%), radial-gradient(ellipse at 88% 80%, rgba(244,114,182,.72), transparent 45%), linear-gradient(135deg, #4c1d4f 0%, #be568d 48%, #f3a6ad 100%)"},
	{ID: "desert", Name: "Desert Dusk", CSSValue: "radial-gradient(ellipse at 82% 12%, #ffd166 0%, rgba(255,209,102,.25) 15%, transparent 32%), linear-gradient(165deg, #311b38 0%, #a44247 48%, #e89b5b 100%)"},
	{ID: "glacier", Name: "Blue Glacier", CSSValue: "radial-gradient(ellipse at 16% 85%, rgba(103,232,249,.78), transparent 46%), radial-gradient(ellipse at 90% 10%, rgba(96,165,250,.74), transparent 40%), linear-gradient(145deg, #082f49 0%, #0e7490 48%, #bae6fd 100%)"},
	{ID: "lavender", Name: "Lavender Haze", CSSValue: "radial-gradient(ellipse at 78% 18%, rgba(216,180,254,.82), transparent 42%), radial-gradient(ellipse at 12% 85%, rgba(129,140,248,.68), transparent 45%), linear-gradient(145deg, #241044 0%, #6843a5 52%, #c084fc 100%)"},
	{ID: "citrus", Name: "Citrus Pop", CSSValue: "radial-gradient(ellipse at 85% 20%, #fde047 0%, transparent 36%), radial-gradient(ellipse at 12% 80%, rgba(74,222,128,.8), transparent 44%), linear-gradient(135deg, #14532d 0%, #16a34a 48%, #a3e635 100%)"},
	{ID: "nebula", Name: "Cosmic Nebula", CSSValue: "radial-gradient(ellipse at 25% 35%, rgba(236,72,153,.78), transparent 38%), radial-gradient(ellipse at 76% 58%, rgba(124,58,237,.88), transparent 44%), radial-gradient(ellipse at 54% 8%, rgba(59,130,246,.58), transparent 38%), linear-gradient(145deg, #09051d 0%, #24104b 52%, #100d33 100%)"},
}

type WallpaperManager struct {
	mu       sync.Mutex
	Selected string
}

var wallpaperManager = &WallpaperManager{
	Selected: "default",
}

// SetWallpaper changes the active wallpaper by ID
func (wm *WallpaperManager) SetWallpaper(id string) (string, bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, opt := range wallpaperOptions {
		if opt.ID == id {
			wm.Selected = id
			return wm.Selected, true
		}
	}
	return wm.Selected, false
}

func (wm *WallpaperManager) GetSelectedID() string {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for _, opt := range wallpaperOptions {
		if opt.ID == wm.Selected {
			return wm.Selected
		}
	}
	return "default"
}

func (wm *WallpaperManager) Restore(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for _, opt := range wallpaperOptions {
		if opt.ID == id {
			wm.Selected = id
			return
		}
	}
	wm.Selected = "default"
}

// GetSelected returns the CSS value for the currently active wallpaper
func (wm *WallpaperManager) GetSelectedCSS() string {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, opt := range wallpaperOptions {
		if opt.ID == wm.Selected {
			return opt.CSSValue
		}
	}
	return wallpaperOptions[0].CSSValue
}

// GetAllOptions returns the full wallpaper list for the settings UI
func (wm *WallpaperManager) GetAllOptions() []WallpaperOption {
	return wallpaperOptions
}
