package server

import (
	"time"

	log "github.com/op/go-logging"
)

func (s *Server) SetPreHTTPHandler(reporter AccessReporter) {
	s.preHTTPHandler = reporter
}

func (s *Server) SetPostHTTPHandler(reporter AccessReporter) {
	s.postHTTPHandler = reporter
}

// SetLogger sets the logger object to which the server logs every request.
func (s *Server) SetLogger(logger *log.Logger) {
	s.SetAccessLogger(logger)
	s.SetStatusLogger(logger)
}

func (s *Server) SetAccessLogger(logger *log.Logger) {
	s.accessLogger = logger
}

func (s *Server) SetStatusLogger(logger *log.Logger) {
	s.statusLogger = logger
}

// SetCloseListenerDelay sets the time to delay closing the TCP listener when
// calling `s.Close()`.
func (s *Server) SetCloseListenerDelay(d int) {
	s.closeListenerDelay = time.Duration(d) * time.Second
}

// SetOsExitDelay sets the time to delay exiting the process when calling
// `s.Close()`.
func (s *Server) SetOsExitDelay(d int) {
	s.osExitDelay = time.Duration(d) * time.Second
}

// SetOsExitCode sets the exit code used in `os.Exit(c)` when calling
// `s.Close()`.
func (s *Server) SetOsExitCode(c int) {
	s.osExitCode = c
}
