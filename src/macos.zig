//! macOS: a Quartz mouse-moved event posted at the HID level.

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

// From the frameworks build.zig links for macOS.
extern fn AXIsProcessTrusted() u8;
extern fn CGEventCreate(source: ?*anyopaque) ?CGEventRef;
extern fn CGEventGetLocation(event: CGEventRef) CGPoint;
extern fn CGEventCreateMouseEvent(source: ?*anyopaque, event_type: u32, position: CGPoint, button: u32) ?CGEventRef;
extern fn CGEventPost(tap: u32, event: CGEventRef) void;
extern fn CFRelease(object: *anyopaque) void;
