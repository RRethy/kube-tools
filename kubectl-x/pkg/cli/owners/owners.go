package owners

import (
	"context"
	"fmt"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	kexec "k8s.io/utils/exec"

	xcontext "github.com/RRethy/kubectl-x/pkg/context"
	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"github.com/RRethy/kubectl-x/pkg/namespace"
	"github.com/RRethy/kubectl-x/pkg/resolver"
)

func Owners(ctx context.Context, configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags, target string) error {
	kubeConfig, err := kubeconfig.NewKubeConfig(kubeconfig.WithConfigFlags(configFlags))
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}

	currentContext := xcontext.NewResolver(kubeConfig, configFlags).Resolve()
	currentNamespace := namespace.NewResolver(kubeConfig, configFlags).Resolve(currentContext)

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	k8sClient := kubernetes.NewClient(configFlags, resourceBuilderFlags)
	fzfClient := fzf.NewFzf(fzf.WithIOStreams(ioStreams))
	resolver := resolver.New(k8sClient, fzfClient)
	exec := kexec.New()

	return (&Ownerser{
		IOStreams: ioStreams,
		Context:   currentContext,
		Namespace: currentNamespace,
		K8sClient: k8sClient,
		Resolver:  resolver,
		Fzf:       fzfClient,
		Exec:      exec,
	}).Owners(ctx, target)
}

