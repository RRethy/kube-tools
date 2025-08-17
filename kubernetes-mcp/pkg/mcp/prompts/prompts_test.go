package prompts

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/kubectl"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testKubectl is a test implementation that can return different responses for different commands
type testKubectl struct {
	contextResponse   string
	namespaceResponse string
	contextError      error
	namespaceError    error
	calls             [][]string
}

func (t *testKubectl) Execute(ctx context.Context, args ...string) (string, string, error) {
	t.calls = append(t.calls, args)

	// Check if this is a current-context call
	for _, arg := range args {
		if arg == "current-context" {
			return t.contextResponse, "", t.contextError
		}
	}

	// Check if this is a namespace query
	if contains(args, "config") && contains(args, "view") && strings.Contains(strings.Join(args, " "), "jsonpath") {
		return t.namespaceResponse, "", t.namespaceError
	}

	return "", "", nil
}

func TestPromptsGetCurrentContext(t *testing.T) {
	tests := []struct {
		name           string
		kubectlOutput  string
		kubectlError   error
		expectedResult string
		expectedError  bool
	}{
		{
			name:           "successful context retrieval",
			kubectlOutput:  "production-cluster\n",
			expectedResult: "production-cluster",
		},
		{
			name:           "context with extra whitespace",
			kubectlOutput:  "  staging-cluster  \n\n",
			expectedResult: "staging-cluster",
		},
		{
			name:          "kubectl error",
			kubectlError:  fmt.Errorf("kubectl not found"),
			expectedError: true,
		},
		{
			name:           "empty context",
			kubectlOutput:  "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &kubectl.FakeKubectl{
				ExecuteStdout: tt.kubectlOutput,
				ExecuteStderr: "",
				ExecuteError:  tt.kubectlError,
			}
			prompts := NewWithKubectl(fake)

			result, err := prompts.getCurrentContext(context.Background())

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get current context")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.True(t, fake.ExecuteCalled)
			assert.Contains(t, fake.ExecuteArgs, "config")
			assert.Contains(t, fake.ExecuteArgs, "current-context")
		})
	}
}

func TestPromptsGetCurrentNamespace(t *testing.T) {
	tests := []struct {
		name           string
		contextName    string
		kubectlOutput  string
		kubectlError   error
		expectedResult string
		expectedError  bool
		expectedArgs   []string
	}{
		{
			name:           "successful namespace retrieval",
			contextName:    "prod-context",
			kubectlOutput:  "production",
			expectedResult: "production",
			expectedArgs:   []string{"--context", "prod-context", "config", "view", "-o", "jsonpath={.contexts[?(@.name==\"prod-context\")].context.namespace}"},
		},
		{
			name:           "empty namespace returns default",
			contextName:    "test-context",
			kubectlOutput:  "",
			expectedResult: "default",
			expectedArgs:   []string{"--context", "test-context", "config", "view", "-o", "jsonpath={.contexts[?(@.name==\"test-context\")].context.namespace}"},
		},
		{
			name:           "namespace with whitespace",
			contextName:    "dev-context",
			kubectlOutput:  "  development  \n",
			expectedResult: "development",
			expectedArgs:   []string{"--context", "dev-context", "config", "view", "-o", "jsonpath={.contexts[?(@.name==\"dev-context\")].context.namespace}"},
		},
		{
			name:          "kubectl error",
			contextName:   "bad-context",
			kubectlError:  fmt.Errorf("context not found"),
			expectedError: true,
			expectedArgs:  []string{"--context", "bad-context", "config", "view", "-o", "jsonpath={.contexts[?(@.name==\"bad-context\")].context.namespace}"},
		},
		{
			name:           "no context name provided",
			contextName:    "",
			kubectlOutput:  "kube-system",
			expectedResult: "kube-system",
			expectedArgs:   []string{"config", "view", "-o", "jsonpath={.contexts[?(@.name==\"\")].context.namespace}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &kubectl.FakeKubectl{
				ExecuteStdout: tt.kubectlOutput,
				ExecuteStderr: "",
				ExecuteError:  tt.kubectlError,
			}
			prompts := NewWithKubectl(fake)

			result, err := prompts.getCurrentNamespace(context.Background(), tt.contextName)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get current namespace")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.True(t, fake.ExecuteCalled)
			assert.Equal(t, tt.expectedArgs, fake.ExecuteArgs)
		})
	}
}

