package server

import (
	"github.com/dchest/uniuri"
)

const (
	requestIDFactoryLen   = 16
	requestIDFactoryChars = `abcdefghijklmnopqrstuvwxyz0123456789`
)

func NewIDFactory() func() string {
	return func() string {
		return uniuri.NewLenChars(requestIDFactoryLen, []byte(requestIDFactoryChars))
	}
}
