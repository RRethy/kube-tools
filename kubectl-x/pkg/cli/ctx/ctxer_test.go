package ctx

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

func TestNewCtxer(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {Cluster: "test-ctx"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	assert.Equal(t, kubeConfig, ctxer.KubeConfig)
	assert.Equal(t, ioStreams, ctxer.IoStreams)
	assert.Equal(t, k8sClient, ctxer.K8sClient)
	assert.Equal(t, fzf, ctxer.Fzf)
	assert.Equal(t, history, ctxer.History)
}

func TestCtxer_Ctx_SelectFromHistory(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
		"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "ns1")
	
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}
	
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{
		Data: map[string][]string{
			"context": {"ctx1", "ctx2"},  // Index 1 will return "ctx2"
		},
	}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "-", "", false)
	
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Switched to context \"ctx2\"")
	assert.Contains(t, outBuf.String(), "Switched to namespace \"ns2\"")
	
	currentCtx, _ := kubeConfig.GetCurrentContext()
	assert.Equal(t, "ctx2", currentCtx)
}

func TestCtxer_Ctx_SelectWithFzf(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
		"ctx2": {Cluster: "ctx2", Namespace: "ns2"}, 
		"ctx3": {Cluster: "ctx3", Namespace: "ns3"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "ns1")
	
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}
	
	// Add namespace objects for namespace selection
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns3"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "selected-ns"}},
		},
	})
	
	// When selecting context with fzf, namespace selection is ALWAYS triggered
	// because selectedNamespace remains empty in the fzf path
	fzf := fzftesting.NewFakeFzfWithMultipleCalls([]fzftesting.CallResult{
		{Items: []string{"ctx3"}},        // First call for context selection
		{Items: []string{"selected-ns"}}, // Second call for namespace selection (always triggered)
	})
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Switched to context \"ctx3\"")
	// Namespace selection is triggered, user selects "selected-ns"
	assert.Contains(t, outBuf.String(), "Switched to namespace \"selected-ns\"")
	
	currentCtx, _ := kubeConfig.GetCurrentContext()
	assert.Equal(t, "ctx3", currentCtx)
	currentNs, _ := kubeConfig.GetCurrentNamespace()
	assert.Equal(t, "selected-ns", currentNs)
}

func TestCtxer_Ctx_FzfError(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, errors.New("fzf failed"))
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selecting context")
}

func TestCtxer_Ctx_NoContextSelected(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf([]string{}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no context selected")
}

func TestCtxer_Ctx_HistoryError(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf(nil, nil)
	history := &historytesting.FakeHistory{
		Data: map[string][]string{},
	}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "-", "", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "getting context from history")
}

func TestCtxer_Ctx_SetContextError(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default")
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf([]string{""}, nil) // Empty context will cause error
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting context")
}

func TestCtxer_Ctx_HistoryWriteError(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
		"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "ns1")
	
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}
	
	// Add namespace objects to avoid errors during namespace selection
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
		},
	})
	fzf := fzftesting.NewFakeFzf([]string{"ctx2"}, nil)
	
	// Create a mock history that returns error on Write
	history := &mockHistoryWithWriteError{
		FakeHistory: historytesting.FakeHistory{Data: make(map[string][]string)},
	}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	// Should not return error, but should write to stderr
	require.NoError(t, err)
	assert.Contains(t, errBuf.String(), "writing history")
	assert.Contains(t, outBuf.String(), "Switched to context \"ctx2\"")
}

func TestCtxer_Ctx_KubeconfigWriteError(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1"},
		"ctx2": {Cluster: "ctx2"},
	}
	
	// Create a mock kubeconfig that returns error on Write
	kubeConfig := &mockKubeconfigWithWriteError{
		FakeKubeConfig: kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default"),
	}
	
	ioStreams := genericiooptions.IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}}
	k8sClient := kubernetestesting.NewFakeClient(nil)
	fzf := fzftesting.NewFakeFzf([]string{"ctx2"}, nil)
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "writing kubeconfig")
}

