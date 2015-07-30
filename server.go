package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	gorillacontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/juju/errgo"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
)

const (
	DefaultCloseListenerDelay = 0
	DefaultOsExitDelay        = 5
	DefaultOsExitCode         = 0

  RequestHeader = "X-Request-ID"
	ScopeKey = 0
)

// Middleware is a http handler method.
type Middleware func(ctx context.Context, res http.ResponseWriter, req *http.Request) error

// Scope is a map getting through all middlewares.
type Scope struct {
	// Contains all placeholders from the route.
	MuxVars map[string]string

	// Helper to quickly write results to the `http.ResponseWriter`.
	Response Response

	// A middleware should call Next() to signal that no problem was encountered and
	// the next middleware in the chain can be executed after this middleware finished.
	// Always returns `nil`, so it can be convieniently used with return to quit the middleware.
	Next func() error

	Logger *logging.Logger
}

type Server struct {
	// The address to listen on.
	addr                string
	accessLogger        *logging.Logger
	statusLogger        *logging.Logger
	listener            net.Listener
	extendAccessLogging bool

	preHTTPHandler  AccessReporter
	postHTTPHandler AccessReporter

	alreadyRegisteredRoutes bool

	Router *mux.Router

	signalCounter      uint32
	closeListenerDelay time.Duration
	osExitDelay        time.Duration
	osExitCode         int

	Uuid func() string
}

func NewServer(host, port string) *Server {
	// We want to apply route names and need the context to be kept.
	router := mux.NewRouter()
	router.KeepContext = true

	s := &Server{
		addr:      host + ":" + port,
		Router:    router,
		Uuid: NewRequestIDFactory(),
	}

  s.SetLogger(MustGetLogger(LoggerOptions{ID: s.Uuid()}))
	s.SetCloseListenerDelay(DefaultCloseListenerDelay)
	s.SetOsExitDelay(DefaultOsExitDelay)
	s.SetOsExitCode(DefaultOsExitCode)

	return s
}

func (s *Server) Serve(method, urlPath string, middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one Middleware-Handler. Aborting...")
	}
	handler := s.NewMiddlewareHandler(middlewares)

	s.Router.Methods(method).Path(urlPath).Handler(handler).Name(method + " " + urlPath)
}

// ServeStatis registers a middleware that serves files from the filesystem.
// Example: s.ServeStatic("/v1/public", "./public_html/v1/")
func (s *Server) ServeStatic(urlPath, fsPath string) {
	handler := http.StripPrefix(urlPath, http.FileServer(http.Dir(fsPath)))
	s.Router.Methods("GET").PathPrefix(urlPath).Handler(handler)
}

func (s *Server) ServeNotFound(middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one NotFound-Handler. Aborting...")
	}

	s.Router.NotFoundHandler = s.NewMiddlewareHandler(middlewares)
}

// ExtendAccessLogging turns on the usage of ExtendedAccessLogger
func (s *Server) ExtendAccessLogging() {
	s.extendAccessLogging = true
}

func (s *Server) RegisterRoutes(mux *http.ServeMux, prefix string) {
	if s.alreadyRegisteredRoutes {
		return
	}

	var handler http.Handler = s.Router

	if s.accessLogger != nil {
		reporter := DefaultAccessReporter(s.accessLogger)
		if s.extendAccessLogging {
			reporter = ExtendedAccessReporter(s.accessLogger)
		}
		handler = NewLogAccessHandler(
			reporter,
			s.preHTTPHandler,
			s.postHTTPHandler,
			handler,
		)
	}

	// Always cleanup gorilla context request variables
	handler = gorillacontext.ClearHandler(handler)

	// http.mux handlers need a trailing slash while gorilla's mux does not need one
	// because they have different matching algorithms.
	prefix = strings.TrimSuffix(prefix, "/")
	mux.Handle(prefix+"/", http.StripPrefix(prefix, handler))

	s.alreadyRegisteredRoutes = true
}

func (s *Server) Listen() {
	mux := http.NewServeMux()
	s.RegisterRoutes(mux, "/")

	var err error
	if s.listener, err = net.Listen("tcp", s.addr); err != nil {
		panic(err)
	}

	go func() {
		s.statusLogger.Info("starting server on " + s.addr)
		if err := http.Serve(s.listener, mux); err != nil {
			if _, ok := err.(*net.OpError); ok {
				// We ignore the error "use of closed network connection", because it is
				// caused by us when shutting down the server.
			} else {
				s.statusLogger.Error("%#v", errgo.Mask(err))
			}
		}
	}()

	s.listenSignals()
}

func (s *Server) listenSignals() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// Block until a signal is received.
	for {
		select {
		case sig := <-c:
			s.statusLogger.Info("server received signal %s", sig)
			go s.Close()
		}
	}
}

func (s *Server) Close() {
	// Interrupt the process when closing is requested twice.
	if atomic.AddUint32(&s.signalCounter, 1) >= 2 {
		s.ExitProcess()
	}

	s.statusLogger.Info("closing tcp listener in %s", s.closeListenerDelay.String())
	time.Sleep(s.closeListenerDelay)
	s.listener.Close()

	s.statusLogger.Info("shutting down server in %s", s.osExitDelay.String())
	time.Sleep(s.osExitDelay)

	s.ExitProcess()
}

func (s *Server) ExitProcess() {
	s.statusLogger.Info("shutting down server with exit code %d", s.osExitCode)
	os.Exit(s.osExitCode)
}

// NewMiddlewareHandler wraps the middlewares in a http.Handler. The handler,
// on activation, calls each middleware in order, if no error was returned and
// `scope.Next()` was called. If a middleware wants to finish the processing, it
// can just write to the `http.ResponseWriter` or use the `scope.Responder` for
// convienience.
func (s *Server) NewMiddlewareHandler(middlewares []Middleware) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		reqID := req.Header.Get(RequestHeader)
		if reqID == "" {
			reqID = s.Uuid()
		}

		logger := MustGetLogger(LoggerOptions{ID: reqID})
    s.SetLogger(logger)

		// Initialize fresh scope variables.
		scope := Scope{
			MuxVars: mux.Vars(req),
			Response: Response{
				w: res,
			},
			Logger: logger,
		}

		ctx := context.Background()

		for _, middleware := range middlewares {
			nextCalled := false
			scope.Next = func() error {
				nextCalled = true
				return nil
			}
			ctx = context.WithValue(ctx, ScopeKey, scope)

			// End the request with an error and stop calling further middlewares.
			if err := middleware(ctx, res, req); err != nil {
				s.statusLogger.Error("%s %s %#v", req.Method, req.URL, errgo.Mask(err))

				scope.Response.Error(err.Error(), http.StatusInternalServerError)
				return
			}

			if !nextCalled {
				break
			}
		}
	})
}
