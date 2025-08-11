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

func TestNew(t *testing.T) {
	tools := New()
	assert.NotNil(t, tools)
	assert.NotNil(t, tools.kubectl)
}

func TestNewWithKubectl(t *testing.T) {
	fake := kubectl.NewFake("", "", nil)
	tools := NewWithKubectl(fake)
	
	assert.NotNil(t, tools)
	assert.Equal(t, fake, tools.kubectl)
}

func TestToolsIsBlockedResource(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		wantBlocked  bool
	}{
		{
			name:         "secret is blocked",
			resourceType: "secret",
			wantBlocked:  true,
		},
		{
			name:         "secrets is blocked",
			resourceType: "secrets",
			wantBlocked:  true,
		},
		{
			name:         "Secret with capital is blocked",
			resourceType: "Secret",
			wantBlocked:  true,
		},
		{
			name:         "SECRETS all caps is blocked",
			resourceType: "SECRETS",
			wantBlocked:  true,
		},
		{
			name:         "SeCrEt mixed case is blocked",
			resourceType: "SeCrEt",
			wantBlocked:  true,
		},
		{
			name:         "pods is not blocked",
			resourceType: "pods",
			wantBlocked:  false,
		},
		{
			name:         "deployments is not blocked",
			resourceType: "deployments",
			wantBlocked:  false,
		},
		{
			name:         "configmaps is not blocked",
			resourceType: "configmaps",
			wantBlocked:  false,
		},
		{
			name:         "services is not blocked",
			resourceType: "services",
			wantBlocked:  false,
		},
		{
			name:         "empty string is not blocked",
			resourceType: "",
			wantBlocked:  false,
		},
		{
			name:         "secretish is not blocked",
			resourceType: "secretish",
			wantBlocked:  false,
		},
		{
			name:         "mysecrets is not blocked",
			resourceType: "mysecrets",
			wantBlocked:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := New()
			result := tools.isBlockedResource(tt.resourceType)
			assert.Equal(t, tt.wantBlocked, result)
		})
	}
}

func TestToolsRunKubectl(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		kubectlStdout  string
		kubectlStderr  string
		kubectlError   error
		expectedStdout string
		expectedStderr string
		expectedError  error
	}{
		{
			name:           "successful command",
			args:           []string{"get", "pods"},
			kubectlStdout:  "pod-1\npod-2",
			kubectlStderr:  "",
			kubectlError:   nil,
			expectedStdout: "pod-1\npod-2",
			expectedStderr: "",
			expectedError:  nil,
		},
		{
			name:           "command with warnings",
			args:           []string{"get", "pods"},
			kubectlStdout:  "pod-1",
			kubectlStderr:  "Warning: deprecated API",
			kubectlError:   nil,
			expectedStdout: "pod-1",
			expectedStderr: "Warning: deprecated API",
			expectedError:  nil,
		},
		{
			name:           "command with error",
			args:           []string{"get", "pods"},
			kubectlStdout:  "",
			kubectlStderr:  "Error from server",
			kubectlError:   fmt.Errorf("command failed"),
			expectedStdout: "",
			expectedStderr: "Error from server",
			expectedError:  fmt.Errorf("command failed"),
		},
		{
			name:           "empty args",
			args:           []string{},
			kubectlStdout:  "kubectl help output",
			kubectlStderr:  "",
			kubectlError:   nil,
			expectedStdout: "kubectl help output",
			expectedStderr: "",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := kubectl.NewFake(tt.kubectlStdout, tt.kubectlStderr, tt.kubectlError)
			tools := NewWithKubectl(fake)

			stdout, stderr, err := tools.runKubectl(context.Background(), tt.args...)

			assert.Equal(t, tt.expectedStdout, stdout)
			assert.Equal(t, tt.expectedStderr, stderr)
			
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, fake.ExecuteCalled)
			assert.Equal(t, tt.args, fake.ExecuteArgs)
		})
	}
}

