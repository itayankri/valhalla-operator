package resource

import (
	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceBuilder interface {
	Build() (client.Object, error)
	Update(client.Object) error
	GetPhase() valhallav1alpha1.LifecyclePhase
}

type ValhallaResourceBuilder struct {
	Instance *valhallav1alpha1.Valhalla
	Scheme   *runtime.Scheme
	Phase    valhallav1alpha1.LifecyclePhase
}

func (builder *ValhallaResourceBuilder) ResourceBuilders(phase valhallav1alpha1.LifecyclePhase) []ResourceBuilder {
	builders := []ResourceBuilder{
		builder.Deployment(),
		builder.Service(),
		builder.Job(),
		builder.HorizontalPodAutoscaler(),
		builder.PersistentVolumeClaim(),
		builder.PodDisruptionBudget(),
	}
	return builders
}
