package status

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ConditionAvailable             = "Available"
	ConditionReconciliationSuccess = "ReconciliationSuccess"
	ConditionNoWarnings            = "NoWarnings"
	ConditionAllReplicasReady      = "AllReplicasReady"
)

func AvailableCondition(resources []runtime.Object) metav1.Condition {
	return metav1.Condition{}
}

func NoWarningsCondition(resources []runtime.Object) metav1.Condition {
	return metav1.Condition{}
}

func AllReplicasReadyCondition(resources []runtime.Object) metav1.Condition {
	return metav1.Condition{}
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
