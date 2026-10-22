package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	webview "github.com/webview/webview_go"
)

// ---------------------------------------------------
// HISTORY SYSTEM (RAM ONLY - NEVER TOUCHES DISK)
// ---------------------------------------------------

type HistoryEntry struct {
	URL   string
	Title string
	Time  time.Time
}

type RAMHistory struct {
	mu      sync.Mutex
	Entries []HistoryEntry
}

var history = &RAMHistory{}

var buildID = "development"

type NavResult struct {
	Blocked bool   `json:"blocked"`
	URL     string `json:"url"`
}

type NavState struct {
	URL        string `json:"url"`
	CanBack    bool   `json:"canBack"`
	CanForward bool   `json:"canForward"`
	TabID      int    `json:"tabId"`
}

func dataURI(html string) string {
	return "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString([]byte(html))
}

func firstRunMarkerPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "Sparrow")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "first_run_done"), nil
}

func isFirstRunForBuild(marker string, currentBuildID string) (bool, error) {
	data, err := os.ReadFile(marker)
	if err == nil && strings.TrimSpace(string(data)) == currentBuildID {
		return false, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err := os.WriteFile(marker, []byte(currentBuildID), 0644); err != nil {
		return false, err
	}
	return true, nil
}

func isFirstRun() (bool, error) {
	marker, err := firstRunMarkerPath()
	if err != nil {
		return false, err
	}
	return isFirstRunForBuild(marker, buildID)
}

func buildPDFViewerURL(rawURL string) string {
	return "sparrow://pdf?url=" + url.QueryEscape(rawURL)
}

func internalPageHTML(sparrowURL string) (string, bool) {
	// Internal routes may contain query parameters (for example sparrow://pdf?url=...).
	// Match the route itself instead of requiring an exact string.
	route := sparrowURL
	if i := strings.IndexByte(route, '?'); i >= 0 {
		route = route[:i]
	}
	switch route {
	case "sparrow://home":
		return homePageHTML, true
	case "sparrow://settings":
		return settingsPageHTML, true
	case "sparrow://bookmarks":
		return bookmarksPageHTML, true
	case "sparrow://favorites":
		return favoritesPageHTML, true
	case "sparrow://downloads":
		return downloadsPageHTML, true
	case "sparrow://history":
		return historyPageHTML, true
	case "sparrow://shortcuts":
		return shortcutsPageHTML, true
	case "sparrow://links":
		return linkSaverPageHTML, true
	case "sparrow://about":
		return aboutPageHTML, true
	case "sparrow://site-settings":
		return siteSettingsPageHTML, true
	case "sparrow://cookies":
		return cookiesPageHTML, true
	case "sparrow://welcome":
		return welcomePageHTML, true
	case "sparrow://autofill":
		return autofillPageHTML, true
	case "sparrow://feedback":
		return feedbackPageHTML, true
	case "sparrow://updates":
		return updatesPageHTML, true
	case "sparrow://backup":
		return backupPageHTML, true
	case "sparrow://pdf":
		return pdfViewerPageHTML, true
	case "sparrow://customize":
		return customizePageHTML, true
	case "sparrow://lock":
		return appLockPageHTML, true
	}
	return "", false
}

func applyWebViewZoom(w webview.WebView, level float64) {
	if level < 0.5 {
		level = 0.5
	}
	if level > 3.0 {
		level = 3.0
	}
	js := fmt.Sprintf(`(function(){if(window.__sparrowSetZoom){window.__sparrowSetZoom(%0.2f);}})();`, level)
	w.Eval(js)
}

func navigateTo(w webview.WebView, target string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return
	}
	if appLock.IsLocked() && target != "sparrow://lock" {
		target = "sparrow://lock"
	}

	// Any PDF (sparrow://pdf?url=..., file:///x.pdf, https://x/y.pdf, C:\\x.pdf)
	// always opens in the embedded viewer, including back/forward/reload/tab switch.
	if strings.HasPrefix(target, "sparrow://pdf") || isPDFURL(target) {
		pdfSource := target
		if strings.HasPrefix(target, "sparrow://pdf") {
			pdfSource = ""
			if raw, err := url.Parse(target); err == nil {
				pdfSource = raw.Query().Get("url")
			}
		}
		if pdfSource != "" {
			if viewer, err := sparrowPDF.ViewerURL(pdfSource); err == nil {
				w.Dispatch(func() { w.Navigate(viewer) })
				return
			}
		}
	}

	if html, ok := internalPageHTML(target); ok {
		// Keep the PDF query available to the PDF page while still using the
		// correct internal HTML document.
		if strings.HasPrefix(target, "sparrow://pdf?") {
			if raw, err := url.Parse(target); err == nil {
				if pdfURL := raw.Query().Get("url"); pdfURL != "" {
					html = strings.ReplaceAll(html, "__SPARROW_PDF_URL__", pdfURL)
				}
			}
		}
		w.Dispatch(func() { w.Navigate(dataURI(html)) })
		return
	}

	navigateExternal(w, target)
}

func (h *RAMHistory) Add(url string, title string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Entries = append(h.Entries, HistoryEntry{
		URL:   url,
		Title: title,
		Time:  time.Now(),
	})
}

func (h *RAMHistory) DeleteSince(cutoff time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	newEntries := []HistoryEntry{}
	for _, entry := range h.Entries {
		if entry.Time.Before(cutoff) {
			newEntries = append(newEntries, entry)
		}
	}
	h.Entries = newEntries
}

func (h *RAMHistory) DeleteAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Entries = []HistoryEntry{}
}

func (h *RAMHistory) GetAll() []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries := make([]HistoryEntry, len(h.Entries))
	copy(entries, h.Entries)
	return entries
}

