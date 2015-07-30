package main

import (
	"net/http"

	"github.com/giantswarm/middleware-server"
)

func middlewareOne(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.Logger.Debug("middleware one")
	return ctx.Next()
}

func middlewareTwo(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.Logger.Debug("middleware two")
	return scope.Next()
}

func middlewareThree(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.Logger.Debug("middleware three")
	return ctx.Response.PlainText("OK", http.StatusOK)
}

func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.Serve("GET", "/", middlewareOne, middlewareTwo, middlewareThree)
	logger.Info("This is the context example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
