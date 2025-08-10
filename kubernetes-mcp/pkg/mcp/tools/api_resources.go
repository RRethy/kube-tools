package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateAPIResourcesTool() mcp.Tool {
	return mcp.NewTool("api-resources",
		mcp.WithDescription("Get supported API resources on the server"),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("namespaced", mcp.Description("Show only namespaced resources")),
		mcp.WithBoolean("no_headers", mcp.Description("Don't show headers")),
		mcp.WithString("output", mcp.Description("Output format: 'wide' or default")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleAPIResources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	cmdArgs := []string{"api-resources"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if namespaced, ok := args["namespaced"].(bool); ok && namespaced {
		cmdArgs = append(cmdArgs, "--namespaced=true")
	}

	if noHeaders, ok := args["no_headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
	}

	if output, ok := args["output"].(string); ok && output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}