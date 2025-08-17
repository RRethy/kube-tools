// Package cur provides functionality to display current context and namespace
package cur

import (
	"context"
	"fmt"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/fatih/color"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"
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
	klog.V(4).Infof("Retrieving current context and namespace, promptFormat=%t", promptFormat)

	currentContext, err := c.KubeConfig.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("getting current context: %w", err)
	}
	klog.V(5).Infof("Retrieved current context: %s", currentContext)

	currentNamespace, err := c.KubeConfig.GetCurrentNamespace()
	if err != nil {
		return fmt.Errorf("getting current namespace: %w", err)
	}
	if currentNamespace == "" {
		currentNamespace = "default"
	}
	klog.V(5).Infof("Retrieved current namespace: %s", currentNamespace)

	contextColor := color.New(color.FgBlue)
	namespaceColor := color.New(color.FgGreen)
	flagColor := color.New(color.FgWhite)

	if promptFormat {
		klog.V(6).Info("Displaying context/namespace in prompt format")
		contextColor.Fprint(c.IoStreams.Out, currentContext)
		fmt.Fprint(c.IoStreams.Out, "/")
		namespaceColor.Fprintf(c.IoStreams.Out, "%s\n", currentNamespace)
	} else {
		klog.V(6).Info("Displaying context/namespace in flag format")
		flagColor.Fprint(c.IoStreams.Out, "--context ")
		contextColor.Fprint(c.IoStreams.Out, currentContext)
		flagColor.Fprint(c.IoStreams.Out, " --namespace ")
		namespaceColor.Fprintf(c.IoStreams.Out, "%s\n", currentNamespace)
	}

	klog.V(2).Infof("Successfully displayed current status: context=%s namespace=%s", currentContext, currentNamespace)
	return nil
}
