package server

import (
	"fmt"
	"net/http"
)

func NewWelcomeMiddleware(appName, version string) Middleware {
	return func(res http.ResponseWriter, rep *http.Request, ctx *Context) error {
		return ctx.Response.PlainText(fmt.Sprintf("This is %s version %s\n", appName, version), http.StatusOK)
	}
}
