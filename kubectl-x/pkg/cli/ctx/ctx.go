package ctx

import (
	"context"
	"os"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/history"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// Ctx switches Kubernetes context based on the provided substring with optional namespace selection
func Ctx(ctx context.Context, configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags, contextSubstring, namespaceSubstring string, exactMatch bool) error {
	kubeConfig, err := kubeconfig.NewKubeConfig(kubeconfig.WithConfigFlags(configFlags))
	if err != nil {
		return err
	}
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	k8sClient := kubernetes.NewClient(configFlags, resourceBuilderFlags)
	fzf := fzf.NewFzf(fzf.WithIOStreams(ioStreams))
	history, err := history.NewHistory(history.NewConfig())
	if err != nil {
		return err
	}
	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)
	return ctxer.Ctx(ctx, contextSubstring, namespaceSubstring, exactMatch)
}
