package ns

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd/api"

	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	historytesting "github.com/RRethy/kubectl-x/pkg/history/testing"
	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
	kubernetestesting "github.com/RRethy/kubectl-x/pkg/kubernetes/testing"
)

func TestNewNser(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	assert.Equal(t, kubeConfig, nser.KubeConfig)
	assert.Equal(t, ioStreams, nser.IoStreams)
	assert.Equal(t, k8sClient, nser.K8sClient)
	assert.Equal(t, fzf, nser.Fzf)
	assert.Equal(t, history, nser.History)
}

func TestNser_Ns_SelectFromHistory(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}
	
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{
		Data: map[string][]string{
			"namespace": {"default", "kube-system"},  // Index 1 will return "kube-system"
		},
	}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "-", false)
	
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Switched to namespace \"kube-system\"")
	
	currentNs, _ := kubeConfig.GetCurrentNamespace()
	assert.Equal(t, "kube-system", currentNs)
}

func TestNser_Ns_SelectWithFzf(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}
	
	// Mock k8s client that returns namespaces
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
		},
	})
	
	fzf := fzftesting.NewFakeFzf([]string{"test-ns"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Switched to namespace \"test-ns\"")
	
	currentNs, _ := kubeConfig.GetCurrentNamespace()
	assert.Equal(t, "test-ns", currentNs)
}

func TestNser_Ns_ListNamespacesError(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	
	// K8s client returns error when listing namespaces
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listing namespaces")
}

func TestNser_Ns_FzfError(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		},
	})
	fzf := fzftesting.NewFakeFzf(nil, errors.New("fzf failed"))
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selecting namespace")
}

func TestNser_Ns_NoNamespaceSelected(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		},
	})
	fzf := fzftesting.NewFakeFzf([]string{}, nil)  // Empty selection
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no namespace selected")
}

func TestNser_Ns_HistoryError(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{
		Data: map[string][]string{},  // Empty history will cause error
	}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "-", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "getting namespace from history")
}

func TestNser_Ns_SetNamespaceError(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		},
	})
	fzf := fzftesting.NewFakeFzf([]string{""}, nil)  // Empty namespace will cause SetNamespace error
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting namespace")
}

func TestNser_Ns_KubeconfigWriteError(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	
	// Create a mock kubeconfig that returns error on Write
	kubeConfig := &mockKubeconfigWithWriteError{
		FakeKubeConfig: kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default"),
	}
	
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
		},
	})
	fzf := fzftesting.NewFakeFzf([]string{"test-ns"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "writing kubeconfig")
}

func TestNser_Ns_HistoryWriteError(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
		},
	})
	fzf := fzftesting.NewFakeFzf([]string{"test-ns"}, nil)
	
	// Create a mock history that returns error on Write
	history := &mockHistoryWithWriteError{
		FakeHistory: historytesting.FakeHistory{Data: make(map[string][]string)},
	}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "test", false)
	
	// Should not return error, but should write to stderr
	require.NoError(t, err)
	assert.Contains(t, errBuf.String(), "writing history")
	assert.Contains(t, outBuf.String(), "Switched to namespace \"test-ns\"")
}

func TestNser_Ns_ExactMatch(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	
	outBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: &bytes.Buffer{}}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-dev"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-prod"}},
		},
	})
	
	// Mock fzf that records the config
	fzfMock := fzftesting.NewFakeFzf([]string{"test"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzfMock, history)

	err := nser.Ns(context.Background(), "test", true)
	
	require.NoError(t, err)
	assert.True(t, fzfMock.LastConfig.ExactMatch)
	assert.Equal(t, "test", fzfMock.LastConfig.Query)
	assert.Contains(t, outBuf.String(), "Switched to namespace \"test\"")
}

func TestNser_Ns_VerifyPrompt(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		},
	})
	
	fzfMock := fzftesting.NewFakeFzf([]string{"default"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzfMock, history)

	err := nser.Ns(context.Background(), "def", false)
	
	require.NoError(t, err)
	// Note: The prompt says "Select context" but should probably say "Select namespace"
	// This is a bug in the original code at line 53
	assert.Equal(t, "Select context", fzfMock.LastConfig.Prompt)
	assert.Equal(t, "def", fzfMock.LastConfig.Query)
	assert.True(t, fzfMock.LastConfig.Sorted)
	assert.False(t, fzfMock.LastConfig.Multi)
}

func TestNser_Ns_HistoryAdded(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "new-ns"}},
		},
	})
	fzf := fzftesting.NewFakeFzf([]string{"new-ns"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "new", false)
	
	require.NoError(t, err)
	// Verify namespace was added to history
	assert.Equal(t, []string{"new-ns"}, history.Data["namespace"])
}

func TestNser_Ns_MultipleNamespaces(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	
	outBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: &bytes.Buffer{}}
	
	// Test with many namespaces
	namespaces := []any{
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-public"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-dev"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-staging"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-prod"}},
	}
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": namespaces,
	})
	
	fzf := fzftesting.NewFakeFzf([]string{"app-staging"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	nser := NewNser(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := nser.Ns(context.Background(), "app", false)
	
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Switched to namespace \"app-staging\"")
	
	currentNs, _ := kubeConfig.GetCurrentNamespace()
	assert.Equal(t, "app-staging", currentNs)
}

// Helper mocks

type mockKubeconfigWithWriteError struct {
	*kubeconfigtesting.FakeKubeConfig
}

func (m *mockKubeconfigWithWriteError) Write() error {
	return errors.New("kubeconfig write error")
}

type mockHistoryWithWriteError struct {
	historytesting.FakeHistory
}

func (m *mockHistoryWithWriteError) Write() error {
	return errors.New("history write error")
}