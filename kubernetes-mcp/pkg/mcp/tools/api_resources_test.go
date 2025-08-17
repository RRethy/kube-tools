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

func TestToolsCreateAPIResourcesTool(t *testing.T) {
	tools := New()
	tool := tools.CreateAPIResourcesTool()

	// Verify basic tool properties
	assert.Equal(t, "api-resources", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify all expected parameters are present
	expectedParams := []string{
		"context",
		"namespaced",
		"api-group",
		"sort-by",
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

func TestToolsHandleAPIResources(t *testing.T) {
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
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND\npods po v1 true Pod",
			wantArgs:      []string{"api-resources"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "pods")
			},
		},
		{
			name: "with context parameter",
			args: map[string]any{
				"context": "test-context",
			},
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND",
			wantArgs:      []string{"--context", "test-context", "api-resources"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				assert.NotNil(t, result)
				assert.False(t, result.IsError)
			},
		},
		{
			name: "with namespaced true",
			args: map[string]any{
				"namespaced": true,
			},
			kubectlStdout: "pods po v1 true Pod",
			wantArgs:      []string{"api-resources", "--namespaced=true"},
		},
		{
			name: "with namespaced false",
			args: map[string]any{
				"namespaced": false,
			},
			kubectlStdout: "namespaces ns v1 false Namespace",
			wantArgs:      []string{"api-resources", "--namespaced=false"},
		},
		{
			name: "with api-group parameter",
			args: map[string]any{
				"api-group": "apps",
			},
			kubectlStdout: "deployments deploy apps/v1 true Deployment",
			wantArgs:      []string{"api-resources", "--api-group", "apps"},
		},
		{
			name: "with sort-by parameter",
			args: map[string]any{
				"sort-by": "name",
			},
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND",
			wantArgs:      []string{"api-resources", "--sort-by", "name"},
		},
		{
			name: "with no-headers true",
			args: map[string]any{
				"no-headers": true,
			},
			kubectlStdout: "pods po v1 true Pod",
			wantArgs:      []string{"api-resources", "--no-headers"},
		},
		{
			name: "with no-headers false does not add flag",
			args: map[string]any{
				"no-headers": false,
			},
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND",
			wantArgs:      []string{"api-resources"},
		},
		{
			name: "with output format",
			args: map[string]any{
				"output": "wide",
			},
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND VERBS",
			wantArgs:      []string{"api-resources", "-o", "wide"},
		},
		{
			name: "with all parameters combined",
			args: map[string]any{
				"context":    "prod",
				"namespaced": true,
				"api-group":  "apps",
				"sort-by":    "kind",
				"no-headers": true,
				"output":     "name",
			},
			kubectlStdout: "deployments.apps\nreplicasets.apps",
			wantArgs: []string{
				"--context", "prod",
				"api-resources",
				"--namespaced=true",
				"--api-group", "apps",
				"--sort-by", "kind",
				"--no-headers",
				"-o", "name",
			},
		},
		{
			name: "empty string parameters are ignored",
			args: map[string]any{
				"context":   "",
				"api-group": "",
				"sort-by":   "",
				"output":    "",
			},
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND",
			wantArgs:      []string{"api-resources"},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown":    "value",
				"context":    "my-context",
				"not-a-flag": 123,
			},
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND",
			wantArgs:      []string{"--context", "my-context", "api-resources"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: "NAME SHORTNAMES APIVERSION NAMESPACED KIND",
			wantArgs:      []string{"api-resources"},
		},
		{
			name:          "kubectl error is handled",
			args:          map[string]any{},
			kubectlStderr: "error: unable to connect to server",
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"api-resources"},
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
			name:         "invalid arguments type returns error",
			args:         "not a map",
			wantError:    true,
			wantErrorMsg: "invalid arguments",
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				assert.Nil(t, result)
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
					Name:      "api-resources",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleAPIResources(context.Background(), req)

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
