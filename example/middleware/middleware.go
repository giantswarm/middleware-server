package main

import (
	"net/http"

	"github.com/giantswarm/middleware-server"
	"github.com/giantswarm/request-context"
)

type middleware struct {
	logger requestcontext.Logger
}

func (m middleware) one(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	m.logger.Debug(nil, "I am hidden")
	m.logger.Info(nil, "middleware one")

	return ctx.Next()
}

func (m middleware) two(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	m.logger.Notice(ctx.Request, "middleware two")

	ctx.Request["foo"] = 12.38

	return ctx.Next()
}

func (m middleware) three(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	m.logger.Error(ctx.Request, "middleware three")

	return ctx.Response.PlainText("OK", http.StatusOK)
}

func main() {
	m := middleware{
		logger: requestcontext.MustGetLogger(requestcontext.LoggerConfig{
			Name:  "middleware-example",
			Level: "info",
      Color: true,
		}),
	}

	srv := server.NewServer("127.0.0.1", "8080")
	srv.SetLogger(m.logger)
	srv.Serve("GET", "/", m.one, m.two, m.three)
	srv.Logger.Info(nil, "This is the middleware example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
