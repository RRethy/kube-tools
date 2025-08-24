package kustomize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/RRethy/kube-tools/kustomizelite/api/v1"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/helm"
)

func TestKustomizeWithHelm(t *testing.T) {
	t.Run("processes helm charts in kustomization", func(t *testing.T) {
		tempDir := t.TempDir()

		helperScript := `#!/bin/sh
echo "---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  type: ClusterIP"
`
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		// Set env var which will be read by NewKustomize()
		t.Setenv("HELM_BINARY_PATH", scriptPath)

		kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: production

helmCharts:
- name: nginx
  releaseName: my-nginx
`
		kustomizationPath := filepath.Join(tempDir, "kustomization.yaml")
		require.NoError(t, os.WriteFile(kustomizationPath, []byte(kustomizationContent), 0644))

		k, err := NewKustomize(nil)
		require.NoError(t, err)
		resources, err := k.Kustomize(tempDir, nil)
		require.NoError(t, err)
		require.Len(t, resources, 1)

		svc := resources[0]
		assert.Equal(t, "Service", svc["kind"])
		assert.Equal(t, "nginx-service", svc["metadata"].(map[string]any)["name"])
		assert.Equal(t, "production", svc["metadata"].(map[string]any)["namespace"])
	})

	t.Run("processes helm charts with fake templater", func(t *testing.T) {
		fakeTemplater := &helm.FakeTemplater{
			TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
				return []map[string]any{
					{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]any{
							"name": "fake-service",
						},
					},
				}, nil
			},
		}

		k := &kustomization{
			helmTemplater: fakeTemplater,
		}

		tempDir := t.TempDir()
		kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: test-ns

helmCharts:
- name: fake-chart
`
		kustomizationPath := filepath.Join(tempDir, "kustomization.yaml")
		require.NoError(t, os.WriteFile(kustomizationPath, []byte(kustomizationContent), 0644))

		resources, err := k.Kustomize(tempDir, nil)
		require.NoError(t, err)
		require.Len(t, resources, 1)

		assert.Equal(t, "Service", resources[0]["kind"])
		assert.Equal(t, "fake-service", resources[0]["metadata"].(map[string]any)["name"])
		assert.Equal(t, "test-ns", resources[0]["metadata"].(map[string]any)["namespace"])
	})
}
