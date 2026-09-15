const std = @import("std");

const zon = @import("build.zig.zon");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    // The version lives in build.zig.zon and nowhere else.
    const options = b.addOptions();
    options.addOption([]const u8, "version", zon.version);

    const mod = b.createModule(.{
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    mod.addOptions("build_options", options);

    // Quartz events and the Accessibility check. Nothing to link elsewhere:
    // Linux is raw syscalls, and Windows' user32 comes from the extern.
    // Zig only finds the macOS SDK for native builds, so macOS is never cross-compiled.
    if (target.result.os.tag == .macos) {
        mod.linkFramework("CoreGraphics", .{});
        mod.linkFramework("CoreFoundation", .{});
        mod.linkFramework("ApplicationServices", .{});
    }

    const exe = b.addExecutable(.{ .name = "sigi", .root_module = mod });
    b.installArtifact(exe);

    const run = b.addRunArtifact(exe);
    if (b.args) |args| run.addArgs(args);
    b.step("run", "Run sigi; pass flags after --").dependOn(&run.step);

    const tests = b.addTest(.{ .root_module = mod });
    b.step("test", "Run the unit tests").dependOn(&b.addRunArtifact(tests).step);
}
