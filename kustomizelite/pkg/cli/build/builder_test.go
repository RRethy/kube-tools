package build

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kube-tools/kustomizelite/pkg/kustomize"
)

func TestBuilder_Build(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		kustomizeFunc func(path string, globalHelmValuesFiles []string) ([]map[string]any, error)
		wantOut       []string
		wantError     bool
		wantErrMsg    string
	}{
		{
			name: "single valid file",
			path: "/path/to/kustomization.yaml",
			kustomizeFunc: func(_ string, _ []string) ([]map[string]any, error) {
				return []map[string]any{{
					"apiVersion": "kustomize.config.k8s.io/v1beta1",
					"kind":       "Kustomization",
				}}, nil
			},
			wantOut: []string{
				"apiVersion: kustomize.config.k8s.io/v1beta1",
				"kind: Kustomization",
			},
		},
		{
			name: "invalid file",
			path: "/path/to/invalid.yaml",
			kustomizeFunc: func(_ string, _ []string) ([]map[string]any, error) {
				return nil, errors.New("file /path/to/invalid.yaml is not a kustomization file")
			},
			wantError:  true,
			wantErrMsg: "file /path/to/invalid.yaml is not a kustomization file",
		},
		{
			name: "multiple resources",
			path: "/path/to/kustomization.yaml",
			kustomizeFunc: func(_ string, _ []string) ([]map[string]any, error) {
				return []map[string]any{
					{"content": "content1"},
					{"content": "content2"},
				}, nil
			},
			wantOut: []string{
				"content: content1",
				"---",
				"content: content2",
			},
		},
		{
			name: "empty resources",
			path: "/path/to/empty.yaml",
			kustomizeFunc: func(_ string, _ []string) ([]map[string]any, error) {
				return []map[string]any{}, nil
			},
			wantOut: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			b := &Builder{
				IOStreams: genericiooptions.IOStreams{
					Out:    &stdout,
					ErrOut: &stderr,
				},
				kustomizer: &kustomize.FakeKustomizer{
					KustomizeFunc: tt.kustomizeFunc,
				},
				useKustomizeFormat: false,
			}

			err := b.Build(t.Context(), tt.path)

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}

			outStr := stdout.String()

			for _, want := range tt.wantOut {
				assert.Contains(t, outStr, want)
			}
		})
	}
}
