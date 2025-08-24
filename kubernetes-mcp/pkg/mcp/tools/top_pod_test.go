package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/RRethy/kube-tools/kubernetes-mcp/pkg/kubectl"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsCreateTopPodTool(t *testing.T) {
	tools := New()
	tool := tools.CreateTopPodTool()

	assert.Equal(t, "top-pod", tool.Name)
	assert.Equal(t, "Display Resource (CPU/Memory) usage for pods", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	assert.Contains(t, tool.InputSchema.Properties, "namespace")
	assert.Contains(t, tool.InputSchema.Properties, "context")
	assert.Contains(t, tool.InputSchema.Properties, "all-namespaces")
	assert.Contains(t, tool.InputSchema.Properties, "selector")
	assert.Contains(t, tool.InputSchema.Properties, "containers")
	assert.Contains(t, tool.InputSchema.Properties, "sort-by")
	assert.Contains(t, tool.InputSchema.Properties, "field-selector")
	assert.Contains(t, tool.InputSchema.Properties, "no-headers")
	assert.Contains(t, tool.InputSchema.Properties, "sum")

	assert.Empty(t, tool.InputSchema.Required)

	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleTopPod(t *testing.T) {
	tests := []struct {
		name          string
		args          interface{}
		expectedArgs  []string
		expectedError string
		kubectlOutput string
		kubectlError  error
	}{
		{
			name:          "nil arguments",
			args:          nil,
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi\npod-2       20m          200Mi",
		},
		{
			name:          "invalid arguments type",
			args:          "invalid",
			expectedError: "invalid arguments",
		},
		{
			name:          "empty arguments",
			args:          map[string]any{},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi",
		},
		{
			name: "with namespace",
			args: map[string]any{
				"namespace": "my-namespace",
			},
			expectedArgs:  []string{"top", "pod", "-n", "my-namespace"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi",
		},
		{
			name: "with context",
			args: map[string]any{
				"context": "my-context",
			},
			expectedArgs:  []string{"--context", "my-context", "top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi",
		},
		{
			name: "with all-namespaces",
			args: map[string]any{
				"all-namespaces": true,
			},
			expectedArgs: []string{"top", "pod", "--all-namespaces"},
			kubectlOutput: `NAMESPACE   NAME        CPU(cores)   MEMORY(bytes)
default     pod-1       10m          100Mi
kube-system pod-2       20m          200Mi`,
		},
		{
			name: "with selector",
			args: map[string]any{
				"selector": "app=nginx",
			},
			expectedArgs:  []string{"top", "pod", "-l", "app=nginx"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\nnginx-1     10m          100Mi",
		},
		{
			name: "with containers",
			args: map[string]any{
				"containers": true,
			},
			expectedArgs: []string{"top", "pod", "--containers"},
			kubectlOutput: `POD         NAME        CPU(cores)   MEMORY(bytes)
pod-1       container1  5m           50Mi
pod-1       container2  5m           50Mi`,
		},
		{
			name: "with sort-by cpu",
			args: map[string]any{
				"sort-by": "cpu",
			},
			expectedArgs: []string{"top", "pod", "--sort-by", "cpu"},
			kubectlOutput: `NAME        CPU(cores)   MEMORY(bytes)
pod-high    100m         200Mi
pod-med     50m          150Mi
pod-low     10m          100Mi`,
		},
		{
			name: "with sort-by memory",
			args: map[string]any{
				"sort-by": "memory",
			},
			expectedArgs: []string{"top", "pod", "--sort-by", "memory"},
			kubectlOutput: `NAME        CPU(cores)   MEMORY(bytes)
pod-heavy   20m          500Mi
pod-med     50m          250Mi
pod-light   100m         100Mi`,
		},
		{
			name: "all-namespaces overrides namespace",
			args: map[string]any{
				"namespace":      "my-namespace",
				"all-namespaces": true,
			},
			expectedArgs: []string{"top", "pod", "--all-namespaces"},
			kubectlOutput: `NAMESPACE   NAME        CPU(cores)   MEMORY(bytes)
default     pod-1       10m          100Mi`,
		},
		{
			name: "all parameters combined",
			args: map[string]any{
				"context":        "production",
				"all-namespaces": true,
				"selector":       "app=web",
				"containers":     true,
				"sort-by":        "cpu",
			},
			expectedArgs: []string{
				"--context", "production",
				"top", "pod",
				"--all-namespaces",
				"-l", "app=web",
				"--containers",
				"--sort-by", "cpu",
			},
			kubectlOutput: `NAMESPACE   POD         NAME        CPU(cores)   MEMORY(bytes)
default     web-1       container1  10m          100Mi`,
		},
		{
			name: "empty string parameters ignored",
			args: map[string]any{
				"namespace": "",
				"context":   "",
				"selector":  "",
				"sort-by":   "",
			},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi",
		},
		{
			name: "false boolean parameters ignored",
			args: map[string]any{
				"all-namespaces": false,
				"containers":     false,
			},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi",
		},
		{
			name: "invalid parameter types ignored",
			args: map[string]any{
				"namespace":      123,
				"context":        true,
				"selector":       456,
				"all-namespaces": "yes",
				"containers":     "true",
				"sort-by":        123,
			},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\npod-1       10m          100Mi",
		},
		{
			name:         "kubectl error - metrics not available",
			args:         map[string]any{},
			expectedArgs: []string{"top", "pod"},
			kubectlError: fmt.Errorf("error: Metrics API not available"),
		},
		{
			name: "kubectl error - invalid context",
			args: map[string]any{
				"context": "non-existent",
			},
			expectedArgs: []string{"--context", "non-existent", "top", "pod"},
			kubectlError: fmt.Errorf("context \"non-existent\" does not exist"),
		},
		{
			name: "no pods found with selector",
			args: map[string]any{
				"selector": "app=nonexistent",
			},
			expectedArgs:  []string{"top", "pod", "-l", "app=nonexistent"},
			kubectlOutput: "No resources found",
		},
		{
			name: "with complex selector",
			args: map[string]any{
				"selector": "app=web,tier=frontend",
			},
			expectedArgs:  []string{"top", "pod", "-l", "app=web,tier=frontend"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\nfrontend-1  10m          100Mi",
		},
		{
			name: "with namespace and context",
			args: map[string]any{
				"namespace": "production",
				"context":   "prod-cluster",
			},
			expectedArgs:  []string{"--context", "prod-cluster", "top", "pod", "-n", "production"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\nprod-pod-1  50m          500Mi",
		},
		{
			name: "sort-by with namespace and selector",
			args: map[string]any{
				"namespace": "kube-system",
				"selector":  "k8s-app=kube-dns",
				"sort-by":   "memory",
			},
			expectedArgs: []string{"top", "pod", "-n", "kube-system", "-l", "k8s-app=kube-dns", "--sort-by", "memory"},
			kubectlOutput: `NAME                CPU(cores)   MEMORY(bytes)
coredns-1           10m          70Mi
coredns-2           8m           65Mi`,
		},
		{
			name: "sort-by with containers",
			args: map[string]any{
				"containers": true,
				"sort-by":    "cpu",
			},
			expectedArgs: []string{"top", "pod", "--containers", "--sort-by", "cpu"},
			kubectlOutput: `POD         NAME        CPU(cores)   MEMORY(bytes)
pod-busy    worker      80m          300Mi
pod-busy    sidecar     20m          100Mi
pod-idle    main        5m           50Mi`,
		},
		{
			name: "with field-selector",
			args: map[string]any{
				"field-selector": "status.phase=Running",
			},
			expectedArgs:  []string{"top", "pod", "--field-selector", "status.phase=Running"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)\nrunning-1   10m          100Mi",
		},
		{
			name: "with no-headers",
			args: map[string]any{
				"no-headers": true,
			},
			expectedArgs:  []string{"top", "pod", "--no-headers"},
			kubectlOutput: "pod-1       10m          100Mi\npod-2       20m          200Mi",
		},
		{
			name: "with sum",
			args: map[string]any{
				"sum": true,
			},
			expectedArgs: []string{"top", "pod", "--sum"},
			kubectlOutput: `NAME        CPU(cores)   MEMORY(bytes)
pod-1       10m          100Mi
pod-2       20m          200Mi
            --------     --------
            30m          300Mi`,
		},
		{
			name: "all new parameters combined",
			args: map[string]any{
				"namespace":      "test",
				"field-selector": "status.phase=Running",
				"no-headers":     true,
				"sum":            true,
				"sort-by":        "memory",
			},
			expectedArgs: []string{
				"top", "pod", "-n", "test",
				"--sort-by", "memory",
				"--field-selector", "status.phase=Running",
				"--no-headers",
				"--sum",
			},
			kubectlOutput: "test-pod-1  5m   50Mi\ntest-pod-2  10m  100Mi\n            ---  ----\n            15m  150Mi",
		},
		{
			name: "empty new parameters ignored",
			args: map[string]any{
				"field-selector": "",
			},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)",
		},
		{
			name: "false boolean new parameters ignored",
			args: map[string]any{
				"no-headers": false,
				"sum":        false,
			},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)",
		},
		{
			name: "invalid types for new parameters ignored",
			args: map[string]any{
				"field-selector": 123,
				"no-headers":     "yes",
				"sum":            "true",
			},
			expectedArgs:  []string{"top", "pod"},
			kubectlOutput: "NAME        CPU(cores)   MEMORY(bytes)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := kubectl.NewFake(tt.kubectlOutput, "", tt.kubectlError)
			tools := NewWithKubectl(fake)

			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleTopPod(context.Background(), req)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			if tt.kubectlError != nil {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				assert.Len(t, result.Content, 1)
				content, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok)
				assert.Contains(t, content.Text, tt.kubectlError.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.True(t, fake.ExecuteCalled, "kubectl should have been called")
			assert.Equal(t, tt.expectedArgs, fake.ExecuteArgs)

			var content []mcp.TextContent
			for _, c := range result.Content {
				if tc, ok := c.(mcp.TextContent); ok {
					content = append(content, tc)
				}
			}
			require.Len(t, content, 1)
			assert.Equal(t, tt.kubectlOutput, content[0].Text)
		})
	}
}
