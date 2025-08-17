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
	pkghistory "github.com/RRethy/kubectl-x/pkg/history"
	historytesting "github.com/RRethy/kubectl-x/pkg/history/testing"
	pkgkubeconfig "github.com/RRethy/kubectl-x/pkg/kubeconfig"
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

func TestNser_Ns(t *testing.T) {
	tests := []struct {
		name               string
		namespaceSubstring string
		exactMatch         bool
		setupContexts      map[string]*api.Context
		currentContext     string
		currentNamespace   string
		historyData        map[string][]string
		namespaces         []any
		fzfResults         []string
		fzfError           error
		customKubeConfig   func(*kubeconfigtesting.FakeKubeConfig) interface{}
		customHistory      func(*historytesting.FakeHistory) interface{}
		expectedError      string
		expectedOutput     string
		expectedErrOutput  string
		expectedNamespace  string
		verifyFzfConfig    func(*testing.T, *fzftesting.FakeFzf)
		verifyHistory      func(*testing.T, *historytesting.FakeHistory)
	}{
		{
			name:               "select from history",
			namespaceSubstring: "-",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			historyData: map[string][]string{
				"namespace": {"default", "kube-system"},
			},
			expectedOutput:    "Switched to namespace \"kube-system\"",
			expectedNamespace: "kube-system",
		},
		{
			name:               "select with fzf",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
			},
			fzfResults:        []string{"test-ns"},
			expectedOutput:    "Switched to namespace \"test-ns\"",
			expectedNamespace: "test-ns",
		},
		{
			name:               "list namespaces error",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces:       nil, // No namespaces will cause listing error
			expectedError:    "listing namespaces",
		},
		{
			name:               "fzf error",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			},
			fzfError:      errors.New("fzf failed"),
			expectedError: "selecting namespace",
		},
		{
			name:               "no namespace selected",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			},
			fzfResults:    []string{}, // Empty selection
			expectedError: "no namespace selected",
		},
		{
			name:               "history error",
			namespaceSubstring: "-",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			historyData:      map[string][]string{}, // Empty history will cause error
			expectedError:    "getting namespace from history",
		},
		{
			name:               "set namespace error",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			},
			fzfResults:    []string{""}, // Empty namespace will cause SetNamespace error
			expectedError: "setting namespace",
		},
		{
			name:               "kubeconfig write error",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
			},
			fzfResults: []string{"test-ns"},
			customKubeConfig: func(k *kubeconfigtesting.FakeKubeConfig) interface{} {
				return &mockKubeconfigWithWriteError{FakeKubeConfig: k}
			},
			expectedError: "writing kubeconfig",
		},
		{
			name:               "history write error",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
			},
			fzfResults: []string{"test-ns"},
			customHistory: func(h *historytesting.FakeHistory) interface{} {
				return &mockHistoryWithWriteError{FakeHistory: *h}
			},
			expectedOutput:    "Switched to namespace \"test-ns\"",
			expectedNamespace: "test-ns",
		},
		{
			name:               "exact match mode",
			namespaceSubstring: "test",
			exactMatch:         true,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-dev"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-prod"}},
			},
			fzfResults:        []string{"test"},
			expectedOutput:    "Switched to namespace \"test\"",
			expectedNamespace: "test",
			verifyFzfConfig: func(t *testing.T, fzf *fzftesting.FakeFzf) {
				assert.True(t, fzf.LastConfig.ExactMatch)
				assert.Equal(t, "test", fzf.LastConfig.Query)
			},
		},
		{
			name:               "verify prompt and config",
			namespaceSubstring: "def",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			},
			fzfResults:        []string{"default"},
			expectedOutput:    "Switched to namespace \"default\"",
			expectedNamespace: "default",
			verifyFzfConfig: func(t *testing.T, fzf *fzftesting.FakeFzf) {
				// Note: The prompt says "Select context" but should probably say "Select namespace"
				// This is a bug in the original code
				assert.Equal(t, "Select context", fzf.LastConfig.Prompt)
				assert.Equal(t, "def", fzf.LastConfig.Query)
				assert.True(t, fzf.LastConfig.Sorted)
				assert.False(t, fzf.LastConfig.Multi)
			},
		},
		{
			name:               "history is updated",
			namespaceSubstring: "new",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "new-ns"}},
			},
			fzfResults:        []string{"new-ns"},
			expectedOutput:    "Switched to namespace \"new-ns\"",
			expectedNamespace: "new-ns",
			verifyHistory: func(t *testing.T, history *historytesting.FakeHistory) {
				assert.Equal(t, []string{"new-ns"}, history.Data["namespace"])
			},
		},
		{
			name:               "multiple namespaces",
			namespaceSubstring: "app",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-public"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-dev"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-staging"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-prod"}},
			},
			fzfResults:        []string{"app-staging"},
			expectedOutput:    "Switched to namespace \"app-staging\"",
			expectedNamespace: "app-staging",
		},
		{
			name:               "namespace with special characters",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-1.2.3"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test_underscore"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-dash"}},
			},
			fzfResults:        []string{"test-1.2.3"},
			expectedOutput:    "Switched to namespace \"test-1.2.3\"",
			expectedNamespace: "test-1.2.3",
		},
		{
			name:               "empty namespace substring selects all",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"test-ctx": {Cluster: "test-ctx", Namespace: "default"},
			},
			currentContext:   "test-ctx",
			currentNamespace: "default",
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns3"}},
			},
			fzfResults:        []string{"ns2"},
			expectedOutput:    "Switched to namespace \"ns2\"",
			expectedNamespace: "ns2",
			verifyFzfConfig: func(t *testing.T, fzf *fzftesting.FakeFzf) {
				assert.Equal(t, "", fzf.LastConfig.Query)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup kubeconfig
			var kubeCfg pkgkubeconfig.Interface = kubeconfigtesting.NewFakeKubeConfig(tt.setupContexts, tt.currentContext, tt.currentNamespace)
			if tt.customKubeConfig != nil {
				kubeCfg = tt.customKubeConfig(kubeconfigtesting.NewFakeKubeConfig(tt.setupContexts, tt.currentContext, tt.currentNamespace)).(pkgkubeconfig.Interface)
			}

			// Setup IO streams
			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			ioStreams := genericiooptions.IOStreams{Out: outBuf, ErrOut: errBuf}

			// Setup k8s client
			k8sObjects := make(map[string][]any)
			if len(tt.namespaces) > 0 {
				k8sObjects["namespace"] = tt.namespaces
			}
			k8sClient := kubernetestesting.NewFakeClient(k8sObjects)

			// Setup fzf
			fzf := fzftesting.NewFakeFzf(tt.fzfResults, tt.fzfError)

			// Setup history
			history := &historytesting.FakeHistory{Data: tt.historyData}
			if history.Data == nil {
				history.Data = make(map[string][]string)
			}
			var historyInterface pkghistory.Interface = history
			if tt.customHistory != nil {
				historyInterface = tt.customHistory(history).(pkghistory.Interface)
			}

			// Create nser and execute
			nser := NewNser(kubeCfg, ioStreams, k8sClient, fzf, historyInterface)
			err := nser.Ns(context.Background(), tt.namespaceSubstring, tt.exactMatch)

			// Check error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Check output
			if tt.expectedOutput != "" {
				assert.Contains(t, outBuf.String(), tt.expectedOutput)
			}

			// Check error output
			if tt.expectedErrOutput != "" {
				assert.Contains(t, errBuf.String(), tt.expectedErrOutput)
			}

			// Check namespace
			if tt.expectedNamespace != "" {
				currentNs, _ := kubeCfg.GetCurrentNamespace()
				assert.Equal(t, tt.expectedNamespace, currentNs)
			}

			// Custom fzf config verification
			if tt.verifyFzfConfig != nil {
				tt.verifyFzfConfig(t, fzf)
			}

			// Custom history verification
			if tt.verifyHistory != nil {
				tt.verifyHistory(t, history)
			}
		})
	}
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