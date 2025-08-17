package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateAPIVersionsTool creates an MCP tool for listing API versions
func (t *Tools) CreateAPIVersionsTool() mcp.Tool {
	return mcp.NewTool("api-versions",
		mcp.WithDescription("Print the supported API versions on the server"),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleAPIVersions processes requests to list API versions
func (t *Tools) HandleAPIVersions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	}

	cmdArgs := []string{"api-versions"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}
