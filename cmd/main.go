package main

import (
	"github.com/semirm-dev/sigi"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := sigi.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
