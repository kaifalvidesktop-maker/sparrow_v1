package main

import (
	"strings"
	"sync"
)

// ---------------------------------------------------
// CUSTOM THEME EDITOR (user-defined accent color + shell tweaks)
// Applies only to the SPARROW shell UI (toolbar/tabs), not page content
// ---------------------------------------------------

type CustomTheme struct {
	mu          sync.Mutex
	AccentColor string
	TabRadius   int // px
	Compact     bool
}

type CustomThemeSnapshot struct {
	AccentColor string `json:"accentColor"`
	TabRadius   int    `json:"tabRadius"`
	Compact     bool   `json:"compact"`
}

var customTheme = &CustomTheme{
	AccentColor: "#000000",
	TabRadius:   10,
	Compact:     false,
}

var validHexColor = "0123456789abcdefABCDEF#"

// SetAccentColor validates and applies a new accent color
func (ct *CustomTheme) SetAccentColor(hex string) string {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	hex = strings.TrimSpace(hex)
	if isValidHex(hex) {
		ct.AccentColor = hex
	}
	return ct.AccentColor
}

func isValidHex(hex string) bool {
	if len(hex) != 7 || hex[0] != '#' {
		return false
	}
	for _, c := range hex[1:] {
		if !strings.ContainsRune(validHexColor, c) {
			return false
		}
	}
	return true
}

// SetTabRadius adjusts how rounded tab corners are (0-20px)
func (ct *CustomTheme) SetTabRadius(radius int) int {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if radius < 0 {
		radius = 0
	}
	if radius > 20 {
		radius = 20
	}
	ct.TabRadius = radius
	return ct.TabRadius
}

// ToggleCompact switches between compact and normal toolbar height
func (ct *CustomTheme) ToggleCompact() bool {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.Compact = !ct.Compact
	return ct.Compact
}

// GetAll returns the current custom theme state
func (ct *CustomTheme) GetAll() map[string]any {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return map[string]any{
		"accentColor": ct.AccentColor,
		"tabRadius":   ct.TabRadius,
		"compact":     ct.Compact,
	}
}

func (ct *CustomTheme) Snapshot() CustomThemeSnapshot {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return CustomThemeSnapshot{
		AccentColor: ct.AccentColor,
		TabRadius:   ct.TabRadius,
		Compact:     ct.Compact,
	}
}

func (ct *CustomTheme) Restore(snapshot CustomThemeSnapshot) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	if isValidHex(snapshot.AccentColor) {
		ct.AccentColor = snapshot.AccentColor
	}
	if snapshot.TabRadius >= 0 && snapshot.TabRadius <= 20 {
		ct.TabRadius = snapshot.TabRadius
	}
	ct.Compact = snapshot.Compact
}

// BuildThemeCSS generates CSS variables to inject into the shell
func (ct *CustomTheme) BuildThemeCSS() string {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	compactHeight := "48px"
	if ct.Compact {
		compactHeight = "38px"
	}

	return `
<style id="sparrow-custom-theme">
  #toolbar { height: ` + compactHeight + ` !important; }
  .tab-chip { border-radius: ` + itoa(ct.TabRadius) + `px ` + itoa(ct.TabRadius) + `px 0 0 !important; }
  #address-wrap:focus-within { border-color: ` + ct.AccentColor + ` !important; }
  .switch.on { background: ` + ct.AccentColor + ` !important; }
  .start-btn { background: ` + ct.AccentColor + ` !important; }
</style>
`
}
