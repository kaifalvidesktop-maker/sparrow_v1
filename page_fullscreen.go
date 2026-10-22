package main

import "sync"

// Native window fullscreen state shared by F11 and page (video) fullscreen.
var fullscreenMu sync.Mutex
var windowIsFullscreen bool
var pageOwnsFullscreen bool

// toggleWindowFullscreen is the F11 / titlebar button path.
func toggleWindowFullscreen(hwnd uintptr) {
	fullscreenMu.Lock()
	defer fullscreenMu.Unlock()
	toggleFullscreenNativeWindow(hwnd)
	windowIsFullscreen = !windowIsFullscreen
	pageOwnsFullscreen = false
}

// setPageFullscreen is called by the page guard when a video/element enters
// or leaves fullscreen. It only undoes what it did itself, so a window the
// user already put in F11 mode stays that way.
func setPageFullscreen(hwnd uintptr, on bool) {
	fullscreenMu.Lock()
	defer fullscreenMu.Unlock()
	if on {
		if windowIsFullscreen {
			return
		}
		toggleFullscreenNativeWindow(hwnd)
		windowIsFullscreen = true
		pageOwnsFullscreen = true
		return
	}
	if pageOwnsFullscreen {
		toggleFullscreenNativeWindow(hwnd)
		windowIsFullscreen = false
		pageOwnsFullscreen = false
	}
}
