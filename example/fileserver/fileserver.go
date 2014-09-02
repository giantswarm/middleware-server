package main

import (
	srvPkg "github.com/catalyst-zero/middleware-server"
)

// $ curl -i http://localhost:8080/v1/public/index.html
//
func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "middleware-example"})

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.ServeStatic("/v1/public/", "./example/fileserver/public/")

	srv.Listen()
}
