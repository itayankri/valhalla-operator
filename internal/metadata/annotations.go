package metadata

import "strings"

func ReconcileAnnotations(existing map[string]string, defaults ...map[string]string) map[string]string {
	return merge(existing, defaults...)
}

func merge(baseAnnotations map[string]string, maps ...map[string]string) map[string]string {
	annotations := map[string]string{}
	if baseAnnotations != nil {
		annotations = baseAnnotations
	}

	for _, m := range maps {
		for k, v := range m {
			annotations[k] = v
		}
	}

	return annotations
}

func isKubernetesAnnotation(k string) bool {
	return strings.Contains(k, "kubernetes.io") || strings.Contains(k, "k8s.io")
}
