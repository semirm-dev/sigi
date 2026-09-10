package keepalive

// The mouse action: a one-pixel nudge right, then back.

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
)

// nudge is how far the cursor moves, in pixels. One is enough for the system
// to count it as activity, and small enough that nobody notices.
const nudge = 1

// settle is the pause between the two halves of the nudge. Without it the pair
// can be coalesced into no movement at all.
const settle = 10 * time.Millisecond

// Mouse moves the cursor a pixel and puts it back, which every platform counts
// as input without anything moving as far as the person watching is concerned.
type Mouse struct{}

// NewMouse refuses where the nudge would do nothing.
//
// On Linux the underlying library talks to X11 through XTest. Under Wayland
// that call succeeds and moves no cursor, and with no display at all there is
// nothing to move -- either way sigi would tick happily for hours while the
// screen locked behind it, which is the one outcome worth preventing.
func NewMouse() (*Mouse, error) {
	if reason := noPointer(); reason != "" {
		return nil, fmt.Errorf("the mouse action needs X11 and %s, so it would move nothing: use --action keyboard", reason)
	}
	return &Mouse{}, nil
}

func (m *Mouse) Name() string { return "mouse" }

func (m *Mouse) Execute() error {
	robotgo.MoveRelative(nudge, 0)
	time.Sleep(settle)
	robotgo.MoveRelative(-nudge, 0)

	return nil
}

// noPointer describes why this machine has no pointer the nudge could reach,
// or is empty when it has one. Only Linux is examined: macOS and Windows always have one,
// and their sessions do not advertise themselves in the environment.
func noPointer() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	if os.Getenv("WAYLAND_DISPLAY") != "" ||
		strings.EqualFold(os.Getenv("XDG_SESSION_TYPE"), "wayland") {
		return "this is a Wayland session"
	}
	if os.Getenv("DISPLAY") == "" {
		return "there is no display here"
	}
	return ""
}
