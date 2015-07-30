package main

import (
	"github.com/giantswarm/middleware-server"
)

// $ curl -i http://localhost:8080/v1/public/
// $ curl -i http://localhost:8080/v1/public/test.html
//
func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.ServeStatic("/", "./example/fileserver/public/")
	srv.Logger.Info("This is the fileserver example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
