package resource

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const PersistentVolumeClaimSuffix = ""

type PersistentVolumeClaimBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) PersistentVolumeClaim() *PersistentVolumeClaimBuilder {
	return &PersistentVolumeClaimBuilder{builder}
}

func (builder *PersistentVolumeClaimBuilder) Build() (client.Object, error) {
	name := builder.Instance.ChildResourceName(PersistentVolumeClaimSuffix)
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: builder.Instance.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: *builder.Instance.Spec.Persistence.Storage,
				},
			},
			VolumeName:       "",
			StorageClassName: &builder.Instance.Spec.Persistence.StorageClassName,
		},
	}, nil
}

func (builder *PersistentVolumeClaimBuilder) Update(object client.Object) error {
	pvc := object.(*corev1.PersistentVolumeClaim)

	if err := controllerutil.SetControllerReference(builder.Instance, pvc, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}

func (*PersistentVolumeClaimBuilder) ShouldDeploy(resources []runtime.Object) bool {
	return true
}
