package main

import (
	"net/http"

	srvPkg "github.com/catalyst-zero/middleware-server"
	logPkg "github.com/op/go-logging"
)

type V1 struct {
	Logger *logPkg.Logger
}

func (this *V1) welcomeMiddleware(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	return ctx.Response.PlainText("This is middleware!", http.StatusOK)
}

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "middleware-example"})
	v1 := &V1{Logger: logger}

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.Serve("GET", "/", v1.welcomeMiddleware)

	srv.Listen()
}
