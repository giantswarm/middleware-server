package server

import (
	"github.com/dchest/uniuri"
)

const (
	requestIDFactoryLen   = 8
	requestIDFactoryChars = `abcdefghijklmnopqrstuvwxyz0123456789`
)

func NewRequestIDFactory() func() string {
	return func() string {
		return uniuri.NewLenChars(requestIDFactoryLen, []byte(requestIDFactoryChars))
	}
}
