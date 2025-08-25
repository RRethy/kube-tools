package owners

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	kexec "k8s.io/utils/exec"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	"github.com/RRethy/kubectl-x/pkg/resolver"
)

type Ownerser struct {
	IOStreams genericclioptions.IOStreams
	Context   string
	Namespace string
	K8sClient kubernetes.Interface
	Resolver  resolver.Resolver
	Fzf       fzf.Interface
	Exec      kexec.Interface
}

type resourceWithParents struct {
	resource metav1.Object
	kind     string
	parents  []*resourceWithParents
}

func (o *Ownerser) Owners(ctx context.Context, target string) error {
	klog.V(1).Infof("Finding owners for target: %s", target)

	resourceKind, resourceName := o.Resolver.ResolveTarget(ctx, target)

	resource, err := o.Resolver.ResolveResource(ctx, resourceKind, resourceName, o.Namespace, fzf.Config{})
	if err != nil {
		return fmt.Errorf("resolving resource %s/%s: %w", resourceKind, resourceName, err)
	}

	klog.V(2).Infof("Found resource: %s/%s", resourceKind, resource.GetName())

	visited := make(map[string]bool)
	tree := o.buildResourceWithParents(ctx, resource, resourceKind, visited)
	fmt.Fprint(o.IOStreams.Out, o.buildOwnershipLines(tree))
	fmt.Fprintln(o.IOStreams.Out)
	return nil
}

func (o *Ownerser) buildResourceWithParents(ctx context.Context, resource metav1.Object, kind string, visited map[string]bool) *resourceWithParents {
	resourceKey := fmt.Sprintf("%s/%s/%s", resource.GetNamespace(), kind, resource.GetName())
	if visited[resourceKey] {
		resource.SetName(resource.GetName() + "[cycle detected]")
		return &resourceWithParents{
			resource: resource,
			kind:     kind,
			parents:  []*resourceWithParents{},
		}
	}
	visited[resourceKey] = true

	var parents []*resourceWithParents
	for _, ownerRef := range resource.GetOwnerReferences() {
		owner, err := o.getResource(ctx, ownerRef.Kind, ownerRef.Name, resource.GetNamespace())
		if err != nil {
			parents = append(parents, &resourceWithParents{
				resource: &metav1.ObjectMeta{
					Name:      ownerRef.Name + "[" + err.Error() + "]",
					Namespace: resource.GetNamespace(),
				},
				kind: ownerRef.Kind,
			})
		} else {
			parents = append(parents, o.buildResourceWithParents(ctx, owner, ownerRef.Kind, visited))
		}
	}

	return &resourceWithParents{
		resource: resource,
		kind:     kind,
		parents:  parents,
	}
}

func (o *Ownerser) getResource(ctx context.Context, kind, name, namespace string) (metav1.Object, error) {
	return o.K8sClient.GetResource(ctx, kind, name, namespace)
}

func (o *Ownerser) buildOwnershipLines(res *resourceWithParents) string {
	if res == nil {
		return ""
	}

	namespaceColor := color.New(color.FgCyan)
	kindColor := color.New(color.FgYellow)
	nameColor := color.New(color.FgGreen, color.Bold)
	prefixColor := color.New(color.FgWhite)

	type node struct {
		res    *resourceWithParents
		prefix string
		depth  int
	}
	stack := []node{{res: res, prefix: "", depth: 1}}
	var lines []string
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		indent := strings.Repeat("  ", n.depth-1)
		prefix := prefixColor.Sprint(n.prefix)
		namespace := namespaceColor.Sprint(n.res.resource.GetNamespace())
		kind := kindColor.Sprint(n.res.kind)
		name := nameColor.Sprint(n.res.resource.GetName())

		lines = append(lines, fmt.Sprintf("%s%s%s/%s/%s", indent, prefix, namespace, kind, name))

		for i := len(n.res.parents) - 1; i >= 0; i-- {
			var newPrefix string
			if i == len(n.res.parents)-1 {
				newPrefix = "↱"
			} else {
				newPrefix = "├"
			}
			stack = append(stack, node{res: n.res.parents[i], prefix: newPrefix, depth: n.depth + 1})
		}
	}

	slices.Reverse(lines)
	return strings.Join(lines, "\n")
}