func (h *RAMHistory) Import(entries []HistoryEntry) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	existing := make(map[HistoryEntry]struct{}, len(h.Entries))
	for _, entry := range h.Entries {
		existing[entry] = struct{}{}
	}

	imported := 0
	for _, entry := range entries {
		if strings.TrimSpace(entry.URL) == "" {
			continue
		}
		if _, found := existing[entry]; found {
			continue
		}
		h.Entries = append(h.Entries, entry)
		existing[entry] = struct{}{}
		imported++
	}
	return imported
}

func (h *RAMHistory) GetRecent(limit int) []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	if limit <= 0 {
		return []HistoryEntry{}
	}

	start := 0
	if len(h.Entries) > limit {
		start = len(h.Entries) - limit
	}

	entries := make([]HistoryEntry, len(h.Entries)-start)
	copy(entries, h.Entries[start:])
	return entries
}

// ---------------------------------------------------
// SPARROW HOME PAGE
// ---------------------------------------------------

// ---------------------------------------------------
// FILE / URL LAUNCH SUPPORT
// ---------------------------------------------------

func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	return fileURLFromSlashPath(filepath.ToSlash(abs))
}

// fileURLFromSlashPath builds file:///C:/dir/a%20b.pdf from "C:/dir/a b.pdf".
// Without the leading slash Go emits file://C:/..., which parses back with
// "C:" as the host and breaks local files.
func fileURLFromSlashPath(p string) string {
	if len(p) >= 2 && p[1] == ':' {
		p = "/" + p
	}
	u := &url.URL{Scheme: "file", Path: p}
	return u.String()
}

func launchTargetFromArgs() string {
	if len(os.Args) < 2 {
		return ""
	}
	arg := strings.TrimSpace(os.Args[1])
	if arg == "" || strings.HasPrefix(arg, "--") {
		return ""
	}

	// Windows file associations pass a local path here.
	if !strings.Contains(arg, "://") {
		if info, err := os.Stat(arg); err == nil && !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(arg))
			switch ext {
			case ".html", ".htm":
				return arg
			case ".pdf":
				return fileURL(arg)
			}
		}
	}

	// A URL launched from a shortcut or command line.
	return arg
}

