// Package cur provides functionality to display current context and namespace
package cur

import (
	"context"
	"fmt"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// Curer displays current Kubernetes context and namespace information
type Curer struct {
	KubeConfig kubeconfig.Interface
	IoStreams  genericiooptions.IOStreams
}

// NewCurer creates a new current status displayer with the provided dependencies
func NewCurer(kubeConfig kubeconfig.Interface, ioStreams genericiooptions.IOStreams) Curer {
	return Curer{
		KubeConfig: kubeConfig,
		IoStreams:  ioStreams,
	}
}

// Cur displays the current context and namespace in standard format
func (c Curer) Cur(ctx context.Context) error {
	return c.CurWithPrompt(ctx, false)
}

// CurWithPrompt displays the current context and namespace with optional prompt format
func (c Curer) CurWithPrompt(ctx context.Context, promptFormat bool) error {
	currentContext, err := c.KubeConfig.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("getting current context: %w", err)
	}

	currentNamespace, err := c.KubeConfig.GetCurrentNamespace()
	if err != nil {
		return fmt.Errorf("getting current namespace: %w", err)
	}
	if currentNamespace == "" {
		currentNamespace = "default"
	}

	if promptFormat {
		fmt.Fprintf(c.IoStreams.Out, "%s/%s\n", currentContext, currentNamespace)
	} else {
		fmt.Fprintf(c.IoStreams.Out, "--context %s --namespace %s\n", currentContext, currentNamespace)
	}

	return nil
}
