// Package kubeconfig provides utilities for managing Kubernetes configuration files
package kubeconfig

import (
	"errors"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// Interface defines methods for interacting with kubeconfig files
type Interface interface {
	// Contexts returns all available context names
	Contexts() []string
	// SetContext switches to the specified context
	SetContext(context string) error
	// SetNamespace changes the namespace for the current context
	SetNamespace(namespace string) error
	// GetCurrentContext returns the name of the active context
	GetCurrentContext() (string, error)
	// GetCurrentNamespace returns the namespace for the current context
	GetCurrentNamespace() (string, error)
	// GetNamespaceForContext returns the namespace configured for a specific context
	GetNamespaceForContext(context string) (string, error)
	// Write persists changes to the kubeconfig file
	Write() error
	// GetKubeconfigPath returns the path to the kubeconfig file being used
	GetKubeconfigPath() string
	// WriteToFile writes the kubeconfig to a specific file path
	WriteToFile(path string) error
}

// KubeConfig provides kubeconfig file operations
type KubeConfig struct {
	configAccess clientcmd.ConfigAccess
	APIConfig    *api.Config
}

// Option configures a KubeConfig
type Option func(*KubeConfig)

// WithConfigFlags sets the config flags for kubeconfig
func WithConfigFlags(configFlags *genericclioptions.ConfigFlags) Option {
	return func(kc *KubeConfig) {
		var kubeconfigPath string
		if configFlags != nil && configFlags.KubeConfig != nil && *configFlags.KubeConfig != "" {
			kubeconfigPath = *configFlags.KubeConfig
		}

		if kubeconfigPath != "" {
			kc.configAccess = &clientcmd.PathOptions{
				GlobalFile:   kubeconfigPath,
				LoadingRules: &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			}
		}
	}
}

// NewKubeConfig creates a kubeconfig manager with options
func NewKubeConfig(opts ...Option) (Interface, error) {
	kc := &KubeConfig{
		configAccess: clientcmd.NewDefaultPathOptions(),
	}

	for _, opt := range opts {
		opt(kc)
	}

	config, err := kc.configAccess.GetStartingConfig()
	if err != nil {
		return KubeConfig{}, err
	}
	kc.APIConfig = config

	return *kc, nil
}

func (kubeConfig KubeConfig) Contexts() []string {
	contexts := make([]string, 0, len(kubeConfig.APIConfig.Contexts))
	for context := range kubeConfig.APIConfig.Contexts {
		contexts = append(contexts, context)
	}
	return contexts
}

func (kubeConfig KubeConfig) SetContext(context string) error {
	klog.V(4).Infof("Setting kubeconfig context: %s", context)
	if len(context) == 0 {
		return errors.New("context cannot be empty")
	}

	ctx, ok := kubeConfig.APIConfig.Contexts[context]
	if !ok {
		return fmt.Errorf("context '%s' not found", context)
	}

	kubeConfig.APIConfig.CurrentContext = context
	if len(ctx.Namespace) > 0 {
		return kubeConfig.SetNamespace(ctx.Namespace)
	}
	return kubeConfig.SetNamespace("default")
}

func (kubeConfig KubeConfig) SetNamespace(namespace string) error {
	klog.V(4).Infof("Setting kubeconfig namespace: %s", namespace)
	if len(namespace) == 0 {
		return errors.New("namespace cannot be empty")
	}

	ctx, ok := kubeConfig.APIConfig.Contexts[kubeConfig.APIConfig.CurrentContext]
	if !ok {
		return errors.New("current context not found")
	}

	ctx.Namespace = namespace
	return nil
}

func (kubeConfig KubeConfig) GetCurrentContext() (string, error) {
	if len(kubeConfig.APIConfig.CurrentContext) == 0 {
		return "", errors.New("current context not set")
	}
	return kubeConfig.APIConfig.CurrentContext, nil
}

func (kubeConfig KubeConfig) GetCurrentNamespace() (string, error) {
	ctx, ok := kubeConfig.APIConfig.Contexts[kubeConfig.APIConfig.CurrentContext]
	if !ok {
		return "", errors.New("current context not found")
	}
	return ctx.Namespace, nil
}

func (kubeConfig KubeConfig) GetNamespaceForContext(context string) (string, error) {
	ctx, ok := kubeConfig.APIConfig.Contexts[context]
	if !ok {
		return "", fmt.Errorf("context '%s' not found", context)
	}
	return ctx.Namespace, nil
}

func (kubeConfig KubeConfig) Write() error {
	klog.V(4).Info("Writing kubeconfig changes to disk")
	return clientcmd.ModifyConfig(kubeConfig.configAccess, *kubeConfig.APIConfig, true)
}

func (kubeConfig KubeConfig) GetKubeconfigPath() string {
	if kubeConfig.configAccess != nil {
		if filename := kubeConfig.configAccess.GetDefaultFilename(); filename != "" {
			return filename
		}
	}
	return clientcmd.RecommendedHomeFile
}

func (kubeConfig KubeConfig) WriteToFile(path string) error {
	return clientcmd.WriteToFile(*kubeConfig.APIConfig, path)
}
