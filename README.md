# middleware-server
Lightweight server module, handling http middlewares.

### Docs
http://godoc.org/github.com/catalyst-zero/middleware-server

### Install
```golang
import "github.com/catalyst-zero/middleware-server"
```

### Usage
```golang
// Optionally define your app context to use across your middlewares.
type AppContext struct {
	Greeting string
}

// Define your version namespace acting as middleware receiver.
type versionOne struct{}

// Define middlewares calling the next middleware.
func (this *versionOne) first(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
  // Optionally manipulate your app context for following middlewares.
	ctx.App.(*AppContext).Greeting = "hello world"
	return ctx.Next()
}

// Define the last middleware in the chain responding to the request.
func (this *versionOne) last(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Response.PlainText(ctx.App.(*AppContext).Greeting, http.StatusOK)
}
```

### Responders
http://godoc.org/github.com/catalyst-zero/middleware-server#Response
