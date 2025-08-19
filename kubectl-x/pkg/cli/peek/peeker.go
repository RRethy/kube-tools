package peek

import (
	"context"
	"fmt"
	"slices"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	kexec "k8s.io/utils/exec"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"github.com/RRethy/kubectl-x/pkg/resolver"
)

// Peeker handles interactive preview of Kubernetes resources
type Peeker struct {
	IOStreams genericclioptions.IOStreams
	Context   string
	Namespace string
	K8sClient kubernetes.Interface
	Resolver  resolver.Resolver
	Fzf       fzf.Interface
	Exec      kexec.Interface
}

// Peek shows an interactive preview of logs or describe output
func (p *Peeker) Peek(ctx context.Context, action string, target string) error {
	if !slices.Contains([]string{"logs", "describe"}, action) {
		return fmt.Errorf("action must be 'logs' or 'describe', got: \"%s\"", action)
	}

	resourceKind, resourceName := p.Resolver.ResolveTarget(ctx, target)

	switch action {
	case "logs":
		return p.peekLogs(ctx, resourceKind, resourceName)
	case "describe":
		return p.peekDescribe(ctx, resourceKind, resourceName)
	}

	return nil
}

func (p *Peeker) peekLogs(ctx context.Context, resourceKind, resourceName string) error {
	resolved, err := p.Resolver.ResolveResource(ctx, resourceKind, resourceName, p.Namespace, fzf.Config{
		Sorted:  true,
		Preview: fmt.Sprintf("kubectl logs %s/{1} -n %s", resourceKind, p.Namespace),
		Height:  "100%",
	})
	if err != nil {
		return fmt.Errorf("resolving pod for logs: %w", err)
	}

	fmt.Fprintf(p.IOStreams.Out, "kubectl logs %s/%s -n %s\n", resourceKind, resolved.GetName(), p.Namespace)
	return nil
}

func (p *Peeker) peekDescribe(ctx context.Context, resourceKind, resourceName string) error {
	resolved, err := p.Resolver.ResolveResource(ctx, resourceKind, resourceName, p.Namespace, fzf.Config{
		Sorted:  true,
		Preview: fmt.Sprintf("kubectl describe %s/{1} -n %s", resourceKind, p.Namespace),
		Height:  "100%",
	})
	if err != nil {
		return fmt.Errorf("resolving resource for describe: %w", err)
	}

	fmt.Fprintf(p.IOStreams.Out, "kubectl describe %s %s -n %s\n", resourceKind, resolved.GetName(), p.Namespace)
	return nil
}
