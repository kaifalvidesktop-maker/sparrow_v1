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

var fullscreenWindowState = struct {
	active bool
	rect   rect
}{
	active: false,
}

func configureNativeWindow(hwnd uintptr) {
	style, _, _ := getWindowLongPtr.Call(hwnd, uintptr(gwlStyle))
	style &^= wsCaption
	style |= wsThickFrame | wsMinimizeBox | wsMaximizeBox | wsSysMenu
	setWindowLongPtr.Call(hwnd, uintptr(gwlStyle), style)
	setWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged)
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

func beginNativeWindowDrag(hwnd uintptr, x, y int) {
	releaseCapture.Call()
	sendMessage.Call(hwnd, wmNCLButtonDown, htCaption, 0)
}
