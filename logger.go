package server

import (
	"os"

	"github.com/juju/errgo"

	log "github.com/op/go-logging"
	Stdlog "log"
)

type LoggerOptions struct {
	Name  string
	Level string
}

// NewSimpleLogger creates a new logger with a default backend logging to `os.Stderr`.
func NewLogger(options LoggerOptions) *log.Logger {
	// Create new logger.
	logger := log.MustGetLogger(options.Name)

	// Configure logger.
	log.SetFormatter(log.MustStringFormatter("[%{level}] %{message}"))
	logBackend := log.NewLogBackend(os.Stderr, "", Stdlog.LstdFlags|Stdlog.Lshortfile)
	logBackend.Color = true
	log.SetBackend(logBackend)

	// Set log level.
	if options.Level != "" {
		logLevel, err := log.LogLevel(options.Level)
		if err != nil {
			panic(errgo.Mask(err))
		}

		log.SetLevel(logLevel, options.Name)
	}

	return logger
}
