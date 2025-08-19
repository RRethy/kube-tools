package peek

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	kexec "k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"

	"github.com/RRethy/kubectl-x/pkg/fzf"
	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	k8stesting "github.com/RRethy/kubectl-x/pkg/kubernetes/testing"
	resolvertesting "github.com/RRethy/kubectl-x/pkg/resolver/testing"
)

func TestPeeker_Peek(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		action         string
		target         string
		setupMocks     func() (*resolvertesting.FakeResolver, *bytes.Buffer)
		expectedOutput string
		expectedError  string
	}{
		{
			name:   "logs action with pod",
			action: "logs",
			target: "pod/my-pod",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: pod,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl logs pod/my-pod -n test-namespace\n",
		},
		{
			name:   "logs action with deployment",
			action: "logs",
			target: "deployment/my-deployment",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				deployment := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-deployment-abc123",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: deployment,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl logs deployment/my-deployment-abc123 -n test-namespace\n",
		},
		{
			name:   "describe action with pod",
			action: "describe",
			target: "pod/my-pod",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: pod,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl describe pod my-pod -n test-namespace\n",
		},
		{
			name:   "describe action with service",
			action: "describe",
			target: "service/my-service",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-service",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: service,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl describe service my-service -n test-namespace\n",
		},
		{
			name:   "invalid action",
			action: "invalid",
			target: "pod/my-pod",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				resolver := &resolvertesting.FakeResolver{}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedError: "action must be 'logs' or 'describe', got: \"invalid\"",
		},
		{
			name:   "empty action",
			action: "",
			target: "pod/my-pod",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				resolver := &resolvertesting.FakeResolver{}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedError: "action must be 'logs' or 'describe', got: \"\"",
		},
		{
			name:   "logs with resource resolution error",
			action: "logs",
			target: "pod/non-existent",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				resolver := &resolvertesting.FakeResolver{
					ReturnError: fmt.Errorf("resource not found"),
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedError: "resolving pod for logs: resource not found",
		},
		{
			name:   "describe with resource resolution error",
			action: "describe",
			target: "deployment/non-existent",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				resolver := &resolvertesting.FakeResolver{
					ReturnError: fmt.Errorf("deployment not found"),
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedError: "resolving resource for describe: deployment not found",
		},
		{
			name:   "logs with shortname",
			action: "logs",
			target: "deploy/nginx",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				deployment := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-deployment",
						Namespace: "production",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: deployment,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl logs deployment/nginx-deployment -n test-namespace\n",
		},
		{
			name:   "describe with shortname",
			action: "describe",
			target: "svc/web",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "web-service",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: service,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl describe service web-service -n test-namespace\n",
		},
		{
			name:   "logs with only resource kind",
			action: "logs",
			target: "pod",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "selected-pod",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: pod,
				}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedOutput: "kubectl logs pod/selected-pod -n test-namespace\n",
		},
		{
			name:   "uppercase action should fail",
			action: "LOGS",
			target: "pod/my-pod",
			setupMocks: func() (*resolvertesting.FakeResolver, *bytes.Buffer) {
				resolver := &resolvertesting.FakeResolver{}
				buf := &bytes.Buffer{}
				return resolver, buf
			},
			expectedError: "action must be 'logs' or 'describe', got: \"LOGS\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, outBuf := tt.setupMocks()

			peeker := &Peeker{
				IOStreams: genericiooptions.IOStreams{
					Out:    outBuf,
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: "test-namespace",
				Resolver:  resolver,
			}

			err := peeker.Peek(ctx, tt.action, tt.target)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, outBuf.String())
			}

			// Verify resolver was called with correct parameters
			if tt.action == "logs" || tt.action == "describe" {
				if tt.expectedError == "" || !strings.Contains(tt.expectedError, "action must be") {
					assert.True(t, resolver.ResolveTargetCalled)
					assert.True(t, resolver.ResolveResourceCalled)
					assert.Equal(t, "test-namespace", resolver.LastNamespace)
				}
			}
		})
	}
}

