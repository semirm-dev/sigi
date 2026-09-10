# sigi

Keep this machine awake.

Idle timers lock the screen, drop the VPN, and mark you away in chat, usually
about ninety seconds into someone else's demo. `sigi` stops that by doing
something small on a timer — nudging the mouse one pixel and putting it back,
or tapping a key that changes nothing you can see.

It runs in the foreground until you stop it. No daemon, no login item, nothing
left behind when you close the terminal.

## Install

```bash
go install github.com/semirm-dev/sigi/cmd/sigi@latest
```

Or from a checkout:

```bash
make build     # .build/sigi for this machine
make install   # into GOBIN
```

`sigi` drives real input devices, so it needs cgo and a C toolchain, and it
cannot be cross-compiled: each binary has to be built on the system it runs on.
Tagged releases are built by CI, one runner per platform — see
[`.github/workflows/release.yml`](.github/workflows/release.yml) — and attached
to the release. `make release` builds one for the machine you are on.

## Making it work

Both actions work on all three platforms, but each one needs something from the
operating system first. This is the whole of it.

### macOS

```bash
xcode-select --install     # once, for the C toolchain
go install github.com/semirm-dev/sigi/cmd/sigi@latest
```

Then grant **Accessibility** permission, or both actions will run and do
nothing: macOS silently drops synthetic input from apps that do not have it.

> System Settings → Privacy & Security → Accessibility → add your terminal

Grant it to the program that *runs* `sigi` — Terminal, iTerm, or your editor —
not to the `sigi` binary. If you switch terminals, grant it again. Both actions
go through Quartz (`CGEventPost`), so the permission covers both.

### Linux

Build needs the X11 headers, even if you only ever use `--action keyboard`,
because robotgo is compiled either way:

```bash
sudo apt install libx11-dev libxtst-dev xorg-dev    # Debian, Ubuntu
sudo dnf install libX11-devel libXtst-devel         # Fedora
```

**`--action mouse` needs an X11 session.** It is Xlib and XTest underneath, so
under Wayland it would move nothing — `sigi` detects that and refuses rather
than ticking uselessly. Check with `echo $XDG_SESSION_TYPE`.

**`--action keyboard` works under both X11 and Wayland**, because it writes to
`/dev/uinput` and never talks to a display server. It needs access to that
device:

```bash
sudo modprobe uinput                        # if /dev/uinput is missing
sudo usermod -aG input "$USER"              # then log out and back in
```

If the group is not enough, a udev rule makes it stick across reboots:

```bash
echo 'KERNEL=="uinput", GROUP="input", MODE="0660"'   | sudo tee /etc/udev/rules.d/99-uinput.rules
sudo udevadm control --reload-rules && sudo udevadm trigger
```

`sudo chmod 0666 /dev/uinput` also works and is what the underlying library
suggests, but it does not survive a reboot and opens the device to everything.

On Wayland, use `--action keyboard`. It is the only one that works there.

### Windows

Install a C toolchain — [MSYS2](https://www.msys2.org/) with
`pacman -S mingw-w64-ucrt-x86_64-gcc`, or TDM-GCC — then:

```powershell
go install github.com/semirm-dev/sigi/cmd/sigi@latest
```

Nothing else. Both actions use Win32 (`SendInput` and `user32!keybd_event`) and
need no special permission. The keyboard action needs no cgo at all here.

### If it runs but nothing happens

| | likely cause |
|---|---|
| macOS | Accessibility not granted, or granted to the wrong app |
| Linux, mouse | a Wayland session — `sigi` should have refused; if not, `echo $DISPLAY` |
| Linux, keyboard | no access to `/dev/uinput` — it reports this rather than failing silently |
| any | an idle policy that ignores synthetic input, e.g. a corporate screen-lock agent |

`sigi --verbose` prints a line per tick, which tells you whether the loop is
running at all as opposed to the input being ignored.

## Use

```bash
sigi                      # nudge the mouse every 2 minutes
sigi --interval 30s       # more often
sigi --action keyboard    # tap caps lock instead
sigi --verbose            # say so on every tick
```

| Flag | | What it does |
|---|---|---|
| `--interval` | `-i` | How long to wait between actions, e.g. `30s` or `2m`. Default `2m`. |
| `--action` | `-a` | `mouse` or `keyboard`. Default `mouse`. |
| `--verbose` | `-v` | Print each action as it happens. |

Ctrl-C stops it. The loop is cancelled first, so it never exits mid-action.

### Which action

**`mouse`** moves the cursor one pixel right, waits 10ms, and moves it back.
Nothing appears to move, and every platform counts it as input. It is the
default because it needs no device permissions — but on Linux it needs an X11
session, so under Wayland use the keyboard.

**`keyboard`** taps caps lock, which registers as input and leaves nothing
behind after the second tap. It is the one that works everywhere, including
Wayland, at the cost of needing `/dev/uinput` access on Linux. It also waits
two seconds there before the first tap, which the library requires.

## Development

```bash
make lint    # gofmt, imports, declaration order, vet
make order   # rewrite declarations into the house order
make test    # go test ./... -race -count=1
```

```
cmd/sigi/            main: signal handling, and one command
internal/keepalive/  the feature: the loop, the actions, and the command
```

`keepalive` owns the command that drives it, so everything about `sigi` is in
one directory. The `Runner` does timing and nothing else: what to do on each
tick is an `Action`, and reporting is a decorator around one, so the loop runs
with no terminal attached and no logging framework underneath it.

Declarations are ordered in every file — constants, variables, exported types,
unexported types, exported functions, exported methods, unexported methods,
unexported functions — and `make lint` fails if one is not.
