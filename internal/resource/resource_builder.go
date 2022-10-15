package resource

import (
	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceBuilder interface {
	Build() (client.Object, error)
	Update(client.Object) error
	ShouldDeploy(resources []runtime.Object) bool
}

type ValhallaResourceBuilder struct {
	Instance *valhallav1alpha1.Valhalla
	Scheme   *runtime.Scheme
}

func (builder *ValhallaResourceBuilder) ResourceBuilders() []ResourceBuilder {
	builders := []ResourceBuilder{
		builder.PersistentVolumeClaim(),
		builder.Job(),
		builder.Deployment(),
		builder.Service(),
		builder.HorizontalPodAutoscaler(),
		builder.PodDisruptionBudget(),
	}
	return builders
}
