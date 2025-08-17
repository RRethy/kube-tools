// Package ns provides namespace switching functionality for kubectl-x
package ns

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/history"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
)

// Nser handles Kubernetes namespace switching operations
type Nser struct {
	KubeConfig kubeconfig.Interface
	IoStreams  genericiooptions.IOStreams
	K8sClient  kubernetes.Interface
	Fzf        fzf.Interface
	History    history.Interface
}

// NewNser creates a new namespace switcher with the provided dependencies
func NewNser(kubeConfig kubeconfig.Interface, ioStreams genericiooptions.IOStreams, k8sClient kubernetes.Interface, fzf fzf.Interface, history history.Interface) Nser {
	return Nser{
		KubeConfig: kubeConfig,
		IoStreams:  ioStreams,
		K8sClient:  k8sClient,
		Fzf:        fzf,
		History:    history,
	}
}

// Ns switches to a namespace matching the given substring
func (n Nser) Ns(ctx context.Context, namespace string, exactMatch bool) error {
	klog.V(2).Infof("Namespace switching operation started: namespace=%s exactMatch=%t", namespace, exactMatch)
	
	var selectedNamespace string
	var err error
	if namespace == "-" {
		klog.V(4).Info("Using previous namespace from history")
		selectedNamespace, err = n.History.Get("namespace", 1)
		if err != nil {
			return fmt.Errorf("getting namespace from history: %s", err)
		}
	} else {
		namespaces, err := kubernetes.List[*corev1.Namespace](ctx, n.K8sClient)
		if err != nil {
			return fmt.Errorf("listing namespaces: %s", err)
		}
		klog.V(6).Infof("Retrieved %d namespaces from cluster", len(namespaces))

		namespaceNames := make([]string, len(namespaces))
		for i, ns := range namespaces {
			namespaceNames[i] = ns.Name
		}

		klog.V(4).Infof("Running fzf for namespace selection: query=%s", namespace)
		fzfCfg := fzf.Config{ExactMatch: exactMatch, Sorted: true, Multi: false, Prompt: "Select context", Query: namespace}
		results, err := n.Fzf.Run(context.Background(), namespaceNames, fzfCfg)
		if err != nil {
			return fmt.Errorf("selecting namespace: %s", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("no namespace selected")
		}
		selectedNamespace = results[0]
		klog.V(5).Infof("User selected namespace: %s", selectedNamespace)
	}

	klog.V(1).Infof("Setting Kubernetes namespace: %s", selectedNamespace)
	err = n.KubeConfig.SetNamespace(selectedNamespace)
	if err != nil {
		return fmt.Errorf("setting namespace: %w", err)
	}

	n.History.Add("namespace", selectedNamespace)

	err = n.KubeConfig.Write()
	if err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	klog.V(4).Info("Successfully wrote kubeconfig after namespace switch")

	err = n.History.Write()
	if err != nil {
		klog.Warningf("Failed to write history: %v", err)
	}

	klog.V(1).Infof("Successfully switched namespace: %s", selectedNamespace)
	fmt.Fprintf(n.IoStreams.Out, "Switched to namespace \"%s\".\n", selectedNamespace)

	return nil
}
