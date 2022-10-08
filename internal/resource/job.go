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

const JobSuffix = ""

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
	name := builder.Instance.ChildResourceName(JobSuffix)
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
						Name:  fmt.Sprintf("%s-%s", name, "map-builder"),
						Image: builder.Instance.Spec.GetImage(),
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("1000M"),
								"cpu":    resource.MustParse("1000m"),
							},
						},
						Command: []string{
							"/bin/sh",
							"-c",
						},
						Args: []string{
							fmt.Sprintf(`
								cd %s && \
								apt update && \
								apt --assume-yes install wget && \
								wget %s && \
								valhalla_build_config --mjolnir-tile-dir ${PWD}/valhalla_tiles --mjolnir-tile-extract ./valhalla_tiles.tar --mjolnir-timezone ./valhalla_tiles/timezones.sqlite --mjolnir-admin ./valhalla_tiles/admins.sqlite > ./conf/valhalla.json && \
								valhalla_build_admins --config ./conf/valhalla.json %s && \
								valhalla_build_timezones > ./valhalla_tiles/timezones.sqlite && \
								valhalla_build_tiles -c ./conf/valhalla.json %s && \
							`,
								valhallaDataPath,
								builder.Instance.Spec.PBFURL,
								builder.Instance.Spec.PBFURL,
								builder.Instance.Spec.PBFURL,
							),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      name,
								MountPath: valhallaDataPath,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: name,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: name,
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
