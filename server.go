package server

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/juju/errgo"

	log "github.com/op/go-logging"
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

	alreadyRegisteredRoutes bool

	Routers map[string]*mux.Router

	ctxConstructor CtxConstructor
}

func NewServer(host, port string) *Server {
	return &Server{
		addr:    host + ":" + port,
		Routers: map[string]*mux.Router{},
	}
}

func (this *Server) Serve(method, urlPath string, middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one Middleware-Handler. Aborting...")
	}
	handler := this.NewMiddlewareHandler(middlewares)

	this.router(urlPath).Methods(method).Path(urlPath).Handler(handler)
}

// ServeStatis registers a middleware that serves files from the filesystem.
// Example: this.ServeStatic("/v1/public", "./public_html/v1/")
func (this *Server) ServeStatic(urlPath, fsPath string) {
	handler := http.StripPrefix(urlPath, http.FileServer(http.Dir(fsPath)))
	this.router(urlPath).Methods("GET").PathPrefix(urlPath).Handler(handler)
}

func (this *Server) router(urlPath string) *mux.Router {
	// Get version by path.
	version := strings.Split(urlPath, "/")[1]

	// Create versioned router if not already set.
	if _, ok := this.Routers[version]; !ok {
		// Set versioned router.
		this.Routers[version] = mux.NewRouter()
	}

	return this.Routers[version]
}

func (this *Server) ServeNotFound(middlewares ...Middleware) {
	if len(middlewares) == 0 {
		panic("Missing at least one NotFound-Handler. Aborting...")
	}

	handler := this.NewMiddlewareHandler(middlewares)
	for version, _ := range this.Routers {
		this.Routers[version].NotFoundHandler = handler
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	if s.alreadyRegisteredRoutes {
		return
	}

	for version, router := range s.Routers {
		var handler http.Handler = router
		if s.accessLogger != nil {
			handler = NewLogAccessHandler(DefaultAccessReporter(s.accessLogger), handler)
		}
		mux.Handle("/"+version+"/", handler)
	}

	s.alreadyRegisteredRoutes = true
}

func (this *Server) Listen() {
	mux := http.NewServeMux()
	this.RegisterRoutes(mux)

	this.statusLogger.Info("starting service on " + this.addr)
	panic(http.ListenAndServe(this.addr, mux))
}

/**
 * SetLogger sets the logger object to which the server logs every request.
 */
func (this *Server) SetLogger(logger *log.Logger) {
	this.SetAccessLogger(logger)
	this.SetStatusLogger(logger)
}

func (this *Server) SetAccessLogger(logger *log.Logger) {
	this.accessLogger = logger
}
func (this *Server) SetStatusLogger(logger *log.Logger) {
	this.statusLogger = logger
}

/**
 * SetAppContext sets the CtxConstructor object, that is called for every request to provide the initial
 * `Context.App` value, which is available to every middleware.
 */
func (this *Server) SetAppContext(ctxConstructor CtxConstructor) {
	this.ctxConstructor = ctxConstructor
}

// NewMiddlewareHandler wraps the middlewares in a http.Handler. The handler, on activation, calls each
// middleware in order, if no error was returned and `ctx.Next()` was called. If a middleware wants to
// finish the processing, it can just write to the `http.ResponseWriter` or use the `ctx.Responder` for
// convienience.
//
// The `Context.App` can be initialized by providing a CtxConstructor via `SetAppContext()`.
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
