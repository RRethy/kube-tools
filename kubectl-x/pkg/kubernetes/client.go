// Package kubernetes provides a client for interacting with Kubernetes resources
package kubernetes

import (
	"context"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/klog/v2"
)

// Interface defines methods for Kubernetes resource operations
type Interface interface {
	// List retrieves all resources of the specified type
	List(ctx context.Context, resourceType string) ([]any, error)
	// ListInNamespace retrieves resources of the specified type in the given namespace
	ListInNamespace(ctx context.Context, resourceType, namespace string) ([]any, error)
}

// Client implements Kubernetes resource operations
type Client struct {
	configFlags          *genericclioptions.ConfigFlags
	resourceBuilderFlags *genericclioptions.ResourceBuilderFlags
}

// NewClient creates a new Kubernetes client with the given configuration
func NewClient(configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags) Interface {
	return &Client{configFlags, resourceBuilderFlags}
}

func (c *Client) List(ctx context.Context, resourceType string) ([]any, error) {
	return c.list(ctx, resourceType, "")
}

// ListInNamespace retrieves resources of the specified type in the given namespace
func (c *Client) ListInNamespace(ctx context.Context, resourceType, namespace string) ([]any, error) {
	return c.list(ctx, resourceType, namespace)
}

func (c *Client) list(ctx context.Context, resourceType, namespace string) ([]any, error) {
	klog.V(6).Infof("Listing Kubernetes resources: resourceType=%s namespace=%s", resourceType, namespace)
	
	builder := resource.NewBuilder(c.configFlags).
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		FieldSelectorParam(*c.resourceBuilderFlags.FieldSelector).
		LabelSelectorParam(*c.resourceBuilderFlags.LabelSelector).
		ContinueOnError().
		ResourceTypeOrNameArgs(true, string(resourceType)).
		Flatten()

	if namespace != "" {
		builder = builder.NamespaceParam(namespace).DefaultNamespace()
	} else {
		builder = builder.AllNamespaces(true)
	}

	infos, err := builder.Do().Infos()
	if err != nil {
		return nil, err
	}

	klog.V(6).Infof("Successfully listed %d Kubernetes resources: resourceType=%s namespace=%s", len(infos), resourceType, namespace)

	var res []any
	for _, info := range infos {
		res = append(res, info.Object)
	}
	return res, nil
}
