package resource

import (
	"fmt"

	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const PodDisruptionBudgetSuffix = ""

type PodDisruptionBudgetBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) PodDisruptionBudget() *PodDisruptionBudgetBuilder {
	return &PodDisruptionBudgetBuilder{builder}
}

func (builder *PodDisruptionBudgetBuilder) Build() (client.Object, error) {
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(HorizontalPodAutoscalerSuffix),
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *PodDisruptionBudgetBuilder) Update(object client.Object) error {
	name := builder.Instance.ChildResourceName(PodDisruptionBudgetSuffix)
	pdb := object.(*policyv1beta1.PodDisruptionBudget)

	pdb.Spec.MinAvailable = builder.Instance.Spec.GetMinAvailable()
	pdb.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": name,
		},
	}

	if err := controllerutil.SetControllerReference(builder.Instance, pdb, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}

func (*PodDisruptionBudgetBuilder) GetPhase() valhallav1alpha1.LifecyclePhase {
	return valhallav1alpha1.Serving
}
