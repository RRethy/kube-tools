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

func TestToolsCreateClusterInfoTool(t *testing.T) {
	tools := New()
	tool := tools.CreateClusterInfoTool()

	// Verify basic tool properties
	assert.Equal(t, "cluster-info", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "cluster information")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	assert.Contains(t, tool.InputSchema.Properties, "context")

	// Verify no required parameters
	assert.Empty(t, tool.InputSchema.Required)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleClusterInfo(t *testing.T) {
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
			kubectlStdout: "Kubernetes control plane is running at https://127.0.0.1:6443\nCoreDNS is running at https://127.0.0.1:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy",
			wantArgs:      []string{"cluster-info"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Kubernetes control plane")
				assert.Contains(t, content.Text, "CoreDNS")
			},
		},
		{
			name: "with context parameter",
			args: map[string]any{
				"context": "production",
			},
			kubectlStdout: "Kubernetes control plane is running at https://prod.example.com:6443",
			wantArgs:      []string{"--context", "production", "cluster-info"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "prod.example.com")
			},
		},
		{
			name: "empty context parameter is ignored",
			args: map[string]any{
				"context": "",
			},
			kubectlStdout: "Kubernetes control plane is running at https://127.0.0.1:6443",
			wantArgs:      []string{"cluster-info"},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown": "value",
				"context": "dev",
				"extra":   123,
			},
			kubectlStdout: "Kubernetes control plane is running at https://dev.example.com:6443",
			wantArgs:      []string{"--context", "dev", "cluster-info"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: "Kubernetes control plane is running at https://127.0.0.1:6443",
			wantArgs:      []string{"cluster-info"},
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
			wantArgs:      []string{"cluster-info"},
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
				"context": "staging",
			},
			kubectlStdout: "Kubernetes control plane is running at https://staging.example.com:6443",
			kubectlStderr: "Warning: context 'staging' is using outdated configuration",
			wantArgs:      []string{"--context", "staging", "cluster-info"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "staging.example.com")
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
			kubectlStdout: "Kubernetes control plane is running at https://127.0.0.1:6443",
			wantArgs:      []string{"cluster-info"},
		},
		{
			name: "context not found error",
			args: map[string]any{
				"context": "nonexistent",
			},
			kubectlStderr: "error: context \"nonexistent\" does not exist",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"--context", "nonexistent", "cluster-info"},
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
			name:          "successful cluster info with multiple services",
			args:          map[string]any{},
			kubectlStdout: "Kubernetes control plane is running at https://10.0.0.1:6443\nCoreDNS is running at https://10.0.0.1:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy\nMetrics-server is running at https://10.0.0.1:6443/api/v1/namespaces/kube-system/services/https:metrics-server:/proxy",
			wantArgs:      []string{"cluster-info"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Kubernetes control plane")
				assert.Contains(t, content.Text, "CoreDNS")
				assert.Contains(t, content.Text, "Metrics-server")
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
					Name:      "cluster-info",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleClusterInfo(context.Background(), req)

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