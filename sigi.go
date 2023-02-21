package sigi

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
	showLogs  bool
)

func init() {
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 120, "interval in seconds")
	rootCmd.Flags().IntVarP(&stopAfter, "stop", "s", 0, "stop after given minutes")
	rootCmd.Flags().BoolVarP(&showLogs, "logs", "l", false, "log each action")
}

var rootCmd = &cobra.Command{
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
		iRunner.RunInterval(runnerCtx, showLogs)

		logrus.Infof("sigi running...")

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		runnerCancel()

		logrus.Info("sigi stopped")
	},
}

// Execute will trigger root command.
func Execute() error {
	return rootCmd.Execute()
}
