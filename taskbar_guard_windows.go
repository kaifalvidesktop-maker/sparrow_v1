//go:build windows

package main

import (
	"time"
	"unsafe"
)

// startTaskbarGuard keeps the window from covering the Windows taskbar when
// it is NOT in fullscreen mode (frameless windows can end up covering the
// whole monitor after snapping / double-click maximize / a stuck fullscreen).
func startTaskbarGuard(hwnd uintptr) {
	go func() {
		for {
			time.Sleep(400 * time.Millisecond)

			fullscreenMu.Lock()
			active := fullscreenWindowState.active
			fullscreenMu.Unlock()
			if active {
				continue
			}

			var cur rect
			ok, _, _ := getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&cur)))
			if ok == 0 {
				continue
			}
			mon, okMon := monitorBounds(hwnd)
			work, okWork := monitorWorkArea(hwnd)
			if !okMon || !okWork {
				continue
			}

			covers := cur.Left <= mon.Left && cur.Top <= mon.Top && cur.Right >= mon.Right && cur.Bottom >= mon.Bottom
			taskbarVisible := work.Left > mon.Left || work.Top > mon.Top || work.Right < mon.Right || work.Bottom < mon.Bottom
			if !covers || !taskbarVisible {
				continue
			}

			if zoomed, _, _ := isZoomed.Call(hwnd); zoomed != 0 {
				showWindow.Call(hwnd, swRestore)
			}
			width := int(work.Right - work.Left)
			height := int(work.Bottom - work.Top)
			setWindowPos.Call(hwnd, 0, uintptr(work.Left), uintptr(work.Top), uintptr(width), uintptr(height), swpNoZOrder|swpFrameChanged)
		}
	}()
}
