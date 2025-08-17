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

func TestToolsCreateAPIVersionsTool(t *testing.T) {
	tools := New()
	tool := tools.CreateAPIVersionsTool()

	// Verify basic tool properties
	assert.Equal(t, "api-versions", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "API versions")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	assert.Contains(t, tool.InputSchema.Properties, "context")

	// Verify no required parameters
	assert.Empty(t, tool.InputSchema.Required)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleAPIVersions(t *testing.T) {
	tests := []struct {
		name          string
		args          any
		kubectlStdout string
		kubectlStderr string
		kubectlError  error
		wantArgs      []string
		wantError     bool
		wantErrorMsg  string
		checkResult   func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name:          "basic command without arguments",
			args:          map[string]any{},
			kubectlStdout: "v1\napps/v1\nautoscaling/v1\nbatch/v1",
			wantArgs:      []string{"api-versions"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "v1")
				assert.Contains(t, content.Text, "apps/v1")
			},
		},
		{
			name: "with context parameter",
			args: map[string]any{
				"context": "staging",
			},
			kubectlStdout: "v1\napps/v1",
			wantArgs:      []string{"--context", "staging", "api-versions"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
			},
		},
		{
			name: "empty context parameter is ignored",
			args: map[string]any{
				"context": "",
			},
			kubectlStdout: "v1",
			wantArgs:      []string{"api-versions"},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown": "value",
				"context": "prod",
			},
			kubectlStdout: "v1",
			wantArgs:      []string{"--context", "prod", "api-versions"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: "v1\napps/v1",
			wantArgs:      []string{"api-versions"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
			},
		},
		{
			name:          "kubectl error is handled",
			args:          map[string]any{},
			kubectlStderr: "error: unable to connect to server",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"api-versions"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "kubectl command failed")
				assert.Contains(t, content.Text, "unable to connect to server")
			},
		},
		{
			name: "kubectl stdout with warnings in stderr",
			args: map[string]any{
				"context": "dev",
			},
			kubectlStdout: "v1\napps/v1",
			kubectlStderr: "Warning: deprecated API",
			wantArgs:      []string{"--context", "dev", "api-versions"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "v1")
				assert.Contains(t, content.Text, "Warning")
			},
		},
		{
			name:         "invalid arguments type returns error",
			args:         "not a map",
			wantError:    true,
			wantErrorMsg: "invalid arguments",
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				assert.Nil(t, result)
			},
		},
		{
			name:         "array arguments returns error",
			args:         []string{"test"},
			wantError:    true,
			wantErrorMsg: "invalid arguments",
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				assert.Nil(t, result)
			},
		},
		{
			name: "numeric context value is ignored",
			args: map[string]any{
				"context": 123,
			},
			kubectlStdout: "v1",
			wantArgs:      []string{"api-versions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake kubectl
			fake := kubectl.NewFake(tt.kubectlStdout, tt.kubectlStderr, tt.kubectlError)

			// Create tools with fake kubectl
			tools := NewWithKubectl(fake)

			// Create request
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "api-versions",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleAPIVersions(context.Background(), req)

			// Check error
			if tt.wantError {
				assert.Error(t, err)
				if tt.wantErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Check result
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			// Check kubectl was called with correct args (if not an invalid args test)
			if tt.wantArgs != nil {
				assert.True(t, fake.ExecuteCalled, "kubectl should have been called")
				assert.Equal(t, tt.wantArgs, fake.ExecuteArgs,
					"kubectl args mismatch")
			}
		})
	}
}
