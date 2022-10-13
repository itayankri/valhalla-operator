package controllers_test

import (
	"context"
	"fmt"
	"time"

	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	"github.com/itayankri/valhalla-operator/internal/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterDeletionTimeout = 5 * time.Second
	MapBuildingTimeout     = 2 * 60 * time.Second
)

var _ = Describe("ValhallaController", func() {
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

	Context("Valhalla CR ReconcileSuccess condition", func() {
		var instance *valhallav1alpha1.Valhalla
		BeforeEach(func() {
			instance = generateValhallaCluster("reconcile-success-condition")
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should keep ReconcileSuccess condition updated", func() {
			By("setting to False when spec is not valid", func() {
				// It is impossible to create a deployment with -1 replicas. Thus we expect reconcilication to fail.
				instance.Spec.MinReplicas = pointer.Int32Ptr(-1)
				Expect(k8sClient.Create(ctx, instance)).To(Succeed())
				waitForValhallaCreation(ctx, instance, k8sClient)

				Eventually(func() metav1.ConditionStatus {
					valhalla := &valhallav1alpha1.Valhalla{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      instance.Name,
						Namespace: instance.Namespace,
					}, valhalla)).To(Succeed())

					for _, condition := range valhalla.Status.Conditions {
						if condition.Type == status.ConditionReconciliationSuccess {
							return condition.Status
						}
					}
					return metav1.ConditionUnknown
				}, 60*time.Second).Should(Equal(metav1.ConditionFalse))
			})
		})
	})

	Context("Pause reconciliation", func() {
		var instance *valhallav1alpha1.Valhalla
		BeforeEach(func() {
			instance = generateValhallaCluster("pause-reconcile")
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			waitForValhallaDeployment(ctx, instance, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
		})

		It("Should skip valhalla instance if pause reconciliation annotation is set to true", func() {
			minReplicas := int32(2)
			originalMinReplicas := *instance.Spec.MinReplicas
			Expect(updateWithRetry(instance, func(v *valhallav1alpha1.Valhalla) {
				v.SetAnnotations(map[string]string{"valhalla.itayankri/operator.paused": "true"})
				v.Spec.MinReplicas = &minReplicas
			})).To(Succeed())

			Eventually(func() int32 {
				return *hpa(ctx, instance, "").Spec.MinReplicas
			}, MapBuildingTimeout).Should(Equal(originalMinReplicas))

			Expect(updateWithRetry(instance, func(v *valhallav1alpha1.Valhalla) {
				v.SetAnnotations(map[string]string{"valhalla.itayankri/operator.paused": "false"})
			})).To(Succeed())

			Eventually(func() int32 {
				return *hpa(ctx, instance, "").Spec.MinReplicas
			}, 10*time.Second).Should(Equal(minReplicas))
		})
	})
})

func generateValhallaCluster(name string) *valhallav1alpha1.Valhalla {
	storage := resource.MustParse("10Mi")
	image := "itayankri/valhalla:latest"
	minReplicas := int32(1)
	maxReplicas := int32(3)
	valhalla := &valhallav1alpha1.Valhalla{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: valhallav1alpha1.ValhallaSpec{
			PBFURL:      "https://download.geofabrik.de/australia-oceania/marshall-islands-latest.osm.pbf",
			Image:       &image,
			MinReplicas: &minReplicas,
			MaxReplicas: &maxReplicas,
			Persistence: valhallav1alpha1.PersistenceSpec{
				StorageClassName: "nfs-csi",
				Storage:          &storage,
			},
			Resources: &corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
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

		if len(instanceCreated.Status.Conditions) == 0 {
			return "not ready"
		}

		return "ready"

	}, MapBuildingTimeout, 1*time.Second).Should(Equal("ready"))
}

func waitForValhallaDeployment(ctx context.Context, instance *valhallav1alpha1.Valhalla, client client.Client) {
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

	}, MapBuildingTimeout, 1*time.Second).Should(Equal("ready"))
}

func hpa(ctx context.Context, v *valhallav1alpha1.Valhalla, hpaName string) autoscalingv1.HorizontalPodAutoscaler {
	name := v.ChildResourceName(hpaName)
	hpa := autoscalingv1.HorizontalPodAutoscaler{}
	EventuallyWithOffset(1, func() error {
		if err := k8sClient.Get(
			ctx,
			types.NamespacedName{Name: name, Namespace: v.Namespace},
			&hpa,
		); err != nil {
			return err
		}
		return nil
	}, MapBuildingTimeout).Should(Succeed())
	return hpa
}
