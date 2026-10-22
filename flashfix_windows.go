//go:build windows

package main

import (
	"syscall"
	"time"
)

var (
	user32dll            = syscall.NewLazyDLL("user32.dll")
	gdi32dll             = syscall.NewLazyDLL("gdi32.dll")
	procSetClassLongPtrW = user32dll.NewProc("SetClassLongPtrW")
	procInvalidateRect   = user32dll.NewProc("InvalidateRect")
	procCreateSolidBrush = gdi32dll.NewProc("CreateSolidBrush")
)

// applyBackdropFix paints the host window's own background white instead of
// the default black brush. Child windows appear asynchronously, so we retry.
func applyBackdropFix(hwnd uintptr) {
	const gclpHbrBackground = ^uintptr(9) // -10
	brush, _, _ := procCreateSolidBrush.Call(0x00FFFFFF) // COLORREF 0x00BBGGRR
	if brush == 0 {
		return
	}
	apply := func() {
		procSetClassLongPtrW.Call(hwnd, gclpHbrBackground, brush)
		procInvalidateRect.Call(hwnd, 0, 1)
	}
	apply()
	go func() {
		for _, d := range []time.Duration{300, 1000, 2500} {
			time.Sleep(d * time.Millisecond)
			apply()
		}
	}()
}
