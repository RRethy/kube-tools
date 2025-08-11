// Package prompts provides MCP prompt handlers for Kubernetes debugging workflows
package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateDebugClusterPrompt creates a prompt for debugging a Kubernetes cluster
func (p *Prompts) CreateDebugClusterPrompt() mcp.Prompt {
	return mcp.NewPrompt(
		"debug-cluster",
		mcp.WithPromptDescription("Debug a Kubernetes cluster"),
		mcp.WithArgument(
			"context",
			mcp.ArgumentDescription("The Kubernetes context/cluster to debug (defaults to the current context)"),
		),
	)
}

// HandleDebugClusterPrompt handles the debug cluster prompt request
func (p *Prompts) HandleDebugClusterPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	contextName := ""
	if req.Params.Arguments != nil {
		if ctxArg, ok := req.Params.Arguments["context"]; ok && ctxArg != "" {
			contextName = ctxArg
		}
	}

	if contextName == "" {
		currentCtx, err := p.getCurrentContext(ctx)
		if err == nil && currentCtx != "" {
			contextName = currentCtx
		}
	}

	promptText := fmt.Sprintf(`You are a Kubernetes cluster debugging assistant. Your task is to systematically analyze and diagnose issues across the entire cluster.

Context/Cluster: %s

CRITICAL INSTRUCTIONS:
- FIRST check what MCP servers are available - there may be other MCP servers beyond kubernetes that could help with debugging (e.g., gcloud, buildkite, monitoring tools, etc.)
- Use the kubernetes MCP server tools (get, describe, logs, events, top-pod, top-node, etc.) to investigate the cluster
- If other relevant MCP servers are available, use them to gather additional context (cloud provider info, CI/CD status, monitoring metrics, etc.)
- Focus EXCLUSIVELY on the Kubernetes cluster resources and related infrastructure
- ALWAYS use filters to limit data retrieval (selectors, limits, field-selector, etc.)
- ALWAYS specify the namespace parameter explicitly - never rely on defaults (use all-namespaces: true or specific namespace names)

IMPORTANT: When you encounter a problem during the analysis, STOP the systematic review and immediately focus on that issue. Think deeply about the root cause, gather relevant logs, events, and configuration details. If it is not an issue, then move on, otherwise present:
1. Clear explanation of the problem
2. Root cause analysis
3. Specific Kubernetes manifest code changes to fix the issue (show actual YAML code that can be modified in the manifest files)
   - ALWAYS use explain tool to verify the correct field structure before suggesting manifest changes
   - Never guess or make up field names - use explain to ensure your suggestions are valid
4. Prevention recommendations
5. STOP analysis.

Gather all results before proceeding with analysis.

## 0. Check Available MCP Servers
- List all available MCP servers and their capabilities
- Identify which servers might provide useful debugging information:
  - Cloud provider MCP servers (AWS, GCP, Azure) for infrastructure details
  - Monitoring/observability MCP servers for metrics and traces
  - CI/CD MCP servers for deployment history
  - Git MCP servers for recent configuration changes

## 1. Cluster Overview
- Check cluster version and basic info
- Verify API server health and availability
- List available API resources
- Check cluster-wide RBAC permissions

## 2. Node Analysis
- List all nodes with their status, roles, and versions
- Identify nodes that are not Ready
- For problematic nodes:
  - Describe the node to get detailed conditions
  - Check for memory/disk pressure
  - Look for network issues
  - Verify kubelet status
- Check node resource allocation and capacity

## 3. System Namespaces Health
- Check kube-system namespace for critical components:
  - API server, controller manager, scheduler
  - Core DNS/kube-dns
  - kube-proxy
  - CNI components
- Verify all system pods are running and healthy
- Check for recent restarts or crashes

## 4. Cluster-wide Events
- Get Warning events from the last 2 hours
- Identify patterns or recurring issues
- Highlight critical system events

## 5. Resource Pressure & Quotas
- Check overall cluster resource usage (if metrics available)
- Identify nodes under resource pressure
- List resource quotas across namespaces
- Check for pods in Pending state due to resource constraints

## 6. Network & Storage
- Verify core networking components are healthy
- Check for persistent volume issues
- Identify any storage class problems
- Look for network policy conflicts

If no critical issues were found, then state that clearly.

If critical issues were found, then provide a detailed analysis of each issue, including:
- CRITICAL: Keep suggestions short and to the point.
- CRITICAL: When suggesting fixes, ALWAYS prefer to show actual Kubernetes manifest YAML code that the user can modify in their files rather than kubectl commands. For example:
- Instead of "kubectl set resources...", show the actual resource limits/requests YAML snippet
- Instead of "kubectl label...", show how to add labels in the metadata section
- Instead of "kubectl scale...", show how to modify the replicas field in the deployment
- Instead of "kubectl taint...", show how to configure taints/tolerations in the manifest
- Focus on what needs to be changed in the Kubernetes manifest files themselves
- ALWAYS use explain tool to verify field paths and structure before suggesting any manifest changes
- Format the output with clear sections, use tables where appropriate, and highlight important information. Focus on actionable insights with actual Kubernetes manifest YAML code that can be modified.

Remember: This is a cluster-focused analysis. Always use filters, selectors, and limits to retrieve only the necessary data.`, contextName)

	return mcp.NewGetPromptResult(
		fmt.Sprintf("Comprehensive cluster debugging for context %s", contextName),
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(strings.TrimSpace(promptText)),
			),
		},
	), nil
}