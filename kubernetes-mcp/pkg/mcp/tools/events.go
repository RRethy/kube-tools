package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateEventsTool creates an MCP tool for retrieving cluster events
func (t *Tools) CreateEventsTool() mcp.Tool {
	return mcp.NewTool("events",
		mcp.WithDescription("Get events from Kubernetes with filtering options"),
		mcp.WithString("namespace", mcp.Description("Namespace to get events from (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("for", mcp.Description("Filter events for a specific resource (e.g., pod/my-pod, deployment/my-deployment)")),
		mcp.WithBoolean("all-namespaces", mcp.Description("Get events from all namespaces (equivalent to kubectl get events --all-namespaces or -A)")),
		mcp.WithString("types", mcp.Description("Comma-separated list of event types to filter (Normal, Warning)")),
		mcp.WithString("output", mcp.Description("Output format: 'json', 'yaml', 'wide', or default table format")),
		mcp.WithBoolean("no-headers", mcp.Description("Don't print headers in table output")),
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

// HandleEvents processes requests to retrieve Kubernetes events
func (t *Tools) HandleEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if req.Params.Arguments != nil {
		var ok bool
		args, ok = req.Params.Arguments.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments")
		}
	}

	cmdArgs := []string{"get", "events", "--sort-by=.lastTimestamp"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if allNamespaces, ok := args["all-namespaces"].(bool); ok && allNamespaces {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	} else if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	if forResource, ok := args["for"].(string); ok && forResource != "" {
		cmdArgs = append(cmdArgs, "--field-selector", fmt.Sprintf("involvedObject.name=%s", forResource))
	}

	if types, ok := args["types"].(string); ok && types != "" {
		// kubectl doesn't have a direct --types flag, we use field-selector
		cmdArgs = append(cmdArgs, "--field-selector", fmt.Sprintf("type=%s", types))
	}

	outputFormat := ""
	if output, ok := args["output"].(string); ok && output != "" {
		outputFormat = output
		cmdArgs = append(cmdArgs, "-o", output)
	}

	if noHeaders, ok := args["no-headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
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
