#import <Cocoa/Cocoa.h>
#include <stdint.h>

static NSWindow *sparrow_window(uintptr_t handle) {
	return (NSWindow *)handle;
}

void sparrow_configure_window(uintptr_t handle) {
	NSWindow *window = sparrow_window(handle);
	[window setTitleVisibility:NSWindowTitleHidden];
	[window setTitlebarAppearsTransparent:YES];
	[[window standardWindowButton:NSWindowCloseButton] setHidden:YES];
	[[window standardWindowButton:NSWindowMiniaturizeButton] setHidden:YES];
	[[window standardWindowButton:NSWindowZoomButton] setHidden:YES];
}

void sparrow_fit_window(uintptr_t handle, int width, int height) {
	NSWindow *window = sparrow_window(handle);
	NSRect visible = [[NSScreen mainScreen] visibleFrame];
	if (width > NSWidth(visible)) width = (int)NSWidth(visible);
	if (height > NSHeight(visible)) height = (int)NSHeight(visible);
	[window setContentSize:NSMakeSize(width, height)];
	[window center];
}

void sparrow_close_window(uintptr_t handle) {
	[sparrow_window(handle) performClose:nil];
}

void sparrow_minimize_window(uintptr_t handle) {
	[sparrow_window(handle) miniaturize:nil];
}

void sparrow_toggle_maximize_window(uintptr_t handle) {
	[sparrow_window(handle) zoom:nil];
}

void sparrow_toggle_fullscreen_window(uintptr_t handle) {
	[sparrow_window(handle) toggleFullScreen:nil];
}

void sparrow_begin_window_drag(uintptr_t handle, int x, int y) {
	NSWindow *window = sparrow_window(handle);
	NSScreen *screen = [window screen] ?: [NSScreen mainScreen];
	NSPoint screenPoint = NSMakePoint(x, NSMaxY([screen frame]) - y);
	NSPoint windowPoint = [window convertPointFromScreen:screenPoint];
	NSEvent *event = [NSEvent mouseEventWithType:NSEventTypeLeftMouseDown
		location:windowPoint
		modifierFlags:0
		timestamp:[[NSProcessInfo processInfo] systemUptime]
		windowNumber:[window windowNumber]
		context:nil
		eventNumber:0
		clickCount:1
		pressure:1.0];
	[window performWindowDragWithEvent:event];
}
