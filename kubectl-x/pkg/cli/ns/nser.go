package ns

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/history"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
)

type Nser struct {
	KubeConfig kubeconfig.Interface
	IoStreams  genericiooptions.IOStreams
	K8sClient  kubernetes.Interface
	Fzf        fzf.Interface
	History    history.Interface
}

func NewNser(kubeConfig kubeconfig.Interface, ioStreams genericiooptions.IOStreams, k8sClient kubernetes.Interface, fzf fzf.Interface, history history.Interface) Nser {
	return Nser{
		KubeConfig: kubeConfig,
		IoStreams:  ioStreams,
		K8sClient:  k8sClient,
		Fzf:        fzf,
		History:    history,
	}
}

func (n Nser) Ns(ctx context.Context, namespace string, exactMatch bool) error {
	var selectedNamespace string
	var err error
	if namespace == "-" {
		selectedNamespace, err = n.History.Get("namespace", 1)
		if err != nil {
			return fmt.Errorf("getting namespace from history: %s", err)
		}
	} else {
		namespaces, err := kubernetes.List[*corev1.Namespace](ctx, n.K8sClient)
		if err != nil {
			return fmt.Errorf("listing namespaces: %s", err)
		}

		namespaceNames := make([]string, len(namespaces))
		for i, ns := range namespaces {
			namespaceNames[i] = ns.Name
		}

		fzfCfg := fzf.Config{ExactMatch: exactMatch, Sorted: true, Multi: false, Prompt: "Select context", Query: namespace}
		results, err := n.Fzf.Run(context.Background(), namespaceNames, fzfCfg)
		if err != nil {
			return fmt.Errorf("selecting namespace: %s", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("no namespace selected")
		}
		selectedNamespace = results[0]
	}

	err = n.KubeConfig.SetNamespace(selectedNamespace)
	if err != nil {
		return fmt.Errorf("setting namespace: %w", err)
	}

	n.History.Add("namespace", selectedNamespace)

	err = n.KubeConfig.Write()
	if err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}

	err = n.History.Write()
	if err != nil {
		fmt.Fprintf(n.IoStreams.ErrOut, "writing history: %s\n", err)
	}

	fmt.Fprintf(n.IoStreams.Out, "Switched to namespace \"%s\".\n", selectedNamespace)

	return nil
}
