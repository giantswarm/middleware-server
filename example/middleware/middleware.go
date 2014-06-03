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

	srv.Serve("GET", "/v1/hello-world", v1.middlewareOne, v1.middlewareTwo)

	srv.Listen()
}
