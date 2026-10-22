package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ---------------------------------------------------
// PERSISTENCE LAYER
// ---------------------------------------------------
//
// Sparrow keeps browsing history and autofill data in RAM. Saved login
// records are persisted as AES-GCM ciphertext and never as plain passwords.
//
// Everything else that a normal browser is expected to remember between
// launches - bookmarks, settings, zoom levels, and the recently-closed
// tab list - is persisted here as a single small JSON file under the
// user's config directory:
//
//   Windows: %AppData%\Sparrow\state.json
//   Linux:   ~/.config/Sparrow/state.json
//   macOS:   ~/Library/Application Support/Sparrow/state.json
//
// The file is written on a 30s autosave tick and once more on shutdown,
// so a crash loses at most ~30s of bookmark/setting changes - nothing
// sensitive is ever at risk since none of it is stored here.

type persistedState struct {
	SavedAt time.Time `json:"savedAt"`

	// Settings snapshot
	Theme               string               `json:"theme"`
	DefaultSearchEngine string               `json:"defaultSearchEngine"`
	Homepage            string               `json:"homepage"`
	JSEnabled           bool                 `json:"jsEnabled"`
	ImagesEnabled       bool                 `json:"imagesEnabled"`
	DownloadPath        string               `json:"downloadPath"`
	CustomTheme         *CustomThemeSnapshot `json:"customTheme,omitempty"`
	HomeAnimation       string               `json:"homeAnimation,omitempty"`
	HomeWallpaper       string               `json:"homeWallpaper,omitempty"`

	// Autofill snapshot is also saved to disk, but not other session data.
	Autofill AutofillData `json:"autofill"`

	// Saved favorites (local to the PC, user-managed list).
	Favorites    []FavoriteEntry  `json:"favorites"`
	SavedLinks   []PinnedShortcut `json:"savedLinks"`
	Passwords    []StoredPassword `json:"passwords"`
	PasswordSalt []byte           `json:"passwordSalt,omitempty"`
	AppLock      AppLockSnapshot  `json:"appLock,omitempty"`
}

var (
	autosaveStopChan chan struct{}
	autosaveOnce     sync.Once
)

func stateFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "Sparrow")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "state.json"), nil
}

// buildPersistedState takes a consistent snapshot of everything we persist.
func buildPersistedState() *persistedState {
	settings.mu.Lock()
	s := persistedState{
		SavedAt:             time.Now(),
		Theme:               settings.Theme,
		DefaultSearchEngine: settings.DefaultSearchEngine,
		Homepage:            settings.Homepage,
		JSEnabled:           settings.JSEnabled,
		ImagesEnabled:       settings.ImagesEnabled,
		DownloadPath:        settings.DownloadPath,
	}
	settings.mu.Unlock()
	customThemeSnapshot := customTheme.Snapshot()

	s.CustomTheme = &customThemeSnapshot
	s.HomeAnimation = homeAnimationManager.Get()
	s.HomeWallpaper = wallpaperManager.GetSelectedID()
	s.Autofill = getAutofillProfile()
	s.Favorites = favoriteManager.GetAll()
	s.SavedLinks = shortcutManagerNewTab.GetShortcuts()
	s.Passwords = passwordVault.Snapshot()
	s.PasswordSalt = passwordVault.SaltSnapshot()
	s.AppLock = appLock.Snapshot()
	return &s
}

// applyPersistedState restores a loaded snapshot into the live managers.
func applyPersistedState(s *persistedState) {
	if s == nil {
		return
	}

	settings.mu.Lock()
	if s.Theme != "" {
		settings.Theme = s.Theme
	}
	if s.DefaultSearchEngine != "" {
		settings.DefaultSearchEngine = s.DefaultSearchEngine
	}
	if s.Homepage != "" {
		settings.Homepage = s.Homepage
	}
	settings.JSEnabled = s.JSEnabled
	settings.ImagesEnabled = s.ImagesEnabled
	if s.DownloadPath != "" {
		settings.DownloadPath = s.DownloadPath
	}
	settings.mu.Unlock()
	if s.CustomTheme != nil {
		customTheme.Restore(*s.CustomTheme)
	}
	homeAnimationManager.Restore(s.HomeAnimation)
	wallpaperManager.Restore(s.HomeWallpaper)

	saveAutofillProfile(s.Autofill)
	favoriteManager.ReplaceAll(s.Favorites)
	shortcutManagerNewTab.ReplaceAll(s.SavedLinks)
	passwordVault.Restore(s.Passwords)
	passwordVault.SetSalt(s.PasswordSalt)
	appLock.Restore(s.AppLock)
}

// loadPersistedState reads state.json if present. Missing file is not
// an error - it just means first run.
func loadPersistedState() *persistedState {
	path, err := stateFilePath()
	if err != nil {
		log.Println("[persistence] could not resolve config dir:", err)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("[persistence] read failed:", err)
		}
		return nil
	}

	var s persistedState
	if err := json.Unmarshal(data, &s); err != nil {
		log.Println("[persistence] corrupt state.json, ignoring:", err)
		return nil
	}
	return &s
}

// savePersistedState writes the current snapshot to disk atomically
// (write to a temp file, then rename) so a crash mid-write can't corrupt it.
func savePersistedState() error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}

	snapshot := buildPersistedState()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// startAutosave saves state every interval until stop is closed.
func startAutosave(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := savePersistedState(); err != nil {
					log.Println("[persistence] autosave failed:", err)
				}
			case <-stop:
				return
			}
		}
	}()
}

// InitPersistence loads existing state and starts the 30s background ticker.
func InitPersistence() {
	if state := loadPersistedState(); state != nil {
		applyPersistedState(state)
		log.Println("[persistence] persisted state loaded successfully")
	}

	autosaveOnce.Do(func() {
		autosaveStopChan = make(chan struct{})
		startAutosave(30*time.Second, autosaveStopChan)
		log.Println("[persistence] 30s autosave timer started")
	})
}

// StopPersistence halts background autosave and does a final write to disk on shutdown.
func StopPersistence() {
	if autosaveStopChan != nil {
		select {
		case <-autosaveStopChan:
		default:
			close(autosaveStopChan)
		}
	}
	if err := savePersistedState(); err != nil {
		log.Println("[persistence] final shutdown save failed:", err)
	} else {
		log.Println("[persistence] state saved successfully on shutdown")
	}
}

// --- tiny local helpers so this file has zero extra imports beyond stdlib ---

func persistIntToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func persistStrToInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	neg := false
	i := 0
	if s[0] == '-' {
		neg = true
		i = 1
	}
	if i == len(s) {
		return 0, false
	}
	n := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, true
}
