# sigi

Keep-alive CLI utility that prevents system idle/sleep by performing periodic mouse or keyboard actions.

## Commands

```shell
make run          # go run cmd/main.go
make build        # GOOS=darwin GOARCH=arm64 go build -o build/sigi-darwin-arm64 ./cmd/main.go
make test         # go test -v ./...
make test-cover   # tests with HTML coverage report
```

## Project Structure

```
cmd/main.go                     — entry point, calls sigi.Execute()
sigi.go                         — Cobra root command, CLI flags (--interval, --logs), signal handling
action/mouse.go                 — MouseMove: moves mouse 1px right/left via robotgo (all platforms)
action/keyboardbutton.go        — KeyboardButton: toggles CAPSLOCK via keybd_event (linux||windows build tag)
runner/intervalrunner.go        — IntervalRunner struct, Action interface, ticker loop with context cancellation
runner/intervalrunner_test.go   — tests using testify/assert and mock Action
```

## Architecture

- **Action interface** (`runner/intervalrunner.go`): `Execute() error` — implemented by `MouseMove` and `KeyboardButton`
- **IntervalRunner** (`runner/intervalrunner.go`): runs an Action on a timer, supports context cancellation and error channel
- **CLI** (`sigi.go`): Cobra command wires IntervalRunner with MouseMove, handles SIGINT/SIGTERM
- **Platform selection**: `keyboardbutton.go` uses `//go:build linux || windows` build tag; `mouse.go` has no build tag (available everywhere). Currently `sigi.go` always uses `MouseMove` regardless of platform.

## Conventions

- Build tags (`//go:build`) for platform-specific code
- `logrus` for all logging
- `testify/assert` for test assertions
- Mock structs implementing Action interface for testing
- Cobra for CLI with `init()` flag registration

## Dependencies

| Package | Purpose |
|---------|---------|
| `go-vgo/robotgo` | Mouse/keyboard control |
| `micmonay/keybd_event` | Keyboard simulation (Linux/Windows) |
| `sirupsen/logrus` | Logging |
| `spf13/cobra` | CLI framework |
| `stretchr/testify` | Testing |
