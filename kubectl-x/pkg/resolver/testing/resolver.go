package testing

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	"github.com/RRethy/kubectl-x/pkg/resolver"
)

var _ resolver.Resolver = &FakeResolver{}

type FakeResolver struct {
	ReturnPod      *corev1.Pod
	ReturnResource metav1.Object
	ReturnError    error

	ResolvePodCalled      bool
	ResolveResourceCalled bool
	ResolveTargetCalled   bool
	LastResourceKind      string
	LastResourceName      string
	LastNamespace         string
	LastFzfConfig         fzf.Config
}

func NewFakeResolver(pod *corev1.Pod, err error) *FakeResolver {
	return &FakeResolver{
		ReturnPod:   pod,
		ReturnError: err,
	}
}

func (f *FakeResolver) ResolveTarget(ctx context.Context, target string) (string, string) {
	f.ResolveTargetCalled = true
	// Simple parsing for testing
	if target == "" {
		return "pod", ""
	}
	return "pod", target
}

func (f *FakeResolver) ResolvePod(ctx context.Context, resourceKind, resourceName, namespace string, fzfConfig fzf.Config) (*corev1.Pod, error) {
	f.ResolvePodCalled = true
	f.LastResourceKind = resourceKind
	f.LastResourceName = resourceName
	f.LastNamespace = namespace
	f.LastFzfConfig = fzfConfig
	return f.ReturnPod, f.ReturnError
}

func (f *FakeResolver) ResolveResource(ctx context.Context, resourceKind, resourceName, namespace string, fzfConfig fzf.Config) (metav1.Object, error) {
	f.ResolveResourceCalled = true
	f.LastResourceKind = resourceKind
	f.LastResourceName = resourceName
	f.LastNamespace = namespace
	f.LastFzfConfig = fzfConfig
	if f.ReturnResource != nil {
		return f.ReturnResource, f.ReturnError
	}
	// For backward compatibility, if ReturnPod is set, return it as a resource
	if f.ReturnPod != nil {
		return f.ReturnPod, f.ReturnError
	}
	return nil, f.ReturnError
}
