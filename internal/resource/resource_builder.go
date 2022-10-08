package resource

import (
	valhallav1alpha1 "github.com/itayankri/Heimdall/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceBuilder interface {
	Build() (client.Object, error)
	Update(client.Object) error
}

type ValhallaResourceBuilder struct {
	Instance *valhallav1alpha1.Valhalla
	Scheme   *runtime.Scheme
}

func (builder *ValhallaResourceBuilder) ResourceBuilders() []ResourceBuilder {
	builders := []ResourceBuilder{
		builder.Deployment(),
		builder.Service(),
		builder.Job(),
		builder.HorizontalPodAutoscaler(),
		builder.PersistentVolumeClaim(),
	}
	return builders
}
