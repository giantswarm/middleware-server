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
// Optionally define your app context to use across your middlewares.
type AppContext struct {
  Greeting string
}

// Define your version namespace acting as middleware receiver.
type versionOne struct{}

// Define middlewares calling the next middleware.
func (this *versionOne) First(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
  // Optionally manipulate your app context for following middlewares.
  ctx.App.(*AppContext).Greeting = "hello world"
  return ctx.Next()
}

// Define the last middleware in the chain responding to the request.
func (this *versionOne) Last(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
  return ctx.Response.PlainText(ctx.App.(*AppContext).Greeting, http.StatusOK)
}

// Create the server.
srv := server.NewServer("127.0.0.1", "8080")
srv.SetLogger(srv.NewLogger("stm-api"))

// Create a version namespace.
v1 := &versionOne{}
srv.Serve("GET", /v1/foo/,
  v1.First,
  v1.Last,
)

// Start the server.
srv.Listen()
```

### Responders
http://godoc.org/github.com/catalyst-zero/middleware-server#Response
