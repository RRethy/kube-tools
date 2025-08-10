package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateLogsTool() mcp.Tool {
	return mcp.NewTool("logs",
		mcp.WithDescription("Get logs from a pod container"),
		mcp.WithString("pod_name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Description("Namespace of the pod (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("container", mcp.Description("Container name (if multiple containers in pod)")),
		mcp.WithNumber("tail", mcp.Description("Number of lines to show from the end of the logs (default: 100)")),
		mcp.WithString("since", mcp.Description("Show logs since this duration (e.g., 5m, 1h)")),
		mcp.WithBoolean("previous", mcp.Description("Show logs from previous instance of container")),
		mcp.WithBoolean("timestamps", mcp.Description("Include timestamps in log output")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	podName, ok := args["pod_name"].(string)
	if !ok {
		return nil, fmt.Errorf("pod_name parameter required")
	}

	cmdArgs := []string{"logs", podName}

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

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}