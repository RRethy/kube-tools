// Package copy provides functionality to copy kubeconfig to $XDG_DATA_HOME
package copy

import (
	"context"
	"os"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// Copy copies the current kubeconfig file to $XDG_DATA_HOME and prints the location
func Copy(ctx context.Context, configFlags *genericclioptions.ConfigFlags) error {
	kubeConfig, err := kubeconfig.NewKubeConfig(kubeconfig.WithConfigFlags(configFlags))
	if err != nil {
		return err
	}

	return (&Copier{
		KubeConfig: kubeConfig,
		IoStreams:  genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}).Copy(ctx)
}
