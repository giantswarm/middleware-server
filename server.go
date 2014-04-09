package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

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
	addr           string
	AccessLogger   *log.Logger
	StatusLogger   *log.Logger
	Routers        map[string]*mux.Router
	ctxConstructor CtxConstructor
}

func NewServer(host, port string) *Server {
	return &Server{
		addr:    host + ":" + port,
		Routers: map[string]*mux.Router{},
	}
}

func (this *Server) Serve(method, urlPath string, middlewares ...Middleware) {
	// Get version by path.
	version := strings.Split(urlPath, "/")[1]

	// Create versioned router if not already set.
	if _, ok := this.Routers[version]; !ok {
		this.Routers[version] = mux.NewRouter()
	}

	// set handler to versioned router
	middlewareHandler := http.HandlerFunc(this.NewMiddlewareHandler(middlewares))
	accessHandler := NewLogAccessHandler(DefaultAccessReporter(this.AccessLogger), middlewareHandler)
	this.Routers[version].Handle(urlPath, accessHandler).Methods(method)
}

func (this *Server) ServeStatic(urlPath, fsPath string) {
	http.Handle(urlPath, http.StripPrefix(urlPath, http.FileServer(http.Dir(fsPath))))
}

func (this *Server) Listen() {
	for version, router := range this.Routers {
		http.Handle("/"+version+"/", router)
	}

	this.StatusLogger.Info("start service on " + this.addr)
	panic(http.ListenAndServe(this.addr, nil))
}

func (this *Server) GetRouter(version string) (*mux.Router, error) {
	if _, ok := this.Routers[version]; !ok {
		return mux.NewRouter(), fmt.Errorf("No router configured for namespace '%s'", version)
	}

	return this.Routers[version], nil
}

/**
 * SetLogger sets the logger object to which the server logs every request.
 */
func (this *Server) SetLogger(logger *log.Logger) {
	this.SetAccessLogger(logger)
	this.SetStatusLogger(logger)
}

func (this *Server) SetAccessLogger(logger *log.Logger) {
	this.AccessLogger = logger
}
func (this *Server) SetStatusLogger(logger *log.Logger) {
	this.StatusLogger = logger
}

/**
 * SetAppContext sets the CtxConstructor object, that is called for every request to provide the initial
 * `Context.App` value, which is available to every middleware.
 */
func (this *Server) SetAppContext(ctxConstructor CtxConstructor) {
	this.ctxConstructor = ctxConstructor
}

// NewLogger calls NewSimpleLogger().
// DEPRECATED
func (this *Server) NewLogger(name string) *log.Logger {
	return NewSimpleLogger(name)
}

// NewMiddlewareHandler wraps the middlewares in a http.Handler. The handler, on activation, calls each
// middleware in order, if no error was returned and `ctx.Next()` was called. If a middleware wants to
// finish the processing, it can just write to the `http.ResponseWriter` or use the `ctx.Responder` for
// convienience.
//
// The `Context.App` can be initialized by providing a CtxConstructor via `SetAppContext()`.
func (this *Server) NewMiddlewareHandler(middlewares []Middleware) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
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
				this.logError(req, ctx, err)
				ctx.Response.Error(err.Error(), http.StatusInternalServerError)
				return
			}

			if !nextCalled {
				break
			}
		}
	}
}

// logError is called by the server-handler in case an error occurs in a middleware.
func (this *Server) logError(req *http.Request, ctx *Context, err error) {
	if this.StatusLogger != nil {
		this.StatusLogger.Error("%s %s %v", req.Method, req.URL, err.Error())
	}
}
