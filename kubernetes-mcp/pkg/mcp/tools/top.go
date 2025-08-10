package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateTopTool() mcp.Tool {
	return mcp.NewTool("top",
		mcp.WithDescription("Display Resource (CPU/Memory) usage"),
		mcp.WithString("resource_type", mcp.Required(), mcp.Description("The type of resource to show metrics for (nodes, pods)")),
		mcp.WithString("namespace", mcp.Description("Namespace for pods (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("all_namespaces", mcp.Description("Show metrics from all namespaces")),
		mcp.WithString("selector", mcp.Description("Label selector to filter results")),
		mcp.WithBoolean("containers", mcp.Description("Show metrics for containers (pods only)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleTop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	resourceType, ok := args["resource_type"].(string)
	if !ok {
		return nil, fmt.Errorf("resource_type parameter required")
	}

	cmdArgs := []string{"top", resourceType}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if allNamespaces, ok := args["all_namespaces"].(bool); ok && allNamespaces {
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

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}