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
		mcp.WithString("sort-by", mcp.Description("Sort by 'cpu' or 'memory'")),
		mcp.WithBoolean("no-headers", mcp.Description("Don't print headers")),
		mcp.WithBoolean("show-capacity", mcp.Description("Show capacity values instead of allocatable")),
		mcp.WithNumber("head_limit", mcp.Description("Limit output to first N lines (default: 50, 0 for all)")),
		mcp.WithNumber("head_offset", mcp.Description("Skip first N lines before applying head_limit")),
		mcp.WithNumber("tail_limit", mcp.Description("Limit output to last N lines")),
		mcp.WithNumber("tail_offset", mcp.Description("Skip last N lines before applying tail_limit")),
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

	if sortBy, ok := args["sort-by"].(string); ok && sortBy != "" {
		cmdArgs = append(cmdArgs, "--sort-by", sortBy)
	}

	if noHeaders, ok := args["no-headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
	}

	if showCapacity, ok := args["show-capacity"].(bool); ok && showCapacity {
		cmdArgs = append(cmdArgs, "--show-capacity")
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	
	// Apply pagination to stdout if successful
	if err == nil && stdout != "" {
		paginationParams := GetPaginationParams(args)
		stdout = ApplyPagination(stdout, paginationParams)
	}
	
	return t.formatOutput(stdout, stderr, err)
}