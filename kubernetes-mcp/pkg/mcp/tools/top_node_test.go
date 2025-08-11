package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/kubectl"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsCreateTopNodeTool(t *testing.T) {
	tools := New()
	tool := tools.CreateTopNodeTool()

	assert.Equal(t, "top-node", tool.Name)
	assert.Equal(t, "Display Resource (CPU/Memory) usage for nodes", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	assert.Contains(t, tool.InputSchema.Properties, "context")
	assert.Contains(t, tool.InputSchema.Properties, "selector")
	assert.Contains(t, tool.InputSchema.Properties, "sort-by")
	assert.Contains(t, tool.InputSchema.Properties, "no-headers")
	assert.Contains(t, tool.InputSchema.Properties, "show-capacity")

	assert.Empty(t, tool.InputSchema.Required)

	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleTopNode(t *testing.T) {
	tests := []struct {
		name           string
		args           interface{}
		expectedArgs   []string
		expectedError  string
		kubectlOutput  string
		kubectlError   error
	}{
		{
			name:         "nil arguments",
			args:         nil,
			expectedArgs: []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         25%    1Gi             50%
node-2               500m         50%    2Gi             75%`,
		},
		{
			name:          "invalid arguments type",
			args:          "invalid",
			expectedError: "invalid arguments",
		},
		{
			name:         "empty arguments",
			args:         map[string]any{},
			expectedArgs: []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         25%    1Gi             50%`,
		},
		{
			name: "with context",
			args: map[string]any{
				"context": "my-context",
			},
			expectedArgs: []string{"--context", "my-context", "top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         25%    1Gi             50%`,
		},
		{
			name: "with selector",
			args: map[string]any{
				"selector": "node-role.kubernetes.io/master=true",
			},
			expectedArgs: []string{"top", "node", "-l", "node-role.kubernetes.io/master=true"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
master-1             100m         10%    500Mi           25%`,
		},
		{
			name: "with context and selector",
			args: map[string]any{
				"context":  "production",
				"selector": "zone=us-west-1",
			},
			expectedArgs: []string{
				"--context", "production",
				"top", "node",
				"-l", "zone=us-west-1",
			},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
prod-node-west-1     1000m        50%    4Gi             50%
prod-node-west-2     1500m        75%    6Gi             75%`,
		},
		{
			name: "empty string parameters ignored",
			args: map[string]any{
				"context":  "",
				"selector": "",
			},
			expectedArgs: []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         25%    1Gi             50%`,
		},
		{
			name: "invalid parameter types ignored",
			args: map[string]any{
				"context":  123,
				"selector": true,
			},
			expectedArgs: []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         25%    1Gi             50%`,
		},
		{
			name:         "kubectl error - metrics not available",
			args:         map[string]any{},
			expectedArgs: []string{"top", "node"},
			kubectlError: fmt.Errorf("error: Metrics API not available"),
		},
		{
			name: "kubectl error - invalid context",
			args: map[string]any{
				"context": "non-existent",
			},
			expectedArgs: []string{"--context", "non-existent", "top", "node"},
			kubectlError: fmt.Errorf("context \"non-existent\" does not exist"),
		},
		{
			name: "no nodes found with selector",
			args: map[string]any{
				"selector": "zone=nonexistent",
			},
			expectedArgs:  []string{"top", "node", "-l", "zone=nonexistent"},
			kubectlOutput: "No resources found",
		},
		{
			name: "with complex selector",
			args: map[string]any{
				"selector": "node-role.kubernetes.io/worker=true,zone=us-east-1",
			},
			expectedArgs: []string{"top", "node", "-l", "node-role.kubernetes.io/worker=true,zone=us-east-1"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
worker-east-1        500m         25%    2Gi             50%
worker-east-2        750m         37%    3Gi             75%`,
		},
		{
			name: "single node cluster",
			args: map[string]any{},
			expectedArgs: []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
minikube             1250m        62%    5Gi             62%`,
		},
		{
			name: "with sort-by cpu",
			args: map[string]any{
				"sort-by": "cpu",
			},
			expectedArgs: []string{"top", "node", "--sort-by", "cpu"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-3               100m         10%    1Gi             25%
node-2               250m         25%    2Gi             50%
node-1               500m         50%    3Gi             75%`,
		},
		{
			name: "with sort-by memory",
			args: map[string]any{
				"sort-by": "memory",
			},
			expectedArgs: []string{"top", "node", "--sort-by", "memory"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-3               100m         10%    1Gi             25%
node-2               250m         25%    2Gi             50%
node-1               500m         50%    3Gi             75%`,
		},
		{
			name: "with no-headers",
			args: map[string]any{
				"no-headers": true,
			},
			expectedArgs: []string{"top", "node", "--no-headers"},
			kubectlOutput: `node-1               250m         25%    1Gi             50%
node-2               500m         50%    2Gi             75%`,
		},
		{
			name: "with show-capacity",
			args: map[string]any{
				"show-capacity": true,
			},
			expectedArgs: []string{"top", "node", "--show-capacity"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         12%    1Gi             25%
node-2               500m         25%    2Gi             50%`,
		},
		{
			name: "all new parameters combined",
			args: map[string]any{
				"context":       "prod",
				"selector":      "zone=us-west",
				"sort-by":       "memory",
				"no-headers":    true,
				"show-capacity": true,
			},
			expectedArgs: []string{
				"--context", "prod",
				"top", "node",
				"-l", "zone=us-west",
				"--sort-by", "memory",
				"--no-headers",
				"--show-capacity",
			},
			kubectlOutput: `prod-node-1          500m         25%    2Gi             50%
prod-node-2          750m         37%    3Gi             75%`,
		},
		{
			name: "empty sort-by ignored",
			args: map[string]any{
				"sort-by": "",
			},
			expectedArgs:  []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%`,
		},
		{
			name: "false boolean parameters ignored",
			args: map[string]any{
				"no-headers":    false,
				"show-capacity": false,
			},
			expectedArgs:  []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%`,
		},
		{
			name: "invalid type for new parameters ignored",
			args: map[string]any{
				"sort-by":       123,
				"no-headers":    "yes",
				"show-capacity": "true",
			},
			expectedArgs:  []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%`,
		},
		{
			name: "unknown parameters ignored",
			args: map[string]any{
				"namespace":      "default",
				"all-namespaces": true,
				"containers":     true,
			},
			expectedArgs: []string{"top", "node"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1               250m         25%    1Gi             50%`,
		},
		{
			name: "high CPU usage nodes",
			args: map[string]any{
				"selector": "status=high-load",
			},
			expectedArgs: []string{"top", "node", "-l", "status=high-load"},
			kubectlOutput: `NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-overloaded-1    1900m        95%    7Gi             87%
node-overloaded-2    1800m        90%    7500Mi          93%`,
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

			result, err := tools.HandleTopNode(context.Background(), req)

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