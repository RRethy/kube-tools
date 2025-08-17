// Package testing provides mock implementations for testing Kubernetes operations
package testing

import (
	"context"
	"errors"

	"github.com/RRethy/kubectl-x/pkg/kubernetes"
)

var _ kubernetes.Interface = &FakeClient{}

// FakeClient is a mock implementation of kubernetes.Interface for testing
type FakeClient struct {
	resources map[string][]any
}

// NewFakeClient creates a new fake Kubernetes client with the given resources
func NewFakeClient(resources map[string][]any) *FakeClient {
	return &FakeClient{resources}
}

func (fake *FakeClient) List(ctx context.Context, resourceType string) ([]any, error) {
	if resources, ok := fake.resources[resourceType]; ok {
		return resources, nil
	}
	return nil, errors.New("resource type not found")
}

func (fake *FakeClient) ListInNamespace(ctx context.Context, resourceType, namespace string) ([]any, error) {
	// For testing purposes, namespace filtering is simplified
	return fake.List(ctx, resourceType)
}