func TestPeeker_PeekLogs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		resourceKind   string
		resourceName   string
		namespace      string
		setupResolver  func() *resolvertesting.FakeResolver
		expectedOutput string
		expectedError  string
		expectedConfig fzf.Config
	}{
		{
			name:         "successful pod logs",
			resourceKind: "pod",
			resourceName: "my-pod",
			namespace:    "default",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-pod",
							Namespace: "default",
						},
					},
				}
			},
			expectedOutput: "kubectl logs pod/my-pod -n default\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl logs pod/{1} -n default",
				Height:  "100%",
			},
		},
		{
			name:         "successful deployment logs",
			resourceKind: "deployment",
			resourceName: "my-deployment",
			namespace:    "production",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-deployment-abc123",
							Namespace: "production",
						},
					},
				}
			},
			expectedOutput: "kubectl logs deployment/my-deployment-abc123 -n production\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl logs deployment/{1} -n production",
				Height:  "100%",
			},
		},
		{
			name:         "resolution error",
			resourceKind: "pod",
			resourceName: "non-existent",
			namespace:    "default",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnError: fmt.Errorf("pod not found"),
				}
			},
			expectedError: "resolving pod for logs: pod not found",
		},
		{
			name:         "empty resource name",
			resourceKind: "pod",
			resourceName: "",
			namespace:    "kube-system",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "selected-pod",
							Namespace: "kube-system",
						},
					},
				}
			},
			expectedOutput: "kubectl logs pod/selected-pod -n kube-system\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl logs pod/{1} -n kube-system",
				Height:  "100%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := tt.setupResolver()
			outBuf := &bytes.Buffer{}

			peeker := &Peeker{
				IOStreams: genericiooptions.IOStreams{
					Out:    outBuf,
					ErrOut: &bytes.Buffer{},
				},
				Namespace: tt.namespace,
				Resolver:  resolver,
			}

			err := peeker.peekLogs(ctx, tt.resourceKind, tt.resourceName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, outBuf.String())

				// Verify the fzf config passed to resolver
				assert.True(t, resolver.ResolveResourceCalled)
				assert.Equal(t, tt.resourceKind, resolver.LastResourceKind)
				assert.Equal(t, tt.resourceName, resolver.LastResourceName)
				assert.Equal(t, tt.namespace, resolver.LastNamespace)
				assert.Equal(t, tt.expectedConfig, resolver.LastFzfConfig)
			}
		})
	}
}

func TestPeeker_PeekDescribe(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		resourceKind   string
		resourceName   string
		namespace      string
		setupResolver  func() *resolvertesting.FakeResolver
		expectedOutput string
		expectedError  string
		expectedConfig fzf.Config
	}{
		{
			name:         "successful pod describe",
			resourceKind: "pod",
			resourceName: "my-pod",
			namespace:    "default",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-pod",
							Namespace: "default",
						},
					},
				}
			},
			expectedOutput: "kubectl describe pod my-pod -n default\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl describe pod/{1} -n default",
				Height:  "100%",
			},
		},
		{
			name:         "successful service describe",
			resourceKind: "service",
			resourceName: "my-service",
			namespace:    "production",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-service",
							Namespace: "production",
						},
					},
				}
			},
			expectedOutput: "kubectl describe service my-service -n production\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl describe service/{1} -n production",
				Height:  "100%",
			},
		},
		{
			name:         "successful deployment describe",
			resourceKind: "deployment",
			resourceName: "nginx",
			namespace:    "web",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "nginx-deployment",
							Namespace: "web",
						},
					},
				}
			},
			expectedOutput: "kubectl describe deployment nginx-deployment -n web\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl describe deployment/{1} -n web",
				Height:  "100%",
			},
		},
		{
			name:         "resolution error",
			resourceKind: "service",
			resourceName: "non-existent",
			namespace:    "default",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnError: fmt.Errorf("service not found"),
				}
			},
			expectedError: "resolving resource for describe: service not found",
		},
		{
			name:         "empty resource name triggers selection",
			resourceKind: "configmap",
			resourceName: "",
			namespace:    "kube-system",
			setupResolver: func() *resolvertesting.FakeResolver {
				return &resolvertesting.FakeResolver{
					ReturnResource: &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "selected-configmap",
							Namespace: "kube-system",
						},
					},
				}
			},
			expectedOutput: "kubectl describe configmap selected-configmap -n kube-system\n",
			expectedConfig: fzf.Config{
				Sorted:  true,
				Preview: "kubectl describe configmap/{1} -n kube-system",
				Height:  "100%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := tt.setupResolver()
			outBuf := &bytes.Buffer{}

			peeker := &Peeker{
				IOStreams: genericiooptions.IOStreams{
					Out:    outBuf,
					ErrOut: &bytes.Buffer{},
				},
				Namespace: tt.namespace,
				Resolver:  resolver,
			}

			err := peeker.peekDescribe(ctx, tt.resourceKind, tt.resourceName)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, outBuf.String())

				// Verify the fzf config passed to resolver
				assert.True(t, resolver.ResolveResourceCalled)
				assert.Equal(t, tt.resourceKind, resolver.LastResourceKind)
				assert.Equal(t, tt.resourceName, resolver.LastResourceName)
				assert.Equal(t, tt.namespace, resolver.LastNamespace)
				assert.Equal(t, tt.expectedConfig, resolver.LastFzfConfig)
			}
		})
	}
}

