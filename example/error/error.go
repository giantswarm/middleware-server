package main

import (
	"fmt"
	"net/http"

	"github.com/giantswarm/middleware-server"
)

func middlewareOne(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return fmt.Errorf("error")
}

func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.Serve("GET", "/", middlewareOne)
	srv.Logger.Info(nil, "This is the error example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
