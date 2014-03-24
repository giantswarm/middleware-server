# middleware-server
Lightweight server module, handling http middlewares.

### Docs
http://godoc.org/github.com/catalyst-zero/middleware-server

### Install
```go
import "github.com/catalyst-zero/middleware-server"
```

### Usage
```go
package main

import (
	serverPkg "github.com/catalyst-zero/middleware-server"
)

// Optionally define your app context to use across your middlewares.
type AppContext struct {
	Greeting string
}

// Define your version namespace acting as middleware receiver.
type V1 struct{}

// Define middlewares calling the next middleware.
func (this *V1) First(res http.ResponseWriter, req *http.Request, ctx *serverPkg.Context) error {
	// Optionally manipulate your app context for following middlewares.
	ctx.App.(AppContext).Greeting = "hello world"
	return ctx.Next()
}

// Define the last middleware in the chain responding to the request.
func (this *V1) Last(res http.ResponseWriter, req *http.Request, ctx *serverPkg.Context) error {
	return ctx.Response.PlainText(ctx.App.(AppContext).Greeting, http.StatusOK)
}

func main() {
	// Create the server.
	server := serverPkg.NewServer("127.0.0.1", "8080")
	server.SetLogger(server.NewLogger("stm-api"))
	server.SetAppContext(AppContext{})

	// Create a version namespace.
	v1 := &V1{}
	server.Serve("GET", "/v1/foo/",
		v1.First,
		v1.Last,
	)

	// Start the server.
	server.Listen()
}
```

### Responders
http://godoc.org/github.com/catalyst-zero/middleware-server#Response
