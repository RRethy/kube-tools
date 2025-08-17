package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateAuthCanITool creates an MCP tool for checking RBAC permissions
func (t *Tools) CreateAuthCanITool() mcp.Tool {
	return mcp.NewTool("auth-can-i",
		mcp.WithDescription("Check whether an action is allowed"),
		mcp.WithString("verb", mcp.Required(), mcp.Description("The verb to check (get, list, create, update, patch, delete, watch)")),
		mcp.WithString("resource", mcp.Required(), mcp.Description("The resource to check (e.g., pods, services, deployments)")),
		mcp.WithString("resource-name", mcp.Description("Specific resource name to check")),
		mcp.WithString("namespace", mcp.Description("Namespace to check in (default: current namespace)")),
		mcp.WithString("context", mcp.Description("Kubernetes context to use (default: current context)")),
		mcp.WithBoolean("all-namespaces", mcp.Description("Check for all namespaces")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// HandleAuthCanI processes authorization check requests
func (t *Tools) HandleAuthCanI(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	verb, ok := args["verb"].(string)
	if !ok {
		return nil, fmt.Errorf("verb parameter required")
	}

	resource, ok := args["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("resource parameter required")
	}

	cmdArgs := []string{"auth", "can-i", verb, resource}

	if contextName, ok := args["context"].(string); ok && contextName != "" {
		cmdArgs = append([]string{"--context", contextName}, cmdArgs...)
	}

	if resourceName, ok := args["resource-name"].(string); ok && resourceName != "" {
		cmdArgs = append(cmdArgs, resourceName)
	}

	if allNamespaces, ok := args["all-namespaces"].(bool); ok && allNamespaces {
		cmdArgs = append(cmdArgs, "--all-namespaces")
	} else if namespace, ok := args["namespace"].(string); ok && namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	stdout, stderr, err := t.runKubectl(ctx, cmdArgs...)
	return t.formatOutput(stdout, stderr, err)
}
