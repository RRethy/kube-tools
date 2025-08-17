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

func TestToolsCreateConfigGetContextsTool(t *testing.T) {
	tools := New()
	tool := tools.CreateConfigGetContextsTool()

	// Verify basic tool properties
	assert.Equal(t, "config-get-contexts", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "contexts")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"no-headers",
		"output",
	}

	for _, param := range expectedParams {
		assert.Contains(t, tool.InputSchema.Properties, param)
	}

	// Verify no required parameters
	assert.Empty(t, tool.InputSchema.Required)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleConfigGetContexts(t *testing.T) {
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
			kubectlStdout: "CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE\n*         docker-desktop       docker-desktop       docker-desktop       default\n          minikube             minikube             minikube             default",
			wantArgs:      []string{"config", "get-contexts"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "docker-desktop")
				assert.Contains(t, content.Text, "minikube")
				assert.Contains(t, content.Text, "CURRENT")
			},
		},
		{
			name: "with no-headers flag",
			args: map[string]any{
				"no-headers": true,
			},
			kubectlStdout: "*         docker-desktop       docker-desktop       docker-desktop       default\n          minikube             minikube             minikube             default",
			wantArgs:      []string{"config", "get-contexts", "--no-headers"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "docker-desktop")
				assert.NotContains(t, content.Text, "CURRENT")
			},
		},
		{
			name: "with output=name",
			args: map[string]any{
				"output": "name",
			},
			kubectlStdout: "docker-desktop\nminikube\nproduction",
			wantArgs:      []string{"config", "get-contexts", "-o", "name"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "docker-desktop")
				assert.Contains(t, content.Text, "minikube")
				assert.Contains(t, content.Text, "production")
			},
		},
		{
			name: "with both no-headers and output=name",
			args: map[string]any{
				"no-headers": true,
				"output":     "name",
			},
			kubectlStdout: "docker-desktop\nminikube",
			wantArgs:      []string{"config", "get-contexts", "--no-headers", "-o", "name"},
		},
		{
			name: "no-headers false is ignored",
			args: map[string]any{
				"no-headers": false,
			},
			kubectlStdout: "CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE",
			wantArgs:      []string{"config", "get-contexts"},
		},
		{
			name: "empty output parameter is ignored",
			args: map[string]any{
				"output": "",
			},
			kubectlStdout: "CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE",
			wantArgs:      []string{"config", "get-contexts"},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown":    "value",
				"no-headers": true,
				"extra":      123,
			},
			kubectlStdout: "docker-desktop       docker-desktop       docker-desktop       default",
			wantArgs:      []string{"config", "get-contexts", "--no-headers"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: "CURRENT   NAME                 CLUSTER              AUTHINFO             NAMESPACE",
			wantArgs:      []string{"config", "get-contexts"},
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
			wantArgs:      []string{"config", "get-contexts"},
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
			name: "numeric no-headers value is ignored",
			args: map[string]any{
				"no-headers": 123,
			},
			kubectlStdout: "CURRENT   NAME                 CLUSTER",
			wantArgs:      []string{"config", "get-contexts"},
		},
		{
			name: "numeric output value is ignored",
			args: map[string]any{
				"output": 456,
			},
			kubectlStdout: "CURRENT   NAME                 CLUSTER",
			wantArgs:      []string{"config", "get-contexts"},
		},
		{
			name:          "empty kubeconfig returns empty result",
			args:          map[string]any{},
			kubectlStdout: "",
			wantArgs:      []string{"config", "get-contexts"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Empty(t, content.Text)
			},
		},
		{
			name: "single context with current marker",
			args: map[string]any{
				"no-headers": true,
			},
			kubectlStdout: "*         docker-desktop       docker-desktop       docker-desktop       default",
			wantArgs:      []string{"config", "get-contexts", "--no-headers"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "*")
				assert.Contains(t, content.Text, "docker-desktop")
			},
		},
		{
			name: "multiple contexts with various namespaces",
			args: map[string]any{
				"output": "name",
			},
			kubectlStdout: "dev-cluster\nstaging-cluster\nproduction-cluster\nlocal-kind",
			wantArgs:      []string{"config", "get-contexts", "-o", "name"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "dev-cluster")
				assert.Contains(t, content.Text, "staging-cluster")
				assert.Contains(t, content.Text, "production-cluster")
				assert.Contains(t, content.Text, "local-kind")
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
					Name:      "config-get-contexts",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleConfigGetContexts(context.Background(), req)

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
