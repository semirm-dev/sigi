package runner

import (
	"context"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

// IntervalRunner will execute actions in configured inervals
type IntervalRunner struct {
	interval time.Duration
	action   Action
}

// Action to execute after each interval tick
type Action interface {
	Execute() error
}

func NewIntervalRunner(action Action, interval time.Duration) *IntervalRunner {
	return &IntervalRunner{
		interval: interval,
		action:   action,
	}
}

// RunInterval will start ticking and execute configured Action.
func (iRunner *IntervalRunner) RunInterval(ctx context.Context, showLogs bool) {
	errors := make(chan error)

	go listenForErrors(ctx, errors)

	for {
		select {
		case <-time.After(iRunner.interval):
			if err := iRunner.action.Execute(); err != nil {
				errors <- err
			}

			if showLogs {
				logrus.Info("action executed: ", reflect.TypeOf(iRunner.action))
			}
		case <-ctx.Done():
			return
		}
	}
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
