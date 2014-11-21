package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/op/go-logging"
)

// Code heavily inspired by https://github.com/streadway/handy/blob/master/report/

type AccessEntry struct {
	RouteName     string
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

// Flush proxies http.Flusher's functionality if it is available on ResponseWriter
func (e *accessEntryWriter) Flush() {
	if f, ok := e.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// CloseNotify proxies http.CloseNotifier functionality
func (e *accessEntryWriter) CloseNotify() <-chan bool {
	cn := e.ResponseWriter.(http.CloseNotifier)
	return cn.CloseNotify()
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
func NewLogAccessHandler(reporter, preHTTP, postHTTP AccessReporter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		entry := AccessEntry{
			RequestMethod: req.Method,
			RequestURI:    req.RequestURI,

			Request:    req,
			StatusCode: 200,
		}
		start := time.Now()

		if preHTTP != nil {
			preHTTP(&entry)
		}

		next.ServeHTTP(&accessEntryWriter{response, &entry}, req)

		// Note, fetching a routes name needs to be done AFTER the routers handler
		// is executed. Otherwise the correct mux context is not given.
		route := mux.CurrentRoute(req)
		if route != nil {
			entry.RouteName = route.GetName()
		}

		if entry.RouteName == "" {
			entry.RouteName = "route-not-found"
		}

		entry.Duration = time.Since(start)

		if postHTTP != nil {
			postHTTP(&entry)
		}

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
