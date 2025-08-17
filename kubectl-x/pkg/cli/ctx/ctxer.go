// Package ctx provides context switching functionality for kubectl-x
package ctx

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"

	"github.com/RRethy/kubectl-x/pkg/cli/ns"
	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/history"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
)

// Ctxer handles Kubernetes context switching operations
type Ctxer struct {
	KubeConfig kubeconfig.Interface
	IoStreams  genericiooptions.IOStreams
	K8sClient  kubernetes.Interface
	Fzf        fzf.Interface
	History    history.Interface
}

// NewCtxer creates a new context switcher with the provided dependencies
func NewCtxer(kubeConfig kubeconfig.Interface, ioStreams genericiooptions.IOStreams, k8sClient kubernetes.Interface, fzf fzf.Interface, history history.Interface) Ctxer {
	return Ctxer{
		KubeConfig: kubeConfig,
		IoStreams:  ioStreams,
		K8sClient:  k8sClient,
		Fzf:        fzf,
		History:    history,
	}
}

// Ctx switches to a context matching the given substring and optionally sets namespace
func (c Ctxer) Ctx(ctx context.Context, contextSubstring, namespaceSubstring string, exactMatch bool) error {
	klog.V(2).Infof("Context switching operation started: contextSubstring=%s namespaceSubstring=%s exactMatch=%t", contextSubstring, namespaceSubstring, exactMatch)

	var selectedContext string
	var selectedNamespace string
	var err error
	if contextSubstring == "-" {
		klog.V(4).Info("Using previous context from history")
		selectedContext, err = c.History.Get("context", 1)
		if err != nil {
			return fmt.Errorf("getting context from history: %s", err)
		}

		selectedNamespace, err = c.KubeConfig.GetNamespaceForContext(selectedContext)
		if err != nil {
			return fmt.Errorf("getting namespace for context: %s", err)
		}
	} else {
		klog.V(4).Infof("Running fzf for context selection: query=%s", contextSubstring)
		fzfCfg := fzf.Config{ExactMatch: exactMatch, Sorted: true, Multi: false, Prompt: "Select context", Query: contextSubstring}
		var results []string
		results, err = c.Fzf.Run(context.Background(), c.KubeConfig.Contexts(), fzfCfg)
		if err != nil {
			return fmt.Errorf("selecting context: %s", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("no context selected")
		}
		selectedContext = results[0]
		klog.V(5).Infof("User selected context: %s", selectedContext)
	}

	c.History.Add("context", selectedContext)

	klog.V(1).Infof("Setting Kubernetes context: %s", selectedContext)
	err = c.KubeConfig.SetContext(selectedContext)
	if err != nil {
		return fmt.Errorf("setting context: %w", err)
	}

	err = c.History.Write()
	if err != nil {
		klog.Warningf("Failed to write history: %v", err)
	}

	err = c.KubeConfig.Write()
	if err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	klog.V(4).Info("Successfully wrote kubeconfig after context switch")

	klog.V(1).Infof("Successfully switched context: %s", selectedContext)
	fmt.Fprintf(c.IoStreams.Out, "Switched to context \"%s\".\n", selectedContext)

	if selectedNamespace == "" {
		klog.V(4).Info("No namespace selected from history, prompting user for namespace selection")
		nser := ns.NewNser(c.KubeConfig, c.IoStreams, c.K8sClient, c.Fzf, c.History)
		return nser.Ns(ctx, namespaceSubstring, exactMatch)
	}

	klog.V(4).Infof("Setting namespace from history: %s", selectedNamespace)
	err = c.KubeConfig.SetNamespace(selectedNamespace)
	if err != nil {
		return fmt.Errorf("setting namespace: %w", err)
	}

	err = c.KubeConfig.Write()
	if err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	klog.V(4).Info("Successfully wrote kubeconfig after namespace switch")

	fmt.Fprintf(c.IoStreams.Out, "Switched to namespace \"%s\".\n", selectedNamespace)
	return nil
}
