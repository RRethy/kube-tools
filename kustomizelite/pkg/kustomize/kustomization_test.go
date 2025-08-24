package kustomize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/RRethy/kube-tools/kustomizelite/api/v1"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/helm"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/testutil"
)

func TestKustomizer_Kustomize(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) string
		wantContent []map[string]any
		wantErr     string
	}{
		{
			name: "valid kustomization.yaml file",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				content := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\n"
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(content), 0644))
				return path
			},
			wantContent: nil, // No resources in the kustomization
		},
		{
			name: "directory without kustomization.yaml",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			wantErr: "stat'ing path",
		},
		{
			name: "directory with kustomization.yaml",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				content := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\n"
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(content), 0644))
				return tmpDir // Return directory, not file
			},
			wantContent: nil, // No resources in the kustomization
		},
		{
			name: "wrong filename",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "wrong.yaml")
				require.NoError(t, os.WriteFile(path, []byte("content"), 0644))
				return path
			},
			wantErr: "is not a kustomization file",
		},
		{
			name: "file does not exist",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "nonexistent.yaml")
			},
			wantErr: "no such file or directory",
		},
		{
			name: "kustomization.yml not supported",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "kustomization.yml")
				require.NoError(t, os.WriteFile(path, []byte("content"), 0644))
				return path
			},
			wantErr: "is not a kustomization file",
		},
		{
			name: "invalid YAML content",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte("invalid: yaml: content:\n  - bad indentation"), 0644))
				return path
			},
			wantErr: "parsing Kustomization YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			content, err := k.Kustomize(path, nil)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, content)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantContent, content)
			}
		})
	}
}

func TestKustomizer_KustomizeWithFixtures(t *testing.T) {
	tests := []struct {
		name        string
		fixturePath string
		wantKeys    []string
		wantErr     string
	}{
		{
			name:        "basic kustomization",
			fixturePath: testutil.ValidFixturePath("basic"),
			wantKeys:    []string{"Deployment", "Service"}, // Has deployment and service resources
		},
		{
			name:        "kustomization with generator",
			fixturePath: filepath.Join("..", "..", "testdata", "fixtures", "generator-example"),
			wantKeys:    []string{"Deployment", "ConfigMap"}, // Has deployment + generated configmap
		},
		{
			name:        "kustomization with namespace",
			fixturePath: testutil.ValidFixturePath("with-namespace"),
			wantKeys:    []string{"ConfigMap"}, // Has configmap.yaml resource
		},
		{
			name:        "kustomization with patches",
			fixturePath: testutil.ValidFixturePath("with-patches"),
			wantKeys:    []string{"Deployment", "Service"}, // Has deployment and service resources
		},
		{
			name:        "kustomization with components",
			fixturePath: testutil.ValidFixturePath("with-components"),
			wantKeys:    []string{"Deployment", "Service", "ConfigMap", "Secret"}, // Has resources + components
		},
		{
			name:        "empty kustomization",
			fixturePath: testutil.ValidFixturePath("empty"),
			wantKeys:    nil, // No resources in the kustomization
		},
		{
			name:        "malformed yaml",
			fixturePath: testutil.InvalidFixturePath("malformed-yaml"),
			wantErr:     "parsing Kustomization YAML",
		},
		{
			name:        "invalid yaml syntax",
			fixturePath: testutil.InvalidFixturePath("invalid-syntax"),
			wantErr:     "parsing Kustomization YAML",
		},
		{
			name:        "tab indentation",
			fixturePath: testutil.InvalidFixturePath("tab-indentation"),
			wantErr:     "parsing Kustomization YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)

			content, err := k.Kustomize(tt.fixturePath, nil)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, content)
			} else {
				require.NoError(t, err)
				if tt.wantKeys != nil {
					// Check we have the expected resource kinds
					kinds := make(map[string]bool)
					for _, doc := range content {
						if kind, ok := doc["kind"].(string); ok {
							kinds[kind] = true
						}
					}
					for _, wantKind := range tt.wantKeys {
						assert.True(t, kinds[wantKind], "Expected to find kind %s", wantKind)
					}
				} else {
					assert.Nil(t, content)
				}
			}
		})
	}
}

