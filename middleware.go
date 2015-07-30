package server

import (
	"fmt"
	"net/http"

	"github.com/juju/errgo"
	"golang.org/x/net/context"
)

// NewWelcomeMiddleware provides a middleware that responds human readable
// information about a service. E.g. one can register this under /.
func NewWelcomeMiddleware(appName, version string) Middleware {
	return func(ctx context.Context, res http.ResponseWriter, rep *http.Request) error {
		scope := ctx.Value(ScopeKey).(Scope)
		return scope.Response.PlainText(fmt.Sprintf("This is %s version %s\n", appName, version), http.StatusOK)
	}
}

// NewHealthcheckMiddleware provides a middleware that responds JSON formatted
// information about a service. E.g. one can register this under /healthcheck.
func NewHealthcheckMiddleware(hc Healthchecker) Middleware {
	return func(ctx context.Context, res http.ResponseWriter, rep *http.Request) error {
		scope := ctx.Value(ScopeKey).(Scope)

		hcRes, err := hc.Status()
		if err != nil {
			return errgo.Mask(err)
		}

		return scope.Response.Json(hcRes, http.StatusOK)
	}
}
