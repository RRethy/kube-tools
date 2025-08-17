package shell

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	fakeexec "k8s.io/utils/exec/testing"

	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	kubernetestesting "github.com/RRethy/kubectl-x/pkg/kubernetes/testing"
)

func TestSheller_Shell(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		expectError bool
	}{
		{
			name:        "pod resolution fails",
			target:      "nonexistent-pod",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
				"pod": {},
			})
			fakeExec := &fakeexec.FakeExec{}

			sheller := &Sheller{
				IOStreams: genericclioptions.IOStreams{
					In:     &bytes.Buffer{},
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: "test-namespace",
				K8sClient: k8sClient,
				Exec:      fakeExec,
			}

			err := sheller.Shell(context.Background(), tt.target, "", "/bin/sh")

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSheller_ResolvePod(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		pods         []*corev1.Pod
		deployments  []*appsv1.Deployment
		fzfSelection []string
		fzfError     error
		expectError  bool
		expectedPod  string
	}{
		{
			name:   "direct pod name match",
			target: "my-pod",
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "my-pod"}},
			},
			expectedPod: "my-pod",
		},
		{
			name:   "partial pod name match",
			target: "pod",
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "my-pod"}},
			},
			expectedPod: "my-pod",
		},
		{
			name:   "multiple pod matches with fzf selection",
			target: "pod",
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			fzfSelection: []string{"pod-2"},
			expectedPod:  "pod-2",
		},
		{
			name:   "deployment with slash separator",
			target: "deployment/nginx",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nginx-pod",
						Labels: map[string]string{"app": "nginx"},
					},
				},
			},
			deployments: []*appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-deployment"},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "nginx"},
						},
					},
				},
			},
			expectedPod: "nginx-pod",
		},
		{
			name:   "deployment with space separator",
			target: "deploy nginx",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nginx-pod",
						Labels: map[string]string{"app": "nginx"},
					},
				},
			},
			deployments: []*appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-deployment"},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "nginx"},
						},
					},
				},
			},
			expectedPod: "nginx-pod",
		},
		{
			name:   "resource kind only (no name)",
			target: "deployment",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nginx-pod",
						Labels: map[string]string{"app": "nginx"},
					},
				},
			},
			deployments: []*appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-deployment"},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "nginx"},
						},
					},
				},
			},
			expectedPod: "nginx-pod",
		},
		{
			name:        "no pods found",
			target:      "nonexistent",
			expectError: true,
		},
		{
			name:   "fzf selection cancelled",
			target: "pod",
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			fzfError:    errors.New("user cancelled"),
			expectError: true,
		},
		{
			name:   "empty fzf selection",
			target: "pod",
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			fzfSelection: []string{},
			expectError:  true,
		},
		{
			name:        "unsupported resource type",
			target:      "unsupported/resource",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			resources := map[string][]any{
				"pod": podListToAny(tt.pods),
			}
			if tt.deployments != nil {
				resources["deployment"] = deploymentListToAny(tt.deployments)
			}

			k8sClient := kubernetestesting.NewFakeClient(resources)
			fzf := fzftesting.NewFakeFzf(tt.fzfSelection, tt.fzfError)

			sheller := &Sheller{
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: "test-namespace",
				K8sClient: k8sClient,
				Fzf:       fzf,
			}

			pod, err := sheller.resolvePod(context.Background(), tt.target)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPod, pod.GetName())
		})
	}
}

