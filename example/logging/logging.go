package main

import (
	"fmt"
	"net/http"

	srvPkg "github.com/catalyst-zero/middleware-server"
	logPkg "github.com/op/go-logging"
)

type V1 struct {
	Logger *logPkg.Logger
}

func (this *V1) middlewareOne(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Critical("middleware %s", "one")
	this.Logger.Error("middleware %s", "one")
	this.Logger.Warning("middleware %s", "one")
	this.Logger.Notice("middleware %s", "one")
	this.Logger.Info("middleware %s", "one")
	this.Logger.Debug("middleware %s", "one")

	fmt.Printf("%#v\n", req.Header["Authorization"])

	return ctx.Response.PlainText("hello world", http.StatusOK)
}

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "middleware-example", Level: "debug"})
	v1 := &V1{Logger: logger}

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.Serve("GET", "/v1/hello-world", v1.middlewareOne)

	srv.Listen()
}
