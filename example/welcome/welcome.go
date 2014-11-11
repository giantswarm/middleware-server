package main

import srvPkg "github.com/catalyst-zero/middleware-server"

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "welcome-example", Level: "debug"})

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	srv.Serve("GET", "/", srvPkg.NewWelcomeMiddleware("welcome example", "0.0.1"))

	srv.Listen()
}
