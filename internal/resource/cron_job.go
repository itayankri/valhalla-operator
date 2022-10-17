package resource

import (
	"fmt"

	"github.com/itayankri/valhalla-operator/internal/status"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type CronJobBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) CronJob() *CronJobBuilder {
	return &CronJobBuilder{builder}
}

func (builder *CronJobBuilder) Build() (client.Object, error) {
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(CronJobSuffix),
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *CronJobBuilder) Update(object client.Object) error {
	cronJob := object.(*batchv1.CronJob)

	cronJob.Spec = batchv1.CronJobSpec{
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      builder.Instance.ChildResourceName(CronJobSuffix),
				Namespace: builder.Instance.Namespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyOnFailure,
						Containers: []corev1.Container{
							{
								Name:  builder.Instance.ChildResourceName(CronJobSuffix),
								Image: hirtoricalTrafficDataFetcherImage,
								Resources: corev1.ResourceRequirements{
									Requests: map[corev1.ResourceName]resource.Quantity{
										"memory": resource.MustParse("100M"),
										"cpu":    resource.MustParse("100m"),
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "ROOT_DIR",
										Value: valhallaDataPath,
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      builder.Instance.Name,
										MountPath: valhallaDataPath,
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: builder.Instance.Name,
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: builder.Instance.Name,
										ReadOnly:  false,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(builder.Instance, cronJob, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}

func (*CronJobBuilder) ShouldDeploy(resources []runtime.Object) bool {
	return status.IsPersistentVolumeClaimBound(resources) && status.IsJobCompleted(resources)
}
