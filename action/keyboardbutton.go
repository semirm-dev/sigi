//go:build linux || windows

package action

import (
	"github.com/micmonay/keybd_event"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

// KeyboardButton action will execute keyboard button.
// VK_CAPSLOCK current implementation.
type KeyboardButton struct {
	bonding keybd_event.KeyBonding
}

func NewKeyboardButton() *KeyboardButton {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		logrus.Fatal("action init failed: ", err)
	}

	// for linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	kb.SetKeys(keybd_event.VK_CAPSLOCK)

	return &KeyboardButton{
		bonding: kb,
	}
}

func (keybd *KeyboardButton) Execute() error {
	return keybd.bonding.Launching()
}
