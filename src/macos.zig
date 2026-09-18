//! macOS: a Quartz mouse-moved event posted at the HID level.
//!
//! Unlike linux.zig (std.os.linux) and windows.zig (std.os.windows), there
//! is no std.os.macos with these bindings, so every function below is a bare
//! `extern fn` -- a C ABI declaration with no @import backing it. They
//! resolve at link time against Apple's frameworks, which build.zig links
//! only when the target is macOS (see the linkFramework calls there). That
//! is also why this file imports nothing.

pub fn init() !void {
    // Without Accessibility, Quartz accepts the events and silently drops
    // them, so refuse up front rather than tick uselessly for hours.
    if (AXIsProcessTrusted() == 0) return error.AccessibilityDenied;
}

pub fn move(dx: i32) !void {
    const here = CGEventCreate(null) orelse return error.EventFailed;
    var point = CGEventGetLocation(here);
    CFRelease(here);

    point.x += @floatFromInt(dx);
    const event = CGEventCreateMouseEvent(null, 5, point, 0) orelse // kCGEventMouseMoved, left button
        return error.EventFailed;
    defer CFRelease(event);
    CGEventPost(0, event); // kCGHIDEventTap
}

const CGPoint = extern struct { x: f64, y: f64 };
const CGEventRef = *opaque {};

// Declared, not imported: resolved at link time against the frameworks
// build.zig links in for macOS.
extern fn AXIsProcessTrusted() u8;
extern fn CGEventCreate(source: ?*anyopaque) ?CGEventRef;
extern fn CGEventGetLocation(event: CGEventRef) CGPoint;
extern fn CGEventCreateMouseEvent(source: ?*anyopaque, event_type: u32, position: CGPoint, button: u32) ?CGEventRef;
extern fn CGEventPost(tap: u32, event: CGEventRef) void;
extern fn CFRelease(object: *anyopaque) void;
