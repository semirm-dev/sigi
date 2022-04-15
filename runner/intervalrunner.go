package runner

import (
	"context"
	"time"
)

const (
	defaultInterval = 120 * time.Second
)

type intervalRunner struct {
	Interval time.Duration
	action   Action
}

type Action interface {
	Execute() error
}

func NewIntervalRunner(action Action) *intervalRunner {
	return &intervalRunner{
		Interval: defaultInterval,
		action:   action,
	}
}

func (aRunner *intervalRunner) RunInterval(ctx context.Context) (chan bool, chan error) {
	finished := make(chan bool)
	errors := make(chan error)

	go func(ctx context.Context, finished chan bool, errors chan error) {
		defer func() {
			finished <- true
		}()

		for {
			select {
			case <-time.After(aRunner.Interval):
				if err := aRunner.action.Execute(); err != nil {
					errors <- err
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, finished, errors)

	return finished, errors
}
