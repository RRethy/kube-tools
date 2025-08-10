package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateEventsTool() mcp.Tool {
	return mcp.NewTool("events",
		mcp.WithDescription("Get events from Kubernetes with filtering options"),
		mcp.WithString("namespace", mcp.Description("Namespace to get events from (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("for", mcp.Description("Filter events for a specific resource (e.g., pod/my-pod, deployment/my-deployment)")),
		mcp.WithBoolean("all_namespaces", mcp.Description("Get events from all namespaces (equivalent to kubectl get events --all-namespaces or -A)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	cmdArgs := []string{"get", "events", "--sort-by=.lastTimestamp"}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if allNamespaces, ok := args["all_namespaces"].(bool); ok && allNamespaces {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	} else if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	if forResource, ok := args["for"].(string); ok && forResource != "" {
		cmdArgs = append(cmdArgs, "--field-selector", fmt.Sprintf("involvedObject.name=%s", forResource))
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}