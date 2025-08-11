package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateGetTool creates an MCP tool for getting Kubernetes resources
func (t *Tools) CreateGetTool() mcp.Tool {
	return mcp.NewTool("get",
		mcp.WithDescription("Get Kubernetes resources from the current context/namespace using kubectl"),
		mcp.WithString("resource-type", mcp.Required(), mcp.Description("The type of Kubernetes resource to get (e.g., pods, deployments, services)")),
		mcp.WithString("resource-name", mcp.Description("Optional specific resource name to get. If not provided, lists all resources of the given type")),
		mcp.WithString("namespace", mcp.Description("Namespace to get resources from (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithString("selector", mcp.Description("Label selector to filter results (e.g., 'app=nginx')")),
		mcp.WithString("output", mcp.Description("Output format: 'json', 'yaml', 'wide', or default table format")),
		mcp.WithBoolean("all-namespaces", mcp.Description("Get resources from all namespaces (equivalent to kubectl get --all-namespaces or -A)")),
		mcp.WithString("field-selector", mcp.Description("Selector (field query) to filter on (e.g., status.phase=Running)")),
		mcp.WithBoolean("show-labels", mcp.Description("Show all labels as the last column")),
		mcp.WithBoolean("no-headers", mcp.Description("Don't print headers in table output")),
		mcp.WithString("sort-by", mcp.Description("Sort using JSONPath expression (e.g., '{.metadata.name}')")),
		mcp.WithBoolean("show-kind", mcp.Description("Show resource type for the requested object(s)")),
		mcp.WithString("label-columns", mcp.Description("Comma-separated list of labels to display as columns")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleGet processes get requests for Kubernetes resources
func (t *Tools) HandleGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	resourceType, ok := args["resource-type"].(string)
	if !ok {
		return nil, fmt.Errorf("resource-type parameter required")
	}

	if t.isBlockedResource(resourceType) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Access to resource type '%s' is blocked for security reasons", resourceType)),
			},
		}, nil
	}

	cmdArgs := []string{"get", resourceType}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if allNamespaces, ok := args["all-namespaces"].(bool); ok && allNamespaces {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	} else if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	if resourceName, ok := args["resource-name"].(string); ok && resourceName != "" {
		cmdArgs = append(cmdArgs, resourceName)
	}

	if selector, ok := args["selector"].(string); ok && selector != "" {
		cmdArgs = append(cmdArgs, "-l", selector)
	}

	if output, ok := args["output"].(string); ok && output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}

	if fieldSelector, ok := args["field-selector"].(string); ok && fieldSelector != "" {
		cmdArgs = append(cmdArgs, "--field-selector", fieldSelector)
	}

	if showLabels, ok := args["show-labels"].(bool); ok && showLabels {
		cmdArgs = append(cmdArgs, "--show-labels")
	}

	if noHeaders, ok := args["no-headers"].(bool); ok && noHeaders {
		cmdArgs = append(cmdArgs, "--no-headers")
	}

	if sortBy, ok := args["sort-by"].(string); ok && sortBy != "" {
		cmdArgs = append(cmdArgs, "--sort-by", sortBy)
	}

	if showKind, ok := args["show-kind"].(bool); ok && showKind {
		cmdArgs = append(cmdArgs, "--show-kind")
	}

	if labelColumns, ok := args["label-columns"].(string); ok && labelColumns != "" {
		cmdArgs = append(cmdArgs, "--label-columns", labelColumns)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}