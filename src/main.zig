//! sigi keeps this machine online by nudging the mouse a pixel and back on a
//! timer, until Ctrl-C.

const builtin = @import("builtin");
const std = @import("std");

const build_options = @import("build_options");

/// One file per platform, picked at compile time.
const mouse = switch (builtin.os.tag) {
    .linux => @import("linux.zig"),
    .macos => @import("macos.zig"),
    .windows => @import("windows.zig"),
    else => @compileError("sigi supports linux, macos and windows"),
};

const usage =
    \\Usage: sigi [-i <interval>] [-v]
    \\
    \\Keeps this machine online by nudging the mouse a pixel and back, until Ctrl-C.
    \\
    \\  -i, --interval <interval>  time between nudges: 30s, 2m, 1h (default 2m)
    \\  -v, --verbose              print each nudge
    \\  -h, --help                 print this help
    \\      --version              print the version
    \\
;

pub fn main(init: std.process.Init) !void {
    const io = init.io;
    var buffer: [256]u8 = undefined;
    var stdout: std.Io.File.Writer = .init(.stdout(), io, &buffer);
    const out = &stdout.interface;

    var interval_text: []const u8 = "2m";
    var verbose = false;

    const args = try init.minimal.args.toSlice(init.arena.allocator());
    var i: usize = 1;
    while (i < args.len) : (i += 1) {
        const arg = args[i];
        if (eql(arg, "-h") or eql(arg, "--help")) {
            try out.writeAll(usage);
            return out.flush();
        } else if (eql(arg, "--version")) {
            try out.print("sigi {s}\n", .{build_options.version});
            return out.flush();
        } else if (eql(arg, "-v") or eql(arg, "--verbose")) {
            verbose = true;
        } else if ((eql(arg, "-i") or eql(arg, "--interval")) and i + 1 < args.len) {
            i += 1;
            interval_text = args[i];
        } else {
            fail("bad argument \"{s}\", see sigi --help", .{arg});
        }
    }
    const interval = parseDuration(interval_text) orelse
        fail("bad interval \"{s}\": use a whole number with ms, s, m or h, e.g. 30s", .{interval_text});

    mouse.init() catch |err| fail("{s}", .{describe(err)});
    try out.print("sigi: mouse every {s}, ctrl-c to stop\n", .{interval_text});
    try out.flush();

    while (true) {
        try io.sleep(.fromNanoseconds(interval), .awake);
        nudge(io) catch |err| fail("{s}", .{describe(err)});
        if (verbose) {
            try out.writeAll("mouse\n");
            try out.flush();
        }
    }
}

/// Right one pixel and back. The pause stops the pair being merged into no
/// movement at all.
fn nudge(io: std.Io) !void {
    try mouse.move(1);
    io.sleep(.fromMilliseconds(10), .awake) catch {};
    try mouse.move(-1);
}

/// A whole number followed by ms, s, m or h, in nanoseconds. Null when it is
/// malformed, zero, or too big.
fn parseDuration(text: []const u8) ?u64 {
    const digits = std.mem.indexOfNone(u8, text, "0123456789") orelse text.len;
    const count = std.fmt.parseInt(u64, text[0..digits], 10) catch return null;
    const suffix = text[digits..];
    const unit: u64 = if (eql(suffix, "ms"))
        1_000_000
    else if (eql(suffix, "s"))
        1_000_000_000
    else if (eql(suffix, "m"))
        60_000_000_000
    else if (eql(suffix, "h"))
        3_600_000_000_000
    else
        return null;
    if (count == 0) return null;
    return std.math.mul(u64, count, unit) catch null;
}

/// What to do about the errors a person can fix. The rest print their name.
fn describe(err: anyerror) []const u8 {
    return switch (err) {
        error.UinputMissing => "/dev/uinput does not exist: run 'sudo modprobe uinput'",
        error.UinputDenied => "no permission to open /dev/uinput: run 'sudo usermod -aG input $USER' and log in again (the README has a udev rule if that is not enough)",
        error.AccessibilityDenied => "macOS ignores synthetic input without Accessibility: System Settings > Privacy & Security > Accessibility, then add your terminal",
        else => @errorName(err),
    };
}

fn fail(comptime format: []const u8, args: anytype) noreturn {
    std.debug.print("sigi: " ++ format ++ "\n", args);
    std.process.exit(1);
}

fn eql(a: []const u8, b: []const u8) bool {
    return std.mem.eql(u8, a, b);
}

test parseDuration {
    try std.testing.expectEqual(250_000_000, parseDuration("250ms"));
    try std.testing.expectEqual(30_000_000_000, parseDuration("30s"));
    try std.testing.expectEqual(120_000_000_000, parseDuration("2m"));
    try std.testing.expectEqual(7_200_000_000_000, parseDuration("2h"));
    for ([_][]const u8{ "", "0s", "30", "s", "-5s", "1.5m", "3d", "18446744073709551615h" }) |bad| {
        try std.testing.expectEqual(null, parseDuration(bad));
    }
}
