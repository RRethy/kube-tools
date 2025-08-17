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

func TestToolsCreateAuthCanITool(t *testing.T) {
	tools := New()
	tool := tools.CreateAuthCanITool()

	// Verify basic tool properties
	assert.Equal(t, "auth-can-i", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "action")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"verb",
		"resource",
		"resource-name",
		"namespace",
		"context",
		"all-namespaces",
	}

	for _, param := range expectedParams {
		assert.Contains(t, tool.InputSchema.Properties, param)
	}

	// Verify required parameters
	assert.Contains(t, tool.InputSchema.Required, "verb")
	assert.Contains(t, tool.InputSchema.Required, "resource")
	assert.Len(t, tool.InputSchema.Required, 2)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleAuthCanI(t *testing.T) {
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
			name: "basic authorization check",
			args: map[string]any{
				"verb":     "get",
				"resource": "pods",
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "get", "pods"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "yes")
			},
		},
		{
			name: "authorization denied",
			args: map[string]any{
				"verb":     "delete",
				"resource": "nodes",
			},
			kubectlStdout: "no",
			wantArgs:      []string{"auth", "can-i", "delete", "nodes"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "no")
			},
		},
		{
			name: "with context parameter",
			args: map[string]any{
				"verb":     "list",
				"resource": "services",
				"context":  "production",
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"--context", "production", "auth", "can-i", "list", "services"},
		},
		{
			name: "with specific resource name",
			args: map[string]any{
				"verb":          "get",
				"resource":      "pods",
				"resource-name": "my-pod",
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "get", "pods", "my-pod"},
		},
		{
			name: "with namespace",
			args: map[string]any{
				"verb":      "create",
				"resource":  "deployments",
				"namespace": "kube-system",
			},
			kubectlStdout: "no",
			wantArgs:      []string{"auth", "can-i", "create", "deployments", "-n", "kube-system"},
		},
		{
			name: "with all-namespaces flag",
			args: map[string]any{
				"verb":           "list",
				"resource":       "pods",
				"all-namespaces": true,
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "list", "pods", "--all-namespaces"},
		},
		{
			name: "namespace ignored when all-namespaces is true",
			args: map[string]any{
				"verb":           "get",
				"resource":       "services",
				"namespace":      "default",
				"all-namespaces": true,
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "get", "services", "--all-namespaces"},
		},
		{
			name: "all parameters combined",
			args: map[string]any{
				"verb":          "update",
				"resource":      "configmaps",
				"resource-name": "my-config",
				"namespace":     "my-namespace",
				"context":       "staging",
			},
			kubectlStdout: "yes",
			wantArgs: []string{
				"--context", "staging",
				"auth", "can-i", "update", "configmaps", "my-config",
				"-n", "my-namespace",
			},
		},
		{
			name: "empty string parameters are ignored",
			args: map[string]any{
				"verb":          "get",
				"resource":      "pods",
				"resource-name": "",
				"namespace":     "",
				"context":       "",
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "get", "pods"},
		},
		{
			name: "all-namespaces false is ignored",
			args: map[string]any{
				"verb":           "list",
				"resource":       "pods",
				"all-namespaces": false,
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "list", "pods"},
		},
		{
			name:         "missing verb parameter",
			args:         map[string]any{"resource": "pods"},
			wantError:    true,
			wantErrorMsg: "verb parameter required",
		},
		{
			name:         "missing resource parameter",
			args:         map[string]any{"verb": "get"},
			wantError:    true,
			wantErrorMsg: "resource parameter required",
		},
		{
			name:         "nil arguments returns error",
			args:         nil,
			wantError:    true,
			wantErrorMsg: "invalid arguments",
		},
		{
			name:         "invalid arguments type",
			args:         "not a map",
			wantError:    true,
			wantErrorMsg: "invalid arguments",
		},
		{
			name: "verb is not a string",
			args: map[string]any{
				"verb":     123,
				"resource": "pods",
			},
			wantError:    true,
			wantErrorMsg: "verb parameter required",
		},
		{
			name: "resource is not a string",
			args: map[string]any{
				"verb":     "get",
				"resource": []string{"pods"},
			},
			wantError:    true,
			wantErrorMsg: "resource parameter required",
		},
		{
			name: "kubectl error is handled",
			args: map[string]any{
				"verb":     "get",
				"resource": "pods",
			},
			kubectlStderr: "error: unable to connect to server",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"auth", "can-i", "get", "pods"},
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
			name: "unknown parameters are ignored",
			args: map[string]any{
				"verb":     "get",
				"resource": "pods",
				"unknown":  "value",
				"extra":    123,
			},
			kubectlStdout: "yes",
			wantArgs:      []string{"auth", "can-i", "get", "pods"},
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
					Name:      "auth-can-i",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleAuthCanI(context.Background(), req)

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
