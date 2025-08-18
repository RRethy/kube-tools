// Package testing provides mock implementations for testing kubeconfig operations
package testing

import (
	"errors"
	"os"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ kubeconfig.Interface = &FakeKubeConfig{}

// FakeKubeConfig is a mock implementation of kubeconfig.Interface for testing
type FakeKubeConfig struct {
	contexts         map[string]*api.Context
	currentContext   string
	currentNamespace string
	kubeconfigPath   string
	writeError       error
}

// NewFakeKubeConfig creates a new fake kubeconfig with the given contexts and current settings
func NewFakeKubeConfig(contexts map[string]*api.Context, currentContext, currentNamespace string) *FakeKubeConfig {
	return &FakeKubeConfig{
		contexts:         contexts,
		currentContext:   currentContext,
		currentNamespace: currentNamespace,
		kubeconfigPath:   "/fake/path/to/kubeconfig",
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

func (fake *FakeKubeConfig) GetKubeconfigPath() string {
	if fake.kubeconfigPath == "" {
		return "/fake/path/to/kubeconfig"
	}
	return fake.kubeconfigPath
}

func (fake *FakeKubeConfig) WriteToFile(path string) error {
	if fake.writeError != nil {
		return fake.writeError
	}
	content := []byte(`apiVersion: v1
kind: Config
current-context: ` + fake.currentContext + `
contexts:
`)
	for name := range fake.contexts {
		content = append(content, []byte("- name: "+name+"\n")...)
	}
	return os.WriteFile(path, content, 0644)
}

// WithKubeconfigPath sets the kubeconfig path for testing
func (fake *FakeKubeConfig) WithKubeconfigPath(path string) *FakeKubeConfig {
	fake.kubeconfigPath = path
	return fake
}

// WithWriteError sets an error to be returned by WriteToFile
func (fake *FakeKubeConfig) WithWriteError(err error) *FakeKubeConfig {
	fake.writeError = err
	return fake
}
