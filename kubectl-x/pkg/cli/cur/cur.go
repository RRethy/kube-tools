// Package cur provides functionality to display current Kubernetes context and namespace
package cur

import (
	"context"
	"os"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// Cur displays the current Kubernetes context and namespace
func Cur(ctx context.Context, configFlags *genericclioptions.ConfigFlags) error {
	return CurWithPrompt(ctx, configFlags, false)
}

// CurWithPrompt displays the current context and namespace with optional prompt formatting
func CurWithPrompt(ctx context.Context, configFlags *genericclioptions.ConfigFlags, promptFormat bool) error {
	kubeConfig, err := kubeconfig.NewKubeConfig(kubeconfig.WithConfigFlags(configFlags))
	if err != nil {
		return err
	}
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	curer := NewCurer(kubeConfig, ioStreams)
	return curer.CurWithPrompt(ctx, promptFormat)
}
