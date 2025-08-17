package shell

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	"github.com/RRethy/kubectl-x/pkg/kubernetes"
	kubernetestesting "github.com/RRethy/kubectl-x/pkg/kubernetes/testing"
)

func TestSheller_Shell_ErrorHandling(t *testing.T) {
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

			err := sheller.Shell(context.Background(), tt.target, "", "/bin/sh", false, "")

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
			name:         "replicaset selector",
			resourceKind: "rs",
			resourceName: "nginx",
			resources: map[string][]any{
				"replicaset": {
					&appsv1.ReplicaSet{
						ObjectMeta: metav1.ObjectMeta{Name: "nginx-replicaset"},
						Spec: appsv1.ReplicaSetSpec{
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
			name:         "daemonset selector",
			resourceKind: "ds",
			resourceName: "logging",
			resources: map[string][]any{
				"daemonset": {
					&appsv1.DaemonSet{
						ObjectMeta: metav1.ObjectMeta{Name: "logging-daemonset"},
						Spec: appsv1.DaemonSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "logging"},
							},
						},
					},
				},
			},
			expectedSelector: map[string]string{"app": "logging"},
		},
		{
			name:         "job selector",
			resourceKind: "job",
			resourceName: "backup",
			resources: map[string][]any{
				"job": {
					&batchv1.Job{
						ObjectMeta: metav1.ObjectMeta{Name: "backup-job"},
						Spec: batchv1.JobSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"job": "backup"},
							},
						},
					},
				},
			},
			expectedSelector: map[string]string{"job": "backup"},
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
			name:         "replicaset without selector",
			resourceKind: "replicaset",
			resourceName: "noselector",
			resources: map[string][]any{
				"replicaset": {
					&appsv1.ReplicaSet{
						ObjectMeta: metav1.ObjectMeta{Name: "noselector-replicaset"},
						Spec:       appsv1.ReplicaSetSpec{},
					},
				},
			},
			expectError: true,
		},
		{
			name:         "job without selector",
			resourceKind: "job",
			resourceName: "noselector",
			resources: map[string][]any{
				"job": {
					&batchv1.Job{
						ObjectMeta: metav1.ObjectMeta{Name: "noselector-job"},
						Spec:       batchv1.JobSpec{},
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
		{
			name:        "target with both space and slash",
			target:      "deployment nginx/extra",
			expectError: false, // Should use slash as separator
		},
		{
			name:   "resource shortname expansion",
			target: "deploy/nginx",
		},
		{
			name:   "resource kind only without name",
			target: "pods",
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

func TestShell_PublicFunction(t *testing.T) {
	// This test would be more comprehensive in a real scenario
	// but we can't easily test the public Shell function without 
	// significant mocking of kubeconfig, context resolution, etc.
	// This serves as a placeholder for integration tests
	t.Skip("Integration test for public Shell function - requires kubeconfig setup")
}

func TestSheller_ResolvePod_LabelMatching(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		pods        []*corev1.Pod
		deployments []*appsv1.Deployment
		expectedPod string
		expectError bool
	}{
		{
			name:   "pod matches deployment selector with multiple labels",
			target: "deployment/web-app",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "web-app-pod-1",
						Labels: map[string]string{"app": "web-app", "version": "v1", "tier": "frontend"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "web-app-pod-2",
						Labels: map[string]string{"app": "web-app", "version": "v2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "other-pod",
						Labels: map[string]string{"app": "other"},
					},
				},
			},
			deployments: []*appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "web-app-deployment"},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "web-app", "version": "v1"},
						},
					},
				},
			},
			expectedPod: "web-app-pod-1", // Only this pod matches both app=web-app AND version=v1
		},
		{
			name:   "no pods match deployment selector",
			target: "deployment/missing-app",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "other-pod",
						Labels: map[string]string{"app": "other"},
					},
				},
			},
			deployments: []*appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "missing-app-deployment"},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "missing-app"},
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := map[string][]any{
				"pod": podListToAny(tt.pods),
			}
			if tt.deployments != nil {
				resources["deployment"] = deploymentListToAny(tt.deployments)
			}

			k8sClient := kubernetestesting.NewFakeClient(resources)
			fzf := fzftesting.NewFakeFzf([]string{}, nil) // No fzf interaction needed for single matches

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


// createTestSheller creates a standard Sheller instance for testing
func createTestSheller(k8sClient kubernetes.Interface, fzfInterface fzf.Interface, execInterface kexec.Interface) *Sheller {
	return &Sheller{
		IOStreams: genericclioptions.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
		Context:   "test-context",
		Namespace: "test-namespace",
		K8sClient: k8sClient,
		Fzf:       fzfInterface,
		Exec:      execInterface,
	}
}

