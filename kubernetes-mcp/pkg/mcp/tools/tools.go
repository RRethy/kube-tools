// Package tools implements kubectl command wrappers for MCP tool handlers
package tools

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/kubectl"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tools provides kubectl command execution for MCP tool handlers
type Tools struct {
	kubectl kubectl.Kubectl
}

// New creates a new Tools instance
func New() *Tools {
	return &Tools{
		kubectl: kubectl.New(),
	}
}

// NewWithKubectl creates a new Tools instance with a custom kubectl implementation
func NewWithKubectl(k kubectl.Kubectl) *Tools {
	return &Tools{
		kubectl: k,
	}
}

func (t *Tools) isBlockedResource(resourceType string) bool {
	blocked := []string{"secret", "secrets"}
	lower := strings.ToLower(resourceType)
	return slices.Contains(blocked, lower)
}

func (t *Tools) runKubectl(ctx context.Context, args ...string) (string, string, error) {
	return t.kubectl.Execute(ctx, args...)
}

func (t *Tools) formatOutput(stdout, stderr string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		errorMsg := fmt.Sprintf("kubectl command failed: %v", err)
		if stderr != "" {
			errorMsg = fmt.Sprintf("%s\nkubectl error: %s", errorMsg, stderr)
		}

		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent(errorMsg),
			},
		}, nil
	}

	output := stdout
	if stderr != "" {
		output = fmt.Sprintf("%s\n\nWarnings:\n%s", stdout, stderr)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(output),
		},
	}, nil
}