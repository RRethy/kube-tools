package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/RRethy/kube-tools/kubernetes-mcp/pkg/kubectl"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsCreateDescribeTool(t *testing.T) {
	tools := New()
	tool := tools.CreateDescribeTool()

	// Verify basic tool properties
	assert.Equal(t, "describe", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "detailed information")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"resource-type",
		"resource-name",
		"namespace",
		"context",
	}

	for _, param := range expectedParams {
		assert.Contains(t, tool.InputSchema.Properties, param)
	}

	// Verify required parameters
	assert.Contains(t, tool.InputSchema.Required, "resource-type")
	assert.Contains(t, tool.InputSchema.Required, "resource-name")
	assert.Len(t, tool.InputSchema.Required, 2)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleDescribe(t *testing.T) {
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
			name: "describe pod successfully",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "nginx-pod",
			},
			kubectlStdout: `Name:         nginx-pod
Namespace:    default
Priority:     0
Node:         minikube/192.168.49.2
Start Time:   Mon, 01 Jan 2024 10:00:00 +0000
Labels:       app=nginx
Status:       Running`,
			wantArgs: []string{"describe", "pod", "nginx-pod"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "nginx-pod")
				assert.Contains(t, content.Text, "Status:       Running")
			},
		},
		{
			name: "describe deployment with namespace",
			args: map[string]any{
				"resource-type": "deployment",
				"resource-name": "web-app",
				"namespace":     "production",
			},
			kubectlStdout: `Name:                   web-app
Namespace:              production
CreationTimestamp:      Mon, 01 Jan 2024 10:00:00 +0000
Labels:                 app=web
Selector:               app=web
Replicas:               3 desired | 3 updated | 3 total | 3 available`,
			wantArgs: []string{"describe", "deployment", "web-app", "-n", "production"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "web-app")
				assert.Contains(t, content.Text, "Namespace:              production")
			},
		},
		{
			name: "describe service with context",
			args: map[string]any{
				"resource-type": "service",
				"resource-name": "api-service",
				"context":       "staging",
			},
			kubectlStdout: `Name:              api-service
Namespace:         default
Labels:            <none>
Type:              ClusterIP
IP:                10.96.0.1
Port:              <unset>  80/TCP`,
			wantArgs: []string{"--context", "staging", "describe", "service", "api-service"},
		},
		{
			name: "describe with all parameters",
			args: map[string]any{
				"resource-type": "configmap",
				"resource-name": "app-config",
				"namespace":     "backend",
				"context":       "production",
			},
			kubectlStdout: `Name:         app-config
Namespace:    backend
Labels:       <none>
Data
====
config.yaml:
----
key: value`,
			wantArgs: []string{"--context", "production", "describe", "configmap", "app-config", "-n", "backend"},
		},
		{
			name: "blocked resource type - secrets",
			args: map[string]any{
				"resource-type": "secret",
				"resource-name": "my-secret",
			},
			wantArgs: nil, // kubectl should not be called
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Access to resource type 'secret' is blocked")
				assert.Contains(t, content.Text, "security reasons")
			},
		},
		{
			name: "blocked resource type - secrets plural",
			args: map[string]any{
				"resource-type": "secrets",
				"resource-name": "my-secret",
			},
			wantArgs: nil, // kubectl should not be called
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Access to resource type 'secrets' is blocked")
			},
		},
		{
			name: "empty namespace parameter is ignored",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "test-pod",
				"namespace":     "",
			},
			kubectlStdout: "Name: test-pod",
			wantArgs:      []string{"describe", "pod", "test-pod"},
		},
		{
			name: "empty context parameter is ignored",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "test-pod",
				"context":       "",
			},
			kubectlStdout: "Name: test-pod",
			wantArgs:      []string{"describe", "pod", "test-pod"},
		},
		{
			name:         "missing resource-type parameter",
			args:         map[string]any{"resource-name": "test-pod"},
			wantError:    true,
			wantErrorMsg: "resource-type parameter required",
		},
		{
			name:         "missing resource-name parameter",
			args:         map[string]any{"resource-type": "pod"},
			wantError:    true,
			wantErrorMsg: "resource-name parameter required",
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
			name: "resource-type is not a string",
			args: map[string]any{
				"resource-type": 123,
				"resource-name": "test-pod",
			},
			wantError:    true,
			wantErrorMsg: "resource-type parameter required",
		},
		{
			name: "resource-name is not a string",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": []string{"test-pod"},
			},
			wantError:    true,
			wantErrorMsg: "resource-name parameter required",
		},
		{
			name: "kubectl error is handled",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "nonexistent-pod",
			},
			kubectlStderr: `Error from server (NotFound): pods "nonexistent-pod" not found`,
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"describe", "pod", "nonexistent-pod"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "kubectl command failed")
				assert.Contains(t, content.Text, "not found")
			},
		},
		{
			name: "describe node (cluster-scoped resource)",
			args: map[string]any{
				"resource-type": "node",
				"resource-name": "minikube",
			},
			kubectlStdout: `Name:               minikube
Roles:              control-plane,master
Labels:             kubernetes.io/arch=amd64
CreationTimestamp:  Mon, 01 Jan 2024 10:00:00 +0000`,
			wantArgs: []string{"describe", "node", "minikube"},
		},
		{
			name: "describe with events section",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "crashing-pod",
			},
			kubectlStdout: `Name:         crashing-pod
Namespace:    default
Status:       CrashLoopBackOff
Events:
  Type     Reason     Age   From               Message
  ----     ------     ----  ----               -------
  Normal   Scheduled  10m   default-scheduler  Successfully assigned default/crashing-pod to minikube
  Normal   Pulled     10m   kubelet            Container image "nginx:latest" already present on machine
  Warning  BackOff    5m    kubelet            Back-off restarting failed container`,
			wantArgs: []string{"describe", "pod", "crashing-pod"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Events:")
				assert.Contains(t, content.Text, "BackOff")
			},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "test-pod",
				"unknown":       "value",
				"extra":         123,
			},
			kubectlStdout: "Name: test-pod",
			wantArgs:      []string{"describe", "pod", "test-pod"},
		},
		{
			name: "describe ingress resource",
			args: map[string]any{
				"resource-type": "ingress",
				"resource-name": "app-ingress",
				"namespace":     "web",
			},
			kubectlStdout: `Name:             app-ingress
Namespace:        web
Address:          192.168.49.2
Default backend:  default-http-backend:80
Rules:
  Host        Path  Backends
  ----        ----  --------
  app.local   /     web-service:80`,
			wantArgs: []string{"describe", "ingress", "app-ingress", "-n", "web"},
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
					Name:      "describe",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleDescribe(context.Background(), req)

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

			// Check kubectl was called with correct args (if not an error test or blocked resource)
			if tt.wantArgs != nil {
				assert.True(t, fake.ExecuteCalled, "kubectl should have been called")
				assert.Equal(t, tt.wantArgs, fake.ExecuteArgs,
					"kubectl args mismatch")
			} else if tt.name == "blocked resource type - secrets" || tt.name == "blocked resource type - secrets plural" {
				// For blocked resources, kubectl should NOT be called
				assert.False(t, fake.ExecuteCalled, "kubectl should not have been called for blocked resource")
			}
		})
	}
}
