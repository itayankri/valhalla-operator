package metadata_test

import (
	"github.com/itayankri/valhalla-operator/internal/metadata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const defaultAnnotationKey = "before"
const defaultAnnotationValue = "each"

var _ = Describe("Annotations", func() {
	Context("ReconcileAnnotations", func() {
		var baseAnnotations map[string]string
		BeforeEach(func() {
			baseAnnotations = map[string]string{defaultAnnotationKey: defaultAnnotationValue}
		})

		It("Should return a merged map of annotations", func() {
			const testAnnotationKey = "valhalla"
			const testAnnotationValue = "operator"
			annotations := map[string]string{testAnnotationKey: testAnnotationValue}
			reconciledAnnotations := metadata.ReconcileAnnotations(baseAnnotations, annotations)

			beforeEachAnnotation, beforeEachAnnotationExists := reconciledAnnotations[defaultAnnotationKey]
			Expect(beforeEachAnnotationExists).To(Equal(true))
			Expect(beforeEachAnnotation).To(Equal(defaultAnnotationValue))

			testAnnotation, testAnnotationExists := reconciledAnnotations[testAnnotationKey]
			Expect(testAnnotationExists).To(Equal(true))
			Expect(testAnnotation).To(Equal(testAnnotationValue))
		})

		It("Should prefer an annotation from maps array rather than existing annotations", func() {
			const testAnnotationKey = "before"
			const testAnnotationValue = "operator"
			annotations := map[string]string{testAnnotationKey: testAnnotationValue}
			reconciledAnnotations := metadata.ReconcileAnnotations(baseAnnotations, annotations)

			beforeEachAnnotation, beforeEachAnnotationExists := reconciledAnnotations[defaultAnnotationKey]
			Expect(beforeEachAnnotationExists).To(Equal(true))
			Expect(beforeEachAnnotation).To(Equal(testAnnotationValue))
		})

		It("Should return an empty annotations map if existins is nil and no other maps supplied", func() {
			reconciledAnnotations := metadata.ReconcileAnnotations(nil)
			Expect(reconciledAnnotations).ToNot(Equal(nil))
			Expect(len(reconciledAnnotations)).To(Equal(0))
		})
	})
})
