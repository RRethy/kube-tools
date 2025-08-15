package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecWrapper is a test implementation of exec.Wrapper.
type mockExecWrapper struct{}

func (m *mockExecWrapper) Command(_ string, _ ...string) *exec.Cmd {
	// Return a command that outputs test YAML
	return exec.Command("echo", `apiVersion: v1
kind: ConfigMap
metadata:
  name: mock-config
data:
  mock: value`)
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    func(t *testing.T, dir string)
		expectedError string
		expectedCount int
	}{
		{
			name: "successful generation",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				// Create generator YAML
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: TestGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./test-generator
  name: test-generator
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "test-generator.yaml"), []byte(generatorYAML), 0644))

				// Create executable generator script
				generatorScript := `#!/bin/sh
cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
type: Opaque
data:
  password: cGFzc3dvcmQ=
EOF`
				scriptPath := filepath.Join(dir, "test-generator")
				require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))
			},
			expectedCount: 2,
		},
		{
			name: "generator with no output",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				// Create generator YAML
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: EmptyGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./empty-generator
  name: empty-generator
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "empty-generator.yaml"), []byte(generatorYAML), 0644))

				// Create executable generator script that produces no output
				generatorScript := `#!/bin/sh
# No output
exit 0`
				scriptPath := filepath.Join(dir, "empty-generator")
				require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))
			},
			expectedCount: 0,
		},
		{
			name: "missing executable",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				// Create generator YAML pointing to non-existent executable
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: MissingGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./missing-generator
  name: missing-generator
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "missing-generator.yaml"), []byte(generatorYAML), 0644))
			},
			expectedError: "executable not found",
		},
		{
			name: "malformed generator YAML",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("not valid yaml: ["), 0644))
			},
			expectedError: "parsing generator YAML",
		},
		{
			name: "missing annotation",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: NoAnnotationGenerator
metadata:
  name: no-annotation
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "no-annotation.yaml"), []byte(generatorYAML), 0644))
			},
			expectedError: "getting annotations",
		},
		{
			name: "file read error",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				// Create a directory instead of a file to trigger read error
				require.NoError(t, os.Mkdir(filepath.Join(dir, "unreadable.yaml"), 0755))
			},
			expectedError: "reading generator file",
		},
		{
			name: "invalid function annotation type",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: InvalidAnnotationType
metadata:
  annotations:
    config.kubernetes.io/function: 123
  name: invalid-annotation-type
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "invalid-annotation.yaml"), []byte(generatorYAML), 0644))
			},
			expectedError: "no config.kubernetes.io/function annotation found",
		},
		{
			name: "malformed exec configuration",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: MalformedExecConfig
metadata:
  annotations:
    config.kubernetes.io/function: "not valid yaml {"
  name: malformed-exec
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "malformed-exec.yaml"), []byte(generatorYAML), 0644))
			},
			expectedError: "parsing exec configuration",
		},
		{
			name: "empty exec path",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: EmptyExecPath
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ""
  name: empty-exec-path
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "empty-exec-path.yaml"), []byte(generatorYAML), 0644))
			},
			expectedError: "no exec path specified",
		},
		{
			name: "generator execution failure",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: FailingGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./failing-generator
  name: failing-generator
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "failing-generator.yaml"), []byte(generatorYAML), 0644))

				// Create executable generator script that fails
				generatorScript := `#!/bin/sh
echo "Error: something went wrong" >&2
exit 1`
				scriptPath := filepath.Join(dir, "failing-generator")
				require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))
			},
			expectedError: "executing generator",
		},
		{
			name: "invalid YAML in generator output",
			setupFiles: func(t *testing.T, dir string) {
				t.Helper()
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: InvalidOutputGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./invalid-output-generator
  name: invalid-output-generator
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "invalid-output-generator.yaml"), []byte(generatorYAML), 0644))

				// Create executable generator script that outputs invalid YAML
				generatorScript := `#!/bin/sh
echo "not valid yaml: {"`
				scriptPath := filepath.Join(dir, "invalid-output-generator")
				require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))
			},
			expectedError: "parsing generator output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()
			tt.setupFiles(t, dir)

			// Create generator
			g := New()

			// Find the generator YAML file
			files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
			require.NoError(t, err)
			require.Len(t, files, 1)

			// Generate resources
			resources, err := g.Generate(files[0])

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resources, tt.expectedCount)
			}
		})
	}
}

func TestGenerateWithMockExec(t *testing.T) {
	// Create temp directory
	dir := t.TempDir()

	// Create generator YAML
	generatorYAML := `apiVersion: stable.shopify.io/v1
kind: TestGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./test-generator
  name: test-generator
spec: {}`
	generatorPath := filepath.Join(dir, "test-generator.yaml")
	require.NoError(t, os.WriteFile(generatorPath, []byte(generatorYAML), 0644))

	// Create a mock executable (won't actually be executed due to our mock)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test-generator"), []byte("#!/bin/sh\necho test"), 0755))

	// Create a mock exec wrapper
	mockWrapper := &mockExecWrapper{}
	g := New(WithExecWrapper(mockWrapper))

	// Generate resources
	resources, err := g.Generate(generatorPath)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	assert.Equal(t, "v1", resources[0]["apiVersion"])
	assert.Equal(t, "ConfigMap", resources[0]["kind"])
}

func TestGenerateEdgeCases(t *testing.T) {
	t.Run("generator with multiple YAML documents including empty ones", func(t *testing.T) {
		dir := t.TempDir()

		// Create generator YAML
		generatorYAML := `apiVersion: stable.shopify.io/v1
kind: MultiDocGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./multi-doc-generator
  name: multi-doc-generator
spec: {}`
		require.NoError(t, os.WriteFile(filepath.Join(dir, "multi-doc-generator.yaml"), []byte(generatorYAML), 0644))

		// Create generator that outputs multiple documents with empty ones
		generatorScript := `#!/bin/sh
cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
data:
  key: value1
---

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
data:
  key: value2
---
EOF`
		scriptPath := filepath.Join(dir, "multi-doc-generator")
		require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))

		g := New()
		resources, err := g.Generate(filepath.Join(dir, "multi-doc-generator.yaml"))
		require.NoError(t, err)
		// Should have 2 resources (empty documents are skipped)
		assert.Len(t, resources, 2)
		assert.Equal(t, "config1", resources[0]["metadata"].(map[string]any)["name"])
		assert.Equal(t, "config2", resources[1]["metadata"].(map[string]any)["name"])
	})
}
