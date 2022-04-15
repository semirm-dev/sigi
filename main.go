package main

import (
	"flag"
	"github.com/sirupsen/logrus"
)

var (
	interval  = flag.Int("interval", 120, "interval in seconds")
	stopAfter = flag.Int("stop", 0, "stop after given minutes")
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
