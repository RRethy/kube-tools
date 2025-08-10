package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateConfigGetContextsTool() mcp.Tool {
	return mcp.NewTool("config-get-contexts",
		mcp.WithDescription("Display all available contexts from kubeconfig"),
		mcp.WithBoolean("no-headers", mcp.Description("Don't show headers")),
		mcp.WithString("output", mcp.Description("Output format: 'name' for context names only, or default")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleConfigGetContexts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	cmdArgs := []string{"config", "get-contexts"}

	if noHeaders, ok := args["no-headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
	}

	if output, ok := args["output"].(string); ok && output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}