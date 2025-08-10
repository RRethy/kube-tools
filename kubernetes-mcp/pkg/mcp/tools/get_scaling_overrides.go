package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (t *Tools) CreateGetScalingOverridesTool() mcp.Tool {
	return mcp.NewTool("get-scaling-overrides",
		mcp.WithDescription("Get scaling overrides for resources (HPA, VPA, etc.)"),
		mcp.WithString("resource_type", mcp.Description("The type of scaling resource (hpa, vpa)")),
		mcp.WithString("namespace", mcp.Description("Namespace to get resources from (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("all_namespaces", mcp.Description("Get resources from all namespaces")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func (t *Tools) HandleGetScalingOverrides(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	// Default to HPA if not specified
	resourceType := "hpa"
	if rt, ok := args["resource_type"].(string); ok && rt != "" {
		resourceType = rt
	}

	cmdArgs := []string{"get", resourceType}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if allNamespaces, ok := args["all_namespaces"].(bool); ok && allNamespaces {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	} else if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}