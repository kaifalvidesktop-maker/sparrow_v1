package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

type ThemeColors struct {
	Background    string `json:"background"`
	Foreground    string `json:"foreground"`
	Primary       string `json:"primary"`
	Secondary     string `json:"secondary"`
	Accent        string `json:"accent"`
	Border        string `json:"border"`
	Hover         string `json:"hover"`
	TextPrimary   string `json:"text_primary"`
	TextSecondary string `json:"text_secondary"`
	Success       string `json:"success"`
	Error         string `json:"error"`
	Warning       string `json:"warning"`
	Info          string `json:"info"`
}

type ThemeManager struct {
	mu       sync.RWMutex
	current  Theme
	colors   map[Theme]ThemeColors
	filePath string
}

// NewThemeManager creates a new theme manager
func NewThemeManager(filePath string) *ThemeManager {
	tm := &ThemeManager{
		current:  ThemeDark,
		colors:   make(map[Theme]ThemeColors),
		filePath: filePath,
	}
	tm.initializeColors()
	tm.loadFromFile()
	return tm
}

// initializeColors sets default color schemes
func (tm *ThemeManager) initializeColors() {
	tm.colors[ThemeDark] = ThemeColors{
		Background:    "#1a1a2e",
		Foreground:    "#eee",
		Primary:       "#e94560",
		Secondary:     "#16213e",
		Accent:        "#6bcb77",
		Border:        "#333",
		Hover:         "#2a2a4e",
		TextPrimary:   "#eee",
		TextSecondary: "#888",
		Success:       "#6bcb77",
		Error:         "#dc3545",
		Warning:       "#ffd93d",
		Info:          "#4dabf7",
	}

	tm.colors[ThemeLight] = ThemeColors{
		Background:    "#f8f9fa",
		Foreground:    "#212529",
		Primary:       "#e94560",
		Secondary:     "#e9ecef",
		Accent:        "#2b8a3e",
		Border:        "#dee2e6",
		Hover:         "#e9ecef",
		TextPrimary:   "#212529",
		TextSecondary: "#6c757d",
		Success:       "#2b8a3e",
		Error:         "#dc3545",
		Warning:       "#f59f00",
		Info:          "#1c7ed6",
	}
}

// SetTheme changes the current theme
func (tm *ThemeManager) SetTheme(theme Theme) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.colors[theme]; !exists {
		theme = ThemeDark
	}

	tm.current = theme
	return tm.saveToFile()
}

// GetCurrentTheme returns the current theme
func (tm *ThemeManager) GetCurrentTheme() Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.current
}

// GetColors returns colors for current theme
func (tm *ThemeManager) GetColors() ThemeColors {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.colors[tm.current]
}

// GetColorsForTheme returns colors for specific theme
func (tm *ThemeManager) GetColorsForTheme(theme Theme) ThemeColors {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.colors[theme]
}

// GetAllThemes returns all available themes
func (tm *ThemeManager) GetAllThemes() []Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]Theme, 0, len(tm.colors))
	for theme := range tm.colors {
		themes = append(themes, theme)
	}
	return themes
}

// saveToFile saves theme preference
func (tm *ThemeManager) saveToFile() error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	data, err := json.MarshalIndent(tm.current, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tm.filePath, data, 0644)
}

// loadFromFile loads theme preference
func (tm *ThemeManager) loadFromFile() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	data, err := os.ReadFile(tm.filePath)
	if err != nil {
		return
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return
	}

	if _, exists := tm.colors[theme]; exists {
		tm.current = theme
	}
}

// GetCSS generates CSS for current theme
func (tm *ThemeManager) GetCSS() string {
	colors := tm.GetColors()
	return fmt.Sprintf(`
:root {
    --bg: %s;
    --fg: %s;
    --primary: %s;
    --secondary: %s;
    --accent: %s;
    --border: %s;
    --hover: %s;
    --text-primary: %s;
    --text-secondary: %s;
    --success: %s;
    --error: %s;
    --warning: %s;
    --info: %s;
}
body {
    background: var(--bg);
    color: var(--fg);
}
`, colors.Background, colors.Foreground, colors.Primary,
		colors.Secondary, colors.Accent, colors.Border,
		colors.Hover, colors.TextPrimary, colors.TextSecondary,
		colors.Success, colors.Error, colors.Warning, colors.Info)
}

// GenerateThemePageHTML generates HTML for theme settings
func (tm *ThemeManager) GenerateThemePageHTML() string {
	themes := tm.GetAllThemes()
	current := tm.GetCurrentTheme()

	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Themes - Sparrow Browser</title>
    <style>
        body { background: #1a1a2e; color: #eee; font-family: 'Segoe UI', Arial, sans-serif; padding: 20px; }
        h1 { color: #e94560; border-bottom: 2px solid #e94560; padding-bottom: 10px; }
        .theme-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .theme-card {
            background: #16213e;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            cursor: pointer;
            border: 2px solid transparent;
            transition: all 0.3s;
        }
        .theme-card:hover { transform: scale(1.02); }
        .theme-card.current { border-color: #6bcb77; }
        .theme-preview {
            width: 100%;
            height: 80px;
            border-radius: 4px;
            margin-bottom: 10px;
        }
        .theme-name { 
            font-weight: bold;
            text-transform: capitalize;
        }
        .theme-status {
            font-size: 0.8em;
            color: #888;
            margin-top: 5px;
        }
        .theme-status.active { color: #6bcb77; }
        .btn {
            background: #e94560;
            color: white;
            border: none;
            padding: 5px 15px;
            border-radius: 5px;
            cursor: pointer;
            margin-top: 10px;
        }
        .btn:hover { background: #c73652; }
    </style>
</head>
<body>
    <h1>🎨 Themes</h1>
    <p style="color:#888;">Choose your preferred theme</p>
    
    <div class="theme-grid">`)

	for _, theme := range themes {
		colors := tm.GetColorsForTheme(theme)
		isCurrent := current == theme
		currentClass := ""
		statusText := "Click to apply"
		if isCurrent {
			currentClass = "current"
			statusText = "✓ Current theme"
		}

		html.WriteString(fmt.Sprintf(`
    <div class="theme-card %s" onclick="applyTheme('%s')">
        <div class="theme-preview" style="background:%s;"></div>
        <div class="theme-name">%s</div>
        <div class="theme-status %s">%s</div>
        <button class="btn">%s</button>
    </div>`,
			currentClass, theme, colors.Background,
			theme,
			func() string {
				if isCurrent {
					return "active"
				}
				return ""
			}(),
			statusText,
			func() string {
				if isCurrent {
					return "✅ Applied"
				}
				return "Apply"
			}()))
	}

	html.WriteString(`
    </div>
    
    <script>
        function applyTheme(theme) {
            window.location.href = 'sparrow://applytheme?theme=' + theme;
        }
    </script>
</body>
</html>`)
	return html.String()
}
