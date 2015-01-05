package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/juju/errgo"

	log "github.com/op/go-logging"
)

const (
	DefaultCloseListenerDelay = 0
	DefaultOsExitDelay        = 3
	DefaultOsExitCode         = 0
)

type CtxConstructor func() interface{}

// Middleware is a http handler method.
type Middleware func(http.ResponseWriter, *http.Request, *Context) error

// Context is a map getting through all middlewares.
type Context struct {
	// Contains all placeholders from the route.
	MuxVars map[string]string

	// Helper to quickly write results to the `http.ResponseWriter`.
	Response Response

	// A middleware should call Next() to signal that no problem was encountered and
	// the next middleware in the chain can be executed after this middleware finished.
	// Always returns `nil`, so it can be convieniently used with return to quit the middleware.
	Next func() error

	// The app context for this request. Gets prefilled by the CtxConstructor, if set in the server.
	App interface{}
}

type Server struct {
	// The address to listen on.
	addr         string
	accessLogger *log.Logger
	statusLogger *log.Logger
	listener     net.Listener

	preHTTPHandler  AccessReporter
	postHTTPHandler AccessReporter

	alreadyRegisteredRoutes bool

	Router *mux.Router

	ctxConstructor CtxConstructor

	signal             os.Signal
	signalCounter      int
	closeListenerDelay time.Duration
	osExitDelay        time.Duration
	osExitCode         int
}

func NewServer(host, port string) *Server {
	// We want to apply route names and need the context to be kept.
	router := mux.NewRouter()
	router.KeepContext = true

	s := &Server{
		addr:   host + ":" + port,
		Router: router,
	}

	s.SetCloseListenerDelay(DefaultCloseListenerDelay)
	s.SetOsExitDelay(DefaultOsExitDelay)
	s.SetOsExitCode(DefaultOsExitCode)

	return s
}

func (this *Server) Serve(method, urlPath string, middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one Middleware-Handler. Aborting...")
	}
	handler := this.NewMiddlewareHandler(middlewares)

	this.Router.Methods(method).Path(urlPath).Handler(handler).Name(method + " " + urlPath)
}

// ServeStatis registers a middleware that serves files from the filesystem.
// Example: this.ServeStatic("/v1/public", "./public_html/v1/")
func (this *Server) ServeStatic(urlPath, fsPath string) {
	handler := http.StripPrefix(urlPath, http.FileServer(http.Dir(fsPath)))
	this.Router.Methods("GET").PathPrefix(urlPath).Handler(handler)
}

func (this *Server) ServeNotFound(middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one NotFound-Handler. Aborting...")
	}

	this.Router.NotFoundHandler = this.NewMiddlewareHandler(middlewares)
}

func (s *Server) RegisterRoutes(mux *http.ServeMux, prefix string) {
	if s.alreadyRegisteredRoutes {
		return
	}

	var handler http.Handler = s.Router

	if s.accessLogger != nil {
		handler = NewLogAccessHandler(
			DefaultAccessReporter(s.accessLogger),
			s.preHTTPHandler,
			s.postHTTPHandler,
			handler,
		)
	}

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
		case s.signal = <-c:
			s.statusLogger.Info("server received signal %s", s.signal)
			go s.Close()
		}
	}
}

func (s *Server) Close() {
	s.signalCounter++

	// Interrupt the process when closing is requested twice.
	if s.signalCounter == 2 {
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
// `ctx.Next()` was called. If a middleware wants to finish the processing, it
// can just write to the `http.ResponseWriter` or use the `ctx.Responder` for
// convienience.
//
// The `Context.App` can be initialized by providing a CtxConstructor via
// `SetAppContext()`.
func (this *Server) NewMiddlewareHandler(middlewares []Middleware) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Initialize fresh scope variables.
		ctx := &Context{
			MuxVars: mux.Vars(req),
			Response: Response{
				w: res,
			},
		}

		if this.ctxConstructor != nil {
			ctx.App = this.ctxConstructor()
		}

		for _, middleware := range middlewares {
			nextCalled := false
			ctx.Next = func() error {
				nextCalled = true
				return nil
			}

			// End the request with an error and stop calling further middlewares.
			if err := middleware(res, req, ctx); err != nil {
				if this.statusLogger != nil {
					this.statusLogger.Error("%s %s %#v", req.Method, req.URL, errgo.Mask(err))
				}
				ctx.Response.Error(err.Error(), http.StatusInternalServerError)
				return
			}

			if !nextCalled {
				break
			}
		}
	})
}
