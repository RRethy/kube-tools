package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateLogsTool creates an MCP tool for retrieving pod logs
func (t *Tools) CreateLogsTool() mcp.Tool {
	return mcp.NewTool("logs",
		mcp.WithDescription("Get logs from a pod container"),
		mcp.WithString("pod-name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Description("Namespace of the pod (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("container", mcp.Description("Container name (if multiple containers in pod)")),
		mcp.WithNumber("tail", mcp.Description("Number of lines to show from the end of the logs (default: 100)")),
		mcp.WithString("since", mcp.Description("Show logs since this duration (e.g., 5m, 1h)")),
		mcp.WithBoolean("previous", mcp.Description("Show logs from previous instance of container")),
		mcp.WithBoolean("timestamps", mcp.Description("Include timestamps in log output")),
		mcp.WithBoolean("all-containers", mcp.Description("Get all containers' logs in the pod")),
		mcp.WithNumber("limit-bytes", mcp.Description("Maximum bytes of logs to return (default: no limit)")),
		mcp.WithString("since-time", mcp.Description("Only return logs after a specific date (RFC3339)")),
		mcp.WithBoolean("prefix", mcp.Description("Prefix each log line with the log source (pod name and container name)")),
		mcp.WithString("selector", mcp.Description("Selector (label query) to filter on")),
		mcp.WithNumber("head_limit", mcp.Description("Limit output to first N lines after kubectl returns (default: 50, 0 for all)")),
		mcp.WithNumber("head_offset", mcp.Description("Skip first N lines before applying head_limit")),
		mcp.WithNumber("tail_limit", mcp.Description("Limit output to last N lines (applied client-side, use 'tail' for kubectl-level limiting)")),
		mcp.WithNumber("tail_offset", mcp.Description("Skip last N lines before applying tail_limit")),
		mcp.WithString("grep", mcp.Description("Filter log lines matching this pattern (regex or literal string)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleLogs processes requests to retrieve pod container logs
func (t *Tools) HandleLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	var cmdArgs []string
	
	if selector, ok := args["selector"].(string); ok && selector != "" {
		cmdArgs = []string{"logs", "-l", selector}
	} else {
		podName, ok := args["pod-name"].(string)
		if !ok || podName == "" {
			return nil, fmt.Errorf("pod-name parameter required")
		}
		cmdArgs = []string{"logs", podName}
	}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	if container, ok := args["container"].(string); ok && container != "" {
		cmdArgs = append(cmdArgs, "-c", container)
	}

	tail := 100.0
	if tailVal, ok := args["tail"].(float64); ok {
		tail = tailVal
	}
	if tail > 0 {
		cmdArgs = append(cmdArgs, "--tail", fmt.Sprintf("%d", int(tail)))
	}

	if since, ok := args["since"].(string); ok && since != "" {
		cmdArgs = append(cmdArgs, "--since", since)
	}

	if previous, ok := args["previous"].(bool); ok && previous {
		cmdArgs = append(cmdArgs, "--previous")
	}

	if timestamps, ok := args["timestamps"].(bool); ok && timestamps {
		cmdArgs = append(cmdArgs, "--timestamps")
	}

	if allContainers, ok := args["all-containers"].(bool); ok && allContainers {
		cmdArgs = append(cmdArgs, "--all-containers")
	}

	if limitBytes, ok := args["limit-bytes"].(float64); ok && limitBytes > 0 {
		cmdArgs = append(cmdArgs, "--limit-bytes", fmt.Sprintf("%d", int(limitBytes)))
	}

	if sinceTime, ok := args["since-time"].(string); ok && sinceTime != "" {
		cmdArgs = append(cmdArgs, "--since-time", sinceTime)
	}

	if prefix, ok := args["prefix"].(bool); ok && prefix {
		cmdArgs = append(cmdArgs, "--prefix")
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	
	if err == nil && stdout != "" {
		filterParams := GetFilterParams(args)
		filteredOutput, filterErr := ApplyFilter(stdout, filterParams, "")
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