func TestCtxer_Ctx_SetNamespaceError(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
		"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
	}
	
	// Create a mock kubeconfig that returns error on SetNamespace with specific value
	kubeConfig := &mockKubeconfigWithSpecificNamespaceError{
		FakeKubeConfig: kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "ns1"),
		errorNamespace: "invalid-ns",
	}
	
	outBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: &bytes.Buffer{}}
	// Add namespace objects for namespace selection
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "invalid-ns"}}, // This will be selected and cause error
		},
	})
	
	// Two fzf calls: one for context, one for namespace
	fzf := fzftesting.NewFakeFzfWithMultipleCalls([]fzftesting.CallResult{
		{Items: []string{"ctx2"}},         // Context selection
		{Items: []string{"invalid-ns"}},   // Namespace selection that will fail
	})
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "", false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting namespace")
}

func TestCtxer_Ctx_TriggerNamespaceSelection(t *testing.T) {
	contexts := map[string]*api.Context{
		"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
		"ctx2": {Cluster: "ctx2", Namespace: ""},  // No namespace, will trigger ns selection
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "ns1")
	
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
	
	// Mock fzf will be called twice: once for context, once for namespace
	fzf := fzftesting.NewFakeFzfWithMultipleCalls([]fzftesting.CallResult{
		{Items: []string{"ctx2"}},     // First call for context selection
		{Items: []string{"test-ns"}},  // Second call for namespace selection
	})
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzf, history)

	err := ctxer.Ctx(context.Background(), "ctx", "test", false)
	
	require.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Switched to context \"ctx2\"")
	assert.Contains(t, outBuf.String(), "Switched to namespace \"test-ns\"")
	assert.True(t, fzf.WasCalledTimes(2))  // Verify fzf was called twice
}

func TestCtxer_Ctx_ExactMatch(t *testing.T) {
	contexts := map[string]*api.Context{
		"dev": {Cluster: "dev", Namespace: "dev-ns"},
		"dev-test": {Cluster: "dev-test", Namespace: "test-ns"},
	}
	kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "dev", "dev-ns")
	
	outBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: &bytes.Buffer{}}
	// Add namespace objects to avoid errors during namespace selection
	k8sClient := kubernetestesting.NewFakeClient(map[string][]any{
		"namespace": {
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "dev-ns"}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
		},
	})
	
	// Mock fzf that records configs for both calls
	fzfMock := fzftesting.NewFakeFzfWithMultipleCalls([]fzftesting.CallResult{
		{Items: []string{"dev"}},      // Context selection
		{Items: []string{"dev-ns"}},   // Namespace selection
	})
	history := &historytesting.FakeHistory{Data: make(map[string][]string)}

	ctxer := NewCtxer(kubeConfig, ioStreams, k8sClient, fzfMock, history)

	err := ctxer.Ctx(context.Background(), "dev", "", true)
	
	require.NoError(t, err)
	// Check the first call (context selection) config
	cfg0, ok := fzfMock.GetConfig(0)
	require.True(t, ok)
	assert.True(t, cfg0.ExactMatch)
	assert.Equal(t, "dev", cfg0.Query)
	// Check the second call (namespace selection) config - uses empty namespace substring
	cfg1, ok := fzfMock.GetConfig(1)
	require.True(t, ok)
	assert.True(t, cfg1.ExactMatch)
	assert.Equal(t, "", cfg1.Query)
}

// Helper mocks for specific test cases

type mockHistoryWithWriteError struct {
	historytesting.FakeHistory
}

func (m *mockHistoryWithWriteError) Write() error {
	return errors.New("history write error")
}

type mockKubeconfigWithWriteError struct {
	*kubeconfigtesting.FakeKubeConfig
}

func (m *mockKubeconfigWithWriteError) Write() error {
	return errors.New("kubeconfig write error")
}

type mockKubeconfigWithSpecificNamespaceError struct {
	*kubeconfigtesting.FakeKubeConfig
	errorNamespace string
}

func (m *mockKubeconfigWithSpecificNamespaceError) SetNamespace(namespace string) error {
	if namespace == m.errorNamespace {
		return errors.New("invalid namespace")
	}
	return m.FakeKubeConfig.SetNamespace(namespace)
}

