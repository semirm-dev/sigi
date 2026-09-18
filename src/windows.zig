//! Windows: SendInput, which needs no special permission.

const std = @import("std");
const windows = std.os.windows;

pub fn init() !void {}

pub fn move(dx: i32) !void {
    const input: INPUT = .{ .mi = .{ .dx = dx } };
    if (SendInput(1, @ptrCast(&input), @sizeOf(INPUT)) != 1) return error.SendInputFailed;
}

/// Only the mouse member of the Win32 union is declared. It is the largest
/// member, so the size still matches.
const INPUT = extern struct {
    type: windows.DWORD = 0, // INPUT_MOUSE
    mi: extern struct {
        dx: windows.LONG,
        dy: windows.LONG = 0,
        mouseData: windows.DWORD = 0,
        dwFlags: windows.DWORD = 0x0001, // MOUSEEVENTF_MOVE
        time: windows.DWORD = 0,
        dwExtraInfo: windows.ULONG_PTR = 0,
    },
};

// Like macos.zig's externs, this is a bare C ABI declaration, not a std
// wrapper. Unlike macOS's frameworks, user32 needs no linkFramework-equivalent
// call in build.zig: Zig's default Windows target libs link it already.
extern "user32" fn SendInput(count: windows.UINT, inputs: [*]const INPUT, size: c_int) callconv(.winapi) windows.UINT;

comptime {
    std.debug.assert(@sizeOf(INPUT) == if (@sizeOf(usize) == 8) 40 else 28);
}
