package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateUseContextTool() mcp.Tool {
	return mcp.NewTool("use-context",
		mcp.WithDescription("Switch to a different Kubernetes context"),
		mcp.WithString("context", mcp.Required(), mcp.Description("The context name to switch to")),
		mcp.WithReadOnlyHintAnnotation(false), // This modifies configuration
	)
}

func (t *Tools) HandleUseContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	contextName, ok := args["context"].(string)
	if !ok {
		return nil, fmt.Errorf("context parameter required")
	}

	cmdArgs := []string{"config", "use-context", contextName}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}