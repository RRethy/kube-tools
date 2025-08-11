package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateDebugNamespacePrompt creates a prompt for debugging a specific Kubernetes namespace
func (p *Prompts) CreateDebugNamespacePrompt() mcp.Prompt {
	return mcp.NewPrompt(
		"debug-namespace",
		mcp.WithPromptDescription("Debug a specific Kubernetes namespace"),
		mcp.WithArgument(
			"namespace",
			mcp.ArgumentDescription("The namespace to debug (defaults to the current namespace)"),
		),
		mcp.WithArgument(
			"context",
			mcp.ArgumentDescription("The Kubernetes context to use (defaults to the current context)"),
		),
	)
}

// HandleDebugNamespacePrompt handles the debug namespace prompt request
func (p *Prompts) HandleDebugNamespacePrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	contextName := ""
	namespace := ""

	if req.Params.Arguments != nil {
		if ctxArg, ok := req.Params.Arguments["context"]; ok && ctxArg != "" {
			contextName = ctxArg
		}
		if nsArg, ok := req.Params.Arguments["namespace"]; ok && nsArg != "" {
			namespace = nsArg
		}
	}

	if contextName == "" {
		currentCtx, err := p.getCurrentContext(ctx)
		if err == nil && currentCtx != "" {
			contextName = currentCtx
		}
	}

	if namespace == "" {
		currentNs, err := p.getCurrentNamespace(ctx, contextName)
		if err == nil && currentNs != "" {
			namespace = currentNs
		} else {
			namespace = "default"
		}
	}

	promptText := fmt.Sprintf(`You are a Kubernetes namespace debugging assistant. Your task is to systematically analyze and diagnose issues in the specified namespace.

Context: %s
Namespace: %s

CRITICAL INSTRUCTIONS:
- ONLY use the kubernetes MCP server tools (get, describe, logs, events, top-pod, top-node, etc.) to investigate the namespace
- Focus EXCLUSIVELY on the Kubernetes resources within the namespace specified above
- ALWAYS use filters to limit data retrieval (selectors, limits, field-selector, etc.)
- ALWAYS specify the namespace parameter explicitly - never rely on defaults

IMPORTANT: When you encounter a problem during the analysis, STOP the systematic review and immediately focus on that issue. Think deeply about the root cause, gather relevant logs, events, and configuration details. If it is not an issue, then move on, otherwise present:
1. Clear explanation of the problem
2. Root cause analysis
3. Specific Kubernetes manifest code changes to fix the issue (show actual YAML code that can be modified in the manifest files)
   - ALWAYS use explain tool to verify the correct field structure before suggesting manifest changes
   - Never guess or make up field names - use explain to ensure your suggestions are valid
4. Prevention recommendations
5. STOP analysis.

Gather all results before proceeding with analysis.

## 1. Namespace Overview
- Check if the namespace exists and its status
- List resource quotas and limits if any
- Check namespace labels and annotations

## 2. Pod Analysis
- List all pods with their status, restart count, and age
- Identify pods that are not in Running state
- For problematic pods:
  - Describe the pod to get detailed status
  - Check recent logs (last 50 lines)
  - Look for container restart reasons
  - Check resource requests/limits vs actual usage

## 3. Recent Events
- Get all events in the namespace from the last 2 hours
- Focus on Warning events
- Look for patterns in pod evictions, failed mounts, or image pulls

## 4. Deployments and ReplicaSets
- List all deployments with their desired vs ready replicas
- Check deployment conditions and rollout status
- For deployments not fully ready:
  - Check the deployment events
  - Look at ReplicaSet status
  - Verify image availability
  - Check resource constraints

## 5. Services and Endpoints
- List all services in the namespace
- Verify each service has endpoints
- For services without endpoints:
  - Check selector labels match pod labels
  - Verify pods are ready
  - Look for network policies blocking traffic

## 6. ConfigMaps and Secrets
- List ConfigMaps (but NOT secrets for security reasons)
- Verify referenced ConfigMaps exist
- Check for recent modifications

## 7. Resource Usage
- If metrics are available, check pod CPU and memory usage
- Compare actual usage against requests/limits
- Identify pods at risk of eviction

If no critical issues were found, then state that clearly.

If critical issues were found, then provide a detailed analysis of each issue, including:
- CRITICAL: Keep suggestions short and to the point.
- CRITICAL: When suggesting fixes, ALWAYS prefer to show actual Kubernetes manifest YAML code that the user can modify in their files rather than kubectl commands. For example:
- Instead of "kubectl set resources...", show the actual resource limits/requests YAML snippet
- Instead of "kubectl label...", show how to add labels in the metadata section
- Instead of "kubectl scale...", show how to modify the replicas field in the deployment
- Focus on what needs to be changed in the Kubernetes manifest files themselves
- ALWAYS use explain tool to verify field paths and structure before suggesting any manifest changes
- Format the output with clear sections, use tables where appropriate, and highlight important information. Focus on actionable insights with actual Kubernetes manifest YAML code that can be modified.

Remember: This is a namespace-focused analysis. Do not investigate resources outside the specified namespace.`, contextName, namespace)

	return mcp.NewGetPromptResult(
		fmt.Sprintf("Comprehensive namespace debugging for %s in context %s", namespace, contextName),
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(strings.TrimSpace(promptText)),
			),
		},
	), nil
}