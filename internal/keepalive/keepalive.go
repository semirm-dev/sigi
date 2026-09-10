// Package keepalive stops a machine going idle by doing something small on a
// timer -- nudging the mouse, or tapping a key that changes nothing.
//
// The Runner does the timing and nothing else. What to do on each tick is an
// Action, and reporting what happened belongs to the command, so the loop can
// run without a terminal attached.
package keepalive

import (
	"context"
	"fmt"
	"io"
	"time"
)

// DefaultInterval is how long ghu waits between actions when nothing says
// otherwise. Idle timers are usually minutes, so two is short enough to beat
// them and long enough to be invisible.
const DefaultInterval = 2 * time.Minute

// Action is the thing a Runner does on each tick.
//
// It is an interface because there are two of them -- the mouse and the
// keyboard -- and a test needs a third that touches neither.
type Action interface {
	Execute() error
	// Name is what the action calls itself, for --verbose to report.
	Name() string
}

// Runner performs an Action on an interval until its context is cancelled.
type Runner struct {
	action   Action
	interval time.Duration
}

// reporting wraps an Action so each tick is announced. It is how --verbose
// works without the Runner knowing anything about output.
type reporting struct {
	Action
	out io.Writer
}

// NewRunner builds a Runner. A non-positive interval falls back to
// DefaultInterval rather than spinning.
func NewRunner(action Action, interval time.Duration) *Runner {
	if interval <= 0 {
		interval = DefaultInterval
	}
	return &Runner{action: action, interval: interval}
}

// Run ticks until ctx is cancelled, and returns nil when it is.
//
// It stops on the first failed Action rather than logging and carrying on: a
// keep-alive whose action does not work is not keeping anything alive, and
// saying so beats a process that looks healthy for hours.
func (r *Runner) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.action.Execute(); err != nil {
				return fmt.Errorf("%s: %w", r.action.Name(), err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (r reporting) Execute() error {
	err := r.Action.Execute()
	if err == nil {
		fmt.Fprintf(r.out, "%s\n", r.Name())
	}
	return err
}
