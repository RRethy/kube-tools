package resolver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	k8stesting "github.com/RRethy/kubectl-x/pkg/kubernetes/testing"
)

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		expectedKind string
		expectedName string
	}{
		{
			name:         "pod with slash separator",
			target:       "pod/my-pod",
			expectedKind: "pod",
			expectedName: "my-pod",
		},
		{
			name:         "deployment with slash separator",
			target:       "deployment/my-deployment",
			expectedKind: "deployment",
			expectedName: "my-deployment",
		},
		{
			name:         "pod with space separator",
			target:       "pod my-pod",
			expectedKind: "pod",
			expectedName: "my-pod",
		},
		{
			name:         "shortname expansion with slash",
			target:       "deploy/my-deployment",
			expectedKind: "deployment",
			expectedName: "my-deployment",
		},
		{
			name:         "shortname expansion with space",
			target:       "svc my-service",
			expectedKind: "service",
			expectedName: "my-service",
		},
		{
			name:         "only resource kind",
			target:       "pod",
			expectedKind: "pod",
			expectedName: "",
		},
		{
			name:         "only shortname",
			target:       "deploy",
			expectedKind: "deployment",
			expectedName: "",
		},
		{
			name:         "empty target",
			target:       "",
			expectedKind: "",
			expectedName: "",
		},
		{
			name:         "target with leading/trailing spaces",
			target:       "  pod/my-pod  ",
			expectedKind: "pod",
			expectedName: "my-pod",
		},
		{
			name:         "resource name with slash",
			target:       "pod/my-pod/with-slash",
			expectedKind: "pod",
			expectedName: "my-pod/with-slash",
		},
		{
			name:         "statefulset shortname",
			target:       "sts/my-statefulset",
			expectedKind: "statefulset",
			expectedName: "my-statefulset",
		},
		{
			name:         "daemonset shortname",
			target:       "ds/my-daemonset",
			expectedKind: "daemonset",
			expectedName: "my-daemonset",
		},
		{
			name:         "replicaset shortname",
			target:       "rs/my-replicaset",
			expectedKind: "replicaset",
			expectedName: "my-replicaset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolver{}
			kind, name := r.ResolveTarget(context.Background(), tt.target)
			assert.Equal(t, tt.expectedKind, kind)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

func TestResolvePod(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		resourceKind  string
		resourceName  string
		namespace     string
		setupMocks    func() (*k8stesting.FakeClient, *fzftesting.FakeFzf)
		expectedPod   *corev1.Pod
		expectedError string
	}{
		{
			name:         "direct pod resolution",
			resourceKind: "pod",
			resourceName: "my-pod",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-pod"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pod",
					Namespace: "default",
				},
			},
		},
		{
			name:         "deployment to pod resolution with single pod",
			resourceKind: "deployment",
			resourceName: "my-deployment",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-app"},
						},
					},
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment-abc123",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-app"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {deployment},
					"pod":        {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-deployment"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-deployment-abc123",
					Namespace: "default",
					Labels:    map[string]string{"app": "my-app"},
				},
			},
		},
		{
			name:         "deployment to pod resolution with multiple pods",
			resourceKind: "deployment",
			resourceName: "my-deployment",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-app"},
						},
					},
				}
				pod1 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment-abc123",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-app"},
					},
				}
				pod2 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment-def456",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-app"},
					},
				}
				pod3 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "other-pod",
						Namespace: "default",
						Labels:    map[string]string{"app": "other-app"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {deployment},
					"pod":        {pod1, pod2, pod3},
				})
				// Need to return deployment first, then the pod
				fzf := fzftesting.NewFakeFzfWithMultipleCalls([]fzftesting.CallResult{
					{Items: []string{"my-deployment"}, Error: nil},
					{Items: []string{"my-deployment-def456"}, Error: nil},
				})
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-deployment-def456",
					Namespace: "default",
					Labels:    map[string]string{"app": "my-app"},
				},
			},
		},
		{
			name:         "statefulset to pod resolution",
			resourceKind: "statefulset",
			resourceName: "my-statefulset",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				statefulset := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-statefulset",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-statefulset"},
						},
					},
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-statefulset-0",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-statefulset"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"statefulset": {statefulset},
					"pod":         {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-statefulset"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-statefulset-0",
					Namespace: "default",
					Labels:    map[string]string{"app": "my-statefulset"},
				},
			},
		},
		{
			name:         "daemonset to pod resolution",
			resourceKind: "daemonset",
			resourceName: "my-daemonset",
			namespace:    "kube-system",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				daemonset := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-daemonset",
						Namespace: "kube-system",
					},
					Spec: appsv1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-daemonset"},
						},
					},
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-daemonset-node1",
						Namespace: "kube-system",
						Labels:    map[string]string{"app": "my-daemonset"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"daemonset": {daemonset},
					"pod":       {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-daemonset"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-daemonset-node1",
					Namespace: "kube-system",
					Labels:    map[string]string{"app": "my-daemonset"},
				},
			},
		},
		{
			name:         "replicaset to pod resolution",
			resourceKind: "replicaset",
			resourceName: "my-replicaset",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				replicaset := &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-replicaset",
						Namespace: "default",
					},
					Spec: appsv1.ReplicaSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-replicaset"},
						},
					},
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-replicaset-xyz",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-replicaset"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"replicaset": {replicaset},
					"pod":        {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-replicaset"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-replicaset-xyz",
					Namespace: "default",
					Labels:    map[string]string{"app": "my-replicaset"},
				},
			},
		},
		{
			name:         "job to pod resolution",
			resourceKind: "job",
			resourceName: "my-job",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-job",
						Namespace: "default",
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"job": "my-job"},
						},
					},
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-job-abc",
						Namespace: "default",
						Labels:    map[string]string{"job": "my-job"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"job": {job},
					"pod": {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-job"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-job-abc",
					Namespace: "default",
					Labels:    map[string]string{"job": "my-job"},
				},
			},
		},
		{
			name:         "service to pod resolution",
			resourceKind: "service",
			resourceName: "my-service",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-service",
						Namespace: "default",
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{"app": "my-service"},
					},
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-service-pod",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-service"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"service": {service},
					"pod":     {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-service"}, nil)
				return client, fzf
			},
			expectedPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-service-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "my-service"},
				},
			},
		},
		{
			name:         "no pods found for deployment",
			resourceKind: "deployment",
			resourceName: "my-deployment",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-app"},
						},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {deployment},
					"pod":        {},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-deployment"}, nil)
				return client, fzf
			},
			expectedError: "no pods found for deployment/my-deployment",
		},
		{
			name:         "unsupported resource type",
			resourceKind: "configmap",
			resourceName: "my-configmap",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				client := k8stesting.NewFakeClient(map[string][]any{})
				fzf := &fzftesting.FakeFzf{}
				return client, fzf
			},
			expectedError: "resource type configmap cannot resolve to pods",
		},
		{
			name:         "resource not found",
			resourceKind: "deployment",
			resourceName: "non-existent",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {},
				})
				fzf := &fzftesting.FakeFzf{}
				return client, fzf
			},
			expectedError: "no deployment found",
		},
		{
			name:         "fzf selection cancelled",
			resourceKind: "deployment",
			resourceName: "my-deployment",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "my-app"},
						},
					},
				}
				pod1 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment-abc123",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-app"},
					},
				}
				pod2 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment-def456",
						Namespace: "default",
						Labels:    map[string]string{"app": "my-app"},
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {deployment},
					"pod":        {pod1, pod2},
				})
				fzf := fzftesting.NewFakeFzf(nil, fmt.Errorf("user cancelled"))
				return client, fzf
			},
			expectedError: "selecting deployment: user cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fzfMock := tt.setupMocks()
			r := New(client, fzfMock)

			pod, err := r.ResolvePod(ctx, tt.resourceKind, tt.resourceName, tt.namespace, fzf.Config{})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, pod)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPod.Name, pod.Name)
				assert.Equal(t, tt.expectedPod.Namespace, pod.Namespace)
				assert.Equal(t, tt.expectedPod.Labels, pod.Labels)
			}
		})
	}
}

