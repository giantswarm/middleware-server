package main

import (
	srvPkg "github.com/catalyst-zero/middleware-server"
)

func main() {
	logger := srvPkg.NewLogger(srvPkg.LoggerOptions{Name: "healthcheck", Level: "debug"})

	srv := srvPkg.NewServer("127.0.0.1", "8080")
	srv.SetLogger(logger)

	hc := func() (srvPkg.HealthInfo, error) {
		info := srvPkg.HealthInfo{
			Status: srvPkg.StatusHealthy,
			Backends: []srvPkg.HealthInfo{
				srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
				srvPkg.HealthInfo{Status: srvPkg.StatusUnhealthy}, // unhealthy
				srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
			},
		}

		return info, nil
	}

	srv.Serve("GET", "/", srvPkg.NewHealthcheckMiddleware(hc))

	logger.Debug("This is the healthcheck example. Try `curl localhost:8080` to see what happens.")
	srv.Listen()
}
