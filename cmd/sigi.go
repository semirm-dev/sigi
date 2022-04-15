package cmd

import (
	"context"
	"github.com/semirm-dev/sigi/action"
	"github.com/semirm-dev/sigi/runner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	interval  int
	stopAfter int
)

func init() {
	sigi.Flags().IntVarP(&interval, "interval", "i", 120, "interval in seconds")
	sigi.Flags().IntVarP(&stopAfter, "stop", "s", 0, "stop after given minutes")
}

var sigi = &cobra.Command{
	Use:   "",
	Short: "Keep alive :)",
	Long:  `Keep alive :)`,
	Run: func(cmd *cobra.Command, args []string) {
		var runnerCtx context.Context
		var runnerCancel context.CancelFunc

		if stopAfter > 0 {
			timeout := time.Duration(stopAfter) * time.Minute
			runnerCtx, runnerCancel = context.WithTimeout(context.Background(), timeout)
		} else {
			runnerCtx, runnerCancel = context.WithCancel(context.Background())
		}

		iRunner := runner.NewIntervalRunner(action.NewMouseMove(), time.Duration(interval)*time.Second)
		finished := iRunner.RunInterval(runnerCtx)

		go listenForShutdown(runnerCancel)

		logrus.Infof("sigi running...")

		<-finished

		logrus.Info("sigi stopped")
	},
}

// Execute will trigger root command.
func Execute() error {
	return sigi.Execute()
}

func listenForShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
}
