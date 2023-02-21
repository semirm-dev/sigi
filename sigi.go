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
	interval int
	showLogs bool
)

func init() {
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 120, "interval in seconds")
	rootCmd.Flags().BoolVarP(&showLogs, "logs", "l", false, "log each action")
}

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "Keep alive :)",
	Long:  `Keep alive :)`,
	Run: func(cmd *cobra.Command, args []string) {
		iRunner := runner.NewIntervalRunner(action.NewMouseMove(), time.Duration(interval)*time.Second)
		go iRunner.RunInterval(context.Background(), showLogs)

		logrus.Infof("sigi running...")

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logrus.Info("sigi stopped")
	},
}

// Execute will trigger root command.
func Execute() error {
	return rootCmd.Execute()
}
