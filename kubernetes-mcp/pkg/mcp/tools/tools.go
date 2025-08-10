// Package tools implements kubectl command wrappers for MCP tool handlers
package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tools provides kubectl command execution for MCP tool handlers
type Tools struct{}

// New creates a new Tools instance
func New() *Tools {
	return &Tools{}
}

func (t *Tools) isBlockedResource(resourceType string) bool {
	blocked := []string{"secret", "secrets"}
	lower := strings.ToLower(resourceType)
	return slices.Contains(blocked, lower)
}

func (t *Tools) runKubectl(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
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