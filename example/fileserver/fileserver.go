package main

import (
	srvPkg "github.com/catalyst-zero/middleware-server"
)

// $ curl -i http://localhost:8080/v1/public/
// $ curl -i http://localhost:8080/v1/public/test.html
//
func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "fileserver-example"})

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.ServeStatic("/v1/public/", "./example/fileserver/public/")

	srv.Listen()
}
