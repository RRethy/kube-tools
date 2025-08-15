package helm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/RRethy/k8s-tools/kustomizelite/api/v1"
)

func TestTemplate(t *testing.T) {
	t.Helper()

	t.Run("validates helm binary exists", func(t *testing.T) {
		templater, err := NewTemplater("/nonexistent/helm", nil)
		require.Error(t, err)
		assert.Nil(t, templater)
		assert.Contains(t, err.Error(), "helm binary not found")
	})

	t.Run("uses custom helm binary", func(t *testing.T) {
		helperScript := `#!/bin/sh
echo "---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value"
`
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		templater, err := NewTemplater(scriptPath, nil)
		require.NoError(t, err)

		chart := v1.HelmChart{
			Name: "test-chart",
		}

		resources, err := templater.Template(".", chart, nil)
		require.NoError(t, err)
		require.Len(t, resources, 1)
		assert.Equal(t, "ConfigMap", resources[0]["kind"])
		assert.Equal(t, "test-config", resources[0]["metadata"].(map[string]any)["name"])
	})

	t.Run("handles multiple documents", func(t *testing.T) {
		helperScript := `#!/bin/sh
echo "---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
apiVersion: v1
kind: Secret
metadata:
  name: secret1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1"
`
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		templater, err := NewTemplater(scriptPath, nil)
		require.NoError(t, err)

		chart := v1.HelmChart{
			Name: "test-chart",
		}

		resources, err := templater.Template(".", chart, nil)
		require.NoError(t, err)
		require.Len(t, resources, 3)

		assert.Equal(t, "ConfigMap", resources[0]["kind"])
		assert.Equal(t, "Secret", resources[1]["kind"])
		assert.Equal(t, "Deployment", resources[2]["kind"])
	})

	t.Run("handles values files and inline values", func(t *testing.T) {
		helperScript := `#!/bin/sh
echo "---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  values-processed: \"true\""
`
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		templater, err := NewTemplater(scriptPath, nil)
		require.NoError(t, err)

		chart := v1.HelmChart{
			Name:       "test-chart",
			ValuesFile: "values.yaml",
		}

		resources, err := templater.Template(".", chart, nil)
		require.NoError(t, err)
		require.Len(t, resources, 1)
	})
}

func TestGetBinaryFromEnv(t *testing.T) {
	t.Run("returns helm by default", func(t *testing.T) {
		os.Unsetenv("HELM_BINARY_PATH")
		assert.Equal(t, "helm", GetBinaryFromEnv())
	})

	t.Run("returns HELM_BINARY_PATH when set", func(t *testing.T) {
		t.Setenv("HELM_BINARY_PATH", "/custom/helm")
		assert.Equal(t, "/custom/helm", GetBinaryFromEnv())
	})
}

func TestNewTemplater(t *testing.T) {
	t.Run("uses provided binary path", func(t *testing.T) {
		// Create a fake helm binary
		helperScript := `#!/bin/sh
echo "test"
`
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		templater, err := NewTemplater(scriptPath, nil)
		require.NoError(t, err)
		assert.NotNil(t, templater)
	})

	t.Run("returns error for non-existent binary", func(t *testing.T) {
		templater, err := NewTemplater("/non/existent/helm", nil)
		assert.Error(t, err)
		assert.Nil(t, templater)
		assert.Contains(t, err.Error(), "helm binary not found")
	})

	t.Run("returns error for non-existent values file", func(t *testing.T) {
		// Create a fake helm binary
		helperScript := `#!/bin/sh
echo "test"
`
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		templater, err := NewTemplater(scriptPath, []string{"/non/existent/values.yaml"})
		assert.Error(t, err)
		assert.Nil(t, templater)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("accepts existing values files", func(t *testing.T) {
		// Create a fake helm binary
		helperScript := `#!/bin/sh
echo "test"
`
		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fake-helm")
		require.NoError(t, os.WriteFile(scriptPath, []byte(helperScript), 0755))

		// Create a temporary values file
		tmpFile, err := os.CreateTemp(tempDir, "values-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		templater, err := NewTemplater(scriptPath, []string{tmpFile.Name()})
		require.NoError(t, err)
		assert.NotNil(t, templater)
	})
}

func TestParseHelmOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int
	}{
		{
			name: "single document",
			output: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test`,
			expected: 1,
		},
		{
			name: "multiple documents",
			output: `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test1
---
apiVersion: v1
kind: Secret
metadata:
  name: test2`,
			expected: 2,
		},
		{
			name: "empty documents filtered",
			output: `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
---
---`,
			expected: 1,
		},
		{
			name:     "only separator",
			output:   "---",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := parseHelmOutput([]byte(tt.output))
			require.NoError(t, err)
			assert.Len(t, resources, tt.expected)
		})
	}
}

func TestHelmChartParsing(t *testing.T) {
	t.Run("parses helm chart with values file", func(t *testing.T) {
		yamlContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

helmCharts:
- name: test-chart
  valuesFile: values.yaml
  additionalValuesFiles:
  - values-prod.yaml
`
		var kustomization v1.Kustomization
		err := yaml.Unmarshal([]byte(yamlContent), &kustomization)
		require.NoError(t, err)

		assert.Len(t, kustomization.HelmCharts, 1)
		chart := kustomization.HelmCharts[0]
		assert.Equal(t, "test-chart", chart.Name)
		assert.Equal(t, "values.yaml", chart.ValuesFile)
		assert.Equal(t, []string{"values-prod.yaml"}, chart.AdditionalValuesFiles)
	})
}
