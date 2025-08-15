package patch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		patchString string
		wantIsJSON  bool
		wantErr     string
	}{
		{
			name:        "empty patch string",
			patchString: "",
			wantErr:     "patch string is empty",
		},
		{
			name: "valid JSON patch",
			patchString: `
- op: replace
  path: /spec/replicas
  value: 3
- op: add
  path: /metadata/labels/app
  value: myapp`,
			wantIsJSON: true,
		},
		{
			name: "valid strategic merge patch",
			patchString: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 5`,
			wantIsJSON: false,
		},
		{
			name: "simple strategic merge patch",
			patchString: `
spec:
  replicas: 10`,
			wantIsJSON: false,
		},
		{
			name: "invalid YAML",
			patchString: `
invalid: yaml: content:
  - bad indentation`,
			wantErr: "parsing patch YAML as neither JSON patch nor merge patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := Parse(tt.patchString)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, obj)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, obj)
			assert.Equal(t, tt.wantIsJSON, obj.IsJSON)
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(t *testing.T) string
		wantIsJSON bool
		wantErr    string
	}{
		{
			name: "valid JSON patch file",
			setupFile: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				content := `
- op: replace
  path: /spec/replicas
  value: 3`
				path := filepath.Join(tmpDir, "patch.yaml")
				require.NoError(t, os.WriteFile(path, []byte(content), 0644))
				return path
			},
			wantIsJSON: true,
		},
		{
			name: "valid strategic merge patch file",
			setupFile: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				content := `
apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 5`
				path := filepath.Join(tmpDir, "patch.yaml")
				require.NoError(t, os.WriteFile(path, []byte(content), 0644))
				return path
			},
			wantIsJSON: false,
		},
		{
			name: "file does not exist",
			setupFile: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "nonexistent.yaml")
			},
			wantErr: "reading patch file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFile(t)
			obj, err := ParseFile(path)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, obj)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, obj)
			assert.Equal(t, tt.wantIsJSON, obj.IsJSON)
		})
	}
}

func TestObject_Apply_JSONPatch(t *testing.T) {
	tests := []struct {
		name        string
		patchString string
		resource    map[string]any
		wantErr     string
		checkResult func(t *testing.T, resource map[string]any)
	}{
		{
			name:        "replace operation",
			patchString: `[{"op": "replace", "path": "/spec/replicas", "value": 5}]`,
			resource: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"spec": map[string]any{
					"replicas": 3,
				},
			},
			checkResult: func(t *testing.T, resource map[string]any) {
				t.Helper()
				// The krepe library returns map[string]interface{} for nested maps
				spec, ok := resource["spec"].(map[string]interface{})
				require.True(t, ok)
				// With krepe, numbers keep their original type
				assert.Equal(t, 5, spec["replicas"])
			},
		},
		{
			name:        "add operation with nested path",
			patchString: `[{"op": "add", "path": "/metadata/labels", "value": {"app": "myapp"}}]`,
			resource: map[string]any{
				"metadata": map[string]any{
					"name": "test",
				},
			},
			checkResult: func(t *testing.T, resource map[string]any) {
				t.Helper()
				metadata, ok := resource["metadata"].(map[string]interface{})
				require.True(t, ok)
				labels, ok := metadata["labels"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "myapp", labels["app"])
			},
		},
		{
			name:        "remove operation",
			patchString: `[{"op": "remove", "path": "/spec/template"}]`,
			resource: map[string]any{
				"spec": map[string]any{
					"replicas": 3,
					"template": map[string]any{
						"spec": map[string]any{
							"containers": []any{},
						},
					},
				},
			},
			checkResult: func(t *testing.T, resource map[string]any) {
				t.Helper()
				spec, ok := resource["spec"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, 3, spec["replicas"])
				_, hasTemplate := spec["template"]
				assert.False(t, hasTemplate, "template should be removed")
			},
		},
		{
			name:        "invalid patch operation",
			patchString: `[{"op": "invalid", "path": "/spec"}]`,
			resource: map[string]any{
				"spec": map[string]any{"replicas": 3},
			},
			wantErr: "parsing patch YAML as neither JSON patch nor merge patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := Parse(tt.patchString)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, obj)
				return
			}

			require.NoError(t, err)
			require.True(t, obj.IsJSON, "Should be detected as JSON patch")

			// Make a copy of the resource
			resourceCopy := make(map[string]any)
			for k, v := range tt.resource {
				resourceCopy[k] = v
			}

			result, err := obj.Apply(resourceCopy)
			require.NoError(t, err)
			resourceCopy = result

			if tt.checkResult != nil {
				tt.checkResult(t, resourceCopy)
			}
		})
	}
}

func TestObject_Apply_StrategicMergePatch(t *testing.T) {
	tests := []struct {
		name        string
		patchString string
		resource    map[string]any
		checkResult func(t *testing.T, resource map[string]any)
	}{
		{
			name: "simple merge",
			patchString: `
spec:
  replicas: 7`,
			resource: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"spec": map[string]any{
					"replicas": 3,
				},
			},
			checkResult: func(t *testing.T, resource map[string]any) {
				t.Helper()
				// Check that existing fields are preserved
				assert.Equal(t, "apps/v1", resource["apiVersion"])
				assert.Equal(t, "Deployment", resource["kind"])

				spec, ok := resource["spec"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, 7, spec["replicas"])
			},
		},
		{
			name: "nested merge with preservation",
			patchString: `
metadata:
  labels:
    app: updated
spec:
  replicas: 5`,
			resource: map[string]any{
				"apiVersion": "apps/v1",
				"metadata": map[string]any{
					"name": "test-deployment",
					"labels": map[string]any{
						"version": "v1",
					},
				},
				"spec": map[string]any{
					"replicas": 3,
				},
			},
			checkResult: func(t *testing.T, resource map[string]any) {
				t.Helper()
				metadata, ok := resource["metadata"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "test-deployment", metadata["name"]) // Preserved

				labels, ok := metadata["labels"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "updated", labels["app"]) // Added from patch
				assert.Equal(t, "v1", labels["version"])  // Preserved from original

				spec, ok := resource["spec"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, 5, spec["replicas"]) // Updated from patch
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := Parse(tt.patchString)
			require.NoError(t, err)
			require.False(t, obj.IsJSON, "Should be detected as strategic merge patch")

			// Make a copy of the resource
			resourceCopy := make(map[string]any)
			for k, v := range tt.resource {
				resourceCopy[k] = v
			}

			result, err := obj.Apply(resourceCopy)
			require.NoError(t, err)
			resourceCopy = result

			if tt.checkResult != nil {
				tt.checkResult(t, resourceCopy)
			}
		})
	}
}

func TestIntegration_EndToEnd(t *testing.T) {
	// Test the full workflow: Parse -> Apply
	t.Run("JSON patch end-to-end", func(t *testing.T) {
		patchYAML := `
- op: replace
  path: /spec/replicas
  value: 10
- op: add
  path: /metadata/labels
  value:
    environment: production`

		resource := map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "test-app",
			},
			"spec": map[string]any{
				"replicas": 3,
			},
		}

		patch, err := Parse(patchYAML)
		require.NoError(t, err)
		assert.True(t, patch.IsJSON)

		resource, err = patch.Apply(resource)
		require.NoError(t, err)

		// Check results
		spec := resource["spec"].(map[string]interface{})
		assert.Equal(t, 10, spec["replicas"])

		metadata := resource["metadata"].(map[string]interface{})
		labels := metadata["labels"].(map[string]interface{})
		assert.Equal(t, "production", labels["environment"])
		assert.Equal(t, "test-app", metadata["name"]) // Preserved
	})

	t.Run("strategic merge patch end-to-end", func(t *testing.T) {
		patchYAML := `
metadata:
  annotations:
    deployment.kubernetes.io/revision: "2"
spec:
  replicas: 8
  template:
    spec:
      containers:
      - name: app
        image: nginx:1.21`

		resource := map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "test-app",
				"labels": map[string]any{
					"app": "nginx",
				},
			},
			"spec": map[string]any{
				"replicas": 3,
			},
		}

		patch, err := Parse(patchYAML)
		require.NoError(t, err)
		assert.False(t, patch.IsJSON)

		resource, err = patch.Apply(resource)
		require.NoError(t, err)

		// Check results
		metadata := resource["metadata"].(map[string]any)
		assert.Equal(t, "test-app", metadata["name"]) // Preserved

		labels := metadata["labels"].(map[string]any)
		assert.Equal(t, "nginx", labels["app"]) // Preserved

		annotations := metadata["annotations"].(map[string]any)
		assert.Equal(t, "2", annotations["deployment.kubernetes.io/revision"]) // Added

		spec := resource["spec"].(map[string]any)
		assert.Equal(t, 8, spec["replicas"]) // Updated
	})
}
