package action

import "github.com/sirupsen/logrus"

// Mock action will only log into console.
type Mock struct{}

func NewMocked() *Mock {
	return &Mock{}
}

func (mock *Mock) Execute() error {
	logrus.Info("mocked action executed...")
	return nil
}
