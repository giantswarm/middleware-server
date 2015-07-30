package server

import (
	"fmt"
	"net/http"

	"github.com/juju/errgo"
)

// NewWelcomeMiddleware provides a middleware that responds human readable
// information about a service. E.g. one can register this under /.
func NewWelcomeMiddleware(appName, version string) Middleware {
	return func(res http.ResponseWriter, rep *http.Request, ctx *Context) error {
		return ctx.Response.PlainText(fmt.Sprintf("This is %s version %s\n", appName, version), http.StatusOK)
	}
}

// NewHealthcheckMiddleware provides a middleware that responds JSON formatted
// information about a service. E.g. one can register this under /healthcheck.
func NewHealthcheckMiddleware(hc Healthchecker) Middleware {
	return func(res http.ResponseWriter, rep *http.Request, ctx *Context) error {
		hcRes, err := hc.Status()
		if err != nil {
			return errgo.Mask(err)
		}

		return ctx.Response.Json(hcRes, http.StatusOK)
	}
}
