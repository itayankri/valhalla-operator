package resource

import (
	"fmt"

	valhallav1alpha1 "github.com/itayankri/valhalla-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const JobSuffix = "builder"

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
								mkdir valhalla_tiles conf
								valhalla_build_config --mjolnir-tile-dir /data/valhalla_tiles \
									--mjolnir-tile-extract /data/valhalla_tiles.tar \
									--mjolnir-timezone /data/valhalla_tiles/timezones.sqlite \
									--mjolnir-admin /data/valhalla_tiles/admins.sqlite \
									--mjolnir-traffic-extract /data/traffic.tar> /data/conf/valhalla.json && \
								valhalla_build_admins --config ./conf/valhalla.json %s && \
								valhalla_build_timezones > ./valhalla_tiles/timezones.sqlite && \
								valhalla_build_tiles --config ./conf/valhalla.json %s && \
								find valhalla_tiles | sort -n | tar -cf "valhalla_tiles.tar" --no-recursion -T -
							`,
								valhallaDataPath,
								builder.Instance.Spec.PBFURL,
								builder.Instance.Spec.GetPbfFileName(),
								builder.Instance.Spec.GetPbfFileName(),
							),
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

func (_ *JobBuilder) GetPhase() valhallav1alpha1.LifecyclePhase {
	return valhallav1alpha1.MapBuilding
}
