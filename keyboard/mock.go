package keyboard

import "github.com/sirupsen/logrus"

type Mock struct {
}

func NewMocked() *Mock {
	return &Mock{}
}

func (mock *Mock) Execute() error {
	logrus.Info("mocked action executed...")
	return nil
}
