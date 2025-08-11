package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateCurrentNamespaceTool() mcp.Tool {
	return mcp.NewTool("current-namespace",
		mcp.WithDescription("Display the current namespace for the current context"),
		mcp.WithString("context", mcp.Description("Kubernetes context to check namespace for (default: current context)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleCurrentNamespace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	}

	cmdArgs := []string{"config", "view", "--minify", "-o", "jsonpath={..namespace}"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	
	// If namespace is empty, it means "default" namespace
	if stdout == "" && err == nil {
		stdout = "default"
	}
	
	return t.formatOutput(stdout, stderr, err)
}