func TestKustomizer_ComponentDetection(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		wantDocCount int
		checkDocs    func(t *testing.T, docs []map[string]any)
	}{
		{
			name: "detects component kind and processes all resources",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a component kustomization.yaml
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
resources:
  - deployment.yaml
  - service.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment resource
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				// Create service resource
				serviceContent := `apiVersion: v1
kind: Service
metadata:
  name: test-service
`
				svcPath := filepath.Join(tmpDir, "service.yaml")
				require.NoError(t, os.WriteFile(svcPath, []byte(serviceContent), 0644))

				return path
			},
			wantDocCount: 2, // deployment + service (component excluded)
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				// Check we have all resources (component excluded)
				kinds := make(map[string]bool)
				for _, doc := range docs {
					kinds[doc["kind"].(string)] = true
				}
				assert.False(t, kinds["Component"]) // Component should not be in results
				assert.True(t, kinds["Deployment"])
				assert.True(t, kinds["Service"])
			},
		},
		{
			name: "regular kustomization is not treated as component",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a regular kustomization.yaml
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment resource
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				return path
			},
			wantDocCount: 1, // deployment only (kustomization excluded)
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				// Should only have deployment, not kustomization
				assert.Equal(t, "Deployment", docs[0]["kind"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			content, err := k.Kustomize(path, nil)
			require.NoError(t, err)
			assert.Len(t, content, tt.wantDocCount)

			if tt.checkDocs != nil {
				tt.checkDocs(t, content)
			}
		})
	}
}

func TestKustomizer_RecursiveResources(t *testing.T) {
	tests := []struct {
		name         string
		fixturePath  string
		wantDocCount int
		checkDocs    func(t *testing.T, docs []map[string]any)
	}{
		{
			name:         "kustomization with resource files",
			fixturePath:  testutil.ValidFixturePath("with-resources"),
			wantDocCount: 2, // deployment + service (kustomization excluded)
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				// Find deployment and service (no kustomization in results)
				var hasDeployment, hasService bool
				for _, doc := range docs {
					switch doc["kind"] {
					case "Deployment":
						hasDeployment = true
						metadata := doc["metadata"].(map[string]any)
						assert.Equal(t, "my-app", metadata["name"])
					case "Service":
						hasService = true
						metadata := doc["metadata"].(map[string]any)
						assert.Equal(t, "my-app-service", metadata["name"])
					}
				}
				assert.True(t, hasDeployment, "Should have found Deployment")
				assert.True(t, hasService, "Should have found Service")
			},
		},
		{
			name:         "kustomization with bases (bases no longer processed)",
			fixturePath:  testutil.ValidFixturePath("with-bases"),
			wantDocCount: 1, // ingress only (kustomization and bases excluded)
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				// Count different resource types
				kindCounts := make(map[string]int)
				for _, doc := range docs {
					kindCounts[doc["kind"].(string)]++
				}

				assert.Equal(t, 0, kindCounts["Kustomization"], "Should have 0 Kustomization docs (excluded from results)")
				assert.Equal(t, 1, kindCounts["Ingress"], "Should have 1 Ingress")
				// ConfigMap from base should not be present
				assert.Equal(t, 0, kindCounts["ConfigMap"], "Should have 0 ConfigMap (from base, not processed)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)

			content, err := k.Kustomize(tt.fixturePath, nil)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(content), tt.wantDocCount)

			if tt.checkDocs != nil {
				tt.checkDocs(t, content)
			}
		})
	}
}

