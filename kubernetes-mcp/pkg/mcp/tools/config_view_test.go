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

func TestToolsCreateConfigViewTool(t *testing.T) {
	tools := New()
	tool := tools.CreateConfigViewTool()

	// Verify basic tool properties
	assert.Equal(t, "config-view", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "kubeconfig")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"minify",
		"flatten",
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

func TestToolsHandleConfigView(t *testing.T) {
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
			name: "basic command without arguments",
			args: map[string]any{},
			kubectlStdout: `apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: docker-desktop
contexts:
- context:
    cluster: docker-desktop
    user: docker-desktop
  name: docker-desktop
current-context: docker-desktop
kind: Config`,
			wantArgs: []string{"config", "view"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "apiVersion")
				assert.Contains(t, content.Text, "docker-desktop")
				assert.Contains(t, content.Text, "current-context")
			},
		},
		{
			name: "with minify flag",
			args: map[string]any{
				"minify": true,
			},
			kubectlStdout: `apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: docker-desktop
contexts:
- context:
    cluster: docker-desktop
    user: docker-desktop
  name: docker-desktop
current-context: docker-desktop`,
			wantArgs: []string{"config", "view", "--minify"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "docker-desktop")
			},
		},
		{
			name: "raw flag is ignored for security",
			args: map[string]any{
				"raw": true,
			},
			kubectlStdout: `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: REDACTED
    server: https://127.0.0.1:6443`,
			wantArgs: []string{"config", "view"}, // raw flag should be ignored
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "REDACTED")
			},
		},
		{
			name: "with flatten flag",
			args: map[string]any{
				"flatten": true,
			},
			kubectlStdout: `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi...
    server: https://127.0.0.1:6443
  name: docker-desktop`,
			wantArgs: []string{"config", "view", "--flatten"},
		},
		{
			name: "with output=json",
			args: map[string]any{
				"output": "json",
			},
			kubectlStdout: `{
  "apiVersion": "v1",
  "clusters": [
    {
      "cluster": {
        "server": "https://127.0.0.1:6443"
      },
      "name": "docker-desktop"
    }
  ]
}`,
			wantArgs: []string{"config", "view", "-o", "json"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "{")
				assert.Contains(t, content.Text, "\"apiVersion\"")
				assert.Contains(t, content.Text, "\"clusters\"")
			},
		},
		{
			name: "with output=yaml",
			args: map[string]any{
				"output": "yaml",
			},
			kubectlStdout: `apiVersion: v1
kind: Config`,
			wantArgs: []string{"config", "view", "-o", "yaml"},
		},
		{
			name: "with multiple flags combined",
			args: map[string]any{
				"minify":  true,
				"raw":     true, // should be ignored
				"flatten": true,
				"output":  "json",
			},
			kubectlStdout: `{"apiVersion":"v1","clusters":[{"cluster":{"certificate-authority-data":"REDACTED"}}]}`,
			wantArgs:      []string{"config", "view", "--minify", "--flatten", "-o", "json"}, // raw is not included
		},
		{
			name: "false boolean flags are ignored",
			args: map[string]any{
				"minify":  false,
				"flatten": false,
			},
			kubectlStdout: "apiVersion: v1",
			wantArgs:      []string{"config", "view"},
		},
		{
			name: "empty output parameter is ignored",
			args: map[string]any{
				"output": "",
			},
			kubectlStdout: "apiVersion: v1",
			wantArgs:      []string{"config", "view"},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown": "value",
				"minify":  true,
				"extra":   123,
			},
			kubectlStdout: "apiVersion: v1",
			wantArgs:      []string{"config", "view", "--minify"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: "apiVersion: v1\nkind: Config",
			wantArgs:      []string{"config", "view"},
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
			wantArgs:      []string{"config", "view"},
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
			name: "numeric boolean values are ignored",
			args: map[string]any{
				"minify":  123,
				"flatten": 789,
			},
			kubectlStdout: "apiVersion: v1",
			wantArgs:      []string{"config", "view"},
		},
		{
			name: "numeric output value is ignored",
			args: map[string]any{
				"output": 456,
			},
			kubectlStdout: "apiVersion: v1",
			wantArgs:      []string{"config", "view"},
		},
		{
			name:          "empty kubeconfig returns empty result",
			args:          map[string]any{},
			kubectlStdout: "",
			wantArgs:      []string{"config", "view"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Empty(t, content.Text)
			},
		},
		{
			name: "large kubeconfig with multiple contexts",
			args: map[string]any{
				"minify": false,
			},
			kubectlStdout: `apiVersion: v1
clusters:
- cluster:
    server: https://dev.example.com:6443
  name: dev-cluster
- cluster:
    server: https://staging.example.com:6443
  name: staging-cluster
- cluster:
    server: https://prod.example.com:6443
  name: prod-cluster
contexts:
- context:
    cluster: dev-cluster
    user: dev-user
  name: dev
- context:
    cluster: staging-cluster
    user: staging-user
  name: staging
- context:
    cluster: prod-cluster
    user: prod-user
  name: production
current-context: dev`,
			wantArgs: []string{"config", "view"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "dev-cluster")
				assert.Contains(t, content.Text, "staging-cluster")
				assert.Contains(t, content.Text, "prod-cluster")
				assert.Contains(t, content.Text, "current-context: dev")
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
					Name:      "config-view",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleConfigView(context.Background(), req)

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
