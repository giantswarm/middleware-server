package main

import (
	"net/http"

	srvPkg "github.com/catalyst-zero/middleware-server"
	logPkg "github.com/op/go-logging"
)

type V1 struct {
	Logger *logPkg.Logger
}

func (this *V1) middlewareOne(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Debug("hello world")
	return ctx.Response.PlainText("hello world", http.StatusOK)
}

func (this *V1) notFound(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Debug("not found")
	return ctx.Response.PlainText("not found", http.StatusOK)
}

func main() {
	logger := srvPkg.NewSimpleLogger("not-found-example")
	v1 := &V1{Logger: logger}

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.Serve("GET", "/v1/hello-world", v1.middlewareOne)
	srv.ServeNotFound(v1.notFound)

	srv.Listen()
}
