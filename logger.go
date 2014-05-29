package server

import (
	"os"
	"strconv"

	"github.com/juju/errgo"

	log "github.com/op/go-logging"
	Stdlog "log"
)

// NewSimpleLogger creates a new logger with a default backend logging to `os.Stderr`.
func NewSimpleLogger(name string, level ...string) *log.Logger {
	// Create new logger.
	logger := log.MustGetLogger(name)

	// Configure logger.
	log.SetFormatter(log.MustStringFormatter("[%{level}] %{message}"))
	logBackend := log.NewLogBackend(os.Stderr, "", Stdlog.LstdFlags|Stdlog.Lshortfile)
	logBackend.Color = true
	log.SetBackend(logBackend)

	// Check for valid log level.
	if len(level) > 1 {
		panic(errgo.Newf("Invalid log level given. Expecting exactly 1 level, got " + strconv.Itoa(len(level)) + ". Aborting..."))
	}

	// Set log level.
	if len(level) == 1 {
		logLevel, err := log.LogLevel(level[0])
		if err != nil {
			panic(errgo.Mask(err))
		}

		log.SetLevel(logLevel, name)
	}

	return logger
}