func TestPromptsCreateDebugClusterPrompt(t *testing.T) {
	prompts := New()
	prompt := prompts.CreateDebugClusterPrompt()

	assert.Equal(t, "debug-cluster", prompt.Name)
	assert.NotEmpty(t, prompt.Description)
	assert.Contains(t, prompt.Description, "Debug a Kubernetes cluster")

	// Check that the context argument exists
	assert.Len(t, prompt.Arguments, 1)
	assert.Equal(t, "context", prompt.Arguments[0].Name)
	assert.NotEmpty(t, prompt.Arguments[0].Description)
}

func TestPromptsCreateDebugNamespacePrompt(t *testing.T) {
	prompts := New()
	prompt := prompts.CreateDebugNamespacePrompt()

	assert.Equal(t, "debug-namespace", prompt.Name)
	assert.NotEmpty(t, prompt.Description)
	assert.Contains(t, prompt.Description, "Debug a specific Kubernetes namespace")

	// Check that both arguments exist
	assert.Len(t, prompt.Arguments, 2)

	// Check argument names
	argNames := make([]string, len(prompt.Arguments))
	for i, arg := range prompt.Arguments {
		argNames[i] = arg.Name
	}
	assert.Contains(t, argNames, "namespace")
	assert.Contains(t, argNames, "context")
}

func TestPromptsHandleDebugClusterPrompt(t *testing.T) {
	tests := []struct {
		name            string
		arguments       map[string]string
		currentContext  string
		kubectlError    error
		expectedContext string
		checkPromptText func(*testing.T, string)
	}{
		{
			name:            "explicit context provided",
			arguments:       map[string]string{"context": "prod-cluster"},
			expectedContext: "prod-cluster",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context/Cluster: prod-cluster")
				assert.Contains(t, text, "Check Available MCP Servers")
				assert.Contains(t, text, "Cluster Overview")
				assert.Contains(t, text, "Node Analysis")
				assert.Contains(t, text, "System Namespaces Health")
			},
		},
		{
			name:            "no context provided, uses current",
			arguments:       map[string]string{},
			currentContext:  "current-cluster",
			expectedContext: "current-cluster",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context/Cluster: current-cluster")
			},
		},
		{
			name:            "empty context argument, uses current",
			arguments:       map[string]string{"context": ""},
			currentContext:  "default-cluster",
			expectedContext: "default-cluster",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context/Cluster: default-cluster")
			},
		},
		{
			name:            "no arguments at all",
			arguments:       nil,
			currentContext:  "fallback-cluster",
			expectedContext: "fallback-cluster",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context/Cluster: fallback-cluster")
				// Check for MCP server awareness
				assert.Contains(t, text, "FIRST check what MCP servers are available")
				assert.Contains(t, text, "Cloud provider MCP servers")
				assert.Contains(t, text, "Monitoring/observability MCP servers")
			},
		},
		{
			name:            "kubectl error getting current context",
			arguments:       map[string]string{},
			kubectlError:    fmt.Errorf("kubectl failed"),
			expectedContext: "",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context/Cluster: ")
				// Should still have the full debugging instructions
				assert.Contains(t, text, "systematically analyze")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := kubectl.NewFake(tt.currentContext, "", tt.kubectlError)
			prompts := NewWithKubectl(fake)

			req := mcp.GetPromptRequest{
				Params: mcp.GetPromptParams{
					Arguments: tt.arguments,
				},
			}

			result, err := prompts.HandleDebugClusterPrompt(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check the description
			expectedDesc := fmt.Sprintf("Comprehensive cluster debugging for context %s", tt.expectedContext)
			assert.Equal(t, expectedDesc, result.Description)

			// Check messages
			require.Len(t, result.Messages, 1)
			assert.Equal(t, mcp.RoleUser, result.Messages[0].Role)

			// Check prompt text content
			textContent, ok := result.Messages[0].Content.(mcp.TextContent)
			require.True(t, ok, "expected text content in message")

			if tt.checkPromptText != nil {
				tt.checkPromptText(t, textContent.Text)
			}
		})
	}
}