func TestKustomizer_CommonLabelsAndAnnotations(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		checkDocs func(t *testing.T, docs []map[string]any)
	}{
		{
			name: "applies common labels to all resources",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with common labels
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: myapp
  environment: production
resources:
  - deployment.yaml
  - service.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment without labels
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				// Create service with existing labels (should be merged)
				serviceContent := `apiVersion: v1
kind: Service
metadata:
  name: test-service
  labels:
    component: backend
`
				svcPath := filepath.Join(tmpDir, "service.yaml")
				require.NoError(t, os.WriteFile(svcPath, []byte(serviceContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 2)

				for _, doc := range docs {
					metadata, ok := doc["metadata"].(map[string]any)
					require.True(t, ok, "Resource should have metadata")

					labels, ok := metadata["labels"].(map[string]any)
					require.True(t, ok, "Resource should have labels")

					// Check that common labels are applied
					assert.Equal(t, "myapp", labels["app"],
						"Resource %s should have app label", doc["kind"])
					assert.Equal(t, "production", labels["environment"],
						"Resource %s should have environment label", doc["kind"])

					// Check that existing labels are preserved
					if doc["kind"] == "Service" {
						assert.Equal(t, "backend", labels["component"],
							"Service should preserve existing component label")
					}
				}
			},
		},
		{
			name: "applies common annotations to all resources",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with common annotations
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonAnnotations:
  managed-by: kustomize
  version: "1.0"
resources:
  - configmap.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create configmap without annotations
				configmapContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`
				cmPath := filepath.Join(tmpDir, "configmap.yaml")
				require.NoError(t, os.WriteFile(cmPath, []byte(configmapContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				annotations, ok := metadata["annotations"].(map[string]any)
				require.True(t, ok, "Resource should have annotations")

				// Check that common annotations are applied
				assert.Equal(t, "kustomize", annotations["managed-by"])
				assert.Equal(t, "1.0", annotations["version"])
			},
		},
		{
			name: "applies labels, annotations, and namespace together",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with all transformations
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: test-namespace
commonLabels:
  app: myapp
  tier: backend
commonAnnotations:
  description: "Test application"
resources:
  - deployment.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment with minimal metadata
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				// Check namespace
				assert.Equal(t, "test-namespace", metadata["namespace"])

				// Check labels
				labels, ok := metadata["labels"].(map[string]any)
				require.True(t, ok, "Resource should have labels")
				assert.Equal(t, "myapp", labels["app"])
				assert.Equal(t, "backend", labels["tier"])

				// Check annotations
				annotations, ok := metadata["annotations"].(map[string]any)
				require.True(t, ok, "Resource should have annotations")
				assert.Equal(t, "Test application", annotations["description"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			content, err := k.Kustomize(path, nil)
			require.NoError(t, err)

			tt.checkDocs(t, content)
		})
	}
}

func TestKustomizer_NamePrefixAndSuffix(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   string
		checkDocs func(t *testing.T, docs []map[string]any)
	}{
		{
			name: "applies namePrefix to all resources with metadata.name",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namePrefix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: dev-
resources:
  - deployment.yaml
  - service.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				// Create service
				serviceContent := `apiVersion: v1
kind: Service
metadata:
  name: my-service
`
				svcPath := filepath.Join(tmpDir, "service.yaml")
				require.NoError(t, os.WriteFile(svcPath, []byte(serviceContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 2)

				for _, doc := range docs {
					metadata, ok := doc["metadata"].(map[string]any)
					require.True(t, ok, "Resource should have metadata")

					name, ok := metadata["name"].(string)
					require.True(t, ok, "Resource should have name")

					// Check that namePrefix is applied
					assert.True(t, strings.HasPrefix(name, "dev-"),
						"Resource name %s should have prefix 'dev-'", name)
				}
			},
		},
		{
			name: "applies nameSuffix to all resources with metadata.name",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with nameSuffix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameSuffix: -prod
resources:
  - deployment.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				name, ok := metadata["name"].(string)
				require.True(t, ok, "Resource should have name")

				// Check that nameSuffix is applied
				assert.Equal(t, "my-app-prod", name)
			},
		},
		{
			name: "applies both namePrefix and nameSuffix",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with both prefix and suffix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: staging-
nameSuffix: -v2
resources:
  - configmap.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create configmap
				configmapContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: config
data:
  key: value
`
				cmPath := filepath.Join(tmpDir, "configmap.yaml")
				require.NoError(t, os.WriteFile(cmPath, []byte(configmapContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				name, ok := metadata["name"].(string)
				require.True(t, ok, "Resource should have name")

				// Check that both prefix and suffix are applied
				assert.Equal(t, "staging-config-v2", name)
			},
		},
		{
			name: "applies namePrefix to metadata.generateName",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namePrefix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: test-
resources:
  - job.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create job with generateName
				jobContent := `apiVersion: batch/v1
kind: Job
metadata:
  generateName: backup-job-
`
				jobPath := filepath.Join(tmpDir, "job.yaml")
				require.NoError(t, os.WriteFile(jobPath, []byte(jobContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				generateName, ok := metadata["generateName"].(string)
				require.True(t, ok, "Resource should have generateName")

				// Check that namePrefix is applied to generateName
				assert.Equal(t, "test-backup-job-", generateName)
			},
		},
		{
			name: "applies nameSuffix to metadata.generateName",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with nameSuffix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameSuffix: -daily
resources:
  - job.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create job with generateName
				jobContent := `apiVersion: batch/v1
kind: Job
metadata:
  generateName: backup-
`
				jobPath := filepath.Join(tmpDir, "job.yaml")
				require.NoError(t, os.WriteFile(jobPath, []byte(jobContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				generateName, ok := metadata["generateName"].(string)
				require.True(t, ok, "Resource should have generateName")

				// Check that nameSuffix is applied to generateName
				assert.Equal(t, "backup--daily", generateName)
			},
		},
		{
			name: "applies to resources with both name and generateName",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with both prefix and suffix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: pre-
nameSuffix: -post
resources:
  - resource.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create resource with both name and generateName
				resourceContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: myconfig
  generateName: config-
`
				resPath := filepath.Join(tmpDir, "resource.yaml")
				require.NoError(t, os.WriteFile(resPath, []byte(resourceContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				// Check name
				name, ok := metadata["name"].(string)
				require.True(t, ok, "Resource should have name")
				assert.Equal(t, "pre-myconfig-post", name)

				// Check generateName
				generateName, ok := metadata["generateName"].(string)
				require.True(t, ok, "Resource should have generateName")
				assert.Equal(t, "pre-config--post", generateName)
			},
		},
		{
			name: "handles resources without metadata gracefully",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namePrefix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: test-
resources:
  - minimal.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create minimal resource without metadata
				minimalContent := `apiVersion: v1
kind: ConfigMap
`
				minPath := filepath.Join(tmpDir, "minimal.yaml")
				require.NoError(t, os.WriteFile(minPath, []byte(minimalContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)
				// Should not error on resources without metadata
				assert.Equal(t, "ConfigMap", docs[0]["kind"])
			},
		},
		{
			name: "returns error when name field has wrong type",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namePrefix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: test-
resources:
  - bad-type.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create resource with name as wrong type
				badContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: 123  # Number instead of string
`
				badPath := filepath.Join(tmpDir, "bad-type.yaml")
				require.NoError(t, os.WriteFile(badPath, []byte(badContent), 0644))
				return path
			},
			wantErr: "getting metadata.name: value at path",
			checkDocs: func(_ *testing.T, _ []map[string]any) {
				// This case expects an error to be handled before getting here
			},
		},
		{
			name: "returns error when generateName field has wrong type",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namePrefix
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: test-
resources:
  - bad-gen.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create resource with generateName as wrong type
				badContent := `apiVersion: v1
kind: ConfigMap
metadata:
  generateName: false  # Boolean instead of string
`
				badPath := filepath.Join(tmpDir, "bad-gen.yaml")
				require.NoError(t, os.WriteFile(badPath, []byte(badContent), 0644))
				return path
			},
			wantErr: "getting metadata.generateName: value at path",
			checkDocs: func(_ *testing.T, _ []map[string]any) {
				// This case expects an error to be handled before getting here
			},
		},
		{
			name: "combined with namespace, labels, and annotations",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with all transformations
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: test-ns
namePrefix: dev-
nameSuffix: -v1
commonLabels:
  env: dev
commonAnnotations:
  note: test
resources:
  - deployment.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				// Check all transformations are applied
				assert.Equal(t, "dev-app-v1", metadata["name"])
				assert.Equal(t, "test-ns", metadata["namespace"])

				labels, ok := metadata["labels"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "dev", labels["env"])

				annotations, ok := metadata["annotations"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "test", annotations["note"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			content, err := k.Kustomize(path, nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				if tt.checkDocs != nil {
					tt.checkDocs(t, content)
				}
			}
		})
	}
}

func TestKustomizer_NamespaceApplication(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		checkDocs func(t *testing.T, docs []map[string]any)
	}{
		{
			name: "applies namespace to all resources",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namespace
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: test-namespace
resources:
  - deployment.yaml
  - service.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create deployment without namespace
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				// Create service with existing namespace (should be overridden)
				serviceContent := `apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: old-namespace
`
				svcPath := filepath.Join(tmpDir, "service.yaml")
				require.NoError(t, os.WriteFile(svcPath, []byte(serviceContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 2)

				for _, doc := range docs {
					metadata, ok := doc["metadata"].(map[string]any)
					require.True(t, ok, "Resource should have metadata")

					// Check that namespace is applied
					assert.Equal(t, "test-namespace", metadata["namespace"],
						"Resource %s should have namespace applied", doc["kind"])
				}
			},
		},
		{
			name: "does not apply namespace when not specified",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml without namespace
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - configmap.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create configmap without namespace
				configmapContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`
				cmPath := filepath.Join(tmpDir, "configmap.yaml")
				require.NoError(t, os.WriteFile(cmPath, []byte(configmapContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata")

				// Check that namespace is not present
				_, hasNamespace := metadata["namespace"]
				assert.False(t, hasNamespace, "Resource should not have namespace when not specified in kustomization")
			},
		},
		{
			name: "creates metadata if not present when applying namespace",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create a kustomization.yaml with namespace
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: metadata-test
resources:
  - minimal.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create resource without metadata section
				minimalContent := `apiVersion: v1
kind: ConfigMap
`
				minPath := filepath.Join(tmpDir, "minimal.yaml")
				require.NoError(t, os.WriteFile(minPath, []byte(minimalContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 1)

				metadata, ok := docs[0]["metadata"].(map[string]any)
				require.True(t, ok, "Resource should have metadata created")

				// Check that namespace is applied
				assert.Equal(t, "metadata-test", metadata["namespace"],
					"Resource should have namespace even when metadata was initially missing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			content, err := k.Kustomize(path, nil)
			require.NoError(t, err)

			tt.checkDocs(t, content)
		})
	}
}

func TestKustomizer_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   string
	}{
		{
			name: "handles YAML parsing error for invalid kustomization file",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "kustomization.yaml")
				// Create file with invalid YAML content
				require.NoError(t, os.WriteFile(path, []byte("invalid: yaml: content"), 0644))
				return path
			},
			wantErr: "parsing Kustomization YAML",
		},
		{
			name: "handles stat error for missing resource",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - missing-file.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))
				return path
			},
			wantErr: "processing resource missing-file.yaml: stat'ing resource",
		},
		{
			name: "handles invalid YAML in resource file",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - invalid.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create invalid YAML resource
				resourcePath := filepath.Join(tmpDir, "invalid.yaml")
				require.NoError(t, os.WriteFile(resourcePath, []byte("invalid:\n  - yaml\n    bad: indentation"), 0644))

				return path
			},
			wantErr: "processing resource invalid.yaml: parsing YAML",
		},
		{
			name: "handles missing component directory",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
components:
  - missing-component
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))
				return path
			},
			wantErr: "processing component missing-component: stat'ing resource",
		},
		{
			name: "handles component that is a file not directory",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
components:
  - component-file.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create a file instead of directory for component
				componentPath := filepath.Join(tmpDir, "component-file.yaml")
				require.NoError(t, os.WriteFile(componentPath, []byte("content"), 0644))

				return path
			},
			wantErr: "processing component component-file.yaml: component",
		},
		{
			name: "handles component with invalid kustomization",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
components:
  - bad-component
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create component directory with invalid kustomization
				componentDir := filepath.Join(tmpDir, "bad-component")
				require.NoError(t, os.Mkdir(componentDir, 0755))
				componentKustPath := filepath.Join(componentDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(componentKustPath, []byte("invalid: yaml: content"), 0644))

				return path
			},
			wantErr: "processing component bad-component: kustomizing component directory",
		},
		{
			name: "collects multiple errors",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - missing1.yaml
  - missing2.yaml
components:
  - missing-comp
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))
				return path
			},
			wantErr: "processing resource missing1.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			_, err2 := k.Kustomize(path, nil)
			require.Error(t, err2)
			assert.Contains(t, err2.Error(), tt.wantErr)
		})
	}
}

func TestKustomizer_Generators(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   string
		checkDocs func(t *testing.T, docs []map[string]any)
	}{
		{
			name: "processes generators successfully",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create kustomization with generators
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - test-generator.yaml
resources:
  - deployment.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

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
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test-generator.yaml"), []byte(generatorYAML), 0644))

				// Create generator script
				generatorScript := `#!/bin/sh
cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: generated-config
data:
  generated: "true"
EOF`
				scriptPath := filepath.Join(tmpDir, "test-generator")
				require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))

				// Create regular resource
				deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`
				depPath := filepath.Join(tmpDir, "deployment.yaml")
				require.NoError(t, os.WriteFile(depPath, []byte(deploymentContent), 0644))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 2) // deployment + generated configmap

				// Find the generated ConfigMap
				var foundGenerated bool
				for _, doc := range docs {
					if doc["kind"] == "ConfigMap" {
						metadata := doc["metadata"].(map[string]any)
						if metadata["name"] == "generated-config" {
							foundGenerated = true
							data := doc["data"].(map[string]any)
							assert.Equal(t, "true", data["generated"])
						}
					}
				}
				assert.True(t, foundGenerated, "Should find generated ConfigMap")
			},
		},
		{
			name: "handles generator errors",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create kustomization with failing generator
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - failing-generator.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create generator YAML
				generatorYAML := `apiVersion: stable.shopify.io/v1
kind: FailingGenerator
metadata:
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./nonexistent-generator
  name: failing-generator
spec: {}`
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "failing-generator.yaml"), []byte(generatorYAML), 0644))

				return path
			},
			wantErr: "processing generator failing-generator.yaml",
		},
		{
			name: "applies transformations to generated resources",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Create kustomization with generators and transformations
				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: test-namespace
commonLabels:
  app: generated-app
generators:
  - test-generator.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

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
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test-generator.yaml"), []byte(generatorYAML), 0644))

				// Create generator script that outputs resources without namespace/labels
				generatorScript := `#!/bin/sh
cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: generated-config
data:
  key: value
---
apiVersion: v1
kind: Secret
metadata:
  name: generated-secret
type: Opaque
data:
  password: cGFzc3dvcmQ=
EOF`
				scriptPath := filepath.Join(tmpDir, "test-generator")
				require.NoError(t, os.WriteFile(scriptPath, []byte(generatorScript), 0755))

				return path
			},
			checkDocs: func(t *testing.T, docs []map[string]any) {
				t.Helper()
				require.Len(t, docs, 2) // configmap + secret

				// Check that transformations are applied to all generated resources
				for _, doc := range docs {
					metadata := doc["metadata"].(map[string]any)

					// Check namespace
					assert.Equal(t, "test-namespace", metadata["namespace"])

					// Check labels
					labels := metadata["labels"].(map[string]any)
					assert.Equal(t, "generated-app", labels["app"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use real generator but fake helm templater
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			content, err := k.Kustomize(path, nil)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				if tt.checkDocs != nil {
					tt.checkDocs(t, content)
				}
			}
		})
	}
}

