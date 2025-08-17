package shell

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
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	kexec "k8s.io/utils/exec"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"github.com/RRethy/kubectl-x/pkg/shortname"
)

// Sheller handles shell command execution in Kubernetes pods
type Sheller struct {
	IOStreams  genericclioptions.IOStreams
	Kubeconfig kubeconfig.Interface
	Context    string
	Namespace  string
	Fzf        fzf.Interface
	K8sClient  kubernetes.Interface
	Exec       kexec.Interface
}

// Shell executes either a shell command in a pod or a debug container
func (s *Sheller) Shell(ctx context.Context, target string, container string, command string, debug bool, image string) error {
	if debug {
		klog.V(2).Infof("Debug operation started: target=%s container=%s command=%s image=%s", target, container, command, image)
	} else {
		klog.V(2).Infof("Shell operation started: target=%s container=%s command=%s", target, container, command)
	}
	
	pod, err := s.resolvePod(ctx, target)
	if err != nil {
		return fmt.Errorf("resolving pod: %w", err)
	}
	
	if debug {
		klog.V(4).Infof("Resolved target to pod for debug: %s", pod.GetName())
	} else {
		klog.V(4).Infof("Resolved target to pod: %s", pod.GetName())
	}

	var args []string
	if debug {
		args = []string{"debug", "-it", pod.GetName()}
		args = append(args, "--context", s.Context)
		args = append(args, "-n", s.Namespace)

		if image != "" {
			args = append(args, "--image", image)
			klog.V(5).Infof("Using debug image: %s", image)
		}

		if container != "" {
			args = append(args, "--target", container)
			klog.V(5).Infof("Targeting container: %s", container)
		}
	} else {
		args = []string{"exec", "-it", pod.GetName()}
		args = append(args, "--context", s.Context)
		args = append(args, "-n", s.Namespace)

		if container != "" {
			args = append(args, "-c", container)
			klog.V(5).Infof("Using specific container: %s", container)
		}
	}

	args = append(args, "--", command)

	if debug {
		klog.V(6).Infof("Executing kubectl debug command: %v", args)
	} else {
		klog.V(6).Infof("Executing kubectl command: %v", args)
	}
	
	cmd := s.Exec.Command("kubectl", args...)
	cmd.SetStdin(s.IOStreams.In)
	cmd.SetStdout(s.IOStreams.Out)
	cmd.SetStderr(s.IOStreams.ErrOut)

	return cmd.Run()
}


func (s *Sheller) resolvePod(ctx context.Context, target string) (*corev1.Pod, error) {
	klog.V(4).Infof("Resolving pod from target: %s", target)
	target = strings.ToLower(strings.TrimSpace(target))
	var resourceKind, resourceName string
	if strings.Contains(target, " ") || strings.Contains(target, "/") {
		sep := " "
		if strings.Contains(target, "/") {
			sep = "/"
		}

		parts := strings.SplitN(target, sep, 2)
		if len(parts) == 1 {
			resourceName = parts[0]
		} else if len(parts) == 2 {
			resourceKind = shortname.Expand(parts[0])
			resourceName = parts[1]
		}
	} else {
		resourceName = target
	}
	if resourceKind == "" {
		if slices.Contains([]string{"pod", "pods", "deployment", "deploy", "statefulset", "sts", "replicaset", "rs", "daemonset", "ds", "job", "jobs", "service", "svc"}, resourceName) {
			resourceKind = resourceName
			resourceName = ""
		} else {
			resourceKind = "pod"
		}
	}
	klog.V(5).Infof("Parsed target: resourceKind=%s resourceName=%s", resourceKind, resourceName)

	allPods, err := kubernetes.ListInNamespace[*corev1.Pod](ctx, s.K8sClient, s.Namespace)
	if err != nil {
		return nil, fmt.Errorf("getting all pods: %w", err)
	}
	klog.V(6).Infof("Retrieved %d pods from namespace %s", len(allPods), s.Namespace)

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
		var selector map[string]string
		selector, err = s.getResourceSelector(ctx, resourceKind, resourceName)
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

	return selectResourceFromList(ctx, s, pods, "", "pod")
}

func (s *Sheller) getResourceSelector(ctx context.Context, resourceKind, resourceName string) (map[string]string, error) {
	switch resourceKind {
	case "deployment", "deploy":
		deployments, err := kubernetes.ListInNamespace[*appsv1.Deployment](ctx, s.K8sClient, s.Namespace)
		if err != nil {
			return nil, fmt.Errorf("listing deployments: %w", err)
		}
		selected, err := selectResourceFromList(ctx, s, deployments, resourceName, "deployment")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("deployment %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil
	case "statefulset", "sts":
		statefulsets, err := kubernetes.ListInNamespace[*appsv1.StatefulSet](ctx, s.K8sClient, s.Namespace)
		if err != nil {
			return nil, fmt.Errorf("listing statefulsets: %w", err)
		}
		selected, err := selectResourceFromList(ctx, s, statefulsets, resourceName, "statefulset")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("statefulset %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil
	case "replicaset", "rs":
		replicasets, err := kubernetes.ListInNamespace[*appsv1.ReplicaSet](ctx, s.K8sClient, s.Namespace)
		if err != nil {
			return nil, fmt.Errorf("listing replicasets: %w", err)
		}
		selected, err := selectResourceFromList(ctx, s, replicasets, resourceName, "replicaset")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("replicaset %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil
	case "daemonset", "ds":
		daemonsets, err := kubernetes.ListInNamespace[*appsv1.DaemonSet](ctx, s.K8sClient, s.Namespace)
		if err != nil {
			return nil, fmt.Errorf("listing daemonsets: %w", err)
		}
		selected, err := selectResourceFromList(ctx, s, daemonsets, resourceName, "daemonset")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("daemonset %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil
	case "job", "jobs":
		jobs, err := kubernetes.ListInNamespace[*batchv1.Job](ctx, s.K8sClient, s.Namespace)
		if err != nil {
			return nil, fmt.Errorf("listing jobs: %w", err)
		}
		selected, err := selectResourceFromList(ctx, s, jobs, resourceName, "job")
		if err != nil {
			return nil, err
		}
		if selected.Spec.Selector == nil || selected.Spec.Selector.MatchLabels == nil {
			return nil, fmt.Errorf("job %s has no selector", selected.GetName())
		}
		return selected.Spec.Selector.MatchLabels, nil
	case "service", "svc":
		services, err := kubernetes.ListInNamespace[*corev1.Service](ctx, s.K8sClient, s.Namespace)
		if err != nil {
			return nil, fmt.Errorf("listing services: %w", err)
		}
		selected, err := selectResourceFromList(ctx, s, services, resourceName, "service")
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

func selectResourceFromList[T metav1.Object](ctx context.Context, s *Sheller, resources []T, resourceName, resourceType string) (T, error) {
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

	selected, err := s.Fzf.Run(ctx, matchNames, fzf.Config{
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
