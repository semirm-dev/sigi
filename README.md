# sigi

Keep this machine online.

Idle timers lock the screen, drop the VPN, and mark you away in chat, usually
about ninety seconds into someone else's demo. `sigi` stops that by nudging the
mouse one pixel and putting it back, on a timer.

It runs in the foreground until you stop it. No daemon, no login item, nothing
left behind when you close the terminal.

## Install

Download a binary for your platform from the
[releases](https://github.com/semirm-dev/sigi/releases), or build one with
[Zig](https://ziglang.org/download/) 0.16.0:

```bash
zig build -Doptimize=ReleaseSafe    # zig-out/bin/sigi
```

There are no dependencies, no C toolchain, and no system libraries to install.
Linux and Windows binaries cross-compile from any machine. macOS links Apple's
frameworks, which Zig only finds for a native build, so macOS binaries are built
on a Mac for that Mac (releases ship arm64).

## Making it work

### macOS

Grant **Accessibility** permission. macOS silently drops synthetic input from
apps that do not have it, so `sigi` checks at startup and refuses with
instructions rather than ticking uselessly.

> System Settings → Privacy & Security → Accessibility → add your terminal

Grant it to the program that *runs* `sigi` — Terminal, iTerm, or your editor —
not to the `sigi` binary. If you switch terminals, grant it again.

### Linux

**X11 and Wayland both work.** `sigi` creates a virtual mouse through
`/dev/uinput`, which the kernel treats as hardware, so it never talks to a
display server. It needs access to that device:

```bash
sudo modprobe uinput                        # if /dev/uinput is missing
sudo usermod -aG input "$USER"              # then log out and back in
```

If the group is not enough, a udev rule makes it stick across reboots:

```bash
echo 'KERNEL=="uinput", GROUP="input", MODE="0660"'   | sudo tee /etc/udev/rules.d/99-uinput.rules
sudo udevadm control --reload-rules && sudo udevadm trigger
```

`sudo chmod 0666 /dev/uinput` also works, but it does not survive a reboot and
opens the device to everything.

### Windows

Nothing. `sigi` uses `SendInput`, which needs no special permission.

### If it runs but nothing happens

| | likely cause |
|---|---|
| macOS | Accessibility granted to a different app than the one running `sigi` |
| Linux | the virtual mouse was created but ignored — check `libinput list-devices` for `sigi` |
| Windows | a UAC prompt or the lock screen was showing, which blocks synthetic input |
| any | an idle policy that ignores synthetic input, e.g. a corporate screen-lock agent |

`sigi --verbose` prints a line per nudge, which tells you whether the loop is
running at all as opposed to the input being ignored.

## Use

```bash
sigi                      # nudge the mouse every 2 minutes
sigi --interval 30s       # more often
sigi --verbose            # say so on every nudge
```

| Flag | | What it does |
|---|---|---|
| `--interval` | `-i` | Time between nudges: a whole number with `ms`, `s`, `m` or `h`. Default `2m`. |
| `--verbose` | `-v` | Print each nudge as it happens. |
| `--help` | `-h` | Print usage. |
| `--version` | | Print the version. |

Ctrl-C stops it.

Each nudge moves the cursor one pixel right, waits 10ms, and moves it back.
Nothing appears to move, and every platform counts it as input.

## Upgrading from the Go version

- `--action` is gone. The mouse is the only action, and on Linux it now works
  under Wayland, which is what the keyboard action existed for.
- `--interval 0` is an error rather than silently falling back to `2m`.
- Linux needs `/dev/uinput` access instead of an X11 session.
- `go install` no longer applies; use a release binary or `zig build`.

## Development

```bash
make build     # zig build
make run ARGS="-i 2s -v"
make test      # zig build test --summary all
make lint      # zig fmt --check
make release   # ReleaseSmall binaries for every target this machine can build
```

```
src/main.zig     flags, the loop, and every message
src/linux.zig    /dev/uinput virtual mouse
src/macos.zig    Quartz event, Accessibility check
src/windows.zig  SendInput
```