func TestKustomizer_MaputilsErrors(t *testing.T) {
	// Test errors from maputils operations
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   string
	}{
		{
			name: "handles nil resource when setting namespace",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: test-namespace
resources:
  - null-resource.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				// Create a YAML file that unmarshals to nil
				nullPath := filepath.Join(tmpDir, "null-resource.yaml")
				require.NoError(t, os.WriteFile(nullPath, []byte("null"), 0644))

				return path
			},
			wantErr: "setting namespace on resource",
		},
		{
			name: "handles nil resource when merging labels",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: test
resources:
  - null-resource.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				nullPath := filepath.Join(tmpDir, "null-resource.yaml")
				require.NoError(t, os.WriteFile(nullPath, []byte("null"), 0644))

				return path
			},
			wantErr: "merging common labels on resource",
		},
		{
			name: "handles nil resource when merging annotations",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonAnnotations:
  note: test
resources:
  - null-resource.yaml
`
				path := filepath.Join(tmpDir, "kustomization.yaml")
				require.NoError(t, os.WriteFile(path, []byte(kustomizationContent), 0644))

				nullPath := filepath.Join(tmpDir, "null-resource.yaml")
				require.NoError(t, os.WriteFile(nullPath, []byte("null"), 0644))

				return path
			},
			wantErr: "merging common annotations on resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fake helm templater for testing
			fakeTemplater := &helm.FakeTemplater{
				TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
					return nil, nil
				},
			}
			k, err := NewKustomize(nil, WithHelmTemplater(fakeTemplater))
			require.NoError(t, err)
			path := tt.setupFunc(t)

			_, err2 := k.Kustomize(path, nil)
			require.Error(t, err2)
			assert.Contains(t, err2.Error(), tt.wantErr)
		})
	}
}
