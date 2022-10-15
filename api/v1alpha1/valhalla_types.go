/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"strings"

	"github.com/itayankri/valhalla-operator/internal/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const OperatorPausedAnnotation = "valhalla.itayankri/operator.paused"

// Phase is the current phase of the deployment
type Phase string

const (
	// PhaseBuildingMap signals that the map building phase is in progress
	PhaseBuildingMap Phase = "BuildingMap"

	// PhaseDeployingWorkers signals that the workers are being deployed
	PhaseDeployingWorkers Phase = "DeployingWorkers"

	// PhaseWorkersDeployed signals that the resources are successfully deployed
	PhaseWorkersDeployed Phase = "WorkersDeployed"

	// PhaseDeleting signals that the resources are being removed
	PhaseDeleting Phase = "Deleting"

	// PhaseDeleted signals that the resources are deleted
	PhaseDeleted Phase = "Deleted"

	// PhaseError signals that the deployment is in an error state
	PhaseError Phase = "Error"

	// PhaseEmpty is an uninitialized phase
	PhaseEmpty Phase = ""
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ValhallaSpec defines the desired state of Valhalla
type ValhallaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PBFURL        string                       `json:"pbfUrl,omitempty"`
	Image         *string                      `json:"image,omitempty"`
	Persistence   PersistenceSpec              `json:"persistence,omitempty"`
	MinReplicas   *int32                       `json:"minReplicas,omitempty"`
	MaxReplicas   *int32                       `json:"maxReplicas,omitempty"`
	MinAvailable  *int32                       `json:"minAvailable,omitempty"`
	ThreadsPerPod *int32                       `json:"threadsPerPod,omitempty"`
	Resources     *corev1.ResourceRequirements `json:"resources,omitempty"`
}

func (spec *ValhallaSpec) GetResources() *corev1.ResourceRequirements {
	if spec.Resources == nil {
		return &corev1.ResourceRequirements{}
	}
	return spec.Resources
}

func (spec *ValhallaSpec) GetThreadsPerPod() int32 {
	if spec.ThreadsPerPod == nil {
		return 2
	}
	return *spec.ThreadsPerPod
}

func (spec *ValhallaSpec) GetMinAvailable() *intstr.IntOrString {
	if spec.MinAvailable != nil {
		return &intstr.IntOrString{IntVal: *spec.MinAvailable}
	}

	return &intstr.IntOrString{IntVal: 1}
}

func (spec *ValhallaSpec) GetPbfFileName() string {
	split := strings.Split(spec.PBFURL, "/")
	return split[len(split)-1]
}

type PersistenceSpec struct {
	StorageClassName string             `json:"storageClassName,omitempty"`
	Storage          *resource.Quantity `json:"storage,omitempty"`
}

// ValhallaStatus defines the observed state of Valhalla
type ValhallaStatus struct {
	// Paused is true when the operator notices paused annotation.
	Paused bool `json:"paused,omitempty"`

	// ObservedGeneration is the latest generation observed by the operator.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	Phase Phase `json:"phase,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func (valhallaStatus *ValhallaStatus) SetConditions(resources []runtime.Object) {
	var oldAvailableCondition *metav1.Condition
	var oldAllReplicasReadyCondition *metav1.Condition
	var oldReconciliationSuccessCondition *metav1.Condition

	for _, condition := range valhallaStatus.Conditions {
		switch condition.Type {
		case status.ConditionAllReplicasReady:
			oldAllReplicasReadyCondition = condition.DeepCopy()
		case status.ConditionAvailable:
			oldAvailableCondition = condition.DeepCopy()
		case status.ConditionReconciliationSuccess:
			oldReconciliationSuccessCondition = condition.DeepCopy()
		}
	}

	var reconciliationSuccessCondition metav1.Condition
	if oldReconciliationSuccessCondition != nil {
		reconciliationSuccessCondition = *oldReconciliationSuccessCondition
	} else {
		reconciliationSuccessCondition = status.ReconcileSuccessCondition(metav1.ConditionUnknown, "Initialising", "")
	}

	availableCondition := status.AvailableCondition(resources, oldAvailableCondition)
	allReplicasReadyCondition := status.AllReplicasReadyCondition(resources, oldAllReplicasReadyCondition)
	valhallaStatus.Conditions = []metav1.Condition{
		availableCondition,
		allReplicasReadyCondition,
		reconciliationSuccessCondition,
	}
}

func (status *ValhallaStatus) SetCondition(condition metav1.Condition) {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condition.Type {
			if status.Conditions[i].Status != condition.Status {
				status.Conditions[i].LastTransitionTime = metav1.Now()
			}
			status.Conditions[i].Status = condition.Status
			status.Conditions[i].Reason = condition.Reason
			status.Conditions[i].Message = condition.Message
			break
		}
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Valhalla is the Schema for the valhallas API
type Valhalla struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValhallaSpec   `json:"spec,omitempty"`
	Status ValhallaStatus `json:"status,omitempty"`
}

func (valhalla Valhalla) ChildResourceName(name string) string {
	return strings.TrimSuffix(strings.Join([]string{valhalla.Name, name}, "-"), "-")
}

//+kubebuilder:object:root=true

// ValhallaList contains a list of Valhalla
type ValhallaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Valhalla `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Valhalla{}, &ValhallaList{})
}
