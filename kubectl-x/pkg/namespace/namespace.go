// Package namespace provides Kubernetes namespace resolution functionality
package namespace

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
)

// Resolver defines methods for resolving Kubernetes namespaces
type Resolver interface {
	// Resolve determines the namespace for the given context
	Resolve(contextName string) string
}

type resolver struct {
	kubeconfig  kubeconfig.Interface
	configFlags *genericclioptions.ConfigFlags
}

// NewResolver creates a new namespace resolver with the given kubeconfig and flags
func NewResolver(kubeconfig kubeconfig.Interface, configFlags *genericclioptions.ConfigFlags) Resolver {
	return &resolver{
		kubeconfig:  kubeconfig,
		configFlags: configFlags,
	}
}

func (r *resolver) Resolve(contextName string) string {
	if r.configFlags != nil && r.configFlags.Namespace != nil && *r.configFlags.Namespace != "" {
		return *r.configFlags.Namespace
	}

	namespace, _ := r.kubeconfig.GetNamespaceForContext(contextName)
	if namespace == "" {
		return "default"
	}
	return namespace
}