func TestToolsFormatOutput(t *testing.T) {
	tests := []struct {
		name           string
		stdout         string
		stderr         string
		err            error
		expectedResult *mcp.CallToolResult
		expectedError  error
	}{
		{
			name:   "successful output with stdout only",
			stdout: "command output",
			stderr: "",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("command output"),
				},
			},
			expectedError: nil,
		},
		{
			name:   "successful output with warnings",
			stdout: "command output",
			stderr: "warning message",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("command output\n\nWarnings:\nwarning message"),
				},
			},
			expectedError: nil,
		},
		{
			name:   "error with stderr",
			stdout: "",
			stderr: "error details",
			err:    fmt.Errorf("command failed"),
			expectedResult: &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.NewTextContent("kubectl command failed: command failed\nkubectl error: error details"),
				},
			},
			expectedError: nil,
		},
		{
			name:   "error without stderr",
			stdout: "",
			stderr: "",
			err:    fmt.Errorf("command failed"),
			expectedResult: &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					mcp.NewTextContent("kubectl command failed: command failed"),
				},
			},
			expectedError: nil,
		},
		{
			name:   "empty stdout and stderr with no error",
			stdout: "",
			stderr: "",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(""),
				},
			},
			expectedError: nil,
		},
		{
			name:   "multiline stdout",
			stdout: "line1\nline2\nline3",
			stderr: "",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("line1\nline2\nline3"),
				},
			},
			expectedError: nil,
		},
		{
			name:   "multiline stderr with stdout",
			stdout: "output",
			stderr: "warning1\nwarning2",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("output\n\nWarnings:\nwarning1\nwarning2"),
				},
			},
			expectedError: nil,
		},
		{
			name:   "json output",
			stdout: `{"key": "value"}`,
			stderr: "",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(`{"key": "value"}`),
				},
			},
			expectedError: nil,
		},
		{
			name:   "yaml output",
			stdout: "key: value\nname: test",
			stderr: "",
			err:    nil,
			expectedResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("key: value\nname: test"),
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := New()
			result, err := tools.formatOutput(tt.stdout, tt.stderr, tt.err)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.expectedResult != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.IsError, result.IsError)
				assert.Equal(t, len(tt.expectedResult.Content), len(result.Content))
				
				for i, expectedContent := range tt.expectedResult.Content {
					if textContent, ok := expectedContent.(mcp.TextContent); ok {
						actualContent, ok := result.Content[i].(mcp.TextContent)
						require.True(t, ok, "expected text content at index %d", i)
						assert.Equal(t, textContent.Text, actualContent.Text)
					}
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestToolsIntegration(t *testing.T) {
	// Test that Tools can be used with different kubectl implementations
	t.Run("with real kubectl", func(t *testing.T) {
		tools := New()
		assert.NotNil(t, tools)
		// Real kubectl.New() would create actual kubectl executor
	})

	t.Run("with fake kubectl", func(t *testing.T) {
		fake := kubectl.NewFake("test output", "", nil)
		tools := NewWithKubectl(fake)
		
		stdout, stderr, err := tools.runKubectl(context.Background(), "test", "command")
		
		assert.NoError(t, err)
		assert.Equal(t, "test output", stdout)
		assert.Empty(t, stderr)
		assert.True(t, fake.ExecuteCalled)
		assert.Equal(t, []string{"test", "command"}, fake.ExecuteArgs)
	})

	t.Run("formatOutput after runKubectl success", func(t *testing.T) {
		fake := kubectl.NewFake("pods list", "", nil)
		tools := NewWithKubectl(fake)
		
		stdout, stderr, err := tools.runKubectl(context.Background(), "get", "pods")
		result, formatErr := tools.formatOutput(stdout, stderr, err)
		
		assert.NoError(t, formatErr)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Len(t, result.Content, 1)
		
		content, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok)
		assert.Equal(t, "pods list", content.Text)
	})

	t.Run("formatOutput after runKubectl error", func(t *testing.T) {
		fake := kubectl.NewFake("", "not found", fmt.Errorf("resource not found"))
		tools := NewWithKubectl(fake)
		
		stdout, stderr, err := tools.runKubectl(context.Background(), "get", "invalid")
		result, formatErr := tools.formatOutput(stdout, stderr, err)
		
		assert.NoError(t, formatErr)
		assert.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Len(t, result.Content, 1)
		
		content, ok := result.Content[0].(mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, content.Text, "kubectl command failed: resource not found")
		assert.Contains(t, content.Text, "kubectl error: not found")
	})
}