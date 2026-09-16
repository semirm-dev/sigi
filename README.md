# sigi

Keep this machine online.

Idle timers lock the screen, drop the VPN, and mark you away in chat, usually
about ninety seconds into someone else's demo. `sigi` stops that by nudging the
mouse one pixel and putting it back, on a timer.

It runs in the foreground until you stop it. No daemon, no login item, nothing
left behind when you close the terminal.

## Install

Every [release](https://github.com/semirm-dev/sigi/releases) carries a binary
per platform -- pick the one for your machine, make it executable, and put it on
your `PATH`:

| | |
|---|---|
| macOS, Apple Silicon | `sigi-macos-arm64` |
| Linux, x86-64 | `sigi-linux-amd64` |
| Linux, ARM64 | `sigi-linux-arm64` |
| Windows, x86-64 | `sigi-windows-amd64.exe` |
| Windows, ARM64 | `sigi-windows-arm64.exe` |

```bash
curl -LO https://github.com/semirm-dev/sigi/releases/latest/download/sigi-macos-arm64
chmod +x sigi-macos-arm64
sudo mv sigi-macos-arm64 /usr/local/bin/sigi
```

`SHA256SUMS` is attached to the same release; verify with
`sha256sum --ignore-missing -c SHA256SUMS`.

Releases ship Apple Silicon for macOS. Anything else -- an Intel Mac, a
platform not in the table -- builds from source:

```bash
zig build -Doptimize=ReleaseSafe    # zig-out/bin/sigi
```

There are no dependencies, no C toolchain, and no system libraries to install.
Linux and Windows binaries cross-compile from any machine.

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
display server. It needs access to that device, and nothing grants it by
default: udev's stock rules put `SUBSYSTEM=="input"` devices in the `input`
group, and `uinput` is in `misc`, so it stays `root:root 0600`. Joining the
group does nothing until a rule puts the device in it, so do both:

```bash
echo 'KERNEL=="uinput", SUBSYSTEM=="misc", GROUP="input", MODE="0660", OPTIONS+="static_node=uinput"' \
  | sudo tee /etc/udev/rules.d/99-uinput.rules
sudo usermod -aG input "$USER"
sudo modprobe uinput                        # if /dev/uinput is missing
```

Then reboot, or apply both halves in place:

```bash
sudo udevadm control --reload-rules && sudo systemctl restart systemd-udevd
newgrp input                                # this shell only; log in again for the rest
```

`OPTIONS+="static_node=uinput"` is the part that is easy to leave out. Until
something loads the module, udev creates `/dev/uinput` ahead of time from
`modules.devname`, and a rule without that option never reaches the node it
made — which is also why restarting udev applies the rule and `udevadm
trigger` does not.

`sudo chmod 0666 /dev/uinput` works for the current boot, and opens the device
to every process on the machine.

**WSL is the exception.** The rule above works there, but the virtual mouse
lives inside the WSL VM, where Windows cannot see it, so it will not hold off a
Windows idle timer. Use the Windows binary from a Windows shell instead.

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
make build     # Debug build for working on sigi -- unoptimised, and large
make build-linux   # bin/sigi-linux-amd64
make build-win     # bin/sigi-windows-amd64.exe
make build-osx     # bin/sigi-macos-<arch> -- native, so a Mac only
make run ARGS="-i 2s -v"
make test      # zig build test --summary all
make lint      # zig fmt --check
make release   # bin/sigi-<os>-<arch> for every target, plus SHA256SUMS
make tag       # bump, commit and tag -- make tag VERSION=2.0.3
```

```
src/main.zig     flags, the loop, and every message
src/linux.zig    /dev/uinput virtual mouse
src/macos.zig    Quartz event, Accessibility check
src/windows.zig  SendInput
```

## Releasing

```bash
make tag VERSION=2.0.3
git push && git push origin v2.0.3
```

The version lives in `build.zig.zon` and nowhere else. `make tag` refuses a
dirty tree, rewrites `.version`, builds, checks `sigi --version` reports the new
number, and only then commits the bump and tags that commit -- so the tag cannot
land on a commit that predates it, which is the way this goes wrong when the
three steps are done by hand.

Pushing the tag runs [`release.yml`](.github/workflows/release.yml): Linux and
Windows cross-compile on one runner, macOS builds natively on a macOS runner,
and the binaries are published with `SHA256SUMS` as a
[release](https://github.com/semirm-dev/sigi/releases).
