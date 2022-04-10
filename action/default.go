//go:build linux || windows

package action

import (
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

type Default struct {
	bonding keybd_event.KeyBonding
}

func NewDefault() *Default {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		logrus.Fatal("action init failed: ", err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	kb.SetKeys(keybd_event.VK_CAPSLOCK)

	return &Default{
		bonding: kb,
	}
}

func (keybd *Default) Execute() error {
	return keybd.bonding.Launching()
}
