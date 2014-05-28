# middleware-server
Lightweight server module, handling http middlewares.

### Docs
http://godoc.org/github.com/catalyst-zero/middleware-server

### Install
```bash
$ go get github.com/catalyst-zero/middleware-server
```

### Import
```go
import "github.com/catalyst-zero/middleware-server"
```

### Usage
See the examples
```bash
make build-examples
```

### Responders
http://godoc.org/github.com/catalyst-zero/middleware-server#Response

### Access Logging
There is a access logging implemented by default when setting a logger.
```bash
# format: date time file:line: [level] METHOD path code bytes milliseconds
2014/05/28 12:51:22 logaccess.go:56: [INFO] GET /v1/hello-world 200 11 0
```
