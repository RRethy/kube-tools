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

func TestToolsCreateGetTool(t *testing.T) {
	tools := New()
	tool := tools.CreateGetTool()

	// Verify basic tool properties
	assert.Equal(t, "get", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.Description, "Get Kubernetes resources")
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Verify expected parameters
	expectedParams := []string{
		"resource-type",
		"resource-name",
		"namespace",
		"context",
		"selector",
		"output",
		"all-namespaces",
		"field-selector",
		"show-labels",
		"no-headers",
		"sort-by",
		"show-kind",
		"label-columns",
	}

	for _, param := range expectedParams {
		assert.Contains(t, tool.InputSchema.Properties, param)
	}

	// Verify required parameters
	assert.Contains(t, tool.InputSchema.Required, "resource-type")
	assert.Len(t, tool.InputSchema.Required, 1)

	// Verify read-only annotation
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)
}

func TestToolsHandleGet(t *testing.T) {
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
			name: "get all pods",
			args: map[string]any{
				"resource-type": "pods",
			},
			kubectlStdout: `NAME                     READY   STATUS    RESTARTS   AGE
nginx-6799fc88d8-nxqwp   1/1     Running   0          5m
redis-7b9f9c6d7f-8xkjl   1/1     Running   0          3m`,
			wantArgs: []string{"get", "pods"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "nginx")
				assert.Contains(t, content.Text, "redis")
				assert.Contains(t, content.Text, "Running")
			},
		},
		{
			name: "get specific pod",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "nginx-pod",
			},
			kubectlStdout: `NAME        READY   STATUS    RESTARTS   AGE
nginx-pod   1/1     Running   0          10m`,
			wantArgs: []string{"get", "pod", "nginx-pod"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "nginx-pod")
			},
		},
		{
			name: "get with namespace",
			args: map[string]any{
				"resource-type": "services",
				"namespace":     "kube-system",
			},
			kubectlStdout: `NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
kube-dns     ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   10d`,
			wantArgs: []string{"get", "services", "-n", "kube-system"},
		},
		{
			name: "get with context",
			args: map[string]any{
				"resource-type": "deployments",
				"context":       "production",
			},
			kubectlStdout: `NAME      READY   UP-TO-DATE   AVAILABLE   AGE
web-app   3/3     3            3           30d`,
			wantArgs: []string{"--context", "production", "get", "deployments"},
		},
		{
			name: "get with label selector",
			args: map[string]any{
				"resource-type": "pods",
				"selector":      "app=nginx",
			},
			kubectlStdout: `NAME                     READY   STATUS    RESTARTS   AGE
nginx-6799fc88d8-nxqwp   1/1     Running   0          5m`,
			wantArgs: []string{"get", "pods", "-l", "app=nginx"},
		},
		{
			name: "get with json output",
			args: map[string]any{
				"resource-type": "configmap",
				"resource-name": "app-config",
				"output":        "json",
			},
			kubectlStdout: `{
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
        "name": "app-config",
        "namespace": "default"
    },
    "data": {
        "config.yaml": "key: value"
    }
}`,
			wantArgs: []string{"get", "configmap", "app-config", "-o", "json"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "\"apiVersion\"")
				assert.Contains(t, content.Text, "\"ConfigMap\"")
			},
		},
		{
			name: "get with yaml output",
			args: map[string]any{
				"resource-type": "service",
				"resource-name": "web-service",
				"output":        "yaml",
			},
			kubectlStdout: `apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  selector:
    app: web`,
			wantArgs: []string{"get", "service", "web-service", "-o", "yaml"},
		},
		{
			name: "get with wide output",
			args: map[string]any{
				"resource-type": "nodes",
				"output":        "wide",
			},
			kubectlStdout: `NAME       STATUS   ROLES                  AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
minikube   Ready    control-plane,master   10d   v1.25.3   192.168.49.2   <none>        Ubuntu 20.04.5 LTS   5.15.49-linuxkit    docker://20.10.20`,
			wantArgs: []string{"get", "nodes", "-o", "wide"},
		},
		{
			name: "get with all-namespaces",
			args: map[string]any{
				"resource-type":  "pods",
				"all-namespaces": true,
			},
			kubectlStdout: `NAMESPACE     NAME                                     READY   STATUS    RESTARTS   AGE
default       nginx-6799fc88d8-nxqwp                   1/1     Running   0          5m
kube-system   coredns-565d847f94-8kqwp                 1/1     Running   0          10d
kube-system   etcd-minikube                            1/1     Running   0          10d`,
			wantArgs: []string{"get", "pods", "--all-namespaces"},
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.False(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "NAMESPACE")
				assert.Contains(t, content.Text, "kube-system")
			},
		},
		{
			name: "all-namespaces overrides namespace",
			args: map[string]any{
				"resource-type":  "services",
				"namespace":      "default",
				"all-namespaces": true,
			},
			kubectlStdout: `NAMESPACE     NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)`,
			wantArgs:      []string{"get", "services", "--all-namespaces"},
		},
		{
			name: "all parameters combined",
			args: map[string]any{
				"resource-type": "pods",
				"namespace":     "production",
				"context":       "staging",
				"selector":      "tier=frontend",
				"output":        "name",
			},
			kubectlStdout: `pod/frontend-1
pod/frontend-2
pod/frontend-3`,
			wantArgs: []string{
				"--context", "staging",
				"get", "pods",
				"-n", "production",
				"-l", "tier=frontend",
				"-o", "name",
			},
		},
		{
			name: "blocked resource type - secrets",
			args: map[string]any{
				"resource-type": "secrets",
			},
			wantArgs: nil, // kubectl should not be called
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)
				content := result.Content[0].(mcp.TextContent)
				assert.Contains(t, content.Text, "Access to resource type 'secrets' is blocked")
				assert.Contains(t, content.Text, "security reasons")
			},
		},
		{
			name: "empty string parameters are ignored",
			args: map[string]any{
				"resource-type": "pods",
				"resource-name": "",
				"namespace":     "",
				"context":       "",
				"selector":      "",
				"output":        "",
			},
			kubectlStdout: `NAME   READY   STATUS`,
			wantArgs:      []string{"get", "pods"},
		},
		{
			name: "all-namespaces false is ignored",
			args: map[string]any{
				"resource-type":  "pods",
				"all-namespaces": false,
			},
			kubectlStdout: `NAME   READY   STATUS`,
			wantArgs:      []string{"get", "pods"},
		},
		{
			name:         "missing resource-type parameter",
			args:         map[string]any{"resource-name": "test"},
			wantError:    true,
			wantErrorMsg: "resource-type parameter required",
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
			},
			wantError:    true,
			wantErrorMsg: "resource-type parameter required",
		},
		{
			name: "kubectl error - resource not found",
			args: map[string]any{
				"resource-type": "pod",
				"resource-name": "nonexistent",
			},
			kubectlStderr: `Error from server (NotFound): pods "nonexistent" not found`,
			kubectlError:  fmt.Errorf("exit status 1"),
			wantArgs:      []string{"get", "pod", "nonexistent"},
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
			name: "no resources found",
			args: map[string]any{
				"resource-type": "pods",
				"namespace":     "empty-namespace",
			},
			kubectlStdout: "No resources found in empty-namespace namespace.",
			wantArgs:      []string{"get", "pods", "-n", "empty-namespace"},
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
				"resource-type": "pods",
				"unknown":       "value",
				"extra":         123,
			},
			kubectlStdout: `NAME   READY   STATUS`,
			wantArgs:      []string{"get", "pods"},
		},
		{
			name: "numeric parameter values are ignored",
			args: map[string]any{
				"resource-type":  "pods",
				"resource-name":  123,
				"namespace":      456,
				"context":        789,
				"selector":       true,
				"output":         false,
				"all-namespaces": "not-bool",
			},
			kubectlStdout: `NAME   READY   STATUS`,
			wantArgs:      []string{"get", "pods"},
		},
		{
			name: "get CRD resources",
			args: map[string]any{
				"resource-type": "virtualservices.networking.istio.io",
				"namespace":     "istio-system",
			},
			kubectlStdout: `NAME          GATEWAYS   HOSTS   AGE
bookinfo-vs   [gateway]  [*]     5d`,
			wantArgs: []string{"get", "virtualservices.networking.istio.io", "-n", "istio-system"},
		},
		{
			name: "with field-selector",
			args: map[string]any{
				"resource-type":  "pods",
				"field-selector": "status.phase=Running",
			},
			kubectlStdout: "pod-1\npod-2",
			wantArgs:      []string{"get", "pods", "--field-selector", "status.phase=Running"},
		},
		{
			name: "with show-labels",
			args: map[string]any{
				"resource-type": "pods",
				"show-labels":   true,
			},
			kubectlStdout: "NAME    READY   STATUS    LABELS\npod-1   1/1     Running   app=web",
			wantArgs:      []string{"get", "pods", "--show-labels"},
		},
		{
			name: "with no-headers",
			args: map[string]any{
				"resource-type": "pods",
				"no-headers":    true,
			},
			kubectlStdout: "pod-1   1/1   Running   0   5m",
			wantArgs:      []string{"get", "pods", "--no-headers"},
		},
		{
			name: "with sort-by",
			args: map[string]any{
				"resource-type": "pods",
				"sort-by":       "{.metadata.name}",
			},
			kubectlStdout: "pod-a\npod-b\npod-c",
			wantArgs:      []string{"get", "pods", "--sort-by", "{.metadata.name}"},
		},
		{
			name: "with show-kind",
			args: map[string]any{
				"resource-type": "pods",
				"show-kind":     true,
			},
			kubectlStdout: "KIND   NAME    READY\nPod    pod-1   1/1",
			wantArgs:      []string{"get", "pods", "--show-kind"},
		},
		{
			name: "with label-columns",
			args: map[string]any{
				"resource-type": "pods",
				"label-columns": "app,version",
			},
			kubectlStdout: "NAME    APP   VERSION\npod-1   web   v1.0",
			wantArgs:      []string{"get", "pods", "--label-columns", "app,version"},
		},
		{
			name: "multiple new parameters combined",
			args: map[string]any{
				"resource-type":  "pods",
				"namespace":      "prod",
				"field-selector": "status.phase=Running",
				"show-labels":    true,
				"no-headers":     true,
				"sort-by":        "{.metadata.name}",
			},
			kubectlStdout: "pod-1   1/1   Running   0   5m   app=web",
			wantArgs: []string{
				"get", "pods", "-n", "prod",
				"--field-selector", "status.phase=Running",
				"--show-labels", "--no-headers",
				"--sort-by", "{.metadata.name}",
			},
		},
		{
			name: "empty field-selector ignored",
			args: map[string]any{
				"resource-type":  "pods",
				"field-selector": "",
			},
			kubectlStdout: "pod-1",
			wantArgs:      []string{"get", "pods"},
		},
		{
			name: "false boolean flags ignored",
			args: map[string]any{
				"resource-type": "pods",
				"show-labels":   false,
				"no-headers":    false,
				"show-kind":     false,
			},
			kubectlStdout: "pod-1",
			wantArgs:      []string{"get", "pods"},
		},
		{
			name: "invalid type for new parameters ignored",
			args: map[string]any{
				"resource-type":  "pods",
				"field-selector": 123,
				"show-labels":    "yes",
				"sort-by":        true,
				"label-columns":  456,
			},
			kubectlStdout: "pod-1",
			wantArgs:      []string{"get", "pods"},
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
					Name:      "get",
					Arguments: tt.args,
				},
			}

			// Call handler
			result, err := tools.HandleGet(context.Background(), req)

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
			} else if tt.name == "blocked resource type - secrets" {
				// For blocked resources, kubectl should NOT be called
				assert.False(t, fake.ExecuteCalled, "kubectl should not have been called for blocked resource")
			}
		})
	}
}
