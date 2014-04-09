package server

import (
	Stdlog "log"
	"os"

	log "github.com/op/go-logging"
)

// NewSimpleLogger creates a new logger with a default backend logging to `os.Stderr`.
func NewSimpleLogger(name string) *log.Logger {
	logger := log.MustGetLogger(name)
	log.SetFormatter(log.MustStringFormatter("[%{level}] %{message}"))
	logBackend := log.NewLogBackend(os.Stderr, "", Stdlog.LstdFlags|Stdlog.Lshortfile)
	logBackend.Color = true
	log.SetBackend(logBackend)

	return logger
}
