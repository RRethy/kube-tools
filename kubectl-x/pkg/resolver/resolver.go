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
	"k8s.io/klog/v2"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"github.com/RRethy/kubectl-x/pkg/shortname"
)

type Interface interface {
	ResolvePod(ctx context.Context, target string, namespace string) (*corev1.Pod, error)
}

type Resolver struct {
	K8sClient kubernetes.Interface
	Fzf       fzf.Interface
}

func New(k8sClient kubernetes.Interface, fzf fzf.Interface) *Resolver {
	return &Resolver{
		K8sClient: k8sClient,
		Fzf:       fzf,
	}
}

func (r *Resolver) ResolvePod(ctx context.Context, target string, namespace string) (*corev1.Pod, error) {
	klog.V(4).Infof("Resolving pod from target: %s", target)
	target = strings.ToLower(strings.TrimSpace(target))

	resourceKind, resourceName := r.parseTarget(target)
	klog.V(5).Infof("Parsed target: resourceKind=%s resourceName=%s", resourceKind, resourceName)

	allPods, err := kubernetes.ListInNamespace[*corev1.Pod](ctx, r.K8sClient, namespace)
	if err != nil {
		return nil, fmt.Errorf("getting all pods: %w", err)
	}
	klog.V(6).Infof("Retrieved %d pods from namespace %s", len(allPods), namespace)

	pods, err := r.filterPods(ctx, allPods, resourceKind, resourceName, namespace)
	if err != nil {
		return nil, err
	}

	return selectResource(ctx, r.Fzf, pods, "", "pod")
}

func (r *Resolver) parseTarget(target string) (resourceKind, resourceName string) {
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
			resourceName = parts[0]
		}
	} else {
		resourceName = target
	}

	if resourceKind == "" {
		supportedTypes := []string{
			"pod", "pods", "deployment", "deploy", "statefulset", "sts",
			"replicaset", "rs", "daemonset", "ds", "job", "jobs", "service", "svc",
		}
		if slices.Contains(supportedTypes, resourceName) {
			resourceKind = resourceName
			resourceName = ""
		}
		if resourceKind == "" {
			resourceKind = "pod"
		}
	}

	return resourceKind, resourceName
}

func (r *Resolver) filterPods(ctx context.Context, allPods []*corev1.Pod, resourceKind, resourceName, namespace string) ([]*corev1.Pod, error) {
	var pods []*corev1.Pod
	klog.V(5).Infof("Filtering pods by resource type: %s", resourceKind)

	switch resourceKind {
	case "pod", "pods":
		for _, pod := range allPods {
			if strings.Contains(pod.GetName(), resourceName) {
				pods = append(pods, pod)
			}
		}
	case "deployment", "deploy", "statefulset", "sts", "replicaset", "rs", "daemonset", "ds", "job", "jobs", "service", "svc":
		selector, err := r.getResourceSelector(ctx, resourceKind, resourceName, namespace)
		if err != nil {
			return nil, fmt.Errorf("getting %s selector: %w", resourceKind, err)
		}
		labelSelector := labels.SelectorFromSet(selector)
		for _, pod := range allPods {
			if labelSelector.Matches(labels.Set(pod.GetLabels())) {
				pods = append(pods, pod)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceKind)
	}

	return pods, nil
}

func (r *Resolver) getResourceSelector(ctx context.Context, resourceKind, resourceName, namespace string) (map[string]string, error) {
	switch resourceKind {
	case "deployment", "deploy":
		deployments, err := kubernetes.ListInNamespace[*appsv1.Deployment](ctx, r.K8sClient, namespace)
		if err != nil {
			return nil, fmt.Errorf("listing deployments: %w", err)
		}
		selected, err := selectResource(ctx, r.Fzf, deployments, resourceName, "deployment")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("deployment %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil

	case "statefulset", "sts":
		statefulsets, err := kubernetes.ListInNamespace[*appsv1.StatefulSet](ctx, r.K8sClient, namespace)
		if err != nil {
			return nil, fmt.Errorf("listing statefulsets: %w", err)
		}
		selected, err := selectResource(ctx, r.Fzf, statefulsets, resourceName, "statefulset")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("statefulset %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil

	case "replicaset", "rs":
		replicasets, err := kubernetes.ListInNamespace[*appsv1.ReplicaSet](ctx, r.K8sClient, namespace)
		if err != nil {
			return nil, fmt.Errorf("listing replicasets: %w", err)
		}
		selected, err := selectResource(ctx, r.Fzf, replicasets, resourceName, "replicaset")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("replicaset %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil

	case "daemonset", "ds":
		daemonsets, err := kubernetes.ListInNamespace[*appsv1.DaemonSet](ctx, r.K8sClient, namespace)
		if err != nil {
			return nil, fmt.Errorf("listing daemonsets: %w", err)
		}
		selected, err := selectResource(ctx, r.Fzf, daemonsets, resourceName, "daemonset")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("daemonset %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil

	case "job", "jobs":
		jobs, err := kubernetes.ListInNamespace[*batchv1.Job](ctx, r.K8sClient, namespace)
		if err != nil {
			return nil, fmt.Errorf("listing jobs: %w", err)
		}
		selected, err := selectResource(ctx, r.Fzf, jobs, resourceName, "job")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("job %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil

	case "service", "svc":
		services, err := kubernetes.ListInNamespace[*corev1.Service](ctx, r.K8sClient, namespace)
		if err != nil {
			return nil, fmt.Errorf("listing services: %w", err)
		}
		selected, err := selectResource(ctx, r.Fzf, services, resourceName, "service")
		if err != nil {
			return nil, err
		}
		if len(selected.Spec.Selector) == 0 {
			return nil, fmt.Errorf("service %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector, nil

	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", resourceKind)
	}
}

func selectResource[T metav1.Object](ctx context.Context, fzfClient fzf.Interface, resources []T, resourceName, resourceType string) (T, error) {
	var zero T
	var matches []T
	for _, resource := range resources {
		if strings.Contains(resource.GetName(), resourceName) {
			matches = append(matches, resource)
		}
	}

	if len(matches) == 0 {
		return zero, fmt.Errorf("no %s found matching '%s'", resourceType, resourceName)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	matchNames := make([]string, len(matches))
	for i, resource := range matches {
		matchNames[i] = resource.GetName()
	}

	selected, err := fzfClient.Run(ctx, matchNames, fzf.Config{
		Prompt: fmt.Sprintf("Select %s: ", resourceType),
		Query:  resourceName,
		Sorted: true,
		Multi:  false,
	})
	if err != nil {
		return zero, fmt.Errorf("selecting %s: %w", resourceType, err)
	}

	if len(selected) == 0 {
		return zero, fmt.Errorf("no %s selected", resourceType)
	}

	for _, resource := range matches {
		if resource.GetName() == selected[0] {
			return resource, nil
		}
	}

	return zero, fmt.Errorf("selected %s not found", resourceType)
}
