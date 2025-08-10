package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateDescribeTool creates an MCP tool for describing Kubernetes resources
func (t *Tools) CreateDescribeTool() mcp.Tool {
	return mcp.NewTool("describe",
		mcp.WithDescription("Describe Kubernetes resources to get detailed information including events"),
		mcp.WithString("resource-type", mcp.Required(), mcp.Description("The type of Kubernetes resource to describe (e.g., pods, deployments, services)")),
		mcp.WithString("resource-name", mcp.Required(), mcp.Description("The name of the resource to describe")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleDescribe processes describe requests for detailed resource information
func (t *Tools) HandleDescribe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	resourceType, ok := args["resource-type"].(string)
	if !ok {
		return nil, fmt.Errorf("resource_type parameter required")
	}

	if t.isBlockedResource(resourceType) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Access to resource type '%s' is blocked for security reasons", resourceType)),
			},
		}, nil
	}

	resourceName, ok := args["resource-name"].(string)
	if !ok {
		return nil, fmt.Errorf("resource_name parameter required")
	}

	cmdArgs := []string{"describe", resourceType, resourceName}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}