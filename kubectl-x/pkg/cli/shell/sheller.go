package shell

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	kexec "k8s.io/utils/exec"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/resolver"
)

// Sheller handles shell command execution in Kubernetes pods
type Sheller struct {
	IOStreams  genericclioptions.IOStreams
	Kubeconfig kubeconfig.Interface
	Context    string
	Namespace  string
	Resolver   resolver.Resolver
	Exec       kexec.Interface
}

// Shell executes either a shell command in a pod or a debug container
func (s *Sheller) Shell(ctx context.Context, target string, container string, command string, debug bool, image string) error {
	if debug {
		klog.V(2).Infof("Debug operation started: target=%s container=%s command=%s image=%s", target, container, command, image)
	} else {
		klog.V(2).Infof("Shell operation started: target=%s container=%s command=%s", target, container, command)
	}

	resourceKind, resourceName := s.Resolver.ResolveTarget(ctx, target)
	if resourceKind == "" {
		resourceKind = "pod"
	}

	pod, err := s.Resolver.ResolvePod(ctx, resourceKind, resourceName, s.Namespace, fzf.Config{
		Sorted: true,
	})
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
