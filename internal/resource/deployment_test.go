package resource_test

import (
	"github.com/itayankri/valhalla-operator/internal/resource"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployment builder", func() {
	Context("ShouldDeploy", func() {
		var builder resource.ResourceBuilder
		BeforeEach(func() {
			builder = valhallaResourceBuilder.Deployment()
		})

		It("Should return 'false' when both PVC is bound and map builder Job is not completed yet", func() {
			resources := generateChildResources(false, false)
			Expect(builder.ShouldDeploy(resources)).To(Equal(false))
		})

		It("Should return 'false' when PVC is bound but map builder Job is not completed yet", func() {
			resources := generateChildResources(true, false)
			Expect(builder.ShouldDeploy(resources)).To(Equal(false))
		})

		It("Should return 'false' when PVC is not bound but map builder Job is completed", func() {
			resources := generateChildResources(false, true)
			Expect(builder.ShouldDeploy(resources)).To(Equal(false))
		})

		It("Should return 'true' when both PVC is bound and map builder Job is compoleted", func() {
			resources := generateChildResources(true, true)
			Expect(builder.ShouldDeploy(resources)).To(Equal(true))
		})
	})
})
