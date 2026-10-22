//go:build linux

package main

/*
#cgo linux pkg-config: gtk+-3.0 webkit2gtk-4.0
#include <gtk/gtk.h>
#include <stdint.h>

static inline GtkWindow *sparrow_window(uintptr_t handle) {
	return (GtkWindow *)handle;
}

static inline void sparrow_configure_window(uintptr_t handle) {
	GtkWindow *window = sparrow_window(handle);
	gtk_window_set_decorated(window, FALSE);
	gtk_window_set_resizable(window, TRUE);
}

static inline void sparrow_fit_window(uintptr_t handle, int width, int height) {
	GtkWindow *window = sparrow_window(handle);
	gtk_window_set_default_size(window, width, height);
	gtk_window_set_position(window, GTK_WIN_POS_CENTER);
}

static inline void sparrow_close_window(uintptr_t handle) {
	gtk_window_close(sparrow_window(handle));
}

static inline void sparrow_minimize_window(uintptr_t handle) {
	gtk_window_iconify(sparrow_window(handle));
}

static inline void sparrow_toggle_maximize_window(uintptr_t handle) {
	GtkWindow *window = sparrow_window(handle);
	if (gtk_window_is_maximized(window)) {
		gtk_window_unmaximize(window);
	} else {
		gtk_window_maximize(window);
	}
}

static inline void sparrow_toggle_fullscreen_window(uintptr_t handle) {
	GtkWindow *window = sparrow_window(handle);
	GdkWindow *gw = gtk_widget_get_window(GTK_WIDGET(window));
	if (gw && (gdk_window_get_state(gw) & GDK_WINDOW_STATE_FULLSCREEN)) {
		gtk_window_unfullscreen(window);
	} else {
		gtk_window_fullscreen(window);
	}
}

static inline void sparrow_begin_window_drag(uintptr_t handle, int x, int y) {
	gtk_window_begin_move_drag(sparrow_window(handle), 1, x, y, GDK_CURRENT_TIME);
}
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
