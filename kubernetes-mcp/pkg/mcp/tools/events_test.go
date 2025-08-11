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

func TestToolsCreateEventsTool(t *testing.T) {
	tools := New()
	tool := tools.CreateEventsTool()

	// Verify basic tool properties
	assert.Equal(t, "events", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "events")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"namespace",
		"context",
		"for",
		"all-namespaces",
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

func TestToolsHandleEvents(t *testing.T) {
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
			name:          "basic events without arguments",
			args:          map[string]any{},
			kubectlStdout: `LAST SEEN   TYPE     REASON              OBJECT                MESSAGE
5m          Normal   Scheduled           pod/nginx             Successfully assigned default/nginx to minikube
5m          Normal   Pulling             pod/nginx             Pulling image "nginx:latest"
4m          Normal   Pulled              pod/nginx             Successfully pulled image "nginx:latest"
4m          Normal   Created             pod/nginx             Created container nginx
4m          Normal   Started             pod/nginx             Started container nginx`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "nginx")
				assert.Contains(t, content.Text, "Scheduled")
				assert.Contains(t, content.Text, "Started")
			},
		},
		{
			name: "events with namespace",
			args: map[string]any{
				"namespace": "kube-system",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON                    OBJECT                          MESSAGE
10m         Normal   RegisteredNode            node/minikube                   Node minikube event: Registered Node minikube in Controller
10m         Normal   Starting                  node/minikube                   Starting kubelet.
10m         Normal   NodeHasSufficientMemory   node/minikube                   Node minikube status is now: NodeHasSufficientMemory`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "-n", "kube-system"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "kubelet")
			},
		},
		{
			name: "events with context",
			args: map[string]any{
				"context": "production",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT              MESSAGE
1m          Warning  BackOff    pod/app-1           Back-off restarting failed container`,
			wantArgs: []string{"--context", "production", "get", "events", "--sort-by=.lastTimestamp"},
		},
		{
			name: "events for specific resource",
			args: map[string]any{
				"for": "my-pod",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT       MESSAGE
30s         Normal   Scheduled  pod/my-pod   Successfully assigned default/my-pod to node-1
29s         Normal   Pulled     pod/my-pod   Container image already present on machine
28s         Normal   Created    pod/my-pod   Created container app
28s         Normal   Started    pod/my-pod   Started container app`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "--field-selector", "involvedObject.name=my-pod"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "my-pod")
				assert.Contains(t, content.Text, "Scheduled")
			},
		},
		{
			name: "events with all-namespaces",
			args: map[string]any{
				"all-namespaces": true,
			},
			kubectlStdout: `NAMESPACE     LAST SEEN   TYPE     REASON              OBJECT                          MESSAGE
default       5m          Normal   Scheduled           pod/nginx                       Successfully assigned default/nginx to minikube
kube-system   10m         Normal   RegisteredNode      node/minikube                   Node minikube event: Registered Node minikube in Controller
monitoring    3m          Normal   ScalingReplicaSet   deployment/prometheus-server    Scaled up replica set prometheus-server-5d9c7b4b9f to 1`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "--all-namespaces"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "default")
				assert.Contains(t, content.Text, "kube-system")
				assert.Contains(t, content.Text, "monitoring")
			},
		},
		{
			name: "all-namespaces overrides namespace",
			args: map[string]any{
				"all-namespaces": true,
				"namespace":      "default",
			},
			kubectlStdout: `NAMESPACE   LAST SEEN   TYPE     REASON     OBJECT       MESSAGE
default     1m          Normal   Scheduled  pod/app      Successfully assigned
kube-system 2m          Normal   Started    pod/coredns  Started container`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "--all-namespaces"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "NAMESPACE")
			},
		},
		{
			name: "all parameters combined",
			args: map[string]any{
				"context":        "staging",
				"all-namespaces": true,
				"for":            "web-app",
			},
			kubectlStdout: `NAMESPACE   LAST SEEN   TYPE     REASON     OBJECT           MESSAGE
production  1h          Normal   Scheduled  pod/web-app      Successfully assigned production/web-app to node-2`,
			wantArgs: []string{
				"--context", "staging",
				"get", "events", "--sort-by=.lastTimestamp",
				"--all-namespaces",
				"--field-selector", "involvedObject.name=web-app",
			},
		},
		{
			name: "empty string parameters are ignored",
			args: map[string]any{
				"namespace": "",
				"context":   "",
				"for":       "",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE
1m          Normal   Pulled     pod/test   Container image pulled`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp"},
		},
		{
			name: "all-namespaces false is ignored",
			args: map[string]any{
				"all-namespaces": false,
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
		},
		{
			name:          "nil arguments creates basic command",
			args:          nil,
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
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
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
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
			name: "numeric parameter values are ignored",
			args: map[string]any{
				"namespace":      123,
				"context":        456,
				"for":            789,
				"all-namespaces": "not-bool",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
		},
		{
			name: "warning events",
			args: map[string]any{
				"namespace": "default",
			},
			kubectlStdout: `LAST SEEN   TYPE      REASON                   OBJECT                 MESSAGE
2m          Warning   FailedScheduling         pod/pending-pod        0/3 nodes are available: 3 Insufficient cpu.
1m          Warning   BackOff                  pod/crashing-pod       Back-off restarting failed container
30s         Warning   FailedMount              pod/volume-pod         Unable to attach or mount volumes`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "-n", "default"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Warning")
				assert.Contains(t, content.Text, "FailedScheduling")
				assert.Contains(t, content.Text, "BackOff")
			},
		},
		{
			name: "no events available",
			args: map[string]any{
				"namespace": "empty-namespace",
			},
			kubectlStdout: "No resources found in empty-namespace namespace.",
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp", "-n", "empty-namespace"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "No resources found")
			},
		},
		{
			name: "unknown parameters are ignored",
			args: map[string]any{
				"unknown":   "value",
				"namespace": "default",
				"extra":     123,
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp", "-n", "default"},
		},
		{
			name: "with types filter for Warning",
			args: map[string]any{
				"types": "Warning",
			},
			kubectlStdout: `LAST SEEN   TYPE      REASON    OBJECT           MESSAGE
2m          Warning   BackOff   pod/crash-loop   Back-off restarting failed container`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "--field-selector", "type=Warning"},
		},
		{
			name: "with types filter for Normal",
			args: map[string]any{
				"types": "Normal",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON      OBJECT        MESSAGE
1m          Normal   Scheduled   pod/my-pod    Successfully assigned`,
			wantArgs: []string{"get", "events", "--sort-by=.lastTimestamp", "--field-selector", "type=Normal"},
		},
		{
			name: "with json output",
			args: map[string]any{
				"output": "json",
			},
			kubectlStdout: `{"items":[{"type":"Normal","reason":"Scheduled"}]}`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp", "-o", "json"},
		},
		{
			name: "with yaml output",
			args: map[string]any{
				"output": "yaml",
			},
			kubectlStdout: `apiVersion: v1
kind: EventList`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp", "-o", "yaml"},
		},
		{
			name: "with wide output",
			args: map[string]any{
				"output": "wide",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON    OBJECT     MESSAGE    FIRST SEEN   COUNT   NAME`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp", "-o", "wide"},
		},
		{
			name: "with no-headers",
			args: map[string]any{
				"no-headers": true,
			},
			kubectlStdout: `5m   Normal   Scheduled   pod/nginx   Successfully assigned`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp", "--no-headers"},
		},
		{
			name: "combined new parameters",
			args: map[string]any{
				"namespace":  "default",
				"types":      "Warning",
				"output":     "json",
				"no-headers": true,
			},
			kubectlStdout: `{"items":[]}`,
			wantArgs: []string{
				"get", "events", "--sort-by=.lastTimestamp",
				"-n", "default",
				"--field-selector", "type=Warning",
				"-o", "json",
				"--no-headers",
			},
		},
		{
			name: "empty new parameters ignored",
			args: map[string]any{
				"types":  "",
				"output": "",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
		},
		{
			name: "false no-headers ignored",
			args: map[string]any{
				"no-headers": false,
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
		},
		{
			name: "invalid types for new parameters ignored",
			args: map[string]any{
				"types":      123,
				"output":     true,
				"no-headers": "yes",
			},
			kubectlStdout: `LAST SEEN   TYPE     REASON     OBJECT     MESSAGE`,
			wantArgs:      []string{"get", "events", "--sort-by=.lastTimestamp"},
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
					Name:      "events",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleEvents(context.Background(), req)

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