func TestResolveResource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		resourceKind  string
		resourceName  string
		namespace     string
		setupMocks    func() (*k8stesting.FakeClient, *fzftesting.FakeFzf)
		expectedName  string
		expectedError string
	}{
		{
			name:         "exact match single resource",
			resourceKind: "pod",
			resourceName: "my-pod",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-pod"}, nil)
				return client, fzf
			},
			expectedName: "my-pod",
		},
		{
			name:         "partial match single resource",
			resourceKind: "pod",
			resourceName: "my",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-pod"}, nil)
				return client, fzf
			},
			expectedName: "my-pod",
		},
		{
			name:         "multiple matches with selection",
			resourceKind: "pod",
			resourceName: "my",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod1 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod-1",
						Namespace: "default",
					},
				}
				pod2 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod-2",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod1, pod2},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-pod-2"}, nil)
				return client, fzf
			},
			expectedName: "my-pod-2",
		},
		{
			name:         "no name provided lists all",
			resourceKind: "pod",
			resourceName: "",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod1 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
				}
				pod2 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default",
					},
				}
				pod3 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-3",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod1, pod2, pod3},
				})
				fzf := fzftesting.NewFakeFzf([]string{"pod-2"}, nil)
				return client, fzf
			},
			expectedName: "pod-2",
		},
		{
			name:         "no matching resources",
			resourceKind: "pod",
			resourceName: "non-existent",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod},
				})
				fzf := fzftesting.NewFakeFzf([]string{"my-pod"}, nil)
				return client, fzf
			},
			expectedError: "no pod found matching 'non-existent'",
		},
		{
			name:         "fzf selection error",
			resourceKind: "pod",
			resourceName: "my",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				pod1 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod-1",
						Namespace: "default",
					},
				}
				pod2 := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod-2",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod1, pod2},
				})
				fzf := fzftesting.NewFakeFzf(nil, fmt.Errorf("user cancelled"))
				return client, fzf
			},
			expectedError: "user cancelled",
		},
		{
			name:         "deployment resource",
			resourceKind: "deployment",
			resourceName: "nginx",
			namespace:    "production",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-deployment",
						Namespace: "production",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {deployment},
				})
				fzf := fzftesting.NewFakeFzf([]string{"nginx-deployment"}, nil)
				return client, fzf
			},
			expectedName: "nginx-deployment",
		},
		{
			name:         "service resource",
			resourceKind: "service",
			resourceName: "web",
			namespace:    "default",
			setupMocks: func() (*k8stesting.FakeClient, *fzftesting.FakeFzf) {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "web-service",
						Namespace: "default",
					},
				}
				client := k8stesting.NewFakeClient(map[string][]any{
					"service": {service},
				})
				fzf := fzftesting.NewFakeFzf([]string{"web-service"}, nil)
				return client, fzf
			},
			expectedName: "web-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fzfMock := tt.setupMocks()
			r := New(client, fzfMock)

			resource, err := r.ResolveResource(ctx, tt.resourceKind, tt.resourceName, tt.namespace, fzf.Config{})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resource)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, resource.GetName())
			}
		})
	}
}

