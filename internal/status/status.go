package status

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ConditionAvailable             = "Available"
	ConditionReconciliationSuccess = "ReconciliationSuccess"
	ConditionAllReplicasReady      = "AllReplicasReady"
)

func AvailableCondition(resources []runtime.Object, old *metav1.Condition) metav1.Condition {
	condition := metav1.Condition{
		Type:    ConditionAvailable,
		Status:  metav1.ConditionFalse,
		Reason:  "DeploymentUnavailable",
		Message: "Deployment does not have minimum availability",
	}

	if old != nil {
		condition.LastTransitionTime = old.LastTransitionTime
	}

	for _, resource := range resources {
		if deployment, ok := resource.(*appsv1.Deployment); ok {
			if deployment != nil {
				for _, cond := range deployment.Status.Conditions {
					if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
						condition.Status = metav1.ConditionTrue
						condition.Message = cond.Message
						condition.Reason = "Available"
					}
				}
			}
			break
		}
	}

	if old == nil || old.Status != condition.Status {
		condition.LastTransitionTime = metav1.Time{
			Time: time.Now(),
		}
	}

	return condition
}

func AllReplicasReadyCondition(resources []runtime.Object, old *metav1.Condition) metav1.Condition {
	condition := metav1.Condition{
		Type:    ConditionAllReplicasReady,
		Status:  metav1.ConditionFalse,
		Reason:  "NotAllReplicasReady",
		Message: "One or more pods are not ready",
	}

	if old != nil {
		condition.LastTransitionTime = old.LastTransitionTime
	}

	if DoAllReplicasReady(resources) {
		condition.Status = metav1.ConditionTrue
		condition.Message = "All pods are ready"
		condition.Reason = "AllReplicasReady"
	}

	if old == nil || old.Status != condition.Status {
		condition.LastTransitionTime = metav1.Time{
			Time: time.Now(),
		}
	}

	return condition
}

func ReconcileSuccessCondition(status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               ConditionReconciliationSuccess,
		Status:             status,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             reason,
		Message:            message,
	}
}

func IsPersistentVolumeClaimBound(resources []runtime.Object) bool {
	pvcBound := false
	for _, resource := range resources {
		if pvc, ok := resource.(*corev1.PersistentVolumeClaim); ok {
			if pvc != nil && pvc.Status.Phase == corev1.ClaimBound {
				pvcBound = true
			}
			break
		}
	}
	return pvcBound
}

func IsJobCompleted(resources []runtime.Object) bool {
	jobCompleted := false
	for _, resource := range resources {
		if job, ok := resource.(*batchv1.Job); ok {
			if job != nil {
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
						jobCompleted = true
					}
				}
				break
			}
		}
	}
	return jobCompleted
}

func DoAllReplicasReady(resources []runtime.Object) bool {
	allReplicasReady := false
	for _, resource := range resources {
		if deployment, ok := resource.(*appsv1.Deployment); ok {
			if deployment != nil && deployment.Spec.Replicas != nil && deployment.Status.ReadyReplicas >= *deployment.Spec.Replicas {
				allReplicasReady = true
			}
			break
		}
	}
	return allReplicasReady
}
