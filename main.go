package main

import (
	"context"
	"flag"
	"github.com/semirm-dev/sigi/keyboard"
	"github.com/semirm-dev/sigi/runner"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	interval  = flag.Int("interval", 120, "interval in seconds")
	stopAfter = flag.Int("stop", 0, "stop after given minutes")
)

func main() {
	flag.Parse()

	var runnerCtx context.Context
	var runnerCancel context.CancelFunc

	if *stopAfter > 0 {
		timeout := time.Duration(*stopAfter) * time.Minute
		runnerCtx, runnerCancel = context.WithTimeout(context.Background(), timeout)
	} else {
		runnerCtx, runnerCancel = context.WithCancel(context.Background())
	}

	actionRunner := runner.NewActionRunner(keyboard.NewDefault())
	actionRunner.Interval = time.Duration(*interval) * time.Second
	finished, errors := actionRunner.RunInterval(runnerCtx)

	go func() {
		defer func() {
			close(errors)
			logrus.Infof("errors listener stopped")
		}()

		for {
			select {
			case err := <-errors:
				logrus.Error(err)
			case <-runnerCtx.Done():
				return
			}
		}
	}()

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		runnerCancel()
	}()

	logrus.Infof("sigi running...")

	<-finished
	close(finished)

	logrus.Info("sigi stopped")
}
