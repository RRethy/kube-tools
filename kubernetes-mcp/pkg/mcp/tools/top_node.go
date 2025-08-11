package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateTopNodeTool creates an MCP tool for displaying node CPU/Memory usage
func (t *Tools) CreateTopNodeTool() mcp.Tool {
	return mcp.NewTool("top-node",
		mcp.WithDescription("Display Resource (CPU/Memory) usage for nodes"),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("selector", mcp.Description("Label selector to filter results")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleTopNode processes requests to display node resource usage
func (t *Tools) HandleTopNode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	}

	cmdArgs := []string{"top", "node"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if selector, ok := args["selector"].(string); ok && selector != "" {
		cmdArgs = append(cmdArgs, "-l", selector)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}