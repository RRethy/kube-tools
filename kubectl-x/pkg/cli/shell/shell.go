package shell

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
)

// Shell executes either a shell command in a pod or a debug container
func Shell(ctx context.Context, configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags, target string, container string, command string, debug bool, image string) error {
	kubeConfig, err := kubeconfig.NewKubeConfig(kubeconfig.WithConfigFlags(configFlags))
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}

	currentContext := xcontext.NewResolver(kubeConfig, configFlags).Resolve()
	namespace := namespace.NewResolver(kubeConfig, configFlags).Resolve(currentContext)

	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	k8sClient := kubernetes.NewClient(configFlags, resourceBuilderFlags)
	fzf := fzf.NewFzf(fzf.WithIOStreams(ioStreams))
	exec := kexec.New()

	return (&Sheller{
		IOStreams:  ioStreams,
		Kubeconfig: kubeConfig,
		Context:    currentContext,
		Namespace:  namespace,
		K8sClient:  k8sClient,
		Fzf:        fzf,
		Exec:       exec,
	}).Shell(ctx, target, container, command, debug, image)
}
