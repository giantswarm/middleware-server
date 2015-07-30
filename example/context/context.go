package main

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/giantswarm/middleware-server"
)

const (
  Key = "key"
  Value = "value"
)

func middlewareOne(ctx context.Context, res http.ResponseWriter, rep *http.Request) error {
  scope := ctx.Value(server.ScopeKey).(server.Scope)
	scope.Logger.Debug("middleware one")
	return scope.Next()
}

func middlewareTwo(ctx context.Context, res http.ResponseWriter, rep *http.Request) error {
  scope := ctx.Value(server.ScopeKey).(server.Scope)
	scope.Logger.Debug("middleware two")

  ctx = context.WithValue(ctx, Key, Value)

	return scope.Next()
}

func middlewareThree(ctx context.Context, res http.ResponseWriter, rep *http.Request) error {
  scope := ctx.Value(server.ScopeKey).(server.Scope)
	scope.Logger.Debug("middleware three")

  if value, ok := ctx.Value(Key).(string); ok {
    scope.Logger.Debug("value: %s", value)
  }

	return scope.Response.PlainText("OK", http.StatusOK)
}

func main() {
	srv := server.NewServer("127.0.0.1", "8080")
	srv.Serve("GET", "/", middlewareOne, middlewareTwo, middlewareThree)
	srv.Listen()
}
