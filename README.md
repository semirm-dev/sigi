![Go](https://img.shields.io/github/go-mod/go-version/semirm-dev/sigi)

# sigi

Stay awake, always appear online on messaging apps, avoid corporate profile blocking your "sleep" options and forced screensaver.

Minimalistic keep-alive utility for macOS, Linux, and Windows.

## How It Works

sigi runs a configurable action at regular intervals to prevent the system from going idle:

- **macOS** — moves the mouse 1px right then 1px left (imperceptible, no side effects)
- **Linux / Windows** — toggles the CAPSLOCK key

## Installation

```shell
go install github.com/semirm-dev/sigi/cmd@latest
```

Or build from source:

```shell
git clone https://github.com/semirm-dev/sigi.git
cd sigi
make build
# binary output: build/sigi-darwin-arm64
```

## Usage

```shell
# run with default 120s interval
sigi

# run with 30s interval
sigi --interval 30
sigi -i 30

# enable action logging
sigi --logs
sigi -l

# combine flags
sigi -i 10 -l
```

Or run directly from source:

```shell
go run cmd/main.go -i 30 -l
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--interval` | `-i` | `120` | Interval in seconds between actions |
| `--logs` | `-l` | `false` | Log each action execution |

Stop with `Ctrl+C` — sigi handles SIGINT/SIGTERM gracefully.

## Project Structure

```
sigi/
├── cmd/main.go                    # entry point
├── sigi.go                        # CLI setup (Cobra), flags, signal handling
├── action/
│   ├── mouse.go                   # MouseMove action (all platforms)
│   └── keyboardbutton.go          # KeyboardButton action (linux/windows only)
├── runner/
│   ├── intervalrunner.go          # IntervalRunner + Action interface
│   └── intervalrunner_test.go     # tests
└── Makefile
```

## Development

```shell
make run          # go run cmd/main.go
make test         # go test -v ./...
make test-cover   # tests + HTML coverage report
make build        # build for darwin/arm64
```
