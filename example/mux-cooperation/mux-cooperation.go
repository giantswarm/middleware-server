package main

import (
	"fmt"
	"net/http"

	srvPkg "github.com/giantswarm/middleware-server"
	logPkg "github.com/op/go-logging"
)

type V1 struct {
	Logger *logPkg.Logger
}

func (this *V1) middlewareOne(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Debug("middleware one")
	return ctx.Next()
}

func (this *V1) middlewareTwo(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Debug("middleware two")
	return ctx.Response.PlainText("hello world", http.StatusOK)
}

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "middleware-example"})
	v1 := &V1{Logger: logger}

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	// mux server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello from custom mux handler")
	})

	// middleware routes
	srv.Serve("GET", "/hello-world", v1.middlewareOne, v1.middlewareTwo)

	// merge middleware handlers into mux
	srv.RegisterRoutes(mux, "/v1")

	// start http server with merged mux
	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		panic(err)
	}
}
