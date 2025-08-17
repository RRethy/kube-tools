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

func TestToolsCreateVersionTool(t *testing.T) {
	tools := New()
	tool := tools.CreateVersionTool()

	assert.Equal(t, "version", tool.Name)
	assert.Equal(t, "Get version information for kubectl client and Kubernetes cluster", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	assert.Contains(t, tool.InputSchema.Properties, "context")
	assert.Contains(t, tool.InputSchema.Properties, "output")

	assert.Empty(t, tool.InputSchema.Required)

	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleVersion(t *testing.T) {
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
			expectedArgs:  []string{"version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name:          "invalid arguments type",
			args:          "invalid",
			expectedError: "invalid arguments",
		},
		{
			name:          "empty arguments",
			args:          map[string]any{},
			expectedArgs:  []string{"version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name: "with context",
			args: map[string]any{
				"context": "my-context",
			},
			expectedArgs:  []string{"--context", "my-context", "version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name: "with json output",
			args: map[string]any{
				"output": "json",
			},
			expectedArgs:  []string{"version", "-o", "json"},
			kubectlOutput: `{"clientVersion":{"major":"1","minor":"27"},"serverVersion":{"major":"1","minor":"27"}}`,
		},
		{
			name: "with yaml output",
			args: map[string]any{
				"output": "yaml",
			},
			expectedArgs: []string{"version", "-o", "yaml"},
			kubectlOutput: `clientVersion:
  major: "1"
  minor: "27"
serverVersion:
  major: "1"
  minor: "27"`,
		},
		{
			name: "with context and json output",
			args: map[string]any{
				"context": "production",
				"output":  "json",
			},
			expectedArgs:  []string{"--context", "production", "version", "-o", "json"},
			kubectlOutput: `{"clientVersion":{"major":"1","minor":"27"},"serverVersion":{"major":"1","minor":"28"}}`,
		},
		{
			name: "empty context ignored",
			args: map[string]any{
				"context": "",
			},
			expectedArgs:  []string{"version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name: "empty output ignored",
			args: map[string]any{
				"output": "",
			},
			expectedArgs:  []string{"version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name: "invalid context type ignored",
			args: map[string]any{
				"context": 123,
			},
			expectedArgs:  []string{"version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name: "invalid output type ignored",
			args: map[string]any{
				"output": true,
			},
			expectedArgs:  []string{"version"},
			kubectlOutput: "Client Version: v1.27.0\nServer Version: v1.27.0",
		},
		{
			name:         "kubectl error - no cluster",
			args:         map[string]any{},
			expectedArgs: []string{"version"},
			kubectlError: fmt.Errorf("Unable to connect to the server"),
		},
		{
			name: "kubectl error - invalid context",
			args: map[string]any{
				"context": "non-existent",
			},
			expectedArgs: []string{"--context", "non-existent", "version"},
			kubectlError: fmt.Errorf("context \"non-existent\" does not exist"),
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

			result, err := tools.HandleVersion(context.Background(), req)

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
