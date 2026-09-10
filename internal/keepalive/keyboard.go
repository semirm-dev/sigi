package keepalive

// The keyboard action: toggling a key that changes nothing you can see.

import (
	"fmt"
	"runtime"
	"time"

	"github.com/micmonay/keybd_event"
)

// linuxSettle is the wait the keybd_event library requires on Linux before its
// first event will register. It is the library's own instruction, not a guess.
const linuxSettle = 2 * time.Second

// Keyboard taps caps lock, which registers as input and changes nothing that
// survives the second tap.
type Keyboard struct {
	bonding keybd_event.KeyBonding
}

// NewKeyboard prepares the key binding, and returns the error rather than
// ending the process: a constructor that calls Fatal cannot be used by
// anything that wants to recover, including a test.
func NewKeyboard() (*Keyboard, error) {
	bonding, err := keybd_event.NewKeyBonding()
	if err != nil {
		return nil, fmt.Errorf("preparing the keyboard: %w", err)
	}

	if runtime.GOOS == "linux" {
		time.Sleep(linuxSettle)
	}
	bonding.SetKeys(keybd_event.VK_CAPSLOCK)

	return &Keyboard{bonding: bonding}, nil
}

func (k *Keyboard) Name() string { return "keyboard" }

func (k *Keyboard) Execute() error { return k.bonding.Launching() }
