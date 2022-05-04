package runner

import (
	"context"
	"github.com/sirupsen/logrus"
	"reflect"
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
func (iRunner *intervalRunner) RunInterval(ctx context.Context, log bool) chan bool {
	finished := make(chan bool)
	errors := make(chan error)

	go listenForErrors(ctx, errors)

	go func(ctx context.Context) {
		defer func() {
			close(finished)
		}()

		for {
			select {
			case <-time.After(iRunner.interval):
				if err := iRunner.action.Execute(); err != nil {
					errors <- err
				}

				if log {
					logrus.Info("action executed: ", reflect.TypeOf(iRunner.action))
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
