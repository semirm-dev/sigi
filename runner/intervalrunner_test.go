package runner_test

import (
	"context"
	"github.com/semirm-dev/sigi/runner"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockAction struct {
	triggerCount int
}

func (a *mockAction) Execute() error {
	a.triggerCount++
	return nil
}

func TestIntervalRunner_RunInterval(t *testing.T) {
	action := &mockAction{}

	iRunner := runner.NewIntervalRunner(action, time.Duration(1)*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	go iRunner.RunInterval(ctx, false)

	select {
	case <-time.After(5 * time.Millisecond):
		cancel()
	}

	assert.True(t, action.triggerCount > 0)
}
