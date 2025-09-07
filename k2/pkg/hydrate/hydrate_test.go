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
	type resource struct {
		kind        string
		name        string
		annotations map[string]string
	}

	tests := []struct {
		name             string
		path             string
		currentResources []*kyaml.RNode
		wantErr          bool
		errMsg           string
		wantResources    []resource
	}{
		{
			name:    "basic empty kustomization with directory path",
			path:    "../../fixtures/basic-empty",
			wantErr: false,
		},
		{
			name:    "basic empty kustomization with file path",
			path:    "../../fixtures/basic-empty/kustomization.yaml",
			wantErr: false,
		},
		{
			name:    "kustomization with two resources",
			path:    "../../fixtures/two-resources",
			wantErr: false,
			wantResources: []resource{
				{kind: "Deployment", name: "nginx-deployment", annotations: map[string]string{}},
				{kind: "Service", name: "nginx-service", annotations: map[string]string{}},
			},
		},
		{
			name:    "mixed resources with file and directories",
			path:    "../../fixtures/mixed-resources",
			wantErr: false,
			wantResources: []resource{
				{kind: "ConfigMap", name: "app-config", annotations: map[string]string{}},
				{kind: "Deployment", name: "web-app", annotations: map[string]string{}},
				{kind: "Service", name: "web-service", annotations: map[string]string{}},
				{kind: "Namespace", name: "dev", annotations: map[string]string{}},
				{kind: "Ingress", name: "web-ingress", annotations: map[string]string{}},
			},
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
			name:    "valid kustomization with valid resource and empty resource file",
			path:    "../../fixtures/empty-yaml",
			wantErr: false,
			wantResources: []resource{
				{kind: "Deployment", name: "valid-deployment", annotations: map[string]string{}},
			},
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
			name:    "kustomization with components",
			path:    "../../fixtures/with-components",
			wantErr: false,
			wantResources: []resource{
				{kind: "Deployment", name: "app", annotations: map[string]string{}},
				{kind: "ConfigMap", name: "logging-config", annotations: map[string]string{}},
			},
		},
		{
			name:    "kustomization with multiple components and resources",
			path:    "../../fixtures/components-and-resources",
			wantErr: false,
			wantResources: []resource{
				{kind: "Deployment", name: "web-app", annotations: map[string]string{}},
				{kind: "Service", name: "web-service", annotations: map[string]string{}},
				{kind: "ServiceMonitor", name: "web-monitor", annotations: map[string]string{}},
				{kind: "ConfigMap", name: "monitoring-config", annotations: map[string]string{}},
				{kind: "Pod", name: "debug-tools", annotations: map[string]string{}},
			},
		},
		{
			name:    "component with missing resource",
			path:    "../../fixtures/component-with-error",
			wantErr: true,
			errMsg:  "no such file or directory",
		},
		// {
		// 	name: "kustomization with component and resource with top level commonAnnotations",
		// 	// TODO: implement this test
		// },
		// {
		// 	name: "kustomization with component with commonAnnotations applies to top level resource",
		// 	// TODO: implement this test
		// },
		// {
		// 	name: "kustomization with resource with commonAnnotations does not apply to top level resource",
		// 	// TODO: implement this test
		// },
		// {
		// 	name:    "kustomization with commonAnnotations",
		// 	path:    "../../fixtures/common-annotations",
		// 	wantErr: false,
		// 	wantResources: []resource{
		// 		{kind: "Deployment", name: "annotated-deployment", annotations: map[string]string{"team": "devops", "environment": "production"}},
		// 		{kind: "Service", name: "annotated-service", annotations: map[string]string{"team": "devops", "environment": "production"}},
		// 	},
		// },
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
				assert.Equal(t, len(tt.wantResources), len(result.Nodes), "Number of resources should match expected")
				for i, expected := range tt.wantResources {
					r := resource{
						kind:        result.Nodes[i].GetKind(),
						name:        result.Nodes[i].GetName(),
						annotations: result.Nodes[i].GetAnnotations(),
					}
					assert.Equal(t, expected, r, "Resource should match expected")

				}
			}
		})
	}
}
