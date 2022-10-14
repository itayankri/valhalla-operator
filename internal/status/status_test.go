package status_test

import (
	"github.com/itayankri/valhalla-operator/internal/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"time"
)

const valhallaName = "test"

var _ = Describe("Status", func() {
	Context("Conditions", func() {
		Context("ConditionAvailable", func() {
			It("Should return a new condition with ConditionTrue status if child deployment is available", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAvailable,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      valhallaName,
							Namespace: "default",
						},
						Spec: appsv1.DeploymentSpec{},
						Status: appsv1.DeploymentStatus{
							Conditions: []appsv1.DeploymentCondition{
								appsv1.DeploymentCondition{
									Type:   appsv1.DeploymentAvailable,
									Status: corev1.ConditionTrue,
									LastTransitionTime: metav1.Time{
										Time: time.Now(),
									},
									LastUpdateTime: metav1.Time{
										Time: time.Now(),
									},
								},
							},
						},
					},
				}
				condition := status.AvailableCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			})

			It("Should return a new condition with ConditionFalse status if child deployment is unavailable", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAvailable,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      valhallaName,
							Namespace: "default",
						},
						Spec: appsv1.DeploymentSpec{},
						Status: appsv1.DeploymentStatus{
							Conditions: []appsv1.DeploymentCondition{
								appsv1.DeploymentCondition{
									Type:   appsv1.DeploymentAvailable,
									Status: corev1.ConditionFalse,
									LastTransitionTime: metav1.Time{
										Time: time.Now(),
									},
									LastUpdateTime: metav1.Time{
										Time: time.Now(),
									},
								},
							},
						},
					},
				}
				condition := status.AvailableCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			})

			It("Should return a new condition with ConditionFalse status if child deployment is not present in child resources slice", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAvailable,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{}
				condition := status.AvailableCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			})

			It("Should update LastTransitionTime if status changed", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAvailable,
					Status: metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{}
				condition := status.AvailableCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
				Expect(oldCondition.LastTransitionTime.Before(&condition.LastTransitionTime)).To(Equal(true))
			})

			It("Should not update LastTransitionTime if status has not changed", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAvailable,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{}
				condition := status.AvailableCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
				Expect(oldCondition.LastTransitionTime.Before(&condition.LastTransitionTime)).To(Equal(false))
			})
		})

		Context("ConditionAllReplicasReady", func() {
			It("Should return a new condition with ConditionTrue status if all child deployment's pods are available", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAllReplicasReady,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      valhallaName,
							Namespace: "default",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: pointer.Int32Ptr(2),
						},
						Status: appsv1.DeploymentStatus{
							ReadyReplicas: 2,
						},
					},
				}

				condition := status.AllReplicasReadyCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAllReplicasReady))
				Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			})

			It("Should return a new condition with ConditionFalse status if not all child deployment's pods are available", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAllReplicasReady,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      valhallaName,
							Namespace: "default",
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: pointer.Int32Ptr(2),
						},
						Status: appsv1.DeploymentStatus{
							ReadyReplicas: 1,
						},
					},
				}

				condition := status.AllReplicasReadyCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAllReplicasReady))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			})

			It("Should return a new condition with ConditionFalse status if child deployment' is not present in child resources slice", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAllReplicasReady,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{}
				condition := status.AllReplicasReadyCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAllReplicasReady))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			})

			It("Should update LastTransitionTime if status changed", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAllReplicasReady,
					Status: metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{}
				condition := status.AllReplicasReadyCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAllReplicasReady))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
				Expect(oldCondition.LastTransitionTime.Before(&condition.LastTransitionTime)).To(Equal(true))
			})

			It("Should not update LastTransitionTime if status has not changed", func() {
				oldCondition := &metav1.Condition{
					Type:   status.ConditionAllReplicasReady,
					Status: metav1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				}

				childResources := []runtime.Object{}
				condition := status.AllReplicasReadyCondition(childResources, oldCondition)
				Expect(condition.Type).To(Equal(status.ConditionAllReplicasReady))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
				Expect(oldCondition.LastTransitionTime.Before(&condition.LastTransitionTime)).To(Equal(false))
			})
		})
	})
})
