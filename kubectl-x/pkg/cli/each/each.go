package each

import (
	"context"
	"fmt"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	xcontext "github.com/RRethy/kubectl-x/pkg/context"
	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/namespace"
)

// Each executes kubectl commands across multiple contexts matching the given pattern
func Each(ctx context.Context, configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags, contextPattern string, outputFormat string, interactive bool, commandArgs []string) error {
	kubeConfig, err := kubeconfig.NewKubeConfig()
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}

	namespace := namespace.NewResolver(kubeConfig, configFlags).Resolve(xcontext.NewResolver(kubeConfig, configFlags).Resolve())

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	fzf := fzf.NewFzf(fzf.WithIOStreams(ioStreams))

	return (&Eacher{
		IOStreams:  ioStreams,
		Kubeconfig: kubeConfig,
		Namespace:  namespace,
		Fzf:        fzf,
	}).Each(ctx, contextPattern, outputFormat, interactive, commandArgs)
}

