package main

import (
	"net/http"

	"github.com/giantswarm/middleware-server"
)

func middlewareOne(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.Logger.Debug("middlewareOne")
	return ctx.Response.PlainText("OK", http.StatusOK)
}

func notFound(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.Logger.Debug("not found")
	return ctx.Response.PlainText("not found", http.StatusOK)
}

func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.Serve("GET", "/", middlewareOne)
	srv.ServeNotFound(notFound)
	srv.Logger.Debug("This is the not-found example. Try `curl localhost:8080`, or `curl localhost:8080/foo` to see what happens.")
	srv.Listen()
}
