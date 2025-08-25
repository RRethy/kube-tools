// Package kubernetes provides a client for interacting with Kubernetes resources
package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/scheme"
)

// Interface defines methods for Kubernetes resource operations
type Interface interface {
	// List retrieves all resources of the specified type
	List(ctx context.Context, resourceType string) ([]any, error)
	// ListInNamespace retrieves resources of the specified type in the given namespace
	ListInNamespace(ctx context.Context, resourceType, namespace string) ([]any, error)
	// GetResource retrieves a specific resource by kind, name, and namespace using dynamic client
	GetResource(ctx context.Context, kind, name, namespace string) (metav1.Object, error)
}

// Client implements Kubernetes resource operations
type Client struct {
	configFlags          *genericclioptions.ConfigFlags
	resourceBuilderFlags *genericclioptions.ResourceBuilderFlags
	dynamicClient        dynamic.Interface
	discoveryClient      discovery.DiscoveryInterface
	restMapper           meta.RESTMapper
}

// NewClient creates a new Kubernetes client with the given configuration
func NewClient(configFlags *genericclioptions.ConfigFlags, resourceBuilderFlags *genericclioptions.ResourceBuilderFlags) Interface {
	// Create REST config
	restConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		klog.V(2).Infof("Failed to create REST config, some features may not work: %v", err)
		return &Client{
			configFlags:          configFlags,
			resourceBuilderFlags: resourceBuilderFlags,
		}
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		klog.V(2).Infof("Failed to create dynamic client, some features may not work: %v", err)
		return &Client{
			configFlags:          configFlags,
			resourceBuilderFlags: resourceBuilderFlags,
		}
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		klog.V(2).Infof("Failed to create discovery client, some features may not work: %v", err)
		return &Client{
			configFlags:          configFlags,
			resourceBuilderFlags: resourceBuilderFlags,
			dynamicClient:        dynamicClient,
		}
	}

	// Create REST mapper
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		klog.V(2).Infof("Failed to get API group resources, some features may not work: %v", err)
		return &Client{
			configFlags:          configFlags,
			resourceBuilderFlags: resourceBuilderFlags,
			dynamicClient:        dynamicClient,
			discoveryClient:      discoveryClient,
		}
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	return &Client{
		configFlags:          configFlags,
		resourceBuilderFlags: resourceBuilderFlags,
		dynamicClient:        dynamicClient,
		discoveryClient:      discoveryClient,
		restMapper:           restMapper,
	}
}

func (c *Client) List(ctx context.Context, resourceType string) ([]any, error) {
	return c.list(resourceType, "")
}

// ListInNamespace retrieves resources of the specified type in the given namespace
func (c *Client) ListInNamespace(ctx context.Context, resourceType, namespace string) ([]any, error) {
	return c.list(resourceType, namespace)
}

func (c *Client) list(resourceType, namespace string) ([]any, error) {
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

// GetResource retrieves a specific resource by kind, name, and namespace using dynamic client
func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (metav1.Object, error) {
	if c.dynamicClient == nil || c.restMapper == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}

	gvk, err := c.restMapper.KindFor(schema.GroupVersionResource{Resource: strings.ToLower(kind + "s")})
	if err != nil {
		gvk, err = c.restMapper.KindFor(schema.GroupVersionResource{Resource: strings.ToLower(kind)})
		if err != nil {
			gvks, err := c.restMapper.KindsFor(schema.GroupVersionResource{Resource: strings.ToLower(kind)})
			if err != nil || len(gvks) == 0 {
				return nil, fmt.Errorf("finding GroupVersionKind for %s: %w", kind, err)
			}
			gvk = gvks[0]
		}
	}

	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("getting REST mapping for %s: %w", kind, err)
	}

	var resource *unstructured.Unstructured
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resource, err = c.dynamicClient.Resource(mapping.Resource).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		resource, err = c.dynamicClient.Resource(mapping.Resource).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("getting %s/%s: %w", kind, name, err)
	}

	return resource, nil
}