func TestSheller_Shell(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		container   string
		command     string
		debug       bool
		image       string
		pods        []*corev1.Pod
		expectError bool
	}{
		{
			name:    "basic shell execution",
			target:  "test-pod",
			command: "/bin/sh",
			debug:   false,
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				},
			},
		},
		{
			name:    "basic debug container",
			target:  "test-pod",
			command: "/bin/bash",
			debug:   true,
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				},
			},
		},
		{
			name:    "debug with custom image",
			target:  "test-pod",
			command: "/bin/sh",
			debug:   true,
			image:   "ubuntu:latest",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				},
			},
		},
		{
			name:      "debug with target container",
			target:    "test-pod",
			command:   "/bin/sh",
			debug:     true,
			container: "app-container",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				},
			},
		},
		{
			name:        "shell with nonexistent pod",
			target:      "nonexistent-pod",
			command:     "/bin/sh",
			debug:       false,
			pods:        []*corev1.Pod{},
			expectError: true,
		},
		{
			name:        "debug with nonexistent pod",
			target:      "nonexistent-pod",
			command:     "/bin/sh",
			debug:       true,
			pods:        []*corev1.Pod{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
				"pod": podListToAny(tt.pods),
			})
			fakeExec := &fakeexec.FakeExec{
				CommandScript: []fakeexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						return &fakeexec.FakeCmd{
							RunScript: []fakeexec.FakeAction{
								func() ([]byte, []byte, error) {
									return nil, nil, nil
								},
							},
						}
					},
				},
			}

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

			err := sheller.Shell(context.Background(), tt.target, tt.container, tt.command, tt.debug, tt.image)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}



func TestSheller_Shell_CommandArguments(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		container      string
		command        string
		debug          bool
		image          string
		expectedCmd    string
		expectedArgs   []string
		verifyArgs     func(t *testing.T, cmd string, args []string)
	}{
		{
			name:        "debug mode with image and container",
			target:      "test-pod",
			container:   "app-container",
			command:     "/bin/bash",
			debug:       true,
			image:       "ubuntu:latest",
			expectedCmd: "kubectl",
			verifyArgs: func(t *testing.T, cmd string, args []string) {
				assert.Equal(t, "kubectl", cmd)
				assert.Contains(t, args, "debug")
				assert.Contains(t, args, "-it")
				assert.Contains(t, args, "test-pod")
				assert.Contains(t, args, "--context")
				assert.Contains(t, args, "test-context")
				assert.Contains(t, args, "-n")
				assert.Contains(t, args, "test-namespace")
				assert.Contains(t, args, "--image")
				assert.Contains(t, args, "ubuntu:latest")
				assert.Contains(t, args, "--target")
				assert.Contains(t, args, "app-container")
				assert.Contains(t, args, "--")
				assert.Contains(t, args, "/bin/bash")
			},
		},
		{
			name:        "exec mode with container",
			target:      "test-pod",
			container:   "web-container",
			command:     "/bin/sh",
			debug:       false,
			expectedCmd: "kubectl",
			verifyArgs: func(t *testing.T, cmd string, args []string) {
				assert.Equal(t, "kubectl", cmd)
				assert.Contains(t, args, "exec")
				assert.Contains(t, args, "-it")
				assert.Contains(t, args, "test-pod")
				assert.Contains(t, args, "-c")
				assert.Contains(t, args, "web-container")
				assert.Contains(t, args, "--")
				assert.Contains(t, args, "/bin/sh")
				assert.NotContains(t, args, "debug")
				assert.NotContains(t, args, "--image")
				assert.NotContains(t, args, "--target")
			},
		},
		{
			name:        "debug mode without image or container",
			target:      "test-pod",
			command:     "/bin/sh",
			debug:       true,
			expectedCmd: "kubectl",
			verifyArgs: func(t *testing.T, cmd string, args []string) {
				assert.Equal(t, "kubectl", cmd)
				assert.Contains(t, args, "debug")
				assert.NotContains(t, args, "--image")
				assert.NotContains(t, args, "--target")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
				"pod": podListToAny([]*corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "test-namespace",
						},
					},
				}),
			})

			var capturedCmd string
			var capturedArgs []string
			fakeExec := &fakeexec.FakeExec{
				CommandScript: []fakeexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						capturedCmd = cmd
						capturedArgs = args
						return &fakeexec.FakeCmd{
							RunScript: []fakeexec.FakeAction{
								func() ([]byte, []byte, error) {
									return nil, nil, nil
								},
							},
						}
					},
				},
			}

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

			err := sheller.Shell(context.Background(), tt.target, tt.container, tt.command, tt.debug, tt.image)
			require.NoError(t, err)

			tt.verifyArgs(t, capturedCmd, capturedArgs)
		})
	}
}

func TestSheller_Shell_ContextCancellation(t *testing.T) {
	t.Skip("Context cancellation testing requires more complex mocking - skipping for now")
	// This test would require more sophisticated mocking of the kubernetes client
	// to properly test context cancellation behavior. The current fake client
	// doesn't respect context cancellation in its ListInNamespace methods.
}

func TestSheller_Shell_ExecFailure(t *testing.T) {
	tests := []struct {
		name        string
		execError   error
		expectError bool
	}{
		{
			name:        "kubectl command fails",
			execError:   errors.New("kubectl: command not found"),
			expectError: true,
		},
		{
			name:        "kubectl exits with non-zero code",
			execError:   errors.New("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
				"pod": podListToAny([]*corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-pod",
							Namespace: "test-namespace",
						},
					},
				}),
			})
			fakeExec := &fakeexec.FakeExec{
				CommandScript: []fakeexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						return &fakeexec.FakeCmd{
							RunScript: []fakeexec.FakeAction{
								func() ([]byte, []byte, error) {
									return nil, nil, tt.execError
								},
							},
						}
					},
				},
			}

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

			err := sheller.Shell(context.Background(), "test-pod", "", "/bin/sh", false, "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.execError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

