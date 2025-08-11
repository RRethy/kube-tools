package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateTopPodTool creates an MCP tool for displaying pod CPU/Memory usage
func (t *Tools) CreateTopPodTool() mcp.Tool {
	return mcp.NewTool("top-pod",
		mcp.WithDescription("Display Resource (CPU/Memory) usage for pods"),
		mcp.WithString("namespace", mcp.Description("Namespace for pods (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("all-namespaces", mcp.Description("Show metrics from all namespaces")),
		mcp.WithString("selector", mcp.Description("Label selector to filter results")),
		mcp.WithBoolean("containers", mcp.Description("Show metrics for containers")),
		mcp.WithString("sort-by", mcp.Description("Sort by 'cpu' or 'memory'")),
		mcp.WithString("field-selector", mcp.Description("Selector (field query) to filter on")),
		mcp.WithBoolean("no-headers", mcp.Description("Don't print headers")),
		mcp.WithBoolean("sum", mcp.Description("Show sum of resource usage")),
		mcp.WithNumber("head_limit", mcp.Description("Limit output to first N lines (default: 50, 0 for all)")),
		mcp.WithNumber("head_offset", mcp.Description("Skip first N lines before applying head_limit")),
		mcp.WithNumber("tail_limit", mcp.Description("Limit output to last N lines")),
		mcp.WithNumber("tail_offset", mcp.Description("Skip last N lines before applying tail_limit")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleTopPod processes requests to display pod resource usage
func (t *Tools) HandleTopPod(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	}

	cmdArgs := []string{"top", "pod"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if allNamespaces, ok := args["all-namespaces"].(bool); ok && allNamespaces {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	} else if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	if selector, ok := args["selector"].(string); ok && selector != "" {
		cmdArgs = append(cmdArgs, "-l", selector)
	}

	if containers, ok := args["containers"].(bool); ok && containers {
		cmdArgs = append(cmdArgs, "--containers")
	}

	if sortBy, ok := args["sort-by"].(string); ok && sortBy != "" {
		cmdArgs = append(cmdArgs, "--sort-by", sortBy)
	}

	if fieldSelector, ok := args["field-selector"].(string); ok && fieldSelector != "" {
		cmdArgs = append(cmdArgs, "--field-selector", fieldSelector)
	}

	if noHeaders, ok := args["no-headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
	}

	if sum, ok := args["sum"].(bool); ok && sum {
		cmdArgs = append(cmdArgs, "--sum")
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	
	// Apply pagination to stdout if successful
	if err == nil && stdout != "" {
		paginationParams := GetPaginationParams(args)
		stdout = ApplyPagination(stdout, paginationParams)
	}
	
	return t.formatOutput(stdout, stderr, err)
}