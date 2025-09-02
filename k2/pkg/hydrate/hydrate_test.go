package hydrate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestNewHydrator(t *testing.T) {
	h := NewHydrator()
	assert.NotNil(t, h, "NewHydrator() should not return nil")
}

func TestHydrate(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		currentResources []*kyaml.RNode
		wantErr          bool
		errMsg           string
		wantNodes        int
		validateFunc     func(t *testing.T, result *HydratedResult)
	}{
		{
			name:      "basic empty kustomization with directory path",
			path:      "../../fixtures/basic-empty",
			wantErr:   false,
			wantNodes: 0,
		},
		{
			name:      "basic empty kustomization with file path",
			path:      "../../fixtures/basic-empty/kustomization.yaml",
			wantErr:   false,
			wantNodes: 0,
		},
		{
			name:      "kustomization with two resources",
			path:      "../../fixtures/two-resources",
			wantErr:   false,
			wantNodes: 2,
		},
		{
			name:      "mixed resources with file and directories",
			path:      "../../fixtures/mixed-resources",
			wantErr:   false,
			wantNodes: 5,
		},
		{
			name:    "non-existent directory",
			path:    "../../fixtures/non-existent",
			wantErr: true,
			errMsg:  "no such file or directory",
		},
		{
			name:    "directory without kustomization.yaml",
			path:    "../../fixtures",
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:    "non-kustomization file",
			path:    "hydrate.go",
			wantErr: true,
			errMsg:  "is not kustomization.yaml",
		},
		{
			name:    "invalid yaml syntax",
			path:    "../../fixtures/invalid-yaml",
			wantErr: true,
			errMsg:  "yaml:",
		},
		{
			name:    "missing resource file",
			path:    "../../fixtures/missing-resource",
			wantErr: true,
			errMsg:  "no such file or directory",
		},
		{
			name:      "valid kustomization with valid resource and empty resource file",
			path:      "../../fixtures/empty-yaml",
			wantErr:   false,
			wantNodes: 1,
		},
		{
			name:    "kustomization with unknown fields",
			path:    "../../fixtures/invalid-kustomization",
			wantErr: true,
			errMsg:  "field invalidField not found",
		},
		{
			name:    "directory resource without kustomization",
			path:    "../../fixtures/directory-no-kustomization",
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:      "kustomization with components",
			path:      "../../fixtures/with-components",
			wantErr:   false,
			wantNodes: 2,
			validateFunc: func(t *testing.T, result *HydratedResult) {
				resources := make(map[string]string)
				for _, node := range result.Nodes {
					meta, err := node.GetMeta()
					require.NoError(t, err)
					resources[meta.Kind] = meta.Name
				}
				assert.Equal(t, "app", resources["Deployment"])
				assert.Equal(t, "logging-config", resources["ConfigMap"])
			},
		},
		{
			name:      "kustomization with multiple components and resources",
			path:      "../../fixtures/components-and-resources",
			wantErr:   false,
			wantNodes: 5,
			validateFunc: func(t *testing.T, result *HydratedResult) {
				resources := make(map[string]string)
				for _, node := range result.Nodes {
					meta, err := node.GetMeta()
					require.NoError(t, err)
					resources[meta.Kind] = meta.Name
				}
				assert.Equal(t, "web-app", resources["Deployment"])
				assert.Equal(t, "web-service", resources["Service"])
				assert.Equal(t, "web-monitor", resources["ServiceMonitor"])
				assert.Equal(t, "monitoring-config", resources["ConfigMap"])
				assert.Equal(t, "debug-tools", resources["Pod"])
			},
		},
		{
			name:    "component with missing resource",
			path:    "../../fixtures/component-with-error",
			wantErr: true,
			errMsg:  "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHydrator()
			ctx := context.Background()

			result, err := h.Hydrate(ctx, tt.path, tt.currentResources)

			if tt.wantErr {
				require.Error(t, err, "Hydrate() should return an error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err, "Hydrate() should not return an error")
				assert.NotNil(t, result, "Hydrate() should not return nil result")
				assert.NotNil(t, result.Nodes, "Hydrate() should not return nil nodes")
				assert.Len(t, result.Nodes, tt.wantNodes, "Hydrate() should return expected number of nodes")

				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}
		})
	}
}
