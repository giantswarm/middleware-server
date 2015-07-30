package main

import (
	"net/http"

	"github.com/giantswarm/middleware-server"
)

func main() {
	srv := server.NewServer("127.0.0.1", "8080")

	srv.SetPreHTTPHandler(func(entry *server.AccessEntry) {
		srv.Logger.Debug("Pre-HTTP-Handler called!")
	})

	srv.Serve("GET", "/", func(res http.ResponseWriter, rep *http.Request, ctx *server.Context) error {
		srv.Logger.Debug("HTTP-Handler called!")
		return ctx.Response.PlainText("This is the request-callback example.\n", http.StatusOK)
	})

	srv.SetPostHTTPHandler(func(entry *server.AccessEntry) {
		srv.Logger.Debug("Post-HTTP-Handler called!")
	})

	srv.Logger.Debug("This is the request-callback example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
