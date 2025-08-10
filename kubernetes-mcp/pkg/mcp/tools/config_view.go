package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateConfigViewTool() mcp.Tool {
	return mcp.NewTool("config-view",
		mcp.WithDescription("Display merged kubeconfig settings"),
		mcp.WithBoolean("minify", mcp.Description("Remove all information not used by current context")),
		mcp.WithBoolean("raw", mcp.Description("Display raw byte data")),
		mcp.WithBoolean("flatten", mcp.Description("Flatten the resulting kubeconfig file into self-contained output")),
		mcp.WithString("output", mcp.Description("Output format: 'json', 'yaml', or default")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleConfigView(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	cmdArgs := []string{"config", "view"}

	if minify, ok := args["minify"].(bool); ok && minify {
		cmdArgs = append(cmdArgs, "--minify")
	}

	if raw, ok := args["raw"].(bool); ok && raw {
		cmdArgs = append(cmdArgs, "--raw")
	}

	if flatten, ok := args["flatten"].(bool); ok && flatten {
		cmdArgs = append(cmdArgs, "--flatten")
	}

	if output, ok := args["output"].(string); ok && output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}