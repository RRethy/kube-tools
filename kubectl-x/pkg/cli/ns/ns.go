package ns

import (
	"context"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/history"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
)

func Ns(ctx context.Context, configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags, namespace string, exactMatch bool) error {
	kubeConfig, err := kubeconfig.NewKubeConfig()
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
	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)
	return nser.Ns(ctx, namespace, exactMatch)
}
