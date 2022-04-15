package runner

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

type intervalRunner struct {
	interval time.Duration
	action   Action
}

// Action to execute after each interval tick
type Action interface {
	Execute() error
}

func NewIntervalRunner(action Action, interval time.Duration) *intervalRunner {
	return &intervalRunner{
		interval: interval,
		action:   action,
	}
}

// RunInterval will start ticking and execute provided Action.
func (aRunner *intervalRunner) RunInterval(ctx context.Context) chan bool {
	finished := make(chan bool)
	errors := make(chan error)

	go listenForErrors(ctx, errors)

	go func(ctx context.Context) {
		defer func() {
			finished <- true
		}()

		for {
			select {
			case <-time.After(aRunner.interval):
				if err := aRunner.action.Execute(); err != nil {
					errors <- err
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return finished
}

func listenForErrors(ctx context.Context, errors chan error) {
	defer func() {
		logrus.Infof("errors listener stopped")
	}()

	for {
		select {
		case err := <-errors:
			logrus.Error(err)
		case <-ctx.Done():
			return
		}
	}
}
