//go:build tools
// +build tools

package tools

import (
	_ "helm.sh/helm/v3"
	_ "sigs.k8s.io/kind"
)
