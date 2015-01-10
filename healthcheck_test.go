package server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	srvPkg "github.com/catalyst-zero/middleware-server"
)

func TestHealtcheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "healtcheck")
}

var _ = Describe("healtcheck", func() {
	var (
		err    error
		status string
		hc     srvPkg.Healthchecker
		info   srvPkg.HealthInfo

		expectedStatus string
	)

	BeforeEach(func() {
		err = nil
		status = ""

		hc = func() (srvPkg.HealthInfo, error) {
			return info, nil
		}

		info = srvPkg.HealthInfo{
			Status:   srvPkg.StatusHealthy,
			App:      "test-app",
			Version:  "test-version",
			Backends: []srvPkg.HealthInfo{},
		}
	})

	AfterEach(func() {
		info, err = hc.Status()

		Expect(info.Status).To(Equal(expectedStatus))
	})

	Describe("status calculation", func() {
		Context("healthy service having no backends", func() {
			It("should calculate status healthy", func() {
				info.Status = srvPkg.StatusHealthy
				info.Backends = []srvPkg.HealthInfo{}

				expectedStatus = srvPkg.StatusHealthy
			})
		})

		Context("healthy service having healthy backends", func() {
			It("should calculate status healthy", func() {
				info.Status = srvPkg.StatusHealthy
				info.Backends = []srvPkg.HealthInfo{
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
				}

				expectedStatus = srvPkg.StatusHealthy
			})
		})

		Context("unhealthy service having healthy backends", func() {
			It("should calculate status healthy", func() {
				info.Status = srvPkg.StatusUnhealthy
				info.Backends = []srvPkg.HealthInfo{
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
				}

				expectedStatus = srvPkg.StatusUnhealthy
			})
		})

		Context("healthy service having unhealthy backends", func() {
			It("should calculate status unhealthy", func() {
				info.Status = srvPkg.StatusHealthy
				info.Backends = []srvPkg.HealthInfo{
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
					srvPkg.HealthInfo{Status: srvPkg.StatusUnhealthy}, // unhealthy
					srvPkg.HealthInfo{Status: srvPkg.StatusHealthy},
				}

				expectedStatus = srvPkg.StatusUnhealthy
			})
		})
	})
})
