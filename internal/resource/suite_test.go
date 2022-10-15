package resource_test

import (
	"testing"

	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	"github.com/itayankri/valhalla-operator/internal/resource"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var valhallaResourceBuilder *resource.ValhallaResourceBuilder

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resource Suite")
}

var _ = BeforeSuite(func() {
	valhallaResourceBuilder = &resource.ValhallaResourceBuilder{
		Instance: &valhallav1alpha1.Valhalla{},
	}
})

func generateChildResources(pvcBound bool, jobCompleted bool) []runtime.Object {
	pvcPhase := corev1.ClaimPending
	if pvcBound {
		pvcPhase = corev1.ClaimBound
	}

	jobConditionStatus := corev1.ConditionFalse
	if jobCompleted {
		jobConditionStatus = corev1.ConditionTrue
	}

	childResources := []runtime.Object{
		&batchv1.Job{
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:   batchv1.JobComplete,
						Status: jobConditionStatus,
					},
				},
			},
		},
		&corev1.PersistentVolumeClaim{
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: pvcPhase,
			},
		},
	}

	return childResources
}