func TestSelectResource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		resources     []*corev1.Pod
		resourceKind  string
		fzfConfig     fzf.Config
		setupFzf      func() *fzftesting.FakeFzf
		expectedName  string
		expectedError string
	}{
		{
			name: "single resource returns directly",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
			},
			resourceKind: "pod",
			setupFzf: func() *fzftesting.FakeFzf {
				return fzftesting.NewFakeFzf([]string{"pod-1"}, nil)
			},
			expectedName: "pod-1",
		},
		{
			name: "multiple resources with selection",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-3"}},
			},
			resourceKind: "pod",
			setupFzf: func() *fzftesting.FakeFzf {
				return fzftesting.NewFakeFzf([]string{"pod-2"}, nil)
			},
			expectedName: "pod-2",
		},
		{
			name:         "empty resources list",
			resources:    []*corev1.Pod{},
			resourceKind: "pod",
			setupFzf: func() *fzftesting.FakeFzf {
				return &fzftesting.FakeFzf{}
			},
			expectedError: "no pod found",
		},
		{
			name: "fzf returns empty selection",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceKind: "pod",
			setupFzf: func() *fzftesting.FakeFzf {
				return fzftesting.NewFakeFzf([]string{}, nil)
			},
			expectedError: "no pod selected",
		},
		{
			name: "fzf error",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceKind: "pod",
			setupFzf: func() *fzftesting.FakeFzf {
				return fzftesting.NewFakeFzf(nil, fmt.Errorf("fzf cancelled"))
			},
			expectedError: "selecting pod: fzf cancelled",
		},
		{
			name: "selected resource not in list",
			resources: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "pod-2"}},
			},
			resourceKind: "pod",
			setupFzf: func() *fzftesting.FakeFzf {
				return fzftesting.NewFakeFzf([]string{"pod-99"}, nil)
			},
			expectedError: "selected pod not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fzfMock := tt.setupFzf()

			result, err := selectResource(ctx, fzfMock, tt.resources, tt.resourceKind, tt.fzfConfig)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, result.GetName())
			}
		})
	}
}

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		expectedKind string
		expectedName string
	}{
		{
			name:         "empty string",
			target:       "",
			expectedKind: "",
			expectedName: "",
		},
		{
			name:         "only spaces",
			target:       "   ",
			expectedKind: "",
			expectedName: "",
		},
		{
			name:         "kind only",
			target:       "pod",
			expectedKind: "pod",
			expectedName: "",
		},
		{
			name:         "kind with slash and name",
			target:       "pod/my-pod",
			expectedKind: "pod",
			expectedName: "my-pod",
		},
		{
			name:         "kind with space and name",
			target:       "pod my-pod",
			expectedKind: "pod",
			expectedName: "my-pod",
		},
		{
			name:         "shortname with slash",
			target:       "deploy/nginx",
			expectedKind: "deployment",
			expectedName: "nginx",
		},
		{
			name:         "multiple slashes in name",
			target:       "pod/my-pod/with/slashes",
			expectedKind: "pod",
			expectedName: "my-pod/with/slashes",
		},
		{
			name:         "multiple spaces in name",
			target:       "pod my pod with spaces",
			expectedKind: "pod",
			expectedName: "my pod with spaces",
		},
		{
			name:         "leading and trailing spaces",
			target:       "  pod/my-pod  ",
			expectedKind: "pod",
			expectedName: "my-pod",
		},
		{
			name:         "only slash",
			target:       "/",
			expectedKind: "",
			expectedName: "",
		},
		{
			name:         "slash at the end",
			target:       "pod/",
			expectedKind: "pod",
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolver{}
			kind, name := r.parseTarget(tt.target)
			assert.Equal(t, tt.expectedKind, kind, "kind mismatch")
			assert.Equal(t, tt.expectedName, name, "name mismatch")
		})
	}
}
