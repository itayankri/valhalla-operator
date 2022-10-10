package resource

import (
	"fmt"

	"github.com/itayankri/valhalla-operator/internal/status"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const HorizontalPodAutoscalerSuffix = ""

type HorizontalPodAutoscalerBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) HorizontalPodAutoscaler() *HorizontalPodAutoscalerBuilder {
	return &HorizontalPodAutoscalerBuilder{builder}
}

func (builder *HorizontalPodAutoscalerBuilder) Build() (client.Object, error) {
	return &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(HorizontalPodAutoscalerSuffix),
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *HorizontalPodAutoscalerBuilder) Update(object client.Object) error {
	name := builder.Instance.ChildResourceName(HorizontalPodAutoscalerSuffix)
	hpa := object.(*autoscalingv1.HorizontalPodAutoscaler)

	targetCPUUtilizationPercentage := int32(85)

	hpa.Spec.ScaleTargetRef = autoscalingv1.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       name,
		APIVersion: "apps/v1",
	}
	hpa.Spec.MinReplicas = builder.Instance.Spec.MinReplicas
	hpa.Spec.MaxReplicas = *builder.Instance.Spec.MaxReplicas
	hpa.Spec.TargetCPUUtilizationPercentage = &targetCPUUtilizationPercentage

	if err := controllerutil.SetControllerReference(builder.Instance, hpa, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}

func (*HorizontalPodAutoscalerBuilder) ShouldDeploy(resources []runtime.Object) bool {
	return status.IsJobCompleted(resources)
}
