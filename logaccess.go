package server

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/giantswarm/request-context"
	"github.com/gorilla/mux"
)

// Code heavily inspired by https://github.com/streadway/handy/blob/master/report/

type AccessEntry struct {
	routeName     string
	requestMethod string
	requestURI    string
	request       *http.Request

	duration   time.Duration
	statusCode int
	size       int64
}

func (ae *AccessEntry) RouteName() string {
	return ae.routeName
}

func (ae *AccessEntry) RequestMethod() string {
	return ae.requestMethod
}

func (ae *AccessEntry) RequestURI() string {
	return ae.requestURI
}

func (ae *AccessEntry) Request() *http.Request {
	return ae.request
}

func (ae *AccessEntry) Duration() time.Duration {
	return ae.duration
}

func (ae *AccessEntry) StatusCode() int {
	return ae.statusCode
}

func (ae *AccessEntry) Size() int64 {
	return ae.size
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
	e.entry.size += int64(n)
	return n, err
}

// WriteHeader captures the status code and writes through to the wrapper ResponseWriter.
func (e *accessEntryWriter) WriteHeader(code int) {
	e.entry.statusCode = code
	e.ResponseWriter.WriteHeader(code)
}

// Hijack lets the caller take over the connection.
// After a call to Hijack(), the HTTP server library
// will not do anything else with the connection.
// It becomes the caller's responsibility to manage
// and close the connection.
func (e *accessEntryWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := e.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

// NewLogAccessHandler executes the next handler and logs the requests statistics afterwards to the logger.
func NewLogAccessHandler(reporter, preHTTP, postHTTP AccessReporter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		entry := AccessEntry{
			requestMethod: req.Method,
			requestURI:    req.RequestURI,

			request:    req,
			statusCode: 200,
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
			entry.routeName = route.GetName()
		}

		if entry.routeName == "" {
			entry.routeName = req.Method + " route-not-found"
		}

		entry.duration = time.Since(start)

		if postHTTP != nil {
			postHTTP(&entry)
		}

		reporter(&entry)
	})
}

type AccessReporter func(entry *AccessEntry)

func DefaultAccessReporter(ctx requestcontext.Ctx, logger requestcontext.Logger) AccessReporter {
	return func(entry *AccessEntry) {
		milliseconds := int(entry.duration / time.Millisecond)
		logger.Info(ctx, "%s %s %d %d %d", entry.requestMethod, entry.requestURI, entry.statusCode, entry.size, milliseconds)
	}
}

// ExtendedAccessReporter createsan access logger that logs everything that DefaultAccessReporter does with the User-Agent added to that
func ExtendedAccessReporter(ctx requestcontext.Ctx, logger requestcontext.Logger) AccessReporter {
	return func(entry *AccessEntry) {
		milliseconds := int(entry.duration / time.Millisecond)
		logger.Info(ctx, "%s %s %d %d %d %s", entry.requestMethod, entry.requestURI, entry.statusCode, entry.size, milliseconds, entry.Request().Header.Get("User-Agent"))
	}
}
