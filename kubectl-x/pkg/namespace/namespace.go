package namespace

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
)

type Resolver interface {
	Resolve(contextName string) string
}

type resolver struct {
	kubeconfig  kubeconfig.Interface
	configFlags *genericclioptions.ConfigFlags
}

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