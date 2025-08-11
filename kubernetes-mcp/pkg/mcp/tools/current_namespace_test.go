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

func TestToolsCreateCurrentNamespaceTool(t *testing.T) {
	tools := New()
	tool := tools.CreateCurrentNamespaceTool()

	// Verify basic tool properties
	assert.Equal(t, "current-namespace", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "namespace")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	assert.Contains(t, tool.InputSchema.Properties, "context")

	// Verify no required parameters
	assert.Empty(t, tool.InputSchema.Required)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleCurrentNamespace(t *testing.T) {
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
			name:          "returns namespace successfully",
			args:          map[string]any{},
			kubectlStdout: "kube-system",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "kube-system", content.Text)
			},
		},
		{
			name:          "empty namespace defaults to 'default'",
			args:          map[string]any{},
			kubectlStdout: "",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "default", content.Text)
			},
		},
		{
			name: "with context parameter",
			args: map[string]any{
				"context": "production",
			},
			kubectlStdout: "prod-namespace",
			wantArgs:      []string{"--context", "production", "config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "prod-namespace", content.Text)
			},
		},
		{
			name: "empty context parameter is ignored",
			args: map[string]any{
				"context": "",
			},
			kubectlStdout: "default",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
		},
		{
			name:          "returns monitoring namespace",
			args:          map[string]any{},
			kubectlStdout: "monitoring",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "monitoring", content.Text)
			},
		},
		{
			name:          "handles namespace with special characters",
			args:          map[string]any{},
			kubectlStdout: "my-app-namespace-123",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "my-app-namespace-123", content.Text)
			},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown": "value",
				"context": "dev",
				"extra":   123,
			},
			kubectlStdout: "dev-namespace",
			wantArgs:      []string{"--context", "dev", "config", "view", "--minify", "-o", "jsonpath={..namespace}"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: "default",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
			},
		},
		{
			name:          "kubectl error is handled",
			args:          map[string]any{},
			kubectlStderr: "error: unable to read kubeconfig",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "kubectl command failed")
				assert.Contains(t, content.Text, "unable to read kubeconfig")
			},
		},
		{
			name: "context not found error",
			args: map[string]any{
				"context": "nonexistent",
			},
			kubectlStderr: "error: context \"nonexistent\" does not exist",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"--context", "nonexistent", "config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "context")
				assert.Contains(t, content.Text, "does not exist")
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
			kubectlStdout: "default",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
		},
		{
			name: "handles warning with successful output",
			args: map[string]any{
				"context": "staging",
			},
			kubectlStdout: "staging-namespace",
			kubectlStderr: "Warning: context 'staging' is using outdated configuration",
			wantArgs:      []string{"--context", "staging", "config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "staging-namespace")
				assert.Contains(t, content.Text, "Warning")
			},
		},
		{
			name: "empty namespace with context defaults to 'default'",
			args: map[string]any{
				"context": "minikube",
			},
			kubectlStdout: "",
			wantArgs:      []string{"--context", "minikube", "config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "default", content.Text)
			},
		},
		{
			name:          "handles istio-system namespace",
			args:          map[string]any{},
			kubectlStdout: "istio-system",
			wantArgs:      []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "istio-system", content.Text)
			},
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
					Name:      "current-namespace",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleCurrentNamespace(context.Background(), req)

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

			// Check kubectl was called with correct args (if not an error test)
			if tt.wantArgs != nil {
				assert.True(t, fake.ExecuteCalled, "kubectl should have been called")
				assert.Equal(t, tt.wantArgs, fake.ExecuteArgs,
					"kubectl args mismatch")
			}
		})
	}
}