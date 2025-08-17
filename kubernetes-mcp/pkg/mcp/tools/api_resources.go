package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateAPIResourcesTool creates an MCP tool for listing API resources
func (t *Tools) CreateAPIResourcesTool() mcp.Tool {
	return mcp.NewTool("api-resources",
		mcp.WithDescription("Get supported API resources on the server"),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("namespaced", mcp.Description("Filter by namespaced resources (true: only namespaced, false: only non-namespaced, unset: all)")),
		mcp.WithString("api-group", mcp.Description("Limit to resources in the specified API group")),
		mcp.WithString("sort-by", mcp.Description("Sort by field: 'name' or 'kind'")),
		mcp.WithBoolean("no-headers", mcp.Description("Don't show headers")),
		mcp.WithString("output", mcp.Description("Output format: 'wide' or 'name'")),
		mcp.WithNumber("head_limit", mcp.Description("Limit output to first N lines (default: 50, 0 for all)")),
		mcp.WithNumber("head_offset", mcp.Description("Skip first N lines before applying head_limit")),
		mcp.WithNumber("tail_limit", mcp.Description("Limit output to last N lines")),
		mcp.WithNumber("tail_offset", mcp.Description("Skip last N lines before applying tail_limit")),
		mcp.WithString("grep", mcp.Description("Filter output lines matching this pattern (regex or literal string)")),
		mcp.WithString("jq", mcp.Description("Apply jq filter to JSON output (requires output format to be 'json')")),
		mcp.WithString("yq", mcp.Description("Apply yq filter to YAML output (requires output format to be 'yaml')")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleAPIResources processes requests to list available API resources
func (t *Tools) HandleAPIResources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	}

	cmdArgs := []string{"api-resources"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if namespaced, ok := args["namespaced"].(bool); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--namespaced=%t", namespaced))
	}

	if apiGroup, ok := args["api-group"].(string); ok && apiGroup != "" {
		cmdArgs = append(cmdArgs, "--api-group", apiGroup)
	}

	if sortBy, ok := args["sort-by"].(string); ok && sortBy != "" {
		cmdArgs = append(cmdArgs, "--sort-by", sortBy)
	}

	if noHeaders, ok := args["no-headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
	}

	outputFormat := ""
	if output, ok := args["output"].(string); ok && output != "" {
		outputFormat = output
		cmdArgs = append(cmdArgs, "-o", output)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)

	if err == nil && stdout != "" {
		filterParams := GetFilterParams(args)
		filteredOutput, filterErr := ApplyFilter(stdout, filterParams, outputFormat)
		if filterErr != nil {
			stderr = fmt.Sprintf("Filter error: %v\nOriginal output preserved", filterErr)
		} else {
			stdout = filteredOutput
		}

		paginationParams := GetPaginationParams(args)
		result := ApplyPagination(stdout, paginationParams)
		stdout = result.Output
		if result.PaginationInfo != "" {
			stdout += result.PaginationInfo
		}
	}

	return t.formatOutput(stdout, stderr, err)
}
