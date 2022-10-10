package resource

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const JobSuffix = "builder"
const mapBuilderImage = "itayankri/valhalla-builder:latest"

type JobBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) Job() *JobBuilder {
	return &JobBuilder{builder}
}

func (builder *JobBuilder) Build() (client.Object, error) {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(JobSuffix),
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *JobBuilder) Update(object client.Object) error {
	job := object.(*batchv1.Job)

	job.Spec = batchv1.JobSpec{
		Selector: job.Spec.Selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: job.Spec.Template.ObjectMeta.Labels,
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyOnFailure,
				Containers: []corev1.Container{
					{
						Name:  "map-builder",
						Image: mapBuilderImage,
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("1000M"),
								"cpu":    resource.MustParse("1000m"),
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "ROOT_DIR",
								Value: valhallaDataPath,
							},
							{
								Name:  "PBF_URL",
								Value: builder.Instance.Spec.PBFURL,
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
	}

	if err := controllerutil.SetControllerReference(builder.Instance, job, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
