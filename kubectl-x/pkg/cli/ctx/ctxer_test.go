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
	pkghistory "github.com/RRethy/kubectl-x/pkg/history"
	historytesting "github.com/RRethy/kubectl-x/pkg/history/testing"
	pkgkubeconfig "github.com/RRethy/kubectl-x/pkg/kubeconfig"
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

func TestCtxer_Ctx(t *testing.T) {
	tests := []struct {
		name                 string
		contextSubstring     string
		namespaceSubstring   string
		exactMatch           bool
		setupContexts        map[string]*api.Context
		currentContext       string
		currentNamespace     string
		historyData          map[string][]string
		fzfResults           []fzftesting.CallResult
		fzfError             error
		namespaces           []any
		customKubeConfig     func(*kubeconfigtesting.FakeKubeConfig) interface{}
		customHistory        func(*historytesting.FakeHistory) interface{}
		expectedError        string
		expectedOutput       []string
		expectedErrOutput    string
		expectedContext      string
		expectedNamespace    string
		expectedFzfCalls     int
		verifyFzfConfigs     func(*testing.T, *fzftesting.FakeFzf)
	}{
		{
			name:               "select from history",
			contextSubstring:   "-",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
				"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
			},
			currentContext:   "ctx1",
			currentNamespace: "ns1",
			historyData: map[string][]string{
				"context": {"ctx1", "ctx2"},
			},
			expectedOutput:    []string{"Switched to context \"ctx2\"", "Switched to namespace \"ns2\""},
			expectedContext:   "ctx2",
			expectedNamespace: "ns2",
		},
		{
			name:               "select with fzf and trigger namespace selection",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
				"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
				"ctx3": {Cluster: "ctx3", Namespace: "ns3"},
			},
			currentContext:   "ctx1",
			currentNamespace: "ns1",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"ctx3"}},
				{Items: []string{"selected-ns"}},
			},
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns3"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "selected-ns"}},
			},
			expectedOutput:    []string{"Switched to context \"ctx3\"", "Switched to namespace \"selected-ns\""},
			expectedContext:   "ctx3",
			expectedNamespace: "selected-ns",
			expectedFzfCalls:  2,
		},
		{
			name:               "fzf error",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			currentContext:   "ctx1",
			currentNamespace: "default",
			fzfError:         errors.New("fzf failed"),
			expectedError:    "selecting context",
		},
		{
			name:               "no context selected",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			currentContext:   "ctx1",
			currentNamespace: "default",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{}},
			},
			expectedError: "no context selected",
		},
		{
			name:               "history error",
			contextSubstring:   "-",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			currentContext:   "ctx1",
			currentNamespace: "default",
			historyData:      map[string][]string{},
			expectedError:    "getting context from history",
		},
		{
			name:               "set context error",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			currentContext:   "ctx1",
			currentNamespace: "default",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{""}}, // Empty context will cause error
			},
			expectedError: "setting context",
		},
		{
			name:               "history write error",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
				"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
			},
			currentContext:   "ctx1",
			currentNamespace: "ns1",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"ctx2"}},
				{Items: []string{"ns2"}}, // Need namespace selection too
			},
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
			},
			customHistory: func(h *historytesting.FakeHistory) interface{} {
				return &mockHistoryWithWriteError{FakeHistory: *h}
			},
			expectedOutput:    []string{"Switched to context \"ctx2\"", "Switched to namespace \"ns2\""},
			expectedErrOutput: "writing history",
			expectedContext:   "ctx2",
			expectedNamespace: "ns2",
		},
		{
			name:               "kubeconfig write error after context switch",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
				"ctx2": {Cluster: "ctx2"},
			},
			currentContext:   "ctx1",
			currentNamespace: "default",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"ctx2"}},
			},
			customKubeConfig: func(k *kubeconfigtesting.FakeKubeConfig) interface{} {
				return &mockKubeconfigWithWriteError{FakeKubeConfig: k}
			},
			expectedError: "writing kubeconfig",
		},
		{
			name:               "trigger namespace selection for context without namespace",
			contextSubstring:   "ctx",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
				"ctx2": {Cluster: "ctx2", Namespace: ""},
			},
			currentContext:   "ctx1",
			currentNamespace: "ns1",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"ctx2"}},
				{Items: []string{"test-ns"}},
			},
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
			},
			expectedOutput:    []string{"Switched to context \"ctx2\"", "Switched to namespace \"test-ns\""},
			expectedContext:   "ctx2",
			expectedNamespace: "test-ns",
			expectedFzfCalls:  2,
		},
		{
			name:               "exact match mode",
			contextSubstring:   "dev",
			namespaceSubstring: "",
			exactMatch:         true,
			setupContexts: map[string]*api.Context{
				"dev":      {Cluster: "dev", Namespace: "dev-ns"},
				"dev-test": {Cluster: "dev-test", Namespace: "test-ns"},
			},
			currentContext:   "dev",
			currentNamespace: "dev-ns",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"dev"}},
				{Items: []string{"dev-ns"}},
			},
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "dev-ns"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
			},
			expectedOutput:    []string{"Switched to context \"dev\""},
			expectedContext:   "dev",
			expectedNamespace: "dev-ns",
			expectedFzfCalls:  2,
			verifyFzfConfigs: func(t *testing.T, fzf *fzftesting.FakeFzf) {
				cfg0, ok := fzf.GetConfig(0)
				require.True(t, ok)
				assert.True(t, cfg0.ExactMatch)
				assert.Equal(t, "dev", cfg0.Query)
				
				cfg1, ok := fzf.GetConfig(1)
				require.True(t, ok)
				assert.True(t, cfg1.ExactMatch)
				assert.Equal(t, "", cfg1.Query)
			},
		},
		{
			name:               "set namespace error during namespace selection",
			contextSubstring:   "ctx",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
				"ctx2": {Cluster: "ctx2", Namespace: "ns2"},
			},
			currentContext:   "ctx1",
			currentNamespace: "ns1",
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"ctx2"}},
				{Items: []string{"invalid-ns"}},
			},
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "invalid-ns"}},
			},
			customKubeConfig: func(k *kubeconfigtesting.FakeKubeConfig) interface{} {
				return &mockKubeconfigWithSpecificNamespaceError{
					FakeKubeConfig: k,
					errorNamespace: "invalid-ns",
				}
			},
			expectedError: "setting namespace",
		},
		{
			name:               "namespace from context is set without triggering selection",
			contextSubstring:   "-",
			namespaceSubstring: "",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: ""},
				"ctx2": {Cluster: "ctx2", Namespace: "ctx2-ns"},
			},
			currentContext:   "ctx1",
			currentNamespace: "default",
			historyData: map[string][]string{
				"context": {"ctx1", "ctx2"},
			},
			expectedOutput:    []string{"Switched to context \"ctx2\"", "Switched to namespace \"ctx2-ns\""},
			expectedContext:   "ctx2",
			expectedNamespace: "ctx2-ns",
			expectedFzfCalls:  0,
		},
		{
			name:               "get namespace for context with empty namespace triggers selection",
			contextSubstring:   "-",
			namespaceSubstring: "test",
			exactMatch:         false,
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1", Namespace: "ns1"},
				"ctx2": {Cluster: "ctx2", Namespace: ""}, // Empty namespace will trigger selection
			},
			currentContext:   "ctx1",
			currentNamespace: "ns1",
			historyData: map[string][]string{
				"context": {"ctx1", "ctx2"},
			},
			fzfResults: []fzftesting.CallResult{
				{Items: []string{"test-ns"}}, // Namespace selection
			},
			namespaces: []any{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-ns"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			},
			expectedOutput:    []string{"Switched to context \"ctx2\"", "Switched to namespace \"test-ns\""},
			expectedContext:   "ctx2",
			expectedNamespace: "test-ns",
			expectedFzfCalls:  1,
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
			var fzf *fzftesting.FakeFzf
			if len(tt.fzfResults) > 0 {
				fzf = fzftesting.NewFakeFzfWithMultipleCalls(tt.fzfResults)
			} else {
				fzf = fzftesting.NewFakeFzf(nil, tt.fzfError)
			}

			// Setup history
			history := &historytesting.FakeHistory{Data: tt.historyData}
			if history.Data == nil {
				history.Data = make(map[string][]string)
			}
			var historyInterface pkghistory.Interface = history
			if tt.customHistory != nil {
				historyInterface = tt.customHistory(history).(pkghistory.Interface)
			}
			
			ctxer := NewCtxer(kubeCfg, ioStreams, k8sClient, fzf, historyInterface)
			err := ctxer.Ctx(context.Background(), tt.contextSubstring, tt.namespaceSubstring, tt.exactMatch)

			// Check error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Check output
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outBuf.String(), expected)
			}

			// Check error output
			if tt.expectedErrOutput != "" {
				assert.Contains(t, errBuf.String(), tt.expectedErrOutput)
			}

			// Check context and namespace
			if tt.expectedContext != "" {
				currentCtx, _ := kubeCfg.GetCurrentContext()
				assert.Equal(t, tt.expectedContext, currentCtx)
			}
			if tt.expectedNamespace != "" {
				currentNs, _ := kubeCfg.GetCurrentNamespace()
				assert.Equal(t, tt.expectedNamespace, currentNs)
			}

			// Check fzf calls
			if tt.expectedFzfCalls > 0 {
				assert.True(t, fzf.WasCalledTimes(tt.expectedFzfCalls))
			}

			// Custom fzf config verification
			if tt.verifyFzfConfigs != nil {
				tt.verifyFzfConfigs(t, fzf)
			}
		})
	}
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