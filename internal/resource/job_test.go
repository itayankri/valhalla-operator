package resource_test

import (
	"github.com/itayankri/valhalla-operator/internal/resource"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job builder", func() {
	Context("ShouldDeploy", func() {
		var builder resource.ResourceBuilder
		BeforeEach(func() {
			builder = valhallaResourceBuilder.Job()
		})

		It("Should return 'false' when both PVC is not bound", func() {
			resources := generateChildResources(false, false)
			Expect(builder.ShouldDeploy(resources)).To(Equal(false))
		})

		It("Should return 'true' when PVC is bound", func() {
			resources := generateChildResources(true, true)
			Expect(builder.ShouldDeploy(resources)).To(Equal(true))
		})
	})
})
