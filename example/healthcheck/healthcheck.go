package main

import (
	"github.com/giantswarm/middleware-server"
)

func main() {
	srv := server.NewServer("127.0.0.1", "8080")

	hc := func() (server.HealthInfo, error) {
		info := server.HealthInfo{
			Status: server.StatusHealthy,
			Backends: []server.HealthInfo{
				server.HealthInfo{Status: server.StatusHealthy},
				server.HealthInfo{Status: server.StatusUnhealthy}, // unhealthy
				server.HealthInfo{Status: server.StatusHealthy},
			},
		}

		return info, nil
	}

	srv.Serve("GET", "/", server.NewHealthcheckMiddleware(hc))

	srv.Logger.Info(nil, "This is the healthcheck example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
