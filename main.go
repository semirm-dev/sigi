package main

import (
	"flag"
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	done := make(chan bool)

	interval := flag.Int("interval", 60, "interval in seconds")
	useLogging := flag.Bool("logs", false, "use logging")
	flag.Parse()

	i := time.Second * time.Duration(*interval)

	logrus.Infof("sigi will ping every %v", i)

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		logrus.Fatal("keyboard init failed: ", err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	key := keybd_event.VK_CONNECT
	kb.SetKeys(key)

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
