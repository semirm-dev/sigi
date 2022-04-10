package action

import (
	"github.com/go-vgo/robotgo"
)

var defaultMouseMovement = 1

type MouseMove struct {
	initialPosition bool
}

func NewMouseMove() *MouseMove {
	return &MouseMove{
		initialPosition: true,
	}
}

func (mouse *MouseMove) Execute() error {
	if !mouse.initialPosition {
		mouse.moveLeft()

		mouse.initialPosition = true

		return nil
	}

	mouse.moveRight()

	mouse.initialPosition = false

	return nil
}

func (mouse *MouseMove) moveLeft() {
	x := defaultMouseMovement * -1
	robotgo.MoveRelative(x, 0)
}

func (mouse *MouseMove) moveRight() {
	x := defaultMouseMovement
	robotgo.MoveRelative(x, 0)
}
