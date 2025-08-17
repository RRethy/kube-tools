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

func TestToolsCreateLogsTool(t *testing.T) {
	tools := New()
	tool := tools.CreateLogsTool()

	assert.Equal(t, "logs", tool.Name)
	assert.Equal(t, "Get logs from a pod container", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	assert.Contains(t, tool.InputSchema.Properties, "pod-name")
	assert.Contains(t, tool.InputSchema.Properties, "namespace")
	assert.Contains(t, tool.InputSchema.Properties, "context")
	assert.Contains(t, tool.InputSchema.Properties, "container")
	assert.Contains(t, tool.InputSchema.Properties, "tail")
	assert.Contains(t, tool.InputSchema.Properties, "since")
	assert.Contains(t, tool.InputSchema.Properties, "previous")
	assert.Contains(t, tool.InputSchema.Properties, "timestamps")
	assert.Contains(t, tool.InputSchema.Properties, "all-containers")
	assert.Contains(t, tool.InputSchema.Properties, "limit-bytes")
	assert.Contains(t, tool.InputSchema.Properties, "since-time")
	assert.Contains(t, tool.InputSchema.Properties, "prefix")
	assert.Contains(t, tool.InputSchema.Properties, "selector")

	assert.Contains(t, tool.InputSchema.Required, "pod-name")

	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleLogs(t *testing.T) {
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
			expectedError: "invalid arguments",
		},
		{
			name:          "invalid arguments type",
			args:          "invalid",
			expectedError: "invalid arguments",
		},
		{
			name:          "missing pod-name",
			args:          map[string]any{},
			expectedError: "pod-name parameter required",
		},
		{
			name: "invalid pod-name type",
			args: map[string]any{
				"pod-name": 123,
			},
			expectedError: "pod-name parameter required",
		},
		{
			name: "empty pod-name",
			args: map[string]any{
				"pod-name": "",
			},
			expectedError: "pod-name parameter required",
		},
		{
			name: "basic logs command",
			args: map[string]any{
				"pod-name": "my-pod",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "log line 1\nlog line 2\nlog line 3",
		},
		{
			name: "logs with namespace",
			args: map[string]any{
				"pod-name":  "my-pod",
				"namespace": "my-namespace",
			},
			expectedArgs:  []string{"logs", "my-pod", "-n", "my-namespace", "--tail", "100"},
			kubectlOutput: "namespace logs",
		},
		{
			name: "logs with context",
			args: map[string]any{
				"pod-name": "my-pod",
				"context":  "my-context",
			},
			expectedArgs:  []string{"--context", "my-context", "logs", "my-pod", "--tail", "100"},
			kubectlOutput: "context logs",
		},
		{
			name: "logs with container",
			args: map[string]any{
				"pod-name":  "my-pod",
				"container": "my-container",
			},
			expectedArgs:  []string{"logs", "my-pod", "-c", "my-container", "--tail", "100"},
			kubectlOutput: "container logs",
		},
		{
			name: "logs with custom tail",
			args: map[string]any{
				"pod-name": "my-pod",
				"tail":     50.0,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "50"},
			kubectlOutput: "limited logs",
		},
		{
			name: "logs with tail zero (no tail)",
			args: map[string]any{
				"pod-name": "my-pod",
				"tail":     0.0,
			},
			expectedArgs:  []string{"logs", "my-pod"},
			kubectlOutput: "all logs",
		},
		{
			name: "logs with since",
			args: map[string]any{
				"pod-name": "my-pod",
				"since":    "5m",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--since", "5m"},
			kubectlOutput: "recent logs",
		},
		{
			name: "logs with previous",
			args: map[string]any{
				"pod-name": "my-pod",
				"previous": true,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--previous"},
			kubectlOutput: "previous container logs",
		},
		{
			name: "logs with timestamps",
			args: map[string]any{
				"pod-name":   "my-pod",
				"timestamps": true,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--timestamps"},
			kubectlOutput: "2024-01-01T00:00:00Z log line",
		},
		{
			name: "logs with all options",
			args: map[string]any{
				"pod-name":   "my-pod",
				"namespace":  "my-namespace",
				"context":    "my-context",
				"container":  "my-container",
				"tail":       25.0,
				"since":      "1h",
				"previous":   true,
				"timestamps": true,
			},
			expectedArgs: []string{
				"--context", "my-context",
				"logs", "my-pod",
				"-n", "my-namespace",
				"-c", "my-container",
				"--tail", "25",
				"--since", "1h",
				"--previous",
				"--timestamps",
			},
			kubectlOutput: "full logs",
		},
		{
			name: "empty namespace ignored",
			args: map[string]any{
				"pod-name":  "my-pod",
				"namespace": "",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "logs",
		},
		{
			name: "empty context ignored",
			args: map[string]any{
				"pod-name": "my-pod",
				"context":  "",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "logs",
		},
		{
			name: "empty container ignored",
			args: map[string]any{
				"pod-name":  "my-pod",
				"container": "",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "logs",
		},
		{
			name: "empty since ignored",
			args: map[string]any{
				"pod-name": "my-pod",
				"since":    "",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "logs",
		},
		{
			name: "previous false ignored",
			args: map[string]any{
				"pod-name": "my-pod",
				"previous": false,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "logs",
		},
		{
			name: "timestamps false ignored",
			args: map[string]any{
				"pod-name":   "my-pod",
				"timestamps": false,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "logs",
		},
		{
			name: "kubectl error",
			args: map[string]any{
				"pod-name": "my-pod",
			},
			expectedArgs: []string{"logs", "my-pod", "--tail", "100"},
			kubectlError: fmt.Errorf("pod not found"),
		},
		{
			name: "with all-containers",
			args: map[string]any{
				"pod-name":       "my-pod",
				"all-containers": true,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--all-containers"},
			kubectlOutput: "container-1 logs\ncontainer-2 logs",
		},
		{
			name: "with limit-bytes",
			args: map[string]any{
				"pod-name":    "my-pod",
				"limit-bytes": 1024.0,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--limit-bytes", "1024"},
			kubectlOutput: "limited log output",
		},
		{
			name: "with since-time",
			args: map[string]any{
				"pod-name":   "my-pod",
				"since-time": "2024-01-01T00:00:00Z",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--since-time", "2024-01-01T00:00:00Z"},
			kubectlOutput: "logs since time",
		},
		{
			name: "with prefix",
			args: map[string]any{
				"pod-name": "my-pod",
				"prefix":   true,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100", "--prefix"},
			kubectlOutput: "[pod/my-pod/container] log line",
		},
		{
			name: "with selector instead of pod-name",
			args: map[string]any{
				"selector": "app=nginx",
			},
			expectedArgs:  []string{"logs", "-l", "app=nginx", "--tail", "100"},
			kubectlOutput: "logs from multiple pods",
		},
		{
			name: "selector with namespace",
			args: map[string]any{
				"selector":  "app=web",
				"namespace": "production",
			},
			expectedArgs:  []string{"logs", "-l", "app=web", "-n", "production", "--tail", "100"},
			kubectlOutput: "production pod logs",
		},
		{
			name: "all new parameters combined",
			args: map[string]any{
				"pod-name":       "my-pod",
				"namespace":      "test",
				"all-containers": true,
				"limit-bytes":    2048.0,
				"since-time":     "2024-01-01T00:00:00Z",
				"prefix":         true,
			},
			expectedArgs: []string{
				"logs", "my-pod", "-n", "test", "--tail", "100",
				"--all-containers", "--limit-bytes", "2048",
				"--since-time", "2024-01-01T00:00:00Z", "--prefix",
			},
			kubectlOutput: "combined output",
		},
		{
			name: "empty new parameters ignored",
			args: map[string]any{
				"pod-name":   "my-pod",
				"since-time": "",
				"selector":   "",
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "normal logs",
		},
		{
			name: "false boolean new parameters ignored",
			args: map[string]any{
				"pod-name":       "my-pod",
				"all-containers": false,
				"prefix":         false,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "normal logs",
		},
		{
			name: "zero limit-bytes ignored",
			args: map[string]any{
				"pod-name":    "my-pod",
				"limit-bytes": 0.0,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "normal logs",
		},
		{
			name: "invalid types for new parameters ignored",
			args: map[string]any{
				"pod-name":       "my-pod",
				"all-containers": "yes",
				"limit-bytes":    "1024",
				"since-time":     123,
				"prefix":         "true",
				"selector":       456,
			},
			expectedArgs:  []string{"logs", "my-pod", "--tail", "100"},
			kubectlOutput: "normal logs",
		},
		{
			name: "selector overrides pod-name",
			args: map[string]any{
				"pod-name": "my-pod",
				"selector": "app=test",
			},
			expectedArgs:  []string{"logs", "-l", "app=test", "--tail", "100"},
			kubectlOutput: "selector-based logs",
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

			result, err := tools.HandleLogs(context.Background(), req)

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
