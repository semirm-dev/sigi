# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What sigi is

A personal foreground CLI that keeps a machine online by nudging the mouse a
pixel and back on a timer, until Ctrl-C. Zig 0.16.0, std only.

## Keep it small

This is a personal tool. The least code that keeps the behaviour wins: no
interfaces, wrappers, option structs or production-app layering. There are
four source files; add a fifth only when one of them stops being readable.

## Commands

```bash
zig build                        # zig-out/bin/sigi
zig build run -- -i 2s -v        # run from source; flags after --
zig build test --summary all
zig fmt --check build.zig build.zig.zon src
make build|run|test|lint         # thin wrappers over the above
make release                     # bin/sigi-<os>-<arch> for every target it can
                                 # build here, plus SHA256SUMS -- exactly what a
                                 # tag publishes, and what release.yml calls, so
                                 # the target list and the names have one home
make build-linux|build-win|build-osx  # one platform each, same names
```

Released binaries are named `sigi-<os>-<arch>` -- `macos`/`linux`/`windows`
and `amd64`/`arm64` -- which is what jq and most single-binary CLIs publish,
and what ghu publishes. Zig's own triples read the other way round
(`x86_64-linux`), so `release.yml` maps each target to a name and fails on a
target it has no name for, rather than publishing a triple.

Cross-compile with `-Dtarget=`. Linux and Windows targets build from any host.
macOS only builds natively (`zig build`, or `-Dtarget=native`) on a Mac: Zig
finds the SDK only for native builds, and an explicit `-Dtarget=aarch64-macos`
fails to find the frameworks. Making that work needs SDK-path code in
`build.zig`; it was left out on purpose. This is why `make build-osx` checks
`uname` and refuses rather than quietly building a native Linux binary and
naming it `sigi-macos`. It is also why releases are built in CI rather
than by hand: `.github/workflows/release.yml` builds macOS natively on a
`macos-latest` runner, so no Mac has to be present for a tag to ship.

**Apple Silicon is the only macOS target, deliberately.** `macos-latest` is
arm64, so that is what a release ships. Intel Macs are not a goal -- do not add
an Intel runner to the matrix or SDK-path code to `build.zig` to cover them. An
Intel Mac builds from source, which the README says.

The version lives in `build.zig.zon` only, and reaches the code as
`build_options.version`.

## Releases

Pushing a `v*` tag publishes one. `.github/workflows/release.yml` builds every
target, checks the tag against `build.zig.zon`, and attaches the binaries with
`SHA256SUMS`. To cut one:

1. Commit the work and `git push`. Let CI go green.
2. `make tag VERSION=x.y.z`, then `git push && git push origin vx.y.z`.

`make tag` refuses a dirty tree, rewrites `.version`, builds, checks the binary
reports the new version, and only then commits and tags. **Use it rather than
doing the three steps by hand.** The hazard it removes is that a tag on a
commit which predates the bump fails the release job -- the version it finds is
the old one -- and ghu lost a release to exactly that. If it happens anyway,
move the tag (delete it locally and on the remote, recreate it, push again)
rather than weakening the check: forcing past it publishes binaries that
disagree with their own tag.

The version lives in `build.zig.zon` because that is Zig's package manifest,
the way `Cargo.toml` is Rust's. ghu has no equivalent file, so it takes its
version from the tag instead and has no bump commit at all -- the two projects
differ here on purpose.

`workflow_dispatch` runs the same builds without publishing, because the
release job is gated on `refs/tags/v*`. Use it to check a build before
spending a tag on it.

## Layout

```
src/main.zig     flags, the loop, every message; picks the platform file at compile time
src/linux.zig    /dev/uinput virtual mouse
src/macos.zig    Quartz event, Accessibility check
src/windows.zig  SendInput
```

Each platform file exports `init() !void` and `move(dx: i32) !void`, and
nothing else. `main.zig` switches on `builtin.os.tag` to pick one.

## Decisions worth keeping

**Any failure exits 1.** A keep-alive whose nudge fails is not keeping
anything alive. Errors a person can fix get a sentence in `describe` in
`main.zig`; the rest print their name.

**No signal handling.** Ctrl-C ends the process, and the kernel removes the
uinput device when its fd closes. Nothing else needs cleaning up.

**Linux uses `/dev/uinput`, not X11.** The kernel treats the device as
hardware, which is why Wayland works and why the build needs no system
headers. Do not reintroduce a display-server library. The device needs
`BTN_LEFT`, `REL_X` and `REL_Y`, or udev will not tag it as a mouse.

**Nothing grants access to `/dev/uinput` by default,** so most of the Linux
support burden is a permissions problem, not a code one. Stock udev groups
`SUBSYSTEM=="input"` devices and uinput is `misc`, so it stays `root:root
0600` and joining the `input` group on its own changes nothing. The rule that
works carries `OPTIONS+="static_node=uinput"`: udev pre-creates the node from
`modules.devname` before the module is loaded, and a rule without that option
never matches the node it made — which is also why restarting `systemd-udevd`
applies the rule and `udevadm trigger` does not. The advice lives in two
places, the README's Linux section and `describe`'s `UinputDenied` string.
Change them together.

**WSL runs sigi but cannot keep Windows awake.** The virtual mouse lives in
the WSL VM, invisible to the Windows session whose idle timer is the actual
target. The answer there is the cross-compiled Windows binary, not a Linux-side
fix — do not chase it as a uinput bug.

**macOS checks `AXIsProcessTrusted()` before the first tick.** Without
Accessibility, Quartz accepts events and silently drops them, so sigi would
tick for hours while the screen locks. Keep the check.

**ABI numbers are `comptime` asserts,** not tests, so every cross-compile
checks them against the C headers' values.

## Conventions

`zig fmt` formats. Imports go `builtin`/`std`, then `build_options`, then
local files.
