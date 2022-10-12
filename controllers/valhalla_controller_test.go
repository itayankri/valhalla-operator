package controllers_test

import (
	"context"
	"fmt"
	"time"

	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	"github.com/itayankri/valhalla-operator/internal/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterCreationTimeout = 10 * 60 * time.Second
	ClusterDeletionTimeout = 5 * time.Second
)

var _ = Describe("ValhallaController", func() {
	var (
		instance *valhallav1alpha1.Valhalla
		// defaultNamespace = "default"
		// defaultImage     = "itayankri/valhalla-operator:latest"
	)

	Context("Service configurations", func() {

	})

	Context("Resource requirements configurations", func() {

	})

	Context("Persistence configurations", func() {

	})

	Context("Custom Resource updates", func() {

	})

	Context("Recreate child resources after deletion", func() {

	})

	Context("CR ReconcileSuccess condition", func() {

	})

	Context("Pause reconciliation", func() {
		BeforeEach(func() {
			instance = generateValhallaCluster()
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForValhallaCreation(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should skip valhalla instance if pause reconciliation annotation is set to true", func() {
			maxReplicas := int32(10)
			Expect(updateWithRetry(instance, func(v *valhallav1alpha1.Valhalla) {
				v.SetAnnotations(map[string]string{"valhalla.itayankri/operator.paused": "true"})
				v.Spec.MaxReplicas = &maxReplicas
			})).To(Succeed())

			Consistently(func() *int32 {
				instanceCreated := valhallav1alpha1.Valhalla{}
				if err := k8sClient.Get(
					ctx,
					types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
					&instanceCreated,
				); err != nil {
					return nil
				}
				return instanceCreated.Spec.MaxReplicas
			}, 10*time.Second).Should(Equal(&maxReplicas))

			Expect(updateWithRetry(instance, func(v *valhallav1alpha1.Valhalla) {
				v.SetAnnotations(map[string]string{"valhalla.itayankri/operator.paused": "false"})
			})).To(Succeed())

			Consistently(func() *int32 {
				instanceCreated := valhallav1alpha1.Valhalla{}
				if err := k8sClient.Get(
					ctx,
					types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
					&instanceCreated,
				); err != nil {
					return nil
				}
				return instanceCreated.Spec.MaxReplicas
			}, 10*time.Second).Should(Equal(nil))
		})
	})
})

func generateValhallaCluster() *valhallav1alpha1.Valhalla {
	storage := resource.MustParse("10Mi")
	image := "itayankri/valhalla:latest"
	valhalla := &valhallav1alpha1.Valhalla{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "valhalla-pause-reconcile",
			Namespace: "default",
		},
		Spec: valhallav1alpha1.ValhallaSpec{
			PBFURL: "https://download.geofabrik.de/europe/andorra-latest.osm.pbf",
			Image:  &image,
			Persistence: valhallav1alpha1.PersistenceSpec{
				StorageClassName: "nfs-csi",
				Storage:          &storage,
			},
		},
	}
	return valhalla
}

func waitForValhallaCreation(ctx context.Context, instance *valhallav1alpha1.Valhalla, client client.Client) {
	EventuallyWithOffset(1, func() string {
		instanceCreated := valhallav1alpha1.Valhalla{}
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
			&instanceCreated,
		); err != nil {
			return fmt.Sprintf("%v+", err)
		}

		for _, condition := range instanceCreated.Status.Conditions {
			if condition.Type == status.ConditionAvailable && condition.Status == metav1.ConditionTrue {
				return "ready"
			}
		}

		return "not ready"

	}, ClusterCreationTimeout, 1*time.Second).Should(Equal("ready"))
}
