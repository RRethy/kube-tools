package shell

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"

	resolvertesting "github.com/RRethy/kubectl-x/pkg/resolver/testing"
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
			fakeResolver := resolvertesting.NewFakeResolver(nil, errors.New("no pod found"))
			fakeExec := &fakeexec.FakeExec{}

			sheller := &Sheller{
				IOStreams: genericclioptions.IOStreams{
					In:     &bytes.Buffer{},
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
				Context:   "test-context",
				Namespace: "test-namespace",
				Resolver:  fakeResolver,
				Exec:      fakeExec,
			}

			err := sheller.Shell(context.Background(), tt.target, "", "/bin/sh", false, "")

			if tt.expectError {
				assert.Error(t, err)
				return
			}
		})
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
		pod         *corev1.Pod
		resolverErr error
		expectError bool
		expectArgs  []string
	}{
		{
			name:    "basic shell execution",
			target:  "test-pod",
			command: "/bin/sh",
			debug:   false,
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
			expectArgs: []string{"exec", "-it", "test-pod", "--context", "test-context", "-n", "test-namespace", "--", "/bin/sh"},
		},
		{
			name:    "basic debug container",
			target:  "test-pod",
			command: "/bin/bash",
			debug:   true,
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
			expectArgs: []string{"debug", "-it", "test-pod", "--context", "test-context", "-n", "test-namespace", "--", "/bin/bash"},
		},
		{
			name:    "debug with custom image",
			target:  "test-pod",
			command: "/bin/sh",
			debug:   true,
			image:   "ubuntu:latest",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
			expectArgs: []string{"debug", "-it", "test-pod", "--context", "test-context", "-n", "test-namespace", "--image", "ubuntu:latest", "--", "/bin/sh"},
		},
		{
			name:      "debug with target container",
			target:    "test-pod",
			command:   "/bin/sh",
			debug:     true,
			container: "app-container",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
			expectArgs: []string{"debug", "-it", "test-pod", "--context", "test-context", "-n", "test-namespace", "--target", "app-container", "--", "/bin/sh"},
		},
		{
			name:      "exec with specific container",
			target:    "test-pod",
			command:   "/bin/sh",
			debug:     false,
			container: "app-container",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
			},
			expectArgs: []string{"exec", "-it", "test-pod", "--context", "test-context", "-n", "test-namespace", "-c", "app-container", "--", "/bin/sh"},
		},
		{
			name:        "shell with pod resolution error",
			target:      "nonexistent-pod",
			command:     "/bin/sh",
			debug:       false,
			resolverErr: errors.New("no pod found"),
			expectError: true,
		},
		{
			name:        "debug with pod resolution error",
			target:      "nonexistent-pod",
			command:     "/bin/sh",
			debug:       true,
			resolverErr: errors.New("no pod found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeResolver := resolvertesting.NewFakeResolver(tt.pod, tt.resolverErr)

			var actualArgs []string
			fakeExec := &fakeexec.FakeExec{
				CommandScript: []fakeexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						actualArgs = args
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
				Resolver:  fakeResolver,
				Exec:      fakeExec,
			}

			err := sheller.Shell(context.Background(), tt.target, tt.container, tt.command, tt.debug, tt.image)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.expectArgs != nil {
				assert.Equal(t, tt.expectArgs, actualArgs)
			}
		})
	}
}

func TestSheller_Shell_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fakeResolver := resolvertesting.NewFakeResolver(nil, context.Canceled)
	fakeExec := &fakeexec.FakeExec{}

	sheller := &Sheller{
		IOStreams: genericclioptions.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
		Context:   "test-context",
		Namespace: "test-namespace",
		Resolver:  fakeResolver,
		Exec:      fakeExec,
	}

	err := sheller.Shell(ctx, "test-pod", "", "/bin/sh", false, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestSheller_Shell_ExecFailure(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
	}

	fakeResolver := resolvertesting.NewFakeResolver(pod, nil)
	fakeExec := &fakeexec.FakeExec{
		CommandScript: []fakeexec.FakeCommandAction{
			func(cmd string, args ...string) kexec.Cmd {
				return &fakeexec.FakeCmd{
					RunScript: []fakeexec.FakeAction{
						func() ([]byte, []byte, error) {
							return nil, []byte("exec failed"), errors.New("exec error")
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
		Resolver:  fakeResolver,
		Exec:      fakeExec,
	}

	err := sheller.Shell(context.Background(), "test-pod", "", "/bin/sh", false, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec error")
}
