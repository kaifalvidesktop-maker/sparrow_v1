//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdint.h>

void sparrow_configure_window(uintptr_t handle);
void sparrow_fit_window(uintptr_t handle, int width, int height);
void sparrow_close_window(uintptr_t handle);
void sparrow_minimize_window(uintptr_t handle);
void sparrow_toggle_maximize_window(uintptr_t handle);
void sparrow_toggle_fullscreen_window(uintptr_t handle);
void sparrow_begin_window_drag(uintptr_t handle, int x, int y);
*/
import "C"

func configureNativeWindow(hwnd uintptr) {
	C.sparrow_configure_window(C.uintptr_t(hwnd))
}

func fitWindowToTaskbarArea(hwnd uintptr, width, height int) {
	C.sparrow_fit_window(C.uintptr_t(hwnd), C.int(width), C.int(height))
}

func closeNativeWindow(hwnd uintptr) {
	C.sparrow_close_window(C.uintptr_t(hwnd))
}

func minimizeNativeWindow(hwnd uintptr) {
	C.sparrow_minimize_window(C.uintptr_t(hwnd))
}

func toggleMaximizeNativeWindow(hwnd uintptr) {
	C.sparrow_toggle_maximize_window(C.uintptr_t(hwnd))
}

func toggleFullscreenNativeWindow(hwnd uintptr) {
	C.sparrow_toggle_fullscreen_window(C.uintptr_t(hwnd))
}

func beginNativeWindowDrag(hwnd uintptr, x, y int) {
	C.sparrow_begin_window_drag(C.uintptr_t(hwnd), C.int(x), C.int(y))
}
