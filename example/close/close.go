package main

import (
	"net/http"

	srvPkg "github.com/catalyst-zero/middleware-server"
)

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "close", Level: "debug"})

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.SetCloseListenerDelay(5)
	srv.SetOsExitDelay(5)
	srv.SetOsExitCode(1)

	srv.Serve("GET", "/", func(res http.ResponseWriter, rep *http.Request, ctx *srvPkg.Context) error {
		go srv.Close()

		return ctx.Response.PlainText("This is the close example.\n", http.StatusOK)
	})

	logger.Debug("This is the close example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
