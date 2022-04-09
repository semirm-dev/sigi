package main

import (
	"context"
	"flag"
	"github.com/semirm-dev/sigi/keyboard"
	"github.com/semirm-dev/sigi/listener"
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

	var listenerCtx context.Context
	var listenerCancel context.CancelFunc

	if *stopAfter > 0 {
		timeout := time.Duration(*stopAfter) * time.Minute
		listenerCtx, listenerCancel = context.WithTimeout(context.Background(), timeout)
	} else {
		listenerCtx, listenerCancel = context.WithCancel(context.Background())
	}

	actionListener := listener.NewActionListener(keyboard.NewDefault())
	actionListener.Interval = time.Duration(*interval) * time.Second
	finished, errors := actionListener.Listen(listenerCtx)

	go func() {
		defer func() {
			close(errors)
			logrus.Infof("errors listener stopped")
		}()

		for {
			select {
			case err := <-errors:
				logrus.Error(err)
			case <-listenerCtx.Done():
				return
			}
		}
	}()

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		listenerCancel()
	}()

	logrus.Infof("sigi running...")

	<-finished
	close(finished)

	logrus.Info("sigi stopped")
}
