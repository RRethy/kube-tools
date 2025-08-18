package testing

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/RRethy/kubectl-x/pkg/resolver"
)

type FakeResolver struct {
	Pod *corev1.Pod
	Err error
}

func NewFakeResolver(pod *corev1.Pod, err error) resolver.Interface {
	return &FakeResolver{
		Pod: pod,
		Err: err,
	}
}

func (f *FakeResolver) ResolvePod(ctx context.Context, target string, namespace string) (*corev1.Pod, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	if f.Pod == nil {
		return nil, fmt.Errorf("no pod found matching '%s'", target)
	}
	return f.Pod, nil
}
