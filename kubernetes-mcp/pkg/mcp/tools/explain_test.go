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

func TestToolsCreateExplainTool(t *testing.T) {
	tools := New()
	tool := tools.CreateExplainTool()

	// Verify basic tool properties
	assert.Equal(t, "explain", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "Explain")
	assert.Contains(t, tool.Description, "resource types")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"resource",
		"context",
		"recursive",
	}

	for _, param := range expectedParams {
		assert.Contains(t, tool.InputSchema.Properties, param)
	}

	// Verify required parameters
	assert.Contains(t, tool.InputSchema.Required, "resource")
	assert.Len(t, tool.InputSchema.Required, 1)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleExplain(t *testing.T) {
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
			name: "explain pods resource",
			args: map[string]any{
				"resource": "pods",
			},
			kubectlStdout: `KIND:     Pod
VERSION:  v1

DESCRIPTION:
     Pod is a collection of containers that can run on a host. This resource is
     created by clients and scheduled onto hosts.

FIELDS:
   apiVersion	<string>
   kind	<string>
   metadata	<Object>
   spec	<Object>
   status	<Object>`,
			wantArgs: []string{"explain", "pods"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "KIND:     Pod")
				assert.Contains(t, content.Text, "collection of containers")
				assert.Contains(t, content.Text, "metadata")
				assert.Contains(t, content.Text, "spec")
			},
		},
		{
			name: "explain nested field",
			args: map[string]any{
				"resource": "pods.spec.containers",
			},
			kubectlStdout: `KIND:     Pod
VERSION:  v1

RESOURCE: containers <[]Object>

DESCRIPTION:
     List of containers belonging to the pod. Containers cannot currently be
     added or removed. There must be at least one container in a Pod. Cannot be
     updated.

     A single application container that you want to run within a pod.

FIELDS:
   args	<[]string>
   command	<[]string>
   env	<[]Object>
   image	<string>
   name	<string> -required-`,
			wantArgs: []string{"explain", "pods.spec.containers"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "RESOURCE: containers")
				assert.Contains(t, content.Text, "image")
				assert.Contains(t, content.Text, "name")
			},
		},
		{
			name: "explain with context",
			args: map[string]any{
				"resource": "deployments",
				"context":  "production",
			},
			kubectlStdout: `KIND:     Deployment
VERSION:  apps/v1

DESCRIPTION:
     Deployment enables declarative updates for Pods and ReplicaSets.`,
			wantArgs: []string{"--context", "production", "explain", "deployments"},
		},
		{
			name: "explain with recursive flag",
			args: map[string]any{
				"resource":  "services",
				"recursive": true,
			},
			kubectlStdout: `KIND:     Service
VERSION:  v1

DESCRIPTION:
     Service is a named abstraction of software service consisting of local port

FIELDS:
   apiVersion	<string>
   kind	<string>
   metadata	<Object>
      annotations	<map[string]string>
      labels	<map[string]string>
      name	<string>
      namespace	<string>
   spec	<Object>
      clusterIP	<string>
      ports	<[]Object>
         name	<string>
         port	<integer> -required-
         protocol	<string>
         targetPort	<string>`,
			wantArgs: []string{"explain", "services", "--recursive"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "KIND:     Service")
				assert.Contains(t, content.Text, "annotations")
				assert.Contains(t, content.Text, "clusterIP")
				assert.Contains(t, content.Text, "targetPort")
			},
		},
		{
			name: "explain with all parameters",
			args: map[string]any{
				"resource":  "ingresses.spec.rules",
				"context":   "staging",
				"recursive": true,
			},
			kubectlStdout: `KIND:     Ingress
VERSION:  networking.k8s.io/v1

RESOURCE: rules <[]Object>

DESCRIPTION:
     A list of host rules used to configure the Ingress.`,
			wantArgs: []string{"--context", "staging", "explain", "ingresses.spec.rules", "--recursive"},
		},
		{
			name: "empty context parameter is ignored",
			args: map[string]any{
				"resource": "configmaps",
				"context":  "",
			},
			kubectlStdout: `KIND:     ConfigMap`,
			wantArgs:      []string{"explain", "configmaps"},
		},
		{
			name: "recursive false is ignored",
			args: map[string]any{
				"resource":  "secrets",
				"recursive": false,
			},
			kubectlStdout: `KIND:     Secret`,
			wantArgs:      []string{"explain", "secrets"},
		},
		{
			name:         "missing resource parameter",
			args:         map[string]any{},
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
			name: "resource is not a string",
			args: map[string]any{
				"resource": 123,
			},
			wantError:    true,
			wantErrorMsg: "resource parameter required",
		},
		{
			name: "kubectl error - unknown resource",
			args: map[string]any{
				"resource": "unknownresource",
			},
			kubectlStderr: `error: couldn't find resource for "unknownresource"`,
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"explain", "unknownresource"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "kubectl command failed")
				assert.Contains(t, content.Text, "couldn't find resource")
			},
		},
		{
			name: "explain CRD resource",
			args: map[string]any{
				"resource": "virtualservices.networking.istio.io",
			},
			kubectlStdout: `KIND:     VirtualService
VERSION:  networking.istio.io/v1beta1

DESCRIPTION:
     Configuration affecting traffic routing.`,
			wantArgs: []string{"explain", "virtualservices.networking.istio.io"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "VirtualService")
				assert.Contains(t, content.Text, "traffic routing")
			},
		},
		{
			name: "explain statefulsets",
			args: map[string]any{
				"resource": "statefulsets",
			},
			kubectlStdout: `KIND:     StatefulSet
VERSION:  apps/v1

DESCRIPTION:
     StatefulSet represents a set of pods with consistent identities.`,
			wantArgs: []string{"explain", "statefulsets"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "StatefulSet")
				assert.Contains(t, content.Text, "consistent identities")
			},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"resource": "pods",
				"unknown":  "value",
				"extra":    123,
			},
			kubectlStdout: `KIND:     Pod`,
			wantArgs:      []string{"explain", "pods"},
		},
		{
			name: "numeric context and recursive values are ignored",
			args: map[string]any{
				"resource":  "nodes",
				"context":   123,
				"recursive": "not-bool",
			},
			kubectlStdout: `KIND:     Node`,
			wantArgs:      []string{"explain", "nodes"},
		},
		{
			name: "explain deeply nested field",
			args: map[string]any{
				"resource": "pods.spec.containers.env.valueFrom.secretKeyRef",
			},
			kubectlStdout: `KIND:     Pod
VERSION:  v1

RESOURCE: secretKeyRef <Object>

DESCRIPTION:
     Selects a key of a secret in the pod's namespace

FIELDS:
   key	<string> -required-
   name	<string>
   optional	<boolean>`,
			wantArgs: []string{"explain", "pods.spec.containers.env.valueFrom.secretKeyRef"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "secretKeyRef")
				assert.Contains(t, content.Text, "key")
				assert.Contains(t, content.Text, "-required-")
			},
		},
		{
			name: "explain with api version",
			args: map[string]any{
				"resource": "deployments.apps",
			},
			kubectlStdout: `KIND:     Deployment
VERSION:  apps/v1

DESCRIPTION:
     Deployment enables declarative updates for Pods and ReplicaSets.`,
			wantArgs: []string{"explain", "deployments.apps"},
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
					Name:      "explain",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleExplain(context.Background(), req)

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
