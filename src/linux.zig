//! Linux: a virtual mouse made through /dev/uinput. The kernel treats it as
//! hardware, so it works under X11 and Wayland with no display library, and
//! it disappears when the process exits.

const std = @import("std");
const linux = std.os.linux;

var fd: linux.fd_t = -1;

pub fn init() !void {
    const rc = linux.open("/dev/uinput", .{ .ACCMODE = .WRONLY, .CLOEXEC = true }, 0);
    switch (linux.errno(rc)) {
        .SUCCESS => fd = @intCast(rc),
        .NOENT, .NODEV => return error.UinputMissing,
        .ACCES, .PERM => return error.UinputDenied,
        else => return error.UinputFailed,
    }
    // udev only tags a device as a mouse if it has a button and both axes.
    try ioctl(UI_SET_EVBIT, EV_KEY);
    try ioctl(UI_SET_KEYBIT, BTN_LEFT);
    try ioctl(UI_SET_EVBIT, EV_REL);
    try ioctl(UI_SET_RELBIT, REL_X);
    try ioctl(UI_SET_RELBIT, REL_Y);

    var setup: Setup = .{};
    @memcpy(setup.name[0..4], "sigi");
    try ioctl(UI_DEV_SETUP, @intFromPtr(&setup));
    try ioctl(UI_DEV_CREATE, 0);
}

pub fn move(dx: i32) !void {
    const events = [_]Event{
        .{ .type = EV_REL, .code = REL_X, .value = dx },
        .{ .type = EV_SYN, .code = SYN_REPORT, .value = 0 },
    };
    const bytes = std.mem.sliceAsBytes(&events);
    if (linux.write(fd, bytes.ptr, bytes.len) != bytes.len) return error.UinputFailed;
}

fn ioctl(request: u32, arg: usize) !void {
    if (linux.errno(linux.ioctl(fd, request, arg)) != .SUCCESS) return error.UinputFailed;
}

// linux/input-event-codes.h
const EV_SYN = 0x00;
const EV_KEY = 0x01;
const EV_REL = 0x02;
const SYN_REPORT = 0x00;
const REL_X = 0x00;
const REL_Y = 0x01;
const BTN_LEFT = 0x110;

// linux/uinput.h
const UI_DEV_CREATE = linux.IOCTL.IO('U', 1);
const UI_DEV_SETUP = linux.IOCTL.IOW('U', 3, Setup);
const UI_SET_EVBIT = linux.IOCTL.IOW('U', 100, c_int);
const UI_SET_KEYBIT = linux.IOCTL.IOW('U', 101, c_int);
const UI_SET_RELBIT = linux.IOCTL.IOW('U', 102, c_int);

/// struct uinput_setup
const Setup = extern struct {
    bustype: u16 = 0x06, // BUS_VIRTUAL
    vendor: u16 = 0,
    product: u16 = 0,
    version: u16 = 1,
    name: [80]u8 = @splat(0),
    ff_effects_max: u32 = 0,
};

/// struct input_event. The kernel ignores the timestamp on written events.
const Event = extern struct {
    time: [2]isize = .{ 0, 0 },
    type: u16,
    code: u16,
    value: i32,
};

// Checked against the C headers on every build for Linux.
comptime {
    std.debug.assert(@sizeOf(Setup) == 92);
    std.debug.assert(UI_DEV_SETUP == 0x405c5503);
    std.debug.assert(UI_SET_EVBIT == 0x40045564);
}
