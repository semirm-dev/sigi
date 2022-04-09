package listener

import (
	"context"
	"time"
)

const (
	defaultInterval = 120 * time.Second
)

type actionListener struct {
	Interval time.Duration

	action Action
}

type Action interface {
	Execute() error
}

func NewActionListener(action Action) *actionListener {
	return &actionListener{
		Interval: defaultInterval,
		action:   action,
	}
}

func (lsnr *actionListener) Listen(ctx context.Context) (chan bool, chan error) {
	finished := make(chan bool)
	errors := make(chan error)

	go func(ctx context.Context) {
		defer func() {
			finished <- true
		}()

		for {
			select {
			case <-time.After(lsnr.Interval):
				if err := lsnr.action.Execute(); err != nil {
					errors <- err
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return finished, errors
}
