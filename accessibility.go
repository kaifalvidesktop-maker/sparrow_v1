package main

import (
	"fmt"
	"sync"
)

// ---------------------------------------------------
// ACCESSIBILITY MANAGER
// ---------------------------------------------------

type AccessibilityManager struct {
	mu             sync.Mutex
	FontScale      float64
	HighContrast   bool
	ReduceMotion   bool
	UnderlineLinks bool
}

func (a *AccessibilityManager) SetFontScale(scale float64) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.FontScale = scale
	return scale
}

func (a *AccessibilityManager) ToggleHighContrast() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.HighContrast = !a.HighContrast
	return a.HighContrast
}

func (a *AccessibilityManager) ToggleReduceMotion() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ReduceMotion = !a.ReduceMotion
	return a.ReduceMotion
}

func (a *AccessibilityManager) ToggleUnderlineLinks() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.UnderlineLinks = !a.UnderlineLinks
	return a.UnderlineLinks
}

func (a *AccessibilityManager) GetAll() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	return map[string]any{
		"fontScale":      a.FontScale,
		"highContrast":   a.HighContrast,
		"reduceMotion":   a.ReduceMotion,
		"underlineLinks": a.UnderlineLinks,
	}
}

var accessibility = &AccessibilityManager{FontScale: 1.0}

func buildAccessibilityJS(scale float64, highContrast bool, reducemotion bool, underlinks bool) string {
	contrastValue := "100%"
	if highContrast {
		contrastValue = "150%"
	}
	return fmt.Sprintf(`
		document.body.style.zoom = "%f";
		document.body.style.filter = "contrast(%s)";
	`, scale, contrastValue)
}
