package main

import (
	"github.com/giantswarm/middleware-server"
)

func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.Serve("GET", "/", server.NewWelcomeMiddleware("welcome example", "0.0.1"))
	srv.Logger.Info("This is the welcome example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
