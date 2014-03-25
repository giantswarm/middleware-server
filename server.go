package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	log "github.com/op/go-logging"
	Stdlog "log"
)

type CtxConstructor func() interface{}

type Logger *log.Logger

type Server struct {
	// The address to listen on.
	addr           string
	Logger         *log.Logger
	Routers        map[string]*mux.Router
	ctxConstructor CtxConstructor
}

// Middleware is a http handler method.
type Middleware func(http.ResponseWriter, *http.Request, *Context) error

// Context is a map getting through all middlewares.
type Context struct {
	MuxVars  map[string]string
	Response Response
	Next     func() error
	App      interface{}
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
	this.Routers[version].HandleFunc(urlPath, this.NewMiddlewareHandler(middlewares)).Methods(method)
}

func (this *Server) ServeStatic(urlPath, fsPath string) {
	http.Handle(urlPath, http.StripPrefix(urlPath, http.FileServer(http.Dir(fsPath))))
}

func (this *Server) Listen() {
	for version, router := range this.Routers {
		http.Handle("/"+version+"/", router)
	}

	this.Logger.Info("start service on " + this.addr)
	panic(http.ListenAndServe(this.addr, nil))
}

func (this *Server) GetRouter(version string) (*mux.Router, error) {
	if _, ok := this.Routers[version]; !ok {
		return mux.NewRouter(), fmt.Errorf("No router configured for namespace '%s'", version)
	}

	return this.Routers[version], nil
}

func (this *Server) SetLogger(logger *log.Logger) {
	this.Logger = logger
}

func (this *Server) SetAppContext(ctxConstructor CtxConstructor) {
	this.ctxConstructor = ctxConstructor
}

// Create a logger with possible levels:
//   Critical
//   Error
//   Warning
//   Notice
//   Info
//   Debug.
func (this *Server) NewLogger(name string) Logger {
	logger := log.MustGetLogger(name)
	log.SetFormatter(log.MustStringFormatter("[%{level}] %{message}"))
	logBackend := log.NewLogBackend(os.Stderr, "", Stdlog.LstdFlags|Stdlog.Lshortfile)
	logBackend.Color = true
	log.SetBackend(logBackend)

	return logger
}

// Receiving http handler methods to prepare the ordered, sequencial execution
// of them.
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
				this.Logger.Error("%s %s %v", req.Method, req.URL, err.Error())
				ctx.Response.Error(err.Error(), http.StatusInternalServerError)
				return
			}

			if !nextCalled {
				break
			}
		}
	}
}
