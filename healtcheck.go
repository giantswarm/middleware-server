package server

import (
	"github.com/juju/errgo"
)

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

type HealthInfo struct {
	Status   string
	App      string
	Version  string
	Backends []HealthInfo
}

type Healthchecker func() (HealthInfo, error)

// Status just accumulates the backends stati to calculate the main status.
func (hc Healthchecker) Status() (HealthInfo, error) {
	info, err := hc()
	if err != nil {
		return HealthInfo{}, errgo.Mask(err)
	}

	info.Status = checkStatus(info)

	return info, nil
}

func IsStatusHealthy(status string) bool {
	return status == StatusHealthy
}

//------------------------------------------------------------------------------
// private

func checkStatus(info HealthInfo) string {
	if !IsStatusHealthy(info.Status) {
		return StatusUnhealthy
	}

	for _, res := range info.Backends {
		if !IsStatusHealthy(checkStatus(res)) {
			return StatusUnhealthy
		}
	}

	return StatusHealthy
}
