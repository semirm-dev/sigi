package main

import (
	"context"
	"github.com/semirm-dev/sigi/keyboard"
	"github.com/semirm-dev/sigi/listener"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	listenerCtx, listenerCancel := context.WithTimeout(context.Background(), 3*time.Second)
	actionListener := listener.NewActionListener(keyboard.NewMocked())
	actionListener.Interval = 1 * time.Second
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
