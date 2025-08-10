package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateVersionTool() mcp.Tool {
	return mcp.NewTool("version",
		mcp.WithDescription("Get version information for kubectl client and Kubernetes cluster"),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("output", mcp.Description("Output format: 'json', 'yaml', or default")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleVersion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	cmdArgs := []string{"version"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if output, ok := args["output"].(string); ok && output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}