package runner

import (
	"context"
	"time"
)

const (
	defaultInterval = 120 * time.Second
)

type actionRunner struct {
	Interval time.Duration

	action Action
}

type Action interface {
	Execute() error
}

func NewActionRunner(action Action) *actionRunner {
	return &actionRunner{
		Interval: defaultInterval,
		action:   action,
	}
}

func (aRunner *actionRunner) RunInterval(ctx context.Context) (chan bool, chan error) {
	finished := make(chan bool)
	errors := make(chan error)

	go func(ctx context.Context) {
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
	}(ctx)

	return finished, errors
}
