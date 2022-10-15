package resource_test

import (
	"github.com/itayankri/valhalla-operator/internal/resource"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("PersistentVolumeClaim builder", func() {
	Context("ShouldDeploy", func() {
		var builder resource.ResourceBuilder
		BeforeEach(func() {
			builder = valhallaResourceBuilder.PersistentVolumeClaim()
		})

		It("Should always return 'true'", func() {
			resources := []runtime.Object{}
			Expect(builder.ShouldDeploy(resources)).To(Equal(true))
		})
	})
})
