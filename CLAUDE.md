# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What sigi is

A foreground CLI that stops a machine going idle, by nudging the mouse a pixel
or tapping caps lock on a timer. It runs until Ctrl-C.

## Commands

```bash
make build    # .build/sigi, version stamped from VERSION
make run      # run from source
make lint     # gofmt, goimports, declaration order, vet -- run before committing
make order    # rewrite declarations into the house order
make test     # go test ./... -race -count=1
```

A single test runs with `go test ./internal/keepalive/ -run TestName -count=1`.

**Building needs cgo and X11 development headers.** `robotgo` drives real input
devices, so `CGO_ENABLED=0` will not build and neither will cross-compilation:
each platform must be built on that platform. On Debian and Ubuntu the headers
are `libx11-dev`, `libxtst-dev`, `xorg-dev`. Without them `go build ./...`
fails on a missing `X11/Xutil.h`, which looks like a code error and is not.

To type-check without those headers, point the module at a stub:

```bash
go mod edit -replace github.com/go-vgo/robotgo=/path/to/stub
```

where the stub is a module named `github.com/go-vgo/robotgo` exposing
`func MoveRelative(x, y int)`. Undo it with `go mod edit -dropreplace`.

## Architecture

```
cmd/sigi/            main: signal.NotifyContext, and one command
internal/keepalive/  the loop, the two actions, and the command that drives them
```

**The command lives with the feature it drives.** `keepalive.Command` builds the
cobra command, owns the flags, and prints; `cmd/sigi` only wires signals to it.

**The Runner does timing and nothing else.** What happens on a tick is an
`Action`; reporting is `reporting`, a decorator around one, which is how
`--verbose` works without the loop knowing what a terminal is. Do not put
writers or flags on `Runner`.

**`Action` is an interface because there are three implementations** — mouse,
keyboard, and the counting one in the tests. That is the bar: do not add an
interface with a single implementation.

**Nothing is a package-level variable** except `Version`. Flags are locals in
`Command`, closed over by `RunE`, so two commands can exist in one process and
tests do not fight each other.

**Run stops on the first failed Action** rather than logging and continuing. A
keep-alive whose action does not work is not keeping anything alive.

**Both actions work on all three platforms.** keybd_event ships
keybd_darwin.go, keybd_linux.go and keybd_windows.go, all defining
VK_CAPSLOCK, so the keyboard action needs no build tag -- the original code
carried a `linux || windows` one that was simply wrong.

**The mouse action needs X11 on Linux.** robotgo's mouse is Xlib and XTest;
its wayland file is inert, and its build tag is misspelled `+bulid` so it never
compiles anyway. `NewMouse` therefore refuses on Linux when `$WAYLAND_DISPLAY`
or `XDG_SESSION_TYPE=wayland` is set, or when `$DISPLAY` is empty, and names
`--action keyboard` as the way out. Keep that check: a keep-alive that ticks
for hours while the screen locks behind it is the failure worth preventing, and
it is invisible without it.

## Declaration order

Every file orders its top-level declarations this way, and `make lint` fails if
one does not:

1. constants
2. variables
3. exported types
4. unexported types
5. exported functions
6. exported methods
7. unexported methods
8. unexported functions

`scripts/order.py` does the rewriting; `--check` reports without rewriting.

## Conventions

Imports are grouped stdlib / third-party / local. Error strings are lowercase
and unpunctuated. No stuttering: the package is `keepalive`, so the types are
`Runner`, `Mouse`, `Keyboard` — not `KeepaliveRunner`.
