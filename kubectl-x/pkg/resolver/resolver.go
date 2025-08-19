package resolver

import (
	"context"
	"fmt"
	"slices"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"github.com/RRethy/kubectl-x/pkg/shortname"
)

type Resolver interface {
	ResolveTarget(ctx context.Context, target string) (string, string)
	ResolvePod(ctx context.Context, resourceKind, resourceName, namespace string, fzfConfig fzf.Config) (*corev1.Pod, error)
	ResolveResource(ctx context.Context, resourceKind, resourceName, namespace string, fzfConfig fzf.Config) (metav1.Object, error)
}

type resolver struct {
	K8sClient kubernetes.Interface
	Fzf       fzf.Interface
}

func New(k8sClient kubernetes.Interface, fzf fzf.Interface) Resolver {
	return &resolver{K8sClient: k8sClient, Fzf: fzf}
}

func (r *resolver) ResolveTarget(ctx context.Context, target string) (string, string) {
	resourceKind, resourceName := r.parseTarget(target)
	resourceKind = shortname.Expand(resourceKind)
	return resourceKind, resourceName
}

func (r *resolver) ResolvePod(ctx context.Context, resourceKind, resourceName, namespace string, fzfConfig fzf.Config) (*corev1.Pod, error) {
	if !slices.Contains([]string{
		"pod", "pods", "deployment", "deployments", "statefulset", "statefulsets",
		"replicaset", "replicasets", "daemonset", "daemonsets", "job", "jobs", "service", "services",
	}, resourceKind) {
		return nil, fmt.Errorf("resource type %s cannot resolve to pods", resourceKind)
	}

	resolved, err := r.ResolveResource(ctx, resourceKind, resourceName, namespace, fzfConfig)
	if err != nil {
		return nil, fmt.Errorf("resolving resource %s/%s: %w", resourceKind, resourceName, err)
	}

	var labelSelector labels.Selector
	switch res := resolved.(type) {
	case *corev1.Pod:
		return res, nil
	case *appsv1.Deployment:
		labelSelector = labels.SelectorFromSet(res.Spec.Selector.MatchLabels)
	case *appsv1.StatefulSet:
		labelSelector = labels.SelectorFromSet(res.Spec.Selector.MatchLabels)
	case *appsv1.ReplicaSet:
		labelSelector = labels.SelectorFromSet(res.Spec.Selector.MatchLabels)
	case *appsv1.DaemonSet:
		labelSelector = labels.SelectorFromSet(res.Spec.Selector.MatchLabels)
	case *batchv1.Job:
		labelSelector = labels.SelectorFromSet(res.Spec.Selector.MatchLabels)
	case *corev1.Service:
		labelSelector = labels.SelectorFromSet(res.Spec.Selector)
	default:
		return nil, fmt.Errorf("resource type %T cannot resolve to a single pod", res)
	}

	// TODO: we could just pass the labelSelector to ListOptions to filter on the server side
	allPods, err := kubernetes.ListInNamespace[*corev1.Pod](ctx, r.K8sClient, namespace)
	if err != nil {
		return nil, fmt.Errorf("getting all pods: %w", err)
	}

	var pods []*corev1.Pod
	for _, pod := range allPods {
		if labelSelector.Matches(labels.Set(pod.GetLabels())) {
			pods = append(pods, pod)
		}
	}

	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods found for %s/%s", resourceKind, resourceName)
	}
	if len(pods) == 1 {
		return pods[0], nil
	}

	return selectResource(ctx, r.Fzf, pods, "pod", fzfConfig)
}

func (r *resolver) ResolveResource(ctx context.Context, resourceKind, resourceName, namespace string, fzfConfig fzf.Config) (metav1.Object, error) {
	objects, err := kubernetes.ListObjectInNamespace(ctx, r.K8sClient, resourceKind, namespace)
	if err != nil {
		return nil, fmt.Errorf("listing resources in namespace %s: %w", namespace, err)
	}

	obj, err := selectResourceByName(ctx, r.Fzf, objects, resourceKind, resourceName, fzfConfig)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (r *resolver) parseTarget(target string) (resourceKind, resourceName string) {
	target = strings.TrimSpace(target)

	var sep string
	if strings.Contains(target, "/") {
		sep = "/"
	} else if strings.Contains(target, " ") {
		sep = " "
	}

	if sep != "" {
		parts := strings.SplitN(target, sep, 2)
		if len(parts) == 2 {
			resourceKind = shortname.Expand(parts[0])
			resourceName = parts[1]
		} else if len(parts) == 1 {
			resourceKind = parts[0]
		}
	} else {
		resourceKind = target
	}

	return resourceKind, resourceName
}

func selectResourceByName[T metav1.Object](ctx context.Context, fzfClient fzf.Interface, resources []T, resourceKind string, resourceName string, fzfConfig fzf.Config) (T, error) {
	var matches []T
	for _, resource := range resources {
		if resourceName == "" || strings.Contains(resource.GetName(), resourceName) {
			matches = append(matches, resource)
		}
	}

	if len(matches) == 0 {
		var zero T
		return zero, fmt.Errorf("no %s found matching '%s'", resourceKind, resourceName)
	}

	return selectResource(ctx, fzfClient, matches, resourceKind, fzfConfig)
}

func selectResource[T metav1.Object](ctx context.Context, fzfClient fzf.Interface, resources []T, resourceKind string, fzfConfig fzf.Config) (T, error) {
	var zero T
	if len(resources) == 0 {
		return zero, fmt.Errorf("no %s found", resourceKind)
	}

	matchNames := make([]string, len(resources))
	for i, resource := range resources {
		matchNames[i] = resource.GetName()
	}

	// TODO: fzf.Config should be configurable
	fzfConfig.Prompt = fmt.Sprintf("Select %s: ", resourceKind)
	fzfConfig.Query = ""
	selected, err := fzfClient.Run(ctx, matchNames, fzfConfig)
	if err != nil {
		return zero, fmt.Errorf("selecting %s: %w", resourceKind, err)
	}

	if len(selected) == 0 {
		return zero, fmt.Errorf("no %s selected", resourceKind)
	}

	for _, resource := range resources {
		if resource.GetName() == selected[0] {
			return resource, nil
		}
	}

	return zero, fmt.Errorf("selected %s not found", resourceKind)
}
