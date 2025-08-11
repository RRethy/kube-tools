package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateExplainTool() mcp.Tool {
	return mcp.NewTool("explain",
		mcp.WithDescription("Explain Kubernetes resource types and their fields"),
		mcp.WithString("resource", mcp.Required(), mcp.Description("Resource type to explain (e.g., pods, deployments, pods.spec.containers)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("recursive", mcp.Description("Print all fields recursively")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleExplain(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	} else {
		return nil, fmt.Errorf("invalid arguments")
	}

	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter required")
	}

	cmdArgs := []string{"explain", resource}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if recursive, ok := args["recursive"].(bool); ok && recursive {
		cmdArgs = append(cmdArgs, "--recursive")
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}