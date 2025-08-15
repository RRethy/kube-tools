package build

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	v1 "github.com/RRethy/k8s-tools/kustomizelite/api/v1"
	"github.com/RRethy/k8s-tools/kustomizelite/pkg/helm"
	"github.com/RRethy/k8s-tools/kustomizelite/pkg/kustomize"
)

func TestBuildCLI(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "..", "testdata", "fixtures")

	tests := []struct {
		name        string
		path        string
		wantOutputs []string
		wantError   bool
		wantErrMsg  string
	}{
		{
			name:        "valid simple kustomization",
			path:        filepath.Join(testdataDir, "valid-simple", "kustomization.yaml"),
			wantOutputs: []string{
				// Empty kustomization produces no output
			},
		},
		{
			name:        "valid kustomization with directory path",
			path:        filepath.Join(testdataDir, "valid-simple"),
			wantOutputs: []string{
				// Empty kustomization produces no output
			},
		},
		{
			name: "valid kustomization with resources",
			path: filepath.Join(testdataDir, "valid-with-resources", "kustomization.yaml"),
			wantOutputs: []string{
				"apiVersion: apps/v1",
				"kind: Deployment",
				"---",
				"kind: Service",
			},
		},
		{
			name: "valid kustomization with patches",
			path: filepath.Join(testdataDir, "valid-with-patches", "kustomization.yaml"),
			wantOutputs: []string{
				"apiVersion: apps/v1",
				"kind: Deployment",
				"replicas: 5",
			},
		},
		{
			name: "valid kustomization with components",
			path: filepath.Join(testdataDir, "valid-with-components", "kustomization.yaml"),
			wantOutputs: []string{
				"kind: Deployment",
				"---",
				"kind: ConfigMap",
			},
		},
		{
			name:       "invalid YAML",
			path:       filepath.Join(testdataDir, "invalid-yaml", "kustomization.yaml"),
			wantError:  true,
			wantErrMsg: "parsing Kustomization YAML",
		},
		{
			name:       "not a kustomization file",
			path:       filepath.Join(testdataDir, "not-kustomization", "deployment.yaml"),
			wantError:  true,
			wantErrMsg: "is not a kustomization file",
		},
		{
			name:       "missing resource",
			path:       filepath.Join(testdataDir, "missing-resource", "kustomization.yaml"),
			wantError:  true,
			wantErrMsg: "processing resource",
		},
		{
			name:       "non-existent file",
			path:       filepath.Join(testdataDir, "does-not-exist", "kustomization.yaml"),
			wantError:  true,
			wantErrMsg: "stat'ing path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			builder := &Builder{
				IOStreams: genericiooptions.IOStreams{
					In:     os.Stdin,
					Out:    &stdout,
					ErrOut: &stderr,
				},
				kustomizer: func() kustomize.Kustomizer {
					// Use a fake helm templater for testing
					fakeTemplater := &helm.FakeTemplater{
						TemplateFunc: func(_ string, _ v1.HelmChart, _ *v1.HelmGlobals) ([]map[string]any, error) {
							return nil, nil
						},
					}
					k, err := kustomize.NewKustomize(nil, kustomize.WithHelmTemplater(fakeTemplater))
					assert.NoError(t, err)
					return k
				}(),
			}

			err := builder.Build(t.Context(), tt.path)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}

			stdoutStr := stdout.String()

			for _, want := range tt.wantOutputs {
				assert.Contains(t, stdoutStr, want, "stdout should contain expected output")
			}
		})
	}
}