func TestPromptsHandleDebugNamespacePrompt(t *testing.T) {
	tests := []struct {
		name              string
		arguments         map[string]string
		currentContext    string
		currentNamespace  string
		contextError      error
		namespaceError    error
		expectedContext   string
		expectedNamespace string
		checkPromptText   func(*testing.T, string)
	}{
		{
			name: "explicit context and namespace",
			arguments: map[string]string{
				"context":   "prod-cluster",
				"namespace": "production",
			},
			expectedContext:   "prod-cluster",
			expectedNamespace: "production",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context: prod-cluster")
				assert.Contains(t, text, "Namespace: production")
			},
		},
		{
			name:              "no arguments, uses current context and namespace",
			arguments:         map[string]string{},
			currentContext:    "dev-cluster",
			currentNamespace:  "development",
			expectedContext:   "dev-cluster",
			expectedNamespace: "development",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context: dev-cluster")
				assert.Contains(t, text, "Namespace: development")
			},
		},
		{
			name:              "no namespace provided, uses current",
			arguments:         map[string]string{"context": "test-cluster"},
			currentNamespace:  "testing",
			expectedContext:   "test-cluster",
			expectedNamespace: "testing",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Context: test-cluster")
				assert.Contains(t, text, "Namespace: testing")
			},
		},
		{
			name:              "namespace error defaults to 'default'",
			arguments:         map[string]string{},
			currentContext:    "some-cluster",
			namespaceError:    fmt.Errorf("namespace not found"),
			expectedContext:   "some-cluster",
			expectedNamespace: "default",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Namespace: default")
			},
		},
		{
			name:              "empty namespace returns default",
			arguments:         map[string]string{},
			currentContext:    "cluster",
			currentNamespace:  "",
			expectedContext:   "cluster",
			expectedNamespace: "default",
			checkPromptText: func(t *testing.T, text string) {
				assert.Contains(t, text, "Namespace: default")
				// Check for MCP server awareness
				assert.Contains(t, text, "FIRST check what MCP servers are available")
				assert.Contains(t, text, "Application-specific MCP servers")
				assert.Contains(t, text, "Database MCP servers")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For namespace tests, we need a more complex fake that can handle multiple calls
			// We'll create a simple mock that tracks calls and returns appropriate responses
			fake := &testKubectl{
				contextResponse:   tt.currentContext,
				namespaceResponse: tt.currentNamespace,
				contextError:      tt.contextError,
				namespaceError:    tt.namespaceError,
			}

			prompts := NewWithKubectl(fake)

			req := mcp.GetPromptRequest{
				Params: mcp.GetPromptParams{
					Arguments: tt.arguments,
				},
			}

			result, err := prompts.HandleDebugNamespacePrompt(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check the description
			expectedDesc := fmt.Sprintf("Comprehensive namespace debugging for %s in context %s",
				tt.expectedNamespace, tt.expectedContext)
			assert.Equal(t, expectedDesc, result.Description)

			// Check messages
			require.Len(t, result.Messages, 1)
			assert.Equal(t, mcp.RoleUser, result.Messages[0].Role)

			// Check prompt text content
			textContent, ok := result.Messages[0].Content.(mcp.TextContent)
			require.True(t, ok, "expected text content in message")

			if tt.checkPromptText != nil {
				tt.checkPromptText(t, textContent.Text)
			}
		})
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
