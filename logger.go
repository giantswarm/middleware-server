package server

import (
	"encoding/json"
	"os"

	"github.com/juju/errgo"
	"github.com/op/go-logging"
)

type LoggerOptions struct {
	ID      string
	Level   string
	NoColor bool
	Meta    interface{}
}

type logFormat struct {
	ID      string      `json:"id"`
	Time    string      `json:"time"`
	Level   string      `json:"level"`
	File    string      `json:"file"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta"`
}

// NewSimpleLogger creates a new logger with a default backend logging to `os.Stderr`.
func MustGetLogger(options LoggerOptions) *logging.Logger {
	// Create new logger. The logger ID is used to identify the logger using the
	// "id" field in the log format.
	logger := logging.MustGetLogger(options.ID)

	// See https://godoc.org/github.com/op/go-logging#NewStringFormatter for format verbs.
	format := logFormat{
		ID:      "%{module}",
		Time:    "%{time}",
		Level:   "%{level}",
		File:    "%{longfile}",
		Message: "%{message}",
		Meta:    options.Meta,
	}

	// Configure logger.
	rawFormat, err := json.Marshal(format)
	if err != nil {
		panic(errgo.Mask(err))
	}

	logging.SetFormatter(logging.MustStringFormatter(string(rawFormat)))
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackend.Color = !options.NoColor
	logging.SetBackend(logBackend)

	// Set log level.
	if options.Level != "" {
		logLevel, err := logging.LogLevel(options.Level)
		if err != nil {
			panic(errgo.Mask(err))
		}

		logging.SetLevel(logLevel, options.ID)
	}

	return logger
}