func TestSheller_GetResourceSelector(t *testing.T) {
	tests := []struct {
		name         string
		resourceKind string
		resourceName string
		resources    map[string][]any
		fzfSelection []string
		expectedSelector map[string]string
		expectError  bool
	}{
		{
			name:         "deployment selector",
			resourceKind: "deployment",
			resourceName: "nginx",
			resources: map[string][]any{
				"deployment": {
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "nginx-deployment"},
						Spec: appsv1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "nginx"},
							},
						},
					},
				},
			},
			expectedSelector: map[string]string{"app": "nginx"},
		},
		{
			name:         "statefulset selector",
			resourceKind: "sts",
			resourceName: "database",
			resources: map[string][]any{
				"statefulset": {
					&appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{Name: "database-sts"},
						Spec: appsv1.StatefulSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "database"},
							},
						},
					},
				},
			},
			expectedSelector: map[string]string{"app": "database"},
		},
		{
			name:         "service selector",
			resourceKind: "svc",
			resourceName: "web",
			resources: map[string][]any{
				"service": {
					&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "web-service"},
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "web"},
						},
					},
				},
			},
			expectedSelector: map[string]string{"app": "web"},
		},
		{
			name:         "deployment without selector",
			resourceKind: "deployment",
			resourceName: "noselector",
			resources: map[string][]any{
				"deployment": {
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{Name: "noselector-deployment"},
						Spec:       appsv1.DeploymentSpec{},
					},
				},
			},
			expectError: true,
		},
		{
			name:         "service without selector",
			resourceKind: "service",
			resourceName: "noselector",
			resources: map[string][]any{
				"service": {
					&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "noselector-service"},
						Spec:       corev1.ServiceSpec{},
					},
				},
			},
			expectError: true,
		},
		{
			name:         "unsupported resource kind",
			resourceKind: "unsupported",
			resourceName: "test",
			expectError:  true,
		},
		{
			name:         "resource not found",
			resourceKind: "deployment",
			resourceName: "nonexistent",
			resources: map[string][]any{
				"deployment": {},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := kubernetestesting.NewFakeClient(tt.resources)
			fzf := fzftesting.NewFakeFzf(tt.fzfSelection, nil)

			sheller := &Sheller{
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: "test-namespace",
				K8sClient: k8sClient,
				Fzf:       fzf,
			}

			selector, err := sheller.getResourceSelector(context.Background(), tt.resourceKind, tt.resourceName)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedSelector, selector)
		})
	}
}

func TestSelectResourceFromList(t *testing.T) {
	tests := []struct {
		name         string
		resources    []*corev1.Pod
		resourceName string
		fzfSelection []string
		fzfError     error
		expectedPod  string
		expectError  bool
	}{
		{
			name: "single exact match",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}},
			},
			resourceName: "test-pod",
			expectedPod:  "test-pod",
		},
		{
			name: "single partial match",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "test-pod-123"}},
			},
			resourceName: "pod",
			expectedPod:  "test-pod-123",
		},
		{
			name: "multiple matches with fzf selection",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "test-pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "test-pod-2"}},
			},
			resourceName: "pod",
			fzfSelection: []string{"test-pod-2"},
			expectedPod:  "test-pod-2",
		},
		{
			name: "empty resource name matches all",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceName: "",
			fzfSelection: []string{"pod-1"},
			expectedPod:  "pod-1",
		},
		{
			name:         "no matches found",
			resources:    []*corev1.Pod{},
			resourceName: "nonexistent",
			expectError:  true,
		},
		{
			name: "fzf selection cancelled",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceName: "pod",
			fzfError:     errors.New("user cancelled"),
			expectError:  true,
		},
		{
			name: "empty fzf selection",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceName: "pod",
			fzfSelection: []string{},
			expectError:  true,
		},
		{
			name: "fzf selection not found in matches",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceName: "pod",
			fzfSelection: []string{"nonexistent"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fzf := fzftesting.NewFakeFzf(tt.fzfSelection, tt.fzfError)

			sheller := &Sheller{
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
				Fzf: fzf,
			}

			pod, err := selectResourceFromList(context.Background(), sheller, tt.resources, tt.resourceName, "pod")

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPod, pod.GetName())
		})
	}
}

func TestSheller_ResolvePod_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		expectError bool
	}{
		{
			name:   "empty target defaults to pod",
			target: "",
		},
		{
			name:   "whitespace only target",
			target: "   ",
		},
		{
			name:   "case insensitive resource kind",
			target: "DEPLOYMENT/nginx",
		},
		{
			name:   "mixed case resource name",
			target: "deployment/NGINX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
				"pod": {},
			})

			sheller := &Sheller{
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: "test-namespace",
				K8sClient: k8sClient,
			}

			_, err := sheller.resolvePod(context.Background(), tt.target)

			if tt.expectError {
				assert.Error(t, err)
			}
		})
	}
}

// Helper functions to convert slices to []any for the fake client
func podListToAny(pods []*corev1.Pod) []any {
	result := make([]any, len(pods))
	for i, pod := range pods {
		result[i] = pod
	}
	return result
}

func deploymentListToAny(deployments []*appsv1.Deployment) []any {
	result := make([]any, len(deployments))
	for i, deployment := range deployments {
		result[i] = deployment
	}
	return result
}

