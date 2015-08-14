package main

import (
	"fmt"
	"net/http"

	"github.com/giantswarm/middleware-server"
)

func middlewareOne(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Next()
}

func middlewareTwo(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Response.PlainText("OK from middleware", http.StatusOK)
}

func muxHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK from mux handler")
}

func main() {
	srv := server.NewServer("127.0.0.1", "8080")

	// mux server
	mux := http.NewServeMux()
	mux.HandleFunc("/", muxHandler)

	// middleware routes
	srv.Serve("GET", "/middleware", middlewareOne, middlewareTwo)

	// merge middleware handlers into mux
	srv.RegisterRoutes(mux, "/v1")

	// start http server with merged mux
	srv.Logger.Info(nil, "This is the mix-cooperation example. Try `curl localhost:8080`, or `curl localhost:8080/v1/middleware` to see what happens.")
	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		panic(err)
	}
}
