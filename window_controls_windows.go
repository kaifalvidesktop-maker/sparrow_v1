//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	gwlStyle                = 0xfffffff0
	wsCaption               = 0x00C00000
	wsThickFrame            = 0x00040000
	wsMinimizeBox           = 0x00020000
	wsMaximizeBox           = 0x00010000
	wsSysMenu               = 0x00080000
	wmClose                 = 0x0010
	wmNCLButtonDown         = 0x00A1
	htCaption               = 0x0002
	swMinimize              = 6
	swMaximize              = 3
	swRestore               = 9
	swpNoMove               = 0x0002
	swpNoSize               = 0x0001
	swpNoZOrder             = 0x0004
	swpFrameChanged         = 0x0020
	monitorDefaultToNearest = 0x00000002
)

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type monitorInfo struct {
	CbSize    int32
	RcMonitor rect
	RcWork    rect
	DwFlags   int32
}

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	getWindowLongPtr  = user32.NewProc("GetWindowLongPtrW")
	setWindowLongPtr  = user32.NewProc("SetWindowLongPtrW")
	setWindowPos      = user32.NewProc("SetWindowPos")
	showWindow        = user32.NewProc("ShowWindow")
	isZoomed          = user32.NewProc("IsZoomed")
	getWindowRect     = user32.NewProc("GetWindowRect")
	postMessage       = user32.NewProc("PostMessageW")
	releaseCapture    = user32.NewProc("ReleaseCapture")
	sendMessage       = user32.NewProc("SendMessageW")
	monitorFromWindow = user32.NewProc("MonitorFromWindow")
	getMonitorInfoW   = user32.NewProc("GetMonitorInfoW")
)

// F11 / "□" button: fills the work area, taskbar stays visible.
var fullscreenWindowState = struct {
	active bool
	rect   rect
}{
	active: false,
}

// Page (video) fullscreen: covers the whole monitor, no frame.
var pageFullscreenWindowState = struct {
	active bool
	rect   rect
	style  uintptr
}{
	active: false,
}

const (
	gwlpWndProc     = 0xfffffffc
	wmGetMinMaxInfo = 0x0024
)

type point struct {
	X int32
	Y int32
}

type minMaxInfo struct {
	PtReserved     point
	PtMaxSize      point
	PtMaxPosition  point
	PtMinTrackSize point
	PtMaxTrackSize point
}

var (
	callWindowProc  = user32.NewProc("CallWindowProcW")
	originalWndProc uintptr
	minMaxHooked    bool
)

// Frameless windows (no WS_CAPTION) cover the whole monitor, including the
// taskbar, when maximized/snapped. Answer WM_GETMINMAXINFO with the work
// area so Windows keeps the taskbar visible.
var minMaxWndProc = syscall.NewCallback(func(hwnd, msg, wparam, lparam uintptr) uintptr {
	ret, _, _ := callWindowProc.Call(originalWndProc, hwnd, msg, wparam, lparam)
	if msg == wmGetMinMaxInfo && lparam != 0 && !pageFullscreenWindowState.active {
		mon, okMon := monitorBounds(hwnd)
		work, okWork := monitorWorkArea(hwnd)
		if okMon && okWork {
			info := (*minMaxInfo)(*(*unsafe.Pointer)(unsafe.Pointer(&lparam)))
			info.PtMaxPosition.X = work.Left - mon.Left
			info.PtMaxPosition.Y = work.Top - mon.Top
			info.PtMaxSize.X = work.Right - work.Left
			info.PtMaxSize.Y = work.Bottom - work.Top
		}
	}
	return ret
})

func installMinMaxHook(hwnd uintptr) {
	if minMaxHooked {
		return
	}
	previous, _, _ := setWindowLongPtr.Call(hwnd, uintptr(gwlpWndProc), minMaxWndProc)
	if previous == 0 {
		return
	}
	originalWndProc = previous
	minMaxHooked = true
}

func configureNativeWindow(hwnd uintptr) {
	style, _, _ := getWindowLongPtr.Call(hwnd, uintptr(gwlStyle))
	style &^= wsCaption
	style |= wsThickFrame | wsMinimizeBox | wsMaximizeBox | wsSysMenu
	setWindowLongPtr.Call(hwnd, uintptr(gwlStyle), style)
	setWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged)
	installMinMaxHook(hwnd)
}

func monitorBounds(hwnd uintptr) (rect, bool) {
	monitorHandle, _, _ := monitorFromWindow.Call(hwnd, monitorDefaultToNearest)
	if monitorHandle == 0 {
		return rect{}, false
	}

	mi := monitorInfo{CbSize: int32(unsafe.Sizeof(monitorInfo{}))}
	ok, _, _ := getMonitorInfoW.Call(monitorHandle, uintptr(unsafe.Pointer(&mi)))
	if ok == 0 {
		return rect{}, false
	}
	return mi.RcMonitor, true
}

