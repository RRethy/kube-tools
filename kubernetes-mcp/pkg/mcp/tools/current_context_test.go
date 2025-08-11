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

func TestToolsCreateCurrentContextTool(t *testing.T) {
	tools := New()
	tool := tools.CreateCurrentContextTool()

	// Verify basic tool properties
	assert.Equal(t, "current-context", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "current")
	assert.Contains(t, tool.Description, "context")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify no parameters
	assert.Empty(t, tool.InputSchema.Properties)

	// Verify no required parameters
	assert.Empty(t, tool.InputSchema.Required)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleCurrentContext(t *testing.T) {
	tests := []struct {
		name          string
		args          any
		kubectlStdout string
		kubectlStderr string
		kubectlError  error
		wantArgs      []string
		checkResult   func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name:          "returns current context successfully",
			args:          map[string]any{},
			kubectlStdout: "docker-desktop",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "docker-desktop", content.Text)
			},
		},
		{
			name:          "returns different context name",
			args:          nil,
			kubectlStdout: "production-cluster",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "production-cluster", content.Text)
			},
		},
		{
			name:          "handles minikube context",
			args:          map[string]any{},
			kubectlStdout: "minikube",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "minikube", content.Text)
			},
		},
		{
			name:          "handles kind context",
			args:          map[string]any{},
			kubectlStdout: "kind-kind",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "kind-kind", content.Text)
			},
		},
		{
			name:          "handles long context name",
			args:          map[string]any{},
			kubectlStdout: "arn:aws:eks:us-west-2:123456789012:cluster/my-production-cluster",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "arn:aws:eks")
				assert.Contains(t, content.Text, "my-production-cluster")
			},
		},
		{
			name:          "ignores any provided arguments",
			args:          map[string]any{"unused": "value", "context": "ignored"},
			kubectlStdout: "current-context-name",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "current-context-name", content.Text)
			},
		},
		{
			name:          "handles no current context error",
			args:          map[string]any{},
			kubectlStderr: "error: current-context is not set",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "kubectl command failed")
				assert.Contains(t, content.Text, "current-context is not set")
			},
		},
		{
			name:          "handles kubeconfig not found error",
			args:          map[string]any{},
			kubectlStderr: "error: unable to read kubeconfig",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"config", "current-context"},
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
			name:          "handles permission denied error",
			args:          map[string]any{},
			kubectlStderr: "error: open /home/user/.kube/config: permission denied",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "permission denied")
			},
		},
		{
			name:          "handles warning with successful output",
			args:          map[string]any{},
			kubectlStdout: "docker-desktop",
			kubectlStderr: "Warning: kubeconfig file is world-readable",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "docker-desktop")
				assert.Contains(t, content.Text, "Warning")
			},
		},
		{
			name:          "handles empty current context",
			args:          map[string]any{},
			kubectlStdout: "",
			kubectlStderr: "",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Empty(t, content.Text)
			},
		},
		{
			name:          "handles context with special characters",
			args:          map[string]any{},
			kubectlStdout: "my-context@cluster-1",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "my-context@cluster-1", content.Text)
			},
		},
		{
			name:          "handles GKE context format",
			args:          map[string]any{},
			kubectlStdout: "gke_my-project_us-central1-a_my-cluster",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "gke_my-project_us-central1-a_my-cluster", content.Text)
			},
		},
		{
			name:          "works with string arguments (ignored)",
			args:          "not a map",
			kubectlStdout: "some-context",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "some-context", content.Text)
			},
		},
		{
			name:          "works with array arguments (ignored)",
			args:          []string{"test"},
			kubectlStdout: "another-context",
			wantArgs:      []string{"config", "current-context"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Equal(t, "another-context", content.Text)
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
					Name:      "current-context",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleCurrentContext(context.Background(), req)

			// No error expected from handler itself (errors are in result)
			assert.NoError(t, err)

			// Check result
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}

			// Check kubectl was called with correct args
			if tt.wantArgs != nil {
				assert.True(t, fake.ExecuteCalled, "kubectl should have been called")
				assert.Equal(t, tt.wantArgs, fake.ExecuteArgs,
					"kubectl args mismatch")
			}
		})
	}
}