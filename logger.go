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
	Meta    map[string]interface{}
}

type logFormat struct {
	ID      string                 `json:"id"`
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	File    string                 `json:"file"`
	Message string                 `json:"message"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// NewSimpleLogger creates a new logger with a default backend logging to `os.Stderr`.
func MustGetLogger(options LoggerOptions) *logging.Logger {
	// Create new logger. The logger ID is used to identify the logger using the
	// "id" field in the log format.
	logger := logging.MustGetLogger(options.ID)

	// Configure logger.
	rawMeta, err := json.Marshal(options.Meta)
	if err != nil {
		panic(errgo.Mask(err))
	}
	// See https://godoc.org/github.com/op/go-logging#NewStringFormatter for format verbs.
	format := "%{module} | %{time:2006-01-02 15:04:05} | %{level} | %{longfile} | %{message} | " + string(rawMeta)

	formatter := logging.MustStringFormatter(format)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backend.Color = !options.NoColor
	backendFormatter := logging.NewBackendFormatter(backend, formatter)
	leveledBackend := logging.AddModuleLevel(backendFormatter)

	// Set log level.
	if options.Level != "" {
		logLevel, err := logging.LogLevel(options.Level)
		if err != nil {
			panic(errgo.Mask(err))
		}

		leveledBackend.SetLevel(logLevel, options.ID)
	}

	logger.SetBackend(leveledBackend)

	return logger
}
