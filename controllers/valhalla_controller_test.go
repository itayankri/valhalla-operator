package controllers_test

import (
	"context"
	"fmt"
	"time"

	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterCreationTimeout = 10 * time.Second
	ClusterDeletionTimeout = 5 * time.Second
)

var _ = Describe("ValhallaController", func() {
	var (
		cluster          *valhallav1alpha1.Valhalla
		defaultNamespace = "default"
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
			cluster = &valhallav1alpha1.Valhalla{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rabbitmq-pause-reconcile",
					Namespace: defaultNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, cluster)).To(Succeed())
			waitForValhallaCreation(ctx, cluster, k8sClient)
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		})

		It("Should skip valhalla instance if pause reconciliation annotation is set to true", func() {

		})
	})
})

func waitForValhallaCreation(ctx context.Context, instance *valhallav1alpha1.Valhalla, client client.Client) {
	EventuallyWithOffset(1, func() string {
		instanceCreated := valhallav1alpha1.Valhalla{}
		if err := client.Get(
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

	}, ClusterCreationTimeout, 1*time.Second).Should(Equal("ready"))

}
