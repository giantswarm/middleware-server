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
	this.Logger.Critical("middleware one")
	this.Logger.Error("middleware one")
	this.Logger.Warning("middleware one")
	this.Logger.Notice("middleware one")
	this.Logger.Info("middleware one")
	this.Logger.Debug("middleware one")

	return ctx.Response.PlainText("hello world", http.StatusOK)
}

func main() {
	logger := srvPkg.NewSimpleLogger("middleware-example", "debug")
	v1 := &V1{Logger: logger}

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.Serve("GET", "/v1/hello-world", v1.middlewareOne)

	srv.Listen()
}
