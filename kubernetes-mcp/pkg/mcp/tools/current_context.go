package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateCurrentContextTool() mcp.Tool {
	return mcp.NewTool("current-context",
		mcp.WithDescription("Display the current Kubernetes context"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleCurrentContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmdArgs := []string{"config", "current-context"}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}