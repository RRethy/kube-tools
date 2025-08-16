// Package testing provides mock implementations for testing kubeconfig operations
package testing

import (
	"errors"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ kubeconfig.Interface = &FakeKubeConfig{}

// FakeKubeConfig is a mock implementation of kubeconfig.Interface for testing
type FakeKubeConfig struct {
	contexts         map[string]*api.Context
	currentContext   string
	currentNamespace string
}

// NewFakeKubeConfig creates a new fake kubeconfig with the given contexts and current settings
func NewFakeKubeConfig(contexts map[string]*api.Context, currentContext, currentNamespace string) *FakeKubeConfig {
	return &FakeKubeConfig{
		contexts,
		currentContext,
		currentNamespace,
	}
}

func (fake *FakeKubeConfig) Contexts() []string {
	contexts := make([]string, 0, len(fake.contexts))
	for _, context := range fake.contexts {
		contexts = append(contexts, context.Cluster)
	}
	return contexts
}

func (fake *FakeKubeConfig) SetContext(context string) error {
	if context == "" {
		return errors.New("context cannot be empty")
	}
	fake.currentContext = context
	return nil
}

func (fake *FakeKubeConfig) SetNamespace(namespace string) error {
	if namespace == "" {
		return errors.New("namespace cannot be empty")
	}
	fake.currentNamespace = namespace
	return nil
}

func (fake *FakeKubeConfig) GetCurrentContext() (string, error) {
	if fake.currentContext == "" {
		return "", errors.New("current context not set")
	}
	return fake.currentContext, nil
}

func (fake *FakeKubeConfig) GetCurrentNamespace() (string, error) {
	if fake.currentNamespace == "" {
		return "", errors.New("current namespace not set")
	}
	return fake.currentNamespace, nil
}

func (fake *FakeKubeConfig) GetNamespaceForContext(context string) (string, error) {
	return fake.contexts[context].Namespace, nil
}

func (fake *FakeKubeConfig) Write() error {
	return nil
}
