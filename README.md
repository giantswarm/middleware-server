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
// Create server.
srv := server.NewServer("127.0.0.1", "8080")
srv.SetLogger(srv.NewLogger("app-name"))

srv.Serve("GET", "/v1/")
```
