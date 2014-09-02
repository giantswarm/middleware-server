package server

import (
	log "github.com/op/go-logging"
	"net/http"
	"time"
)

// Code heavily inspired by https://github.com/streadway/handy/blob/master/report/

type AccessEntry struct {
	RequestMethod string
	RequestURI    string
	Request       *http.Request

	Duration   time.Duration
	StatusCode int
	Size       int64
}

type accessEntryWriter struct {
	http.ResponseWriter
	entry *AccessEntry
}

// Write sums the writes to produce the actual number of bytes written
func (e *accessEntryWriter) Write(b []byte) (int, error) {
	n, err := e.ResponseWriter.Write(b)
	e.entry.Size += int64(n)
	return n, err
}

// WriteHeader captures the status code and writes through to the wrapper ResponseWriter.
func (e *accessEntryWriter) WriteHeader(code int) {
	e.entry.StatusCode = code
	e.ResponseWriter.WriteHeader(code)
}

// NewLogAccessHandler executes the next handler and logs the requests statistics afterwards to the logger.
func NewLogAccessHandler(reporter AccessReporter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		entry := AccessEntry{
			RequestMethod: req.Method,
			RequestURI:    req.RequestURI,

			Request:    req,
			StatusCode: 200,
		}
		start := time.Now()

		next.ServeHTTP(&accessEntryWriter{response, &entry}, req)

		entry.Duration = time.Since(start)

		reporter(&entry)
	})
}

type AccessReporter func(entry *AccessEntry)

func DefaultAccessReporter(logger *log.Logger) AccessReporter {
	return func(entry *AccessEntry) {
		milliseconds := int(entry.Duration / time.Millisecond)
		logger.Info("%s %s %d %d %d", entry.RequestMethod, entry.RequestURI, entry.StatusCode, entry.Size, milliseconds)
	}
}
