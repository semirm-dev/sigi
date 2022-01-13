package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
)

func main() {
	done := make(chan bool)

	key := flag.Int("key", keybd_event.VK_CAPSLOCK, "key to trigger")
	interval := flag.Int("interval", 120, "interval in seconds")
	useLogging := flag.Bool("logs", false, "use logging")
	stopAfter := flag.Int("stop", 0, "stop after given minutes")
	flag.Parse()

	i := time.Second * time.Duration(*interval)
	startedAt := time.Now()
	stopAt := startedAt.Add(time.Minute * time.Duration(*stopAfter))

	logrus.Infof("zombie %v", i)

	if *stopAfter > 0 {
		logrus.Infof("stopAt: %v", stopAt.Format("2006-01-02 15:04:05"))
	}

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		logrus.Fatal("keyboard init failed: ", err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	kb.SetKeys(*key)

	go func() {
		for {
			select {
			case <-time.After(i):
				err = kb.Launching()
				if err != nil {
					logrus.Error("key press failed: ", err)
				}

				if *useLogging {
					t := time.Now()
					logrus.Infof("[%s] - key trigger: %v", t.Format("2006-01-02 15:04:05"), key)
				}

				if *stopAfter > 0 && time.Now().Sub(stopAt) > 1 {
					close(done)
					return
				}
			}
		}
	}()

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		close(done)
	}()

	logrus.Infof("sigi running...")

	<-done

	logrus.Info("sigi stopped")
}
