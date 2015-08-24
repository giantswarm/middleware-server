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

	"github.com/giantswarm/request-context"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/juju/errgo"
)

const (
	DefaultCloseListenerDelay = 0
	DefaultOsExitDelay        = 5
	DefaultOsExitCode         = 0

	RequestIDKey    = "request-id"
	RequestIDHeader = "X-Request-ID"
)

type CtxConstructor func() interface{}

// Middleware is a http handler method.
type Middleware func(res http.ResponseWriter, req *http.Request, ctx *Context) error

// Context is a map getting through all middlewares.
type Context struct {
	// Contains all placeholders from the route.
	MuxVars map[string]string

	// Helper to quickly write results to the `http.ResponseWriter`.
	Response Response

	// A middleware should call Next() to signal that no problem was encountered
	// and the next middleware in the chain can be executed after this middleware
	// finished. Always returns `nil`, so it can be convieniently used with
	// return to quit the middleware.
	Next func() error

	// The app context for this request. Gets prefilled by the
	// CtxConstructor, if set in the server.
	App     interface{}
	Request requestcontext.Ctx

	RequestID    func() string
	SetRequestID func(ID string)
}

type Server struct {
	// The address to listen on.
	addr                string
	logLevel            string
	logColor            bool
	Logger              requestcontext.Logger
	listener            net.Listener
	extendAccessLogging bool

	preHTTPHandler  AccessReporter
	postHTTPHandler AccessReporter

	alreadyRegisteredRoutes bool

	Router *mux.Router

	ctxConstructor CtxConstructor

	signalCounter      uint32
	closeListenerDelay time.Duration
	osExitDelay        time.Duration
	osExitCode         int

	IDFactory func() string
}

func NewServer(host, port string) *Server {
	// We want to apply route names and need the context to be kept.
	router := mux.NewRouter()
	router.KeepContext = true

	s := &Server{
		addr:      host + ":" + port,
		Router:    router,
		IDFactory: NewIDFactory(),
		logColor:  true,
	}

	s.SetLogger(requestcontext.MustGetLogger(requestcontext.LoggerConfig{Name: "server", Color: s.logColor}))
	s.SetCloseListenerDelay(DefaultCloseListenerDelay)
	s.SetOsExitDelay(DefaultOsExitDelay)
	s.SetOsExitCode(DefaultOsExitCode)

	return s
}

func (s *Server) Serve(method, urlPath string, middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one Middleware-Handler.")
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

	// Always cleanup gorilla context request variables
	handler = context.ClearHandler(handler)

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
		if err := http.Serve(s.listener, mux); err != nil {
			if _, ok := err.(*net.OpError); ok {
				// We ignore the error "use of closed network connection", because it is
				// caused by us when shutting down the server.
			} else {
				s.Logger.Error(nil, "%#v", errgo.Mask(err))
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
			s.Logger.Info(nil, "server received signal %s", sig)
			go s.Close()
		}
	}
}

func (s *Server) Close() {
	// Interrupt the process when closing is requested twice.
	if atomic.AddUint32(&s.signalCounter, 1) >= 2 {
		s.ExitProcess()
	}

	s.Logger.Info(nil, "closing tcp listener in %s", s.closeListenerDelay.String())
	time.Sleep(s.closeListenerDelay)
	s.listener.Close()

	s.Logger.Info(nil, "shutting down server in %s", s.osExitDelay.String())
	time.Sleep(s.osExitDelay)

	s.ExitProcess()
}

func (s *Server) ExitProcess() {
	s.Logger.Info(nil, "shutting down server with exit code %d", s.osExitCode)
	os.Exit(s.osExitCode)
}

// NewMiddlewareHandler wraps the middlewares in a http.Handler. The handler,
// on activation, calls each middleware in order, if no error was returned and
// `ctx.Next()` was called. If a middleware wants to finish the processing, it
// can just write to the `http.ResponseWriter` or use the `ctx.Response` for
// convienience.
func (s *Server) NewMiddlewareHandler(middlewares []Middleware) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// prepare request
		requestID := req.Header.Get(RequestIDHeader)

		// TODO: This is just for backward compatibility. Currently clients are
		// sending both, client and request ID's. We just changed our concept and
		// pushed client changes too fast, so we need to support them for a while
		// to not be confused by our received data.
		clientID := req.Header.Get("X-Client-ID")
		if clientID != "" && requestID != "" {
			requestID = clientID
		}

		if requestID == "" {
			requestID = s.IDFactory()
		} else {
			requestID += ", " + s.IDFactory()
		}
		requestCtx := requestcontext.Ctx{
			RequestIDKey: requestID,
		}

		// create handler that actually processes the middlewares
		middlewareHandler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := &Context{
				MuxVars: mux.Vars(req),
				Request: requestCtx,
				Response: Response{
					w: res,
				},
				RequestID: func() string {
					idRaw := requestCtx[RequestIDKey]
					if id, ok := idRaw.(string); ok {
						return id
					} else {
						// ID not found in map
						return ""
					}
				},
				SetRequestID: func(ID string) {
					requestCtx[RequestIDKey] = ID
				},
			}

			if s.ctxConstructor != nil {
				ctx.App = s.ctxConstructor()
			}

			for _, middleware := range middlewares {
				nextCalled := false
				ctx.Next = func() error {
					nextCalled = true
					return nil
				}

				// End the request with an error and stop calling further middlewares.
				if err := middleware(res, req, ctx); err != nil {
					s.Logger.Error(requestCtx, "%s %s %#v", req.Method, req.URL, errgo.Mask(err))

					ctx.Response.Error(err.Error(), http.StatusInternalServerError)
					break
				}

				if !nextCalled {
					break
				}
			}
		})

		// do access-logging by wrapping the middleware handler
		reporter := DefaultAccessReporter(requestCtx, s.Logger)
		if s.extendAccessLogging {
			reporter = ExtendedAccessReporter(requestCtx, s.Logger)
		}

		handler := NewLogAccessHandler(
			reporter,
			s.preHTTPHandler,
			s.postHTTPHandler,
			middlewareHandler,
		).(http.HandlerFunc)

		// handle request
		handler.ServeHTTP(res, req)
	})
}
