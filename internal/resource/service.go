package resource

import (
	"fmt"

	"github.com/itayankri/valhalla-operator/internal/metadata"
	"github.com/itayankri/valhalla-operator/internal/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) Service() *ServiceBuilder {
	return &ServiceBuilder{builder}
}

func (builder *ServiceBuilder) Build() (client.Object, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(ServiceSuffix),
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *ServiceBuilder) Update(object client.Object) error {
	name := builder.Instance.ChildResourceName(ServiceSuffix)

	service := object.(*corev1.Service)

	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:     "default",
			Protocol: corev1.ProtocolTCP,
			Port:     80,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: containerPort,
			},
		},
	}
	service.Spec.Selector = map[string]string{
		"app": name,
	}

	if builder.Instance.Spec.Service != nil {
		service.Spec.Type = builder.Instance.Spec.Service.Type
	}

	if err := controllerutil.SetControllerReference(builder.Instance, service, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}

func (*ServiceBuilder) ShouldDeploy(resources []runtime.Object) bool {
	return status.IsPersistentVolumeClaimBound(resources) && status.IsJobCompleted(resources)
}

func (builder *ServiceBuilder) setAnnotations(service *corev1.Service) {
	if builder.Instance.Spec.Service.Annotations != nil {
		service.Annotations = metadata.ReconcileAnnotations(service.Annotations, builder.Instance.Spec.Service.Annotations)
	}
}