// ---------------------------------------------------
// MAIN FUNCTION - CREATES BROWSER INSTANCE
// ---------------------------------------------------

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(strings.TrimSpace(os.Args[1])) {
		case "--register-file-associations":
			if err := registerSparrowFileAssociations(); err != nil {
				fmt.Println("Sparrow association registration failed:", err)
				return
			}
			fmt.Println("Sparrow HTML/PDF file associations registered.")
			return
		case "--unregister-file-associations":
			_ = unregisterSparrowFileAssociations()
			fmt.Println("Sparrow file associations removed.")
			return
		}
	}
	// Initialize disk persistence engine
	InitPersistence()
	defer StopPersistence()
	defer ShutdownDownloads()

	var autofillEnabled = true
	var autofillEnabledMu sync.Mutex

	w := webview.New(true)
	defer w.Destroy()

	windowHandle := uintptr(w.Window())
	configureNativeWindow(windowHandle)
	applyBackdropFix(windowHandle)

	w.SetTitle("SPARROW")
	w.SetSize(1200, 800, 0)
	fitWindowToTaskbarArea(windowHandle, 1200, 800)
	startTaskbarGuard(windowHandle)

	// -------------------------------
	// GLOBAL SCRIPTS & INJECTIONS
	// -------------------------------

	w.Init(BuildFullscreenGuardJS())
	w.Init("window.__SHORTCUTS__ = " + shortcutsJSON() + ";")
	w.Init(errorPageHookJS)
	w.Init(chromeInjectionJS)
	w.Init(BuildFixedHeaderJS())
	w.Init(BuildAdblockJS())
	w.Init(BuildPrivacyJS())
	w.Init(BuildPrivacy2JS())
	w.Init(BuildPrivacy3JS())
	w.Init(cookieHookJS)
	w.Init(autocompleteHTML())

	w.Bind("reportBlocked", func(kind string, url string) {
		shield.ReportBlocked(kind, url)
	})

	w.Bind("reportBlockedBatch", func(entries []BlockedReport) {
		shield.ReportBlockedBatch(entries)
	})

	w.Bind("closeWindow", func() {
		closeNativeWindow(windowHandle)
	})

	w.Bind("minimizeWindow", func() {
		minimizeNativeWindow(windowHandle)
	})

	w.Bind("toggleMaximizeWindow", func() {
		toggleMaximizeNativeWindow(windowHandle)
	})

	w.Bind("toggleFullscreenWindow", func() {
		toggleWindowFullscreen(windowHandle)
	})

	w.Bind("setPageFullscreen", func(on bool) {
		setPageFullscreen(windowHandle, on)
	})

	w.Bind("beginWindowDrag", func(x int, y int) {
		beginNativeWindowDrag(windowHandle, x, y)
	})

	// -------------------------------
	// HISTORY CONTROLS
	// -------------------------------

	w.Bind("clearHistory", func(scope string) {
		switch scope {
		case "1hour":
			history.DeleteSince(time.Now().Add(-1 * time.Hour))
		case "today":
			now := time.Now()
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			history.DeleteSince(startOfDay)
		case "all":
			history.DeleteAll()
		}
		fmt.Println("History cleared:", scope)
	})

	w.Bind("getRecentHistory", func(limit int) []HistoryEntry {
		return history.GetRecent(limit)
	})

	// -------------------------------
	// REAL NAVIGATION WITH SECURITY SHIELD
	// -------------------------------

	appLockReturnTarget := ""
	navigateInput := func(input string, newTab bool, offline bool) NavResult {
		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			return NavResult{}
		}
		if appLock.IsLocked() && trimmed != "sparrow://lock" {
			return NavResult{Blocked: true, URL: "sparrow://lock"}
		}
		if trimmed == "sparrow://lock" && !appLock.IsLocked() {
			if tab := tabManager.GetActiveTab(); tab != nil && tab.URL != "" && tab.URL != "sparrow://lock" {
				appLockReturnTarget = tab.URL
			}
		}

		tabForNavigation := func(target string) *Tab {
			if newTab {
				tab := tabManager.NewTab(target)
				if incognito.IsActive() {
					incognito.MarkTab(tab.ID)
				}
				return tab
			}
			tab := tabManager.GetActiveTab()
			if tab == nil {
				tab = tabManager.NewTab(target)
			}
			return tab
		}

		if _, ok := internalPageHTML(trimmed); ok {
			tab := tabForNavigation(trimmed)
			tabManager.RecordNavigation(tab.ID, trimmed)
			navigateTo(w, trimmed)
			return NavResult{URL: trimmed}
		}

		if isPDFURL(trimmed) {
			target := trimmed
			tab := tabForNavigation(target)
			tabManager.RecordNavigation(tab.ID, target)
			if !incognito.IsActive() {
				history.Add(target, target)
			}
			navigateTo(w, "sparrow://pdf?url="+url.QueryEscape(target))
			return NavResult{URL: target}
		}

		target := resolveInput(input)
		target = shield.EnforceHTTPS(target)

		// Ads and trackers are blocked inside pages (see adblock.go);
		// websites themselves are never blocked from opening.

		tab := tabForNavigation(target)
		tabManager.RecordNavigation(tab.ID, target)

		if !incognito.IsActive() {
			history.Add(target, target)
		}

		if offline {
			w.Dispatch(func() { w.Navigate(dataURI(buildOfflinePageHTML(target))) })
		} else {
			navigateTo(w, target)
		}
		return NavResult{URL: target}
	}

	w.Bind("realNavigate", func(input string) NavResult {
		return navigateInput(input, false, false)
	})

	w.Bind("navigateOffline", func(input string, newTab bool) NavResult {
		return navigateInput(input, newTab, true)
	})

	w.Bind("showNetworkError", func(code string) {
		if appLock.IsLocked() {
			navigateTo(w, "sparrow://lock")
			return
		}
		target := ""
		if tab := tabManager.GetActiveTab(); tab != nil {
			target = tab.URL
		}
		if strings.HasPrefix(target, "sparrow:") || strings.HasPrefix(target, "data:") {
			target = ""
		}
		page := buildErrorPageForCode(target, extractErrCode(code))
		w.Dispatch(func() { w.Navigate(dataURI(page)) })
	})

	w.Bind("openLinkInNewTab", func(input string) NavResult {
		parsed, err := url.Parse(strings.TrimSpace(input))
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return NavResult{}
		}
		return navigateInput(parsed.String(), true, false)
	})

	resumeApp := func() {
		target := appLockReturnTarget
		appLockReturnTarget = ""
		if target == "" || target == "sparrow://lock" {
			target = "sparrow://home"
		}
		if tab := tabManager.GetActiveTab(); tab != nil {
			tabManager.RecordNavigation(tab.ID, target)
		}
		navigateTo(w, target)
	}
	w.Bind("getAppLockStatus", func() map[string]bool {
		return appLock.Status()
	})
	w.Bind("setAppLockPassword", func(password string) error {
		return appLock.Configure(password)
	})
	w.Bind("lockApp", func() bool {
		if tab := tabManager.GetActiveTab(); tab != nil && tab.URL != "" && tab.URL != "sparrow://lock" {
			appLockReturnTarget = tab.URL
		}
		if !appLock.Lock() {
			navigateTo(w, "sparrow://lock")
			return false
		}
		navigateTo(w, "sparrow://lock")
		return true
	})
	w.Bind("unlockApp", func(password string) bool {
		if !appLock.Unlock(password) {
			return false
		}
		resumeApp()
		return true
	})
	w.Bind("disableAppLock", func(password string) error {
		if err := appLock.Disable(password); err != nil {
			return err
		}
		resumeApp()
		return nil
	})

	// -------------------------------
	// TAB CONTROLS
	// -------------------------------

	w.Bind("openNewTab", func() int {
		tab := tabManager.NewTab("")
		tabManager.RecordNavigation(tab.ID, "sparrow://home")
		if incognito.IsActive() {
			incognito.MarkTab(tab.ID)
		}
		navigateTo(w, "sparrow://home")
		return tab.ID
	})

	w.Bind("closeTab", func(id int) {
		var closedTitle, closedURL string
		wasIncognito := incognito.IsTabIncognito(id)
		for _, t := range tabManager.GetAllTabs() {
			if t.ID == id {
				closedTitle = t.Title
				closedURL = t.URL
				break
			}
		}
		tabManager.CloseTab(id)
		if !wasIncognito {
			sessionManager.RecordClosed(closedTitle, closedURL)
		}
		incognito.UnmarkTab(id)
		pinManager.Clear(id)
		readerMode.Clear(id)

		active := tabManager.GetActiveTab()
		if active == nil {
			tab := tabManager.NewTab("")
			tabManager.RecordNavigation(tab.ID, "sparrow://home")
			navigateTo(w, "sparrow://home")
			return
		}
		target := active.URL
		if target == "" {
			target = "sparrow://home"
		}
		navigateTo(w, target)
	})

	// Ctrl+Shift+W: close every tab and start fresh on the Sparrow home page.
	// Nothing is added to "recently closed", so the old tabs cannot be reopened.
	w.Bind("closeAllTabs", func() {
		ids := []int{}
		for _, t := range tabManager.GetAllTabs() {
			ids = append(ids, t.ID)
		}
		tabManager.CloseAllTabs()
		for _, id := range ids {
			incognito.UnmarkTab(id)
			pinManager.Clear(id)
			readerMode.Clear(id)
		}

		tab := tabManager.NewTab("")
		tabManager.RecordNavigation(tab.ID, "sparrow://home")
		if incognito.IsActive() {
			incognito.MarkTab(tab.ID)
		}
		navigateTo(w, "sparrow://home")
	})

	w.Bind("switchTab", func(id int) {
		tab := tabManager.SwitchTab(id)
		if tab == nil {
			return
		}
		target := tab.URL
		if target == "" {
			target = "sparrow://home"
		}
		navigateTo(w, target)
		applyWebViewZoom(w, zoomManager.Get(tab.ID))
	})

	w.Bind("goBackNav", func() NavState {
		tab := tabManager.GetActiveTab()
		if tab == nil {
			return NavState{}
		}
		from := tab.URL
		if url, ok := tabManager.GoBack(tab.ID); ok {
			stepHistory(w, -1, from, url)
		}
		return NavState{
			URL:        tab.URL,
			CanBack:    tabManager.CanGoBack(tab.ID),
			CanForward: tabManager.CanGoForward(tab.ID),
			TabID:      tab.ID,
		}
	})

	w.Bind("goForwardNav", func() NavState {
		tab := tabManager.GetActiveTab()
		if tab == nil {
			return NavState{}
		}
		from := tab.URL
		if url, ok := tabManager.GoForward(tab.ID); ok {
			stepHistory(w, 1, from, url)
		}
		return NavState{
			URL:        tab.URL,
			CanBack:    tabManager.CanGoBack(tab.ID),
			CanForward: tabManager.CanGoForward(tab.ID),
			TabID:      tab.ID,
		}
	})

	w.Bind("getNavState", func() NavState {
		tab := tabManager.GetActiveTab()
		if tab == nil {
			return NavState{}
		}
		return NavState{
			URL:        tab.URL,
			CanBack:    tabManager.CanGoBack(tab.ID),
			CanForward: tabManager.CanGoForward(tab.ID),
			TabID:      tab.ID,
		}
	})

	w.Bind("getAllTabsUI", func() []*Tab {
		return tabManager.GetAllTabs()
	})

	w.Bind("updateTabTitle", func(id int, title string, url string, favicon string) {
		if strings.TrimSpace(title) != "" {
			tabManager.SetTitle(id, title)
		}
		tabManager.SetFavicon(id, favicon)
	})

	w.Bind("syncTabLocation", func(id int, location string, sameDocument bool) {
		trimmed := strings.TrimSpace(location)
		if trimmed == "" || (!strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://")) {
			return
		}
		tabManager.SyncLocation(id, sparrowPDF.OriginalFor(trimmed), sameDocument)
	})

	// -------------------------------
	// SECURITY SHIELD CONTROLS
	// -------------------------------

	w.Bind("toggleAdBlock", func() bool {
		return shield.ToggleAdBlock()
	})

	w.Bind("toggleTrackerBlock", func() bool {
		return shield.ToggleTrackerBlock()
	})

	w.Bind("toggleHTTPSOnly", func() bool {
		return shield.ToggleHTTPSOnly()
	})

	w.Bind("getShieldStats", func() map[string]any {
		return shield.GetStats()
	})

	// -------------------------------
	// SETTINGS CONTROLS
	// -------------------------------

	w.Bind("getSettings", func() map[string]any {
		return settings.GetAll()
	})

	w.Bind("setTheme", func(theme string) string {
		return settings.SetTheme(theme)
	})

	w.Bind("setSearchEngine", func(engine string) string {
		return settings.SetSearchEngine(engine)
	})

	w.Bind("setHomepage", func(url string) string {
		return settings.SetHomepage(url)
	})

	w.Bind("toggleJS", func() bool {
		return settings.ToggleJS()
	})

	w.Bind("toggleImages", func() bool {
		return settings.ToggleImages()
	})

	w.Bind("resetSettings", func() map[string]any {
		return settings.ResetToDefault()
	})

	// -------------------------------
	// BOOKMARK CONTROLS
	// -------------------------------

	w.Bind("toggleBookmark", func(title string, url string) bool {
		return bookmarkManager.Toggle(title, url)
	})

	w.Bind("toggleFavorite", func(title string, url string) bool {
		state := favoriteManager.Toggle(title, url)
		if err := savePersistedState(); err != nil {
			log.Println("[favorites] save failed:", err)
		}
		return state
	})

	w.Bind("addFavorite", func(title string, url string) bool {
		if favoriteManager.Add(title, url) == nil {
			// already saved (or empty URL): treat saved-state as success
			return favoriteManager.IsFavorite(url)
		}
		if err := savePersistedState(); err != nil {
			log.Println("[favorites] save failed:", err)
		}
		return true
	})

	w.Bind("removeFavoriteByURL", func(url string) bool {
		removed := favoriteManager.RemoveByURL(url)
		if removed {
			if err := savePersistedState(); err != nil {
				log.Println("[favorites] save failed:", err)
			}
		}
		return removed
	})

	w.Bind("generateQRCode", generateQRCodeDataURL)

	w.Bind("isFavorite", func(url string) bool {
		return favoriteManager.IsFavorite(url)
	})

	w.Bind("getFavorites", func() []FavoriteEntry {
		return favoriteManager.GetAll()
	})

	w.Bind("removeFavorite", func(id int) bool {
		removed := favoriteManager.RemoveByID(id)
		if removed {
			_ = savePersistedState()
		}
		return removed
	})

	w.Bind("isBookmarked", func(url string) bool {
		return bookmarkManager.IsBookmarked(url)
	})

	w.Bind("getBookmarks", func() []*Bookmark {
		return bookmarkManager.GetAll()
	})

	w.Bind("removeBookmark", func(id int) bool {
		return bookmarkManager.Remove(id)
	})

	w.Bind("searchBookmarks", func(keyword string) []*Bookmark {
		return bookmarkManager.Search(keyword)
	})

	// -------------------------------
	// DOWNLOAD CONTROLS
	// -------------------------------

	w.Bind("getDownloads", func() []*DownloadItem {
		return downloadManager.GetAll()
	})

	w.Bind("clearFinishedDownloads", func() {
		downloadManager.ClearFinished()
	})

	w.Bind("cancelDownload", func(id int) {
		CancelActiveDownload(id)
	})

	// -------------------------------
	// INCOGNITO MODE
	// -------------------------------

	w.Bind("enableIncognito", func() {
		incognito.Enable()
	})

	w.Bind("disableIncognito", func() {
		for _, t := range tabManager.GetAllTabs() {
			if incognito.IsTabIncognito(t.ID) {
				tabManager.CloseTab(t.ID)
			}
		}
		incognito.Disable()

		active := tabManager.GetActiveTab()
		if active == nil {
			active = tabManager.NewTab("sparrow://home")
			tabManager.RecordNavigation(active.ID, "sparrow://home")
		}
		navigateTo(w, active.URL)
	})

	w.Bind("isIncognitoActive", func() bool {
		return incognito.IsActive()
	})

	w.Bind("getIncognitoDuration", func() string {
		return incognito.Duration()
	})

	// -------------------------------
	// PERMISSIONS
	// -------------------------------

	w.Bind("getSitePermissions", func(domain string) map[string]string {
		perms := permissionManager.GetFor(domain)
		return map[string]string{
			"camera":        string(perms.Camera),
			"microphone":    string(perms.Microphone),
			"location":      string(perms.Location),
			"notifications": string(perms.Notifications),
		}
	})

	// read-only lookup used by the in-page permission guard (never creates entries)
	w.Bind("getPagePermissions", func(host string) map[string]string {
		return permissionManager.Lookup(host)
	})

	w.Bind("setSitePermission", func(domain string, permType string, state string) map[string]string {
		return permissionManager.Set(domain, permType, state)
	})

	w.Bind("resetSitePermissions", func(domain string) {
		permissionManager.ResetDomain(domain)
	})

	w.Bind("resetAllPermissions", func() {
		permissionManager.ResetAll()
	})

	// -------------------------------
	// READER MODE
	// -------------------------------

	w.Bind("toggleReaderMode", func(tabID int) bool {
		isOn := readerMode.Toggle(tabID)
		w.Eval(BuildReaderModeJS(isOn))
		return isOn
	})
	w.Bind("isReaderModeActive", func(tabID int) bool {
		return readerMode.IsActive(tabID)
	})

	// -------------------------------
	// SHORTCUTS
	// -------------------------------

	w.Bind("getShortcuts", func() []Shortcut {
		return getAllShortcuts()
	})

	// -------------------------------
	// SESSION (RECENTLY CLOSED TABS)
	// -------------------------------

	w.Bind("getRecentlyClosed", func() []ClosedTab {
		return sessionManager.GetAll()
	})

	w.Bind("reopenLastClosed", func() *ClosedTab {
		return sessionManager.PopLastClosed()
	})

	w.Bind("clearRecentlyClosed", func() {
		sessionManager.Clear()
	})

	// -------------------------------
	// ZOOM CONTROLS
	// -------------------------------

	w.Bind("zoomIn", func(tabID int) float64 {
		level := zoomManager.ZoomIn(tabID)
		applyWebViewZoom(w, level)
		return level
	})

	w.Bind("zoomOut", func(tabID int) float64 {
		level := zoomManager.ZoomOut(tabID)
		applyWebViewZoom(w, level)
		return level
	})

	w.Bind("zoomReset", func(tabID int) float64 {
		level := zoomManager.Reset(tabID)
		applyWebViewZoom(w, level)
		return level
	})

	w.Bind("getZoom", func(tabID int) float64 {
		return zoomManager.Get(tabID)
	})

	// -------------------------------
	// REAL DOWNLOADS
	// -------------------------------

	w.Bind("downloadFile", func(url string) *DownloadItem {
		return StartDownload(url, settings.DownloadPath)
	})

	w.Bind("saveDownloadData", func(name string, b64 string) *DownloadItem {
		return SaveDownloadData(name, b64, settings.DownloadPath)
	})

	w.Bind("getDownloadStats", func() []DownloadStat {
		return GetDownloadStats()
	})

	w.Bind("toggleDownloadPause", func(id int) bool {
		return ToggleDownloadPause(id)
	})

	w.Bind("openDownloadedFile", func(id int) {
		OpenDownloadedFile(id)
	})

	w.Bind("showDownloadInFolder", func(id int) {
		RevealDownloadedFile(id)
	})

	w.Bind("openDownloadsFolder", func() {
		OpenFolder(resolveSaveDir(settings.DownloadPath))
	})

	w.Bind("cancelActiveDownload", func(id int) {
		CancelActiveDownload(id)
	})

	w.Bind("isDownloadableLink", func(url string) bool {
		return IsDownloadableURL(url)
	})

	// -------------------------------
	// CONTEXT MENU HELPERS
	// -------------------------------

	w.Bind("openDevTools", func() {
		fmt.Println("DevTools requested")
	})

	// -------------------------------
	// TAB PINNING
	// -------------------------------

	w.Bind("togglePinTab", func(tabID int) bool {
		return pinManager.Toggle(tabID)
	})

	w.Bind("isPinnedTab", func(tabID int) bool {
		return pinManager.IsPinned(tabID)
	})

	w.Bind("getOrderedTabs", func() []*Tab {
		return GetOrderedTabs()
	})

	// -------------------------------
	// NOTIFICATION CENTER
	// -------------------------------

	w.Bind("pushNotification", func(title string, message string, kind string) *NotificationItem {
		return notifCenter.Push(title, message, kind)
	})

	w.Bind("getNotifications", func() []*NotificationItem {
		return notifCenter.GetAll()
	})

	w.Bind("markNotifRead", func(id int) {
		notifCenter.MarkRead(id)
	})

	w.Bind("markAllNotifsRead", func() {
		notifCenter.MarkAllRead()
	})

	w.Bind("getUnreadNotifCount", func() int {
		return notifCenter.UnreadCount()
	})

	w.Bind("clearAllNotifs", func() {
		notifCenter.ClearAll()
	})

	// -------------------------------
	// AUTOCOMPLETE / SUGGESTIONS
	// -------------------------------

	w.Bind("getAutocomplete", func(query string) []AutocompleteResult {
		return getAutocompleteResults(query)
	})

	w.Bind("getSuggestions", func(partial string) []AutocompleteResult {
		results := getAutocompleteResults(partial)
		if len(results) > 6 {
			return results[:6]
		}
		return results
	})

	// -------------------------------
	// SITE INFO POPUP
	// -------------------------------

	w.Bind("getSiteInfo", func(url string) SiteInfo {
		return BuildSiteInfo(url)
	})

	// -------------------------------
	// FIND IN PAGE & CONTEXT MENU
	// -------------------------------

	w.Bind("injectFindInPage", func() {
		w.Eval(findInPageJS)
	})

	w.Bind("injectContextMenu", func() {
		w.Eval(contextMenuJS)
	})

	// -------------------------------
	// BOOKMARK EXPORT / IMPORT
	// -------------------------------

	w.Bind("exportBookmarks", func() string {
		return ExportBookmarksJSON()
	})

	w.Bind("importBookmarks", func(jsonStr string) int {
		return ImportBookmarksJSON(jsonStr)
	})

	// -------------------------------
	// LOCAL HTML SERVER (localhost-only)
	// -------------------------------
	w.Bind("startLocalHTMLServer", func(filePath string) string {
		url, err := localHTMLServer.Start(filePath)
		if err != nil {
			return ""
		}
		return url
	})

	w.Bind("startLocalHTMLServerWithContent", func(fileName string, content string) string {
		url, err := localHTMLServer.StartFromContents(fileName, content)
		if err != nil {
			return ""
		}
		return url
	})

	w.Bind("stopLocalHTMLServer", func() {
		localHTMLServer.Stop()
	})

	w.Bind("getLocalHTMLServerURL", func() string {
		return localHTMLServer.URL()
	})

	// -------------------------------
	// SITE SETTINGS PAGE
	// -------------------------------

	w.Bind("getAllSitePermissions", func() map[string]map[string]string {
		return permissionManager.GetAllAsMap()
	})

	// -------------------------------
	// SMART TOP SITES & SHORTCUTS
	// -------------------------------

	w.Bind("getTopSites", func(limit int) []TopSite {
		return topSitesEngine.GetTopSites(limit)
	})

	w.Bind("addShortcut", func(title string, url string) []PinnedShortcut {
		shortcuts := shortcutManagerNewTab.AddShortcut(title, url)
		if err := savePersistedState(); err != nil {
			fmt.Println("[links] failed to save added shortcut:", err)
		}
		return shortcuts
	})

	w.Bind("removeShortcut", func(index int) []PinnedShortcut {
		shortcuts := shortcutManagerNewTab.RemoveShortcut(index)
		if err := savePersistedState(); err != nil {
			fmt.Println("[links] failed to save removed shortcut:", err)
		}
		return shortcuts
	})

	w.Bind("getShortcutsNewTab", func() []PinnedShortcut {
		return shortcutManagerNewTab.GetShortcuts()
	})

	w.Bind("saveLink", func(title string, url string) (*PinnedShortcut, error) {
		link, err := shortcutManagerNewTab.AddLink(title, url)
		if err != nil {
			return nil, err
		}
		if err := savePersistedState(); err != nil {
			shortcutManagerNewTab.RemoveLink(link.ID)
			return nil, fmt.Errorf("could not save link: %w", err)
		}
		return link, nil
	})

	w.Bind("updateLink", func(id int, title string, url string) error {
		previous := shortcutManagerNewTab.GetShortcuts()
		if err := shortcutManagerNewTab.UpdateLink(id, title, url); err != nil {
			return err
		}
		if err := savePersistedState(); err != nil {
			shortcutManagerNewTab.ReplaceAll(previous)
			return fmt.Errorf("could not save link changes: %w", err)
		}
		return nil
	})

	w.Bind("deleteLink", func(id int) error {
		previous := shortcutManagerNewTab.GetShortcuts()
		if !shortcutManagerNewTab.RemoveLink(id) {
			return fmt.Errorf("saved link not found")
		}
		if err := savePersistedState(); err != nil {
			shortcutManagerNewTab.ReplaceAll(previous)
			return fmt.Errorf("could not delete saved link: %w", err)
		}
		return nil
	})

	// -------------------------------
	// COOKIE MANAGER
	// -------------------------------

	w.Bind("reportCookie", func(domain string, name string) bool {
		return cookieManager.Record(domain, name)
	})

	w.Bind("reportCookies", func(entries []CookieReport) {
		for _, entry := range entries {
			cookieManager.Record(entry.Domain, entry.Name)
		}
	})

	w.Bind("getCookieSummary", func() map[string]int {
		return cookieManager.GetSummary()
	})

	w.Bind("clearDomainCookies", func(domain string) {
		cookieManager.ClearDomain(domain)
	})

	w.Bind("clearAllCookies", func() {
		cookieManager.ClearAll()
	})

	w.Bind("setCookieBlocked", func(domain string, blocked bool) {
		cookieManager.SetBlocked(domain, blocked)
	})

	// -------------------------------
	// FORCE DARK MODE
	// -------------------------------

	w.Bind("toggleForceDarkGlobal", func() bool {
		return forceDark.ToggleGlobal()
	})

	w.Bind("toggleForceDarkTab", func(tabID int) bool {
		on := forceDark.ToggleForTab(tabID)
		w.Eval(forceDarkJS)
		return on
	})

	// -------------------------------
	// PERFORMANCE / BATTERY SAVER
	// -------------------------------

	w.Bind("toggleSaverMode", func() bool {
		return perfManager.ToggleSaver()
	})

	w.Bind("isSaverModeOn", func() bool {
		return perfManager.IsSaverOn()
	})

	w.Bind("suspendTab", func(tabID int) {
		perfManager.SuspendTab(tabID)
	})

	w.Bind("wakeTab", func(tabID int) {
		perfManager.WakeTab(tabID)
	})

	w.Bind("getSuspendedCount", func() int {
		return perfManager.SuspendedCount()
	})

	// -------------------------------
	// PRINT
	// -------------------------------

	w.Bind("printPage", func() {
		w.Eval(printPageJS)
	})

	// -------------------------------
	// AUTOFILL
	// -------------------------------

	w.Bind("saveAutofillProfile", func(data AutofillData) AutofillData {
		data.Name = strings.TrimSpace(data.Name)
		data.Email = strings.TrimSpace(data.Email)
		data.Phone = strings.TrimSpace(data.Phone)
		data.Address = strings.TrimSpace(data.Address)
		saveAutofillProfile(data)
		return data
	})

	w.Bind("getAutofillProfile", func() AutofillData {
		return getAutofillProfile()
	})

	w.Bind("clearAutofillProfile", func() {
		clearAutofillProfile()
	})

	w.Bind("toggleAutofillEnabled", func() bool {
		autofillEnabledMu.Lock()
		defer autofillEnabledMu.Unlock()
		autofillEnabled = !autofillEnabled
		return autofillEnabled
	})

	// -------------------------------
	// WALLPAPER
	// -------------------------------

	w.Bind("setWallpaper", func(id string) string {
		previous := wallpaperManager.GetSelectedID()
		selected, ok := wallpaperManager.SetWallpaper(id)
		if !ok {
			return previous
		}
		if err := savePersistedState(); err != nil {
			wallpaperManager.Restore(previous)
			fmt.Println("[wallpaper] failed to save selection:", err)
			return previous
		}
		return selected
	})

	w.Bind("getWallpaperCSS", func() string {
		return wallpaperManager.GetSelectedCSS()
	})

	w.Bind("getSelectedWallpaper", func() string {
		return wallpaperManager.GetSelectedID()
	})

	w.Bind("getWallpaperOptions", func() []WallpaperOption {
		return wallpaperManager.GetAllOptions()
	})

	w.Bind("getHomeAnimation", func() string {
		return homeAnimationManager.Get()
	})

	w.Bind("getHomeAnimationOptions", func() []HomeAnimationOption {
		return homeAnimationManager.Options()
	})

	w.Bind("setHomeAnimation", func(id string) (string, error) {
		previous := homeAnimationManager.Get()
		selected, err := homeAnimationManager.Set(id)
		if err != nil {
			return "", err
		}
		if err := savePersistedState(); err != nil {
			homeAnimationManager.Restore(previous)
			return "", fmt.Errorf("could not save home animation: %w", err)
		}
		return selected, nil
	})

	// -------------------------------
	// ACCESSIBILITY
	// -------------------------------

	w.Bind("setFontScale", func(scale float64) float64 {
		return accessibility.SetFontScale(scale)
	})

	w.Bind("toggleHighContrast", func() bool {
		return accessibility.ToggleHighContrast()
	})

	w.Bind("toggleReduceMotion", func() bool {
		return accessibility.ToggleReduceMotion()
	})

	w.Bind("toggleUnderlineLinks", func() bool {
		return accessibility.ToggleUnderlineLinks()
	})

	w.Bind("getAccessibilitySettings", func() map[string]any {
		return accessibility.GetAll()
	})

	w.Bind("applyAccessibility", func() {
		s := accessibility
		w.Eval(buildAccessibilityJS(s.FontScale, s.HighContrast, s.ReduceMotion, s.UnderlineLinks))
	})

	// -------------------------------
	// FEEDBACK
	// -------------------------------

	w.Bind("submitFeedback", func(message string, category string) *FeedbackEntry {
		return feedbackManager.Submit(message, category)
	})

	w.Bind("getAllFeedback", func() []*FeedbackEntry {
		return feedbackManager.GetAll()
	})

	// -------------------------------
	// PASSWORD MANAGER
	// -------------------------------

	w.Bind("unlockVault", func(masterPassword string) bool {
		return passwordVault.Unlock(masterPassword)
	})

	w.Bind("lockVault", func() {
		passwordVault.Lock()
	})

	w.Bind("isVaultUnlocked", func() bool {
		return passwordVault.IsUnlocked()
	})

	w.Bind("savePasswordEntry", func(domain string, username string, password string) (*PasswordEntry, error) {
		return passwordVault.Save(domain, username, password)
	})

	w.Bind("revealPasswordEntry", func(id int) (string, error) {
		return passwordVault.Reveal(id)
	})

	w.Bind("deletePasswordEntry", func(id int) bool {
		return passwordVault.Remove(id)
	})

	w.Bind("getPasswordList", func() []PasswordMetadata {
		return passwordVault.GetMetadataOnly()
	})

	w.Bind("findPasswordsForDomain", func(domain string) []PasswordMetadata {
		return passwordVault.FindForDomain(domain)
	})

	w.Bind("wipeAllPasswords", func() {
		passwordVault.WipeAll()
	})

	w.Bind("generatePassword", func(length int) string {
		return GenerateStrongPassword(length)
	})

	// -------------------------------
	// UPDATE CHECKER
	// -------------------------------

	w.Bind("getChangelog", func() []ChangelogEntry {
		return sparrowChangelog
	})

	// -------------------------------
	// PAGE EXPORT
	// -------------------------------

	w.Bind("savePageText", func(title string, text string) (string, error) {
		return SavePageText(title, text)
	})

	w.Bind("extractPageText", func() {
		w.Eval(extractPageTextJS)
	})

	// -------------------------------
	// BACKUP / RESTORE
	// -------------------------------

	w.Bind("exportBackup", func() string {
		return ExportBackupJSON()
	})

	w.Bind("saveBackupToFile", func() (string, error) {
		return SaveBackupToFile()
	})

	w.Bind("restoreBackup", func(jsonStr string) (int, error) {
		return RestoreBackupJSON(jsonStr)
	})

	// -------------------------------
	// PICTURE-IN-PICTURE
	// -------------------------------

	w.Bind("requestPiP", func(tabID int) {
		pipManager.SetActive(tabID, true)
		w.Eval(pipRequestJS)
	})

	w.Bind("isPiPActive", func(tabID int) bool {
		return pipManager.IsActive(tabID)
	})

	// -------------------------------
	// CUSTOM THEME
	// -------------------------------

	w.Bind("setAccentColor", func(hex string) string {
		return customTheme.SetAccentColor(hex)
	})

	w.Bind("setTabRadius", func(radius int) int {
		return customTheme.SetTabRadius(radius)
	})

	w.Bind("toggleCompactToolbar", func() bool {
		return customTheme.ToggleCompact()
	})

	w.Bind("getCustomTheme", func() map[string]any {
		return customTheme.GetAll()
	})

	w.Bind("getThemeCSS", func() string {
		return customTheme.BuildThemeCSS()
	})

	w.Bind("saveCustomTheme", func() bool {
		return savePersistedState() == nil
	})

	// -------------------------------
	// CRASH RECOVERY
	// -------------------------------

	w.Bind("updateRecoverySnapshot", func(urls []string) {
		recoverySnapshot.UpdateSnapshot(urls)
	})

	w.Bind("getRecoverySnapshot", func() []string {
		return recoverySnapshot.GetSnapshot()
	})

	w.Bind("hasRecoverableSession", func() bool {
		return recoverySnapshot.HasRecoverableSession()
	})

	w.Bind("clearRecoverySnapshot", func() {
		recoverySnapshot.ClearSnapshot()
	})

	w.Bind("toggleCrashRecovery", func() bool {
		return recoverySnapshot.ToggleEnabled()
	})

	// -------------------------------
	// SCREENSHOT
	// -------------------------------

	w.Bind("captureScreenshot", func() {
		w.Eval(screenshotCaptureJS)
	})

	w.Bind("reportScreenshot", func(base64Data string, title string) (string, error) {
		path, err := SaveScreenshot(base64Data, title)
		if err == nil {
			notifCenter.Push("Screenshot saved", path, "success")
		}
		return path, err
	})

	// -------------------------------
	// TAB SEARCH
	// -------------------------------

	w.Bind("searchTabs", func(query string) []TabSearchResult {
		return tabSearchEngine.Search(query)
	})

	// -------------------------------
	// CUSTOM SEARCH ENGINES
	// -------------------------------

	w.Bind("addCustomSearchEngine", func(name string, template string) (*CustomSearchEngine, bool) {
		return customSearchManager.AddEngine(name, template)
	})

	w.Bind("removeCustomSearchEngine", func(id int) bool {
		return customSearchManager.RemoveEngine(id)
	})

	w.Bind("setActiveCustomSearchEngine", func(id int) bool {
		return customSearchManager.SetActive(id)
	})

	w.Bind("getCustomSearchEngines", func() []*CustomSearchEngine {
		return customSearchManager.GetAll()
	})

	// -------------------------------
	// SPEED DIAL
	// -------------------------------

	w.Bind("addSpeedDialTile", func(title string, url string) *SpeedDialTile {
		return speedDialManager.AddTile(title, url)
	})

	w.Bind("removeSpeedDialTile", func(id int) bool {
		return speedDialManager.RemoveTile(id)
	})

	w.Bind("reorderSpeedDial", func(fromIndex int, toIndex int) bool {
		return speedDialManager.Reorder(fromIndex, toIndex)
	})

	w.Bind("getSpeedDialTiles", func() []*SpeedDialTile {
		return speedDialManager.GetAll()
	})

	// -------------------------------
	// OPEN INITIAL TAB / FILE ASSOCIATION
	// -------------------------------

	launchTarget := launchTargetFromArgs()
	if launchTarget == "" {
		firstRun, err := isFirstRun()
		if err != nil {
			log.Println("[startup] could not check first-run welcome state:", err)
		} else if firstRun {
			settings.SetSearchEngine("google")
			launchTarget = "sparrow://welcome"
		}
		if launchTarget == "" {
			launchTarget = "sparrow://home"
		}
	}

	firstTab := tabManager.NewTab(launchTarget)
	historyTarget := launchTarget

	if ext := strings.ToLower(filepath.Ext(launchTarget)); ext == ".html" || ext == ".htm" {
		// HTML files opened from Explorer are served from a localhost origin.
		if localURL, err := localHTMLServer.Start(launchTarget); err == nil {
			launchTarget = localURL
			firstTab.URL = launchTarget
			tabManager.RecordNavigation(firstTab.ID, launchTarget)
		}
	}

	if appLock.IsEnabled() {
		appLockReturnTarget = launchTarget
		launchTarget = "sparrow://lock"
		firstTab.URL = launchTarget
		tabManager.RecordNavigation(firstTab.ID, launchTarget)
	} else if !incognito.IsActive() && !strings.HasPrefix(historyTarget, "sparrow://") {
		history.Add(historyTarget, historyTarget)
	}

	if isPDFURL(launchTarget) || strings.EqualFold(filepath.Ext(launchTarget), ".pdf") {
		navigateTo(w, buildPDFViewerURL(launchTarget))
	} else if strings.HasPrefix(launchTarget, "sparrow://") {
		navigateTo(w, launchTarget)
	} else {
		navigateTo(w, launchTarget)
	}

	w.Run()
}
