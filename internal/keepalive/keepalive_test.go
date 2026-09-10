package keepalive_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/semirm-dev/sigi/internal/keepalive"
)

// counting is the third Action, the one that touches no hardware. It is what
// justifies Action being an interface at all.
type counting struct {
	mu   sync.Mutex
	runs int
	err  error
}

// The loop must tick, and must stop when its context is cancelled -- which is
// what ctrl-c does. Cancelling and waiting for Run to return proves it, where
// asserting on a count after a sleep only proves the clock moved.
func TestRunTicksUntilCancelled(t *testing.T) {
	action := &counting{}
	runner := keepalive.NewRunner(action, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	// Wait for real ticks rather than for a duration.
	for action.count() == 0 {
		time.Sleep(time.Millisecond)
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil on cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after its context was cancelled")
	}

	if after := action.count(); after == 0 {
		t.Fatal("the action never ran")
	}
}

// A keep-alive whose action fails is not keeping anything alive, so Run stops
// and says which action broke rather than logging forever.
func TestRunStopsOnActionError(t *testing.T) {
	boom := errors.New("boom")
	runner := keepalive.NewRunner(&counting{err: boom}, time.Millisecond)

	err := runner.Run(context.Background())

	if !errors.Is(err, boom) {
		t.Fatalf("Run returned %v, want it to wrap %v", err, boom)
	}
	if got := err.Error(); got == boom.Error() {
		t.Errorf("Run returned the bare error %q; it should name the action", got)
	}
}

// An interval of zero would spin the CPU, so it falls back rather than trusting
// the caller.
func TestNonPositiveIntervalFallsBack(t *testing.T) {
	action := &counting{}
	runner := keepalive.NewRunner(action, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := runner.Run(ctx); err != nil {
		t.Fatalf("Run returned %v", err)
	}
	if action.count() != 0 {
		t.Errorf("action ran %d times in 20ms; the interval did not fall back to %s",
			action.count(), keepalive.DefaultInterval)
	}
}

// The mouse action must refuse where it would tick without moving anything:
// under Wayland, and with no display at all. Silently doing nothing for hours
// is the failure this whole package exists to avoid.
func TestMouseRefusesWhereItWouldDoNothing(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the check only examines Linux sessions")
	}

	for _, tc := range []struct {
		name string
		env  map[string]string
		want bool // want a refusal
	}{
		{"wayland via WAYLAND_DISPLAY", map[string]string{"WAYLAND_DISPLAY": "wayland-0", "DISPLAY": ":0"}, true},
		{"wayland via XDG_SESSION_TYPE", map[string]string{"XDG_SESSION_TYPE": "wayland", "DISPLAY": ":0"}, true},
		{"no display at all", map[string]string{}, true},
		{"an X11 session", map[string]string{"DISPLAY": ":0"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"WAYLAND_DISPLAY", "XDG_SESSION_TYPE", "DISPLAY"} {
				t.Setenv(key, tc.env[key])
			}

			_, err := keepalive.NewMouse()

			if tc.want && err == nil {
				t.Fatal("NewMouse accepted a session where the nudge would do nothing")
			}
			if !tc.want && err != nil {
				t.Fatalf("NewMouse refused an X11 session: %v", err)
			}
			if tc.want && !strings.Contains(err.Error(), "--action keyboard") {
				t.Errorf("the refusal does not say what to do instead: %v", err)
			}
		})
	}
}

func (c *counting) Name() string { return "counting" }

func (c *counting) Execute() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs++
	return c.err
}

func (c *counting) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.runs
}
