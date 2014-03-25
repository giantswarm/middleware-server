package server_test

import (
	"net/http"
	"net/http/httptest"

	server "github.com/catalyst-zero/middleware-server"

	"github.com/catalyst-zero/middleware-server/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Define testing middlewares v1.
type versionOne struct{}

func (this *versionOne) first(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Next()
}

func (this *versionOne) last(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Response.PlainText("hello world", http.StatusOK)
}

// Define testing middlewares v2.
type AppContext struct {
	Greeting string
}

type versionTwo struct{}

func (this *versionTwo) first(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	ctx.App.(*AppContext).Greeting = "hello world"
	return ctx.Next()
}

func (this *versionTwo) last(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Response.PlainText(ctx.App.(*AppContext).Greeting, http.StatusOK)
}

// Test the server.
var _ = Describe("Server", func() {
	var (
		ts   *httptest.Server
		code int
		body string
		srv  *server.Server
	)

	BeforeEach(func() {
		// Create test server.
		ts = test.NewServer(nil)

		// Create app server.
		srv = server.NewServer("", "")
	})

	AfterEach(func() {
		// Close test server.
		ts.Close()
	})

	Context("Correct middleware handling", func() {
		BeforeEach(func() {
			v1 := &versionOne{}
			srv.Serve("GET", "/v1/hello/",
				v1.first,
				v1.last,
			)

			// Configure test server router.
			ts.Config.Handler = srv.Routers["v1"]

			code, body, _ = test.NewGetRequest(ts.URL + "/v1/hello/")
		})

		It("Should respond with status code 200", func() {
			Expect(code).To(Equal(http.StatusOK))
		})

		It("Should respond with 'hello world'", func() {
			Expect(body).To(Equal("hello world"))
		})
	})

	Context("App Context", func() {
		BeforeEach(func() {
			v2 := &versionTwo{}
			srv.Serve("GET", "/v2/hello/",
				v2.first,
				v2.last,
			)

			srv.SetAppContext(func() interface{} {
				return &AppContext{}
			})

			// Configure test server router.
			ts.Config.Handler = srv.Routers["v2"]

			code, body, _ = test.NewGetRequest(ts.URL + "/v2/hello/")
		})

		It("Should respond with status code 200", func() {
			Expect(code).To(Equal(http.StatusOK))
		})

		It("Should respond with 'hello world'", func() {
			Expect(body).To(Equal("hello world"))
		})
	})
})
