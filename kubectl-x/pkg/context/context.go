package context

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
)

type Resolver interface {
	Resolve() string
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

func (r *resolver) Resolve() string {
	if r.configFlags != nil && r.configFlags.Context != nil && *r.configFlags.Context != "" {
		return *r.configFlags.Context
	}

	currentContext, err := r.kubeconfig.GetCurrentContext()
	if err != nil || currentContext == "" {
		return ""
	}

	return currentContext
}
