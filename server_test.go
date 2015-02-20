package server_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/giantswarm/middleware-server/test"

	srvPkg "github.com/giantswarm/middleware-server"
	log "github.com/op/go-logging"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Define testing middlewares v1.
type V1 struct {
	Logger *log.Logger
}

func (this *V1) first(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Debug("test message")
	return ctx.Next()
}

func (this *V1) last(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	return ctx.Response.PlainText("hello world", http.StatusOK)
}

// Define testing middlewares v2.
type AppContext struct {
	Greeting string
}

type V2 struct {
	Logger *log.Logger
}

func (this *V2) first(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	this.Logger.Info("test message")
	ctx.App.(*AppContext).Greeting = "hello world"
	return ctx.Next()
}

func (this *V2) last(res http.ResponseWriter, req *http.Request, ctx *srvPkg.Context) error {
	return ctx.Response.PlainText(ctx.App.(*AppContext).Greeting, http.StatusOK)
}

// Test the server.
var _ = Describe("Server", func() {
	var (
		ts     *httptest.Server
		code1  int
		code2  int
		body1  string
		body2  string
		srv    *srvPkg.Server
		logger *log.Logger
	)

	BeforeEach(func() {
		// Create test server.
		ts = test.NewServer(nil)

		// Create app server.
		logger = srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "test", Level: "info"})
		srv = srvPkg.NewServer("", "")
		srv.SetLogger(logger)
	})

	AfterEach(func() {
		// Close test server.
		ts.Close()
	})

	Context("Correct middleware handling", func() {
		BeforeEach(func() {
			v1 := &V1{Logger: logger}
			srv.Serve("GET", "/v1/hello/", v1.first, v1.last)

			// Configure test server router.
			ts.Config.Handler = srv.Router

			code1, body1, _ = test.NewGetRequest(ts.URL + "/v1/hello/")
		})

		It("Should respond with status code 200", func() {
			Expect(code1).To(Equal(http.StatusOK))
		})

		It("Should respond with 'hello world'", func() {
			Expect(body1).To(Equal("hello world"))
		})
	})

	Context("App Context", func() {
		BeforeEach(func() {
			v2 := &V2{Logger: logger}
			srv.Serve("GET", "/v2/hello/", v2.first, v2.last)
			srv.Serve("GET", "/v2/empty/", v2.last)

			srv.SetAppContext(func() interface{} {
				return &AppContext{}
			})

			// Configure test server router.
			ts.Config.Handler = srv.Router

			code1, body1, _ = test.NewGetRequest(ts.URL + "/v2/hello/")
			code2, body2, _ = test.NewGetRequest(ts.URL + "/v2/empty/")
		})

		Context("Writing to app context", func() {
			It("Should respond with status code 200", func() {
				Expect(code1).To(Equal(http.StatusOK))
			})

			It("Should write to app context and respond with 'hello world'", func() {
				Expect(body1).To(Equal("hello world"))
			})
		})

		Context("Not writing to app context when it was written before", func() {
			It("Should respond with status code 200", func() {
				Expect(code2).To(Equal(http.StatusOK))
			})

			It("Should not write to app context and respond with empty body", func() {
				Expect(body2).To(Equal(""))
			})
		})
	})
})
