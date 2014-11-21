package main

import (
	"net/http"

	srvPkg "github.com/catalyst-zero/middleware-server"
)

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "request-callback", Level: "debug"})

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.SetPreHTTPHandler(func(entry *srvPkg.AccessEntry) {
		logger.Debug("Pre-HTTP-Handler called!")
	})

	srv.Serve("GET", "/", func(res http.ResponseWriter, rep *http.Request, ctx *srvPkg.Context) error {
		logger.Debug("HTTP-Handler called!")
		return ctx.Response.PlainText("This is the request-callback example.\n", http.StatusOK)
	})

	srv.SetPostHTTPHandler(func(entry *srvPkg.AccessEntry) {
		logger.Debug("Post-HTTP-Handler called!")
	})

	logger.Debug("This is the request-callback example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