func TestPeeker_Integration(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		action        string
		target        string
		namespace     string
		setupMocks    func() (*resolvertesting.FakeResolver, kexec.Interface, *fzftesting.FakeFzf, *k8stesting.FakeClient)
		expectedCalls func(t *testing.T, resolver *resolvertesting.FakeResolver)
	}{
		{
			name:      "full logs flow with pod selection",
			action:    "logs",
			target:    "pod",
			namespace: "default",
			setupMocks: func() (*resolvertesting.FakeResolver, kexec.Interface, *fzftesting.FakeFzf, *k8stesting.FakeClient) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "selected-pod",
						Namespace: "default",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: pod,
				}
				exec := &testingexec.FakeExec{}
				fzf := fzftesting.NewFakeFzf([]string{"selected-pod"}, nil)
				client := k8stesting.NewFakeClient(map[string][]any{
					"pod": {pod},
				})
				return resolver, exec, fzf, client
			},
			expectedCalls: func(t *testing.T, resolver *resolvertesting.FakeResolver) {
				assert.True(t, resolver.ResolveTargetCalled)
				assert.True(t, resolver.ResolveResourceCalled)
				// Note: The FakeResolver stores the values from ResolveResource call
				// not from ResolveTarget, so these are what were passed to ResolveResource
				assert.Equal(t, "pod", resolver.LastResourceKind)
				assert.Equal(t, "", resolver.LastResourceName)
				assert.Equal(t, "default", resolver.LastNamespace)
			},
		},
		{
			name:      "full describe flow with deployment",
			action:    "describe",
			target:    "deployment/nginx",
			namespace: "production",
			setupMocks: func() (*resolvertesting.FakeResolver, kexec.Interface, *fzftesting.FakeFzf, *k8stesting.FakeClient) {
				deployment := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-deployment",
						Namespace: "production",
					},
				}
				resolver := &resolvertesting.FakeResolver{
					ReturnResource: deployment,
				}
				exec := &testingexec.FakeExec{}
				fzf := fzftesting.NewFakeFzf([]string{"nginx-deployment"}, nil)
				client := k8stesting.NewFakeClient(map[string][]any{
					"deployment": {deployment},
				})
				return resolver, exec, fzf, client
			},
			expectedCalls: func(t *testing.T, resolver *resolvertesting.FakeResolver) {
				assert.True(t, resolver.ResolveTargetCalled)
				assert.True(t, resolver.ResolveResourceCalled)
				// The FakeResolver stores values from ResolveResource call
				assert.Equal(t, "deployment", resolver.LastResourceKind)
				assert.Equal(t, "nginx", resolver.LastResourceName)
				assert.Equal(t, "production", resolver.LastNamespace)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, exec, fzf, client := tt.setupMocks()
			outBuf := &bytes.Buffer{}

			peeker := &Peeker{
				IOStreams: genericiooptions.IOStreams{
					Out:    outBuf,
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: tt.namespace,
				K8sClient: client,
				Resolver:  resolver,
				Fzf:       fzf,
				Exec:      exec,
			}

			err := peeker.Peek(ctx, tt.action, tt.target)
			require.NoError(t, err)

			tt.expectedCalls(t, resolver)
		})
	}
}

func TestPeeker_OutputFormatting(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		action         string
		resourceKind   string
		resourceName   string
		namespace      string
		expectedOutput string
	}{
		{
			name:           "logs with standard namespace",
			action:         "logs",
			resourceKind:   "pod",
			resourceName:   "my-pod",
			namespace:      "default",
			expectedOutput: "kubectl logs pod/my-pod -n default\n",
		},
		{
			name:           "logs with custom namespace",
			action:         "logs",
			resourceKind:   "deployment",
			resourceName:   "app",
			namespace:      "production",
			expectedOutput: "kubectl logs deployment/app -n production\n",
		},
		{
			name:           "describe with hyphenated namespace",
			action:         "describe",
			resourceKind:   "service",
			resourceName:   "api",
			namespace:      "kube-system",
			expectedOutput: "kubectl describe service api -n kube-system\n",
		},
		{
			name:           "describe with long resource name",
			action:         "describe",
			resourceKind:   "pod",
			resourceName:   "very-long-pod-name-with-many-segments",
			namespace:      "test",
			expectedOutput: "kubectl describe pod very-long-pod-name-with-many-segments -n test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.resourceName,
					Namespace: tt.namespace,
				},
			}
			resolver := &resolvertesting.FakeResolver{
				ReturnResource: resource,
			}
			outBuf := &bytes.Buffer{}

			peeker := &Peeker{
				IOStreams: genericiooptions.IOStreams{
					Out:    outBuf,
					ErrOut: &bytes.Buffer{},
				},
				Namespace: tt.namespace,
				Resolver:  resolver,
			}

			if tt.action == "logs" {
				err := peeker.peekLogs(ctx, tt.resourceKind, tt.resourceName)
				require.NoError(t, err)
			} else {
				err := peeker.peekDescribe(ctx, tt.resourceKind, tt.resourceName)
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}
