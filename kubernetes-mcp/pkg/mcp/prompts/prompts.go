// Package prompts provides MCP prompt handlers for Kubernetes debugging workflows
package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/kubectl"
)

// Prompts holds the dependencies for prompt handlers
type Prompts struct {
	kubectl kubectl.Kubectl
}

// New creates a new Prompts instance with the default kubectl implementation
func New() *Prompts {
	return &Prompts{
		kubectl: kubectl.New(),
	}
}

// NewWithKubectl creates a new Prompts instance with a custom kubectl implementation
func NewWithKubectl(k kubectl.Kubectl) *Prompts {
	return &Prompts{
		kubectl: k,
	}
}

// getCurrentContext gets the current kubectl context
func (p *Prompts) getCurrentContext(ctx context.Context) (string, error) {
	stdout, _, err := p.kubectl.Execute(ctx, "--no-headers", "config", "current-context")
	if err != nil {
		return "", fmt.Errorf("failed to get current context: %w", err)
	}
	return strings.TrimSpace(stdout), nil
}

// getCurrentNamespace gets the current namespace for a given context
func (p *Prompts) getCurrentNamespace(ctx context.Context, contextName string) (string, error) {
	args := []string{"config", "view", "-o", "jsonpath={.contexts[?(@.name==\"" + contextName + "\")].context.namespace}"}
	if contextName != "" {
		args = append([]string{"--context", contextName}, args...)
	}
	
	stdout, _, err := p.kubectl.Execute(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get current namespace: %w", err)
	}
	
	ns := strings.TrimSpace(stdout)
	if ns == "" {
		return "default", nil
	}
	return ns, nil
}