func monitorWorkArea(hwnd uintptr) (rect, bool) {
	monitorHandle, _, _ := monitorFromWindow.Call(hwnd, monitorDefaultToNearest)
	if monitorHandle == 0 {
		return rect{}, false
	}

	mi := monitorInfo{CbSize: int32(unsafe.Sizeof(monitorInfo{}))}
	ok, _, _ := getMonitorInfoW.Call(monitorHandle, uintptr(unsafe.Pointer(&mi)))
	if ok == 0 {
		return rect{}, false
	}
	return mi.RcWork, true
}

func fitWindowToTaskbarArea(hwnd uintptr, width, height int) {
	workArea, ok := monitorWorkArea(hwnd)
	if !ok {
		return
	}

	availableW := int(workArea.Right - workArea.Left)
	availableH := int(workArea.Bottom - workArea.Top)
	if width > availableW {
		width = availableW
	}
	if height > availableH {
		height = availableH
	}
	if width <= 0 || height <= 0 {
		return
	}

	left := int(workArea.Left) + (availableW-width)/2
	top := int(workArea.Top) + (availableH-height)/2
	setWindowPos.Call(hwnd, 0, uintptr(left), uintptr(top), uintptr(width), uintptr(height), swpNoZOrder|swpFrameChanged)
}

func closeNativeWindow(hwnd uintptr) {
	postMessage.Call(hwnd, wmClose, 0, 0)
}

func minimizeNativeWindow(hwnd uintptr) {
	showWindow.Call(hwnd, swMinimize)
}

func toggleMaximizeNativeWindow(hwnd uintptr) {
	zoomed, _, _ := isZoomed.Call(hwnd)
	if zoomed != 0 {
		showWindow.Call(hwnd, swRestore)
		return
	}

	workArea, ok := monitorWorkArea(hwnd)
	if !ok {
		showWindow.Call(hwnd, swMaximize)
		return
	}

	width := int(workArea.Right - workArea.Left)
	height := int(workArea.Bottom - workArea.Top)
	left := int(workArea.Left)
	top := int(workArea.Top)
	setWindowPos.Call(hwnd, 0, uintptr(left), uintptr(top), uintptr(width), uintptr(height), swpNoZOrder|swpFrameChanged)
}

func toggleFullscreenNativeWindow(hwnd uintptr) {
	if fullscreenWindowState.active {
		setWindowPos.Call(hwnd, 0, uintptr(fullscreenWindowState.rect.Left), uintptr(fullscreenWindowState.rect.Top), uintptr(fullscreenWindowState.rect.Right-fullscreenWindowState.rect.Left), uintptr(fullscreenWindowState.rect.Bottom-fullscreenWindowState.rect.Top), swpNoZOrder|swpFrameChanged)
		fullscreenWindowState.active = false
		return
	}

	var current rect
	windowRectOk, _, _ := getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&current)))
	if windowRectOk == 0 {
		return
	}
	fullscreenWindowState.rect = current
	fullscreenWindowState.active = true

	bounds, monitorOk := monitorWorkArea(hwnd)
	if !monitorOk {
		fullscreenWindowState.active = false
		return
	}

	width := int(bounds.Right - bounds.Left)
	height := int(bounds.Bottom - bounds.Top)
	setWindowPos.Call(hwnd, 0, uintptr(bounds.Left), uintptr(bounds.Top), uintptr(width), uintptr(height), swpNoZOrder|swpFrameChanged)
}

// togglePageFullscreenNativeWindow is used only for page/video fullscreen.
func togglePageFullscreenNativeWindow(hwnd uintptr) {
	if pageFullscreenWindowState.active {
		setWindowLongPtr.Call(hwnd, uintptr(gwlStyle), pageFullscreenWindowState.style)
		r := pageFullscreenWindowState.rect
		setWindowPos.Call(hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), swpNoZOrder|swpFrameChanged)
		pageFullscreenWindowState.active = false
		return
	}

	var current rect
	ok, _, _ := getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&current)))
	if ok == 0 {
		return
	}
	bounds, monitorOk := monitorBounds(hwnd)
	if !monitorOk {
		return
	}

	style, _, _ := getWindowLongPtr.Call(hwnd, uintptr(gwlStyle))
	pageFullscreenWindowState.rect = current
	pageFullscreenWindowState.style = style
	pageFullscreenWindowState.active = true

	setWindowLongPtr.Call(hwnd, uintptr(gwlStyle), style&^uintptr(wsThickFrame))
	width := int(bounds.Right - bounds.Left)
	height := int(bounds.Bottom - bounds.Top)
	setWindowPos.Call(hwnd, 0, uintptr(bounds.Left), uintptr(bounds.Top), uintptr(width), uintptr(height), swpNoZOrder|swpFrameChanged)
}

func beginNativeWindowDrag(hwnd uintptr, x, y int) {
	releaseCapture.Call()
	sendMessage.Call(hwnd, wmNCLButtonDown, htCaption, 0)
}
