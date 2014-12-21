package server

import (
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

// SetAppContext sets the CtxConstructor object, that is called for every
// request to provide the initial `Context.App` value, which is available to
// every middleware.
func (s *Server) SetAppContext(ctxConstructor CtxConstructor) {
	s.ctxConstructor = ctxConstructor
}
