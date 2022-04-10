package action

import (
	"github.com/go-vgo/robotgo"
	"time"
)

var mouseMovementIndex = 1

type MouseMove struct {
}

func NewMouseMove() *MouseMove {
	return &MouseMove{}
}

func (mouse *MouseMove) Execute() error {
	mouse.moveRight()
	time.Sleep(10 * time.Millisecond)
	mouse.moveLeft()

	return nil
}

func (mouse *MouseMove) moveRight() {
	x := mouseMovementIndex
	robotgo.MoveRelative(x, 0)
}

func (mouse *MouseMove) moveLeft() {
	x := mouseMovementIndex * -1
	robotgo.MoveRelative(x, 0)
}
