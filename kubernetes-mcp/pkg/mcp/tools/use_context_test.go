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

func TestToolsCreateUseContextTool(t *testing.T) {
	tools := New()
	tool := tools.CreateUseContextTool()

	assert.Equal(t, "use-context", tool.Name)
	assert.Equal(t, "Switch to a different Kubernetes context", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	assert.Contains(t, tool.InputSchema.Properties, "context")
	assert.Contains(t, tool.InputSchema.Required, "context")

	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.False(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleUseContext(t *testing.T) {
	tests := []struct {
		name           string
		args           interface{}
		expectedArgs   []string
		expectedError  string
		kubectlOutput  string
		kubectlError   error
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
			name:          "missing context",
			args:          map[string]any{},
			expectedError: "context parameter required",
		},
		{
			name: "invalid context type",
			args: map[string]any{
				"context": 123,
			},
			expectedError: "context parameter required",
		},
		{
			name: "empty context",
			args: map[string]any{
				"context": "",
			},
			expectedError: "context parameter required",
		},
		{
			name: "valid context switch",
			args: map[string]any{
				"context": "my-context",
			},
			expectedArgs:  []string{"config", "use-context", "my-context"},
			kubectlOutput: "Switched to context \"my-context\".",
		},
		{
			name: "switch to production context",
			args: map[string]any{
				"context": "production-cluster",
			},
			expectedArgs:  []string{"config", "use-context", "production-cluster"},
			kubectlOutput: "Switched to context \"production-cluster\".",
		},
		{
			name: "context with special characters",
			args: map[string]any{
				"context": "my-context-with-dashes",
			},
			expectedArgs:  []string{"config", "use-context", "my-context-with-dashes"},
			kubectlOutput: "Switched to context \"my-context-with-dashes\".",
		},
		{
			name: "kubectl error - context not found",
			args: map[string]any{
				"context": "non-existent-context",
			},
			expectedArgs: []string{"config", "use-context", "non-existent-context"},
			kubectlError: fmt.Errorf("error: no context exists with the name: \"non-existent-context\""),
		},
		{
			name: "kubectl error - config issue",
			args: map[string]any{
				"context": "my-context",
			},
			expectedArgs: []string{"config", "use-context", "my-context"},
			kubectlError: fmt.Errorf("error loading config file"),
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

			result, err := tools.HandleUseContext(context.Background(), req)

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