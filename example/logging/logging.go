package main

import (
	"net/http"

	"github.com/giantswarm/middleware-server"
)

func middlewareOne(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.Logger.Critical("middleware %s", "one")
	ctx.Logger.Error("middleware %s", "one")
	ctx.Logger.Warning("middleware %s", "one")
	ctx.Logger.Notice("middleware %s", "one")
	ctx.Logger.Info("middleware %s", "one")
	ctx.Logger.Debug("middleware %s", "one")

	return ctx.Next()
}

func middlewareTwo(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.AddLoggerMeta("price", 12.50)
	ctx.Logger.Info("middleware %s", "two")

	return ctx.Next()
}

func middlewareThree(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.AddLoggerMeta("price", 18.25)
	ctx.Logger.Info("middleware %s", "three")

	return ctx.Response.PlainText("OK OK OK", http.StatusOK)
}

func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.SetLogLevel("info")
	srv.Serve("GET", "/", middlewareOne, middlewareTwo, middlewareThree)
	srv.Logger.Info("This is the logging example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
