package server_test

import (
	"net/http"
	"net/http/httptest"

	server "github.com/catalyst-zero/middleware-server"

	"github.com/catalyst-zero/middleware-server/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Define testing middlewares.
type versionOne struct{}

func (this *versionOne) first(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Next()
}

func (this *versionOne) last(res http.ResponseWriter, req *http.Request, ctx *server.Context) error {
	return ctx.Response.PlainText("hello world", http.StatusOK)
}

// Test the server.
var _ = Describe("Server", func() {
	var (
		ts   *httptest.Server
		code int
		body string
	)

	BeforeEach(func() {
		// Create test server.
		ts = test.CreateServer(nil)

		// Create v1 server.
		srv := server.NewServer("", "")
		v1 := &versionOne{}
		srv.Serve("GET", "/v1/hello/",
			v1.first,
			v1.last,
		)

		// Configure test server router.
		ts.Config.Handler = srv.Routers["v1"]
	})

	AfterEach(func() {
		// Close test server.
		ts.Close()
	})

	Describe("Check health", func() {
		BeforeEach(func() {
			code, body, _ = test.GetRequest(ts.URL + "/v1/hello/")
		})

		It("Should respond with status code 200", func() {
			Expect(code).To(Equal(http.StatusOK))
		})

		It("Should respond with 'hello world'", func() {
			Expect(body).To(Equal("hello world"))
		})
	})
})
