package each

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd/api"

	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
)

func TestEacher_Each_InvalidOutputFormat(t *testing.T) {
	eacher := &Eacher{
		IOStreams: genericclioptions.IOStreams{
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}

	err := eacher.Each(context.Background(), "", "invalid", false, []string{"get", "pods"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

func TestEacher_SelectContexts(t *testing.T) {
	tests := []struct {
		name             string
		pattern          string
		interactive      bool
		allContexts      []string
		fzfReturn        []string
		expectedContexts []string
		expectError      bool
	}{
		{
			name:             "empty pattern returns all contexts",
			pattern:          "",
			interactive:      false,
			allContexts:      []string{"ctx1", "ctx2", "ctx3"},
			expectedContexts: []string{"ctx1", "ctx2", "ctx3"},
		},
		{
			name:             "regex pattern filters contexts",
			pattern:          "ctx[12]",
			interactive:      false,
			allContexts:      []string{"ctx1", "ctx2", "ctx3"},
			expectedContexts: []string{"ctx1", "ctx2"},
		},
		{
			name:             "invalid regex returns error",
			pattern:          "[invalid",
			interactive:      false,
			allContexts:      []string{"ctx1"},
			expectError:      true,
		},
		{
			name:             "interactive mode uses fzf",
			pattern:          "ctx",
			interactive:      true,
			allContexts:      []string{"ctx1", "ctx2", "ctx3"},
			fzfReturn:        []string{"ctx1", "ctx3"},
			expectedContexts: []string{"ctx1", "ctx3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contexts := make(map[string]*api.Context)
			for _, ctx := range tt.allContexts {
				contexts[ctx] = &api.Context{
					Cluster: ctx,
				}
			}

			mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default")
			mockFzf := fzftesting.NewFakeFzf(tt.fzfReturn, nil)

			eacher := &Eacher{
				Kubeconfig: mockKubeconfig,
				Fzf:        mockFzf,
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
			}

			result, err := eacher.selectContexts(context.Background(), tt.pattern, tt.interactive)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedContexts, result)
			}
		})
	}
}

func TestEacher_SelectContextsByPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		allContexts []string
		expected    []string
		expectError bool
	}{
		{
			name:        "empty pattern returns all",
			pattern:     "",
			allContexts: []string{"dev", "staging", "prod"},
			expected:    []string{"dev", "staging", "prod"},
		},
		{
			name:        "exact match",
			pattern:     "^dev$",
			allContexts: []string{"dev", "dev-test", "staging"},
			expected:    []string{"dev"},
		},
		{
			name:        "prefix match",
			pattern:     "^dev",
			allContexts: []string{"dev", "dev-test", "staging"},
			expected:    []string{"dev", "dev-test"},
		},
		{
			name:        "suffix match",
			pattern:     "test$",
			allContexts: []string{"dev", "dev-test", "staging-test"},
			expected:    []string{"dev-test", "staging-test"},
		},
		{
			name:        "no matches",
			pattern:     "^prod",
			allContexts: []string{"dev", "staging"},
			expected:    nil,
		},
		{
			name:        "invalid regex",
			pattern:     "[invalid",
			allContexts: []string{"dev"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eacher := &Eacher{}
			result, err := eacher.selectContextsByPattern(tt.allContexts, tt.pattern)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid regex")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEacher_OutputResults(t *testing.T) {
	tests := []struct {
		name           string
		outputFormat   string
		items          []clusterResult
		expectedOutput string
		expectError    bool
	}{
		{
			name:         "json output format",
			outputFormat: "json",
			items: []clusterResult{
				{
					Context: "ctx1",
					Args:    "get pods",
					Output:  "pod1",
				},
			},
			expectedOutput: `"apiVersion": "v1"`,
		},
		{
			name:         "yaml output format",
			outputFormat: "yaml",
			items: []clusterResult{
				{
					Context: "ctx1",
					Output:  "pod1",
				},
			},
			expectedOutput: "apiVersion: v1",
		},
		{
			name:         "raw output format",
			outputFormat: "raw",
			items: []clusterResult{
				{
					Context: "ctx1",
					Output:  "pod1\npod2",
				},
			},
			expectedOutput: "ctx1:",
		},
		{
			name:         "raw output with error",
			outputFormat: "raw",
			items: []clusterResult{
				{
					Context: "ctx1",
					Error:   "command failed",
					Output:  "error output",
				},
			},
			expectedOutput: "Error:",
		},
		{
			name:         "unsupported format",
			outputFormat: "xml",
			items:        []clusterResult{},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			eacher := &Eacher{
				IOStreams: genericclioptions.IOStreams{
					Out:    out,
					ErrOut: &bytes.Buffer{},
				},
			}

			err := eacher.outputResults(tt.items, tt.outputFormat)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, out.String(), tt.expectedOutput)
			}
		})
	}
}

func TestEacher_ExecuteCommand(t *testing.T) {
	tests := []struct {
		name         string
		contextName  string
		namespace    string
		outputFormat string
		args         []string
		expectedArgs []string
	}{
		{
			name:         "basic command with context and namespace",
			contextName:  "test-ctx",
			namespace:    "test-ns",
			outputFormat: "raw",
			args:         []string{"get", "pods"},
			expectedArgs: []string{"get", "pods", "--context", "test-ctx", "--namespace", "test-ns"},
		},
		{
			name:         "command with json output",
			contextName:  "test-ctx",
			namespace:    "default",
			outputFormat: "json",
			args:         []string{"get", "pods"},
			expectedArgs: []string{"get", "pods", "--context", "test-ctx", "--namespace", "default", "-ojson"},
		},
		{
			name:         "command with yaml output",
			contextName:  "test-ctx",
			namespace:    "kube-system",
			outputFormat: "yaml",
			args:         []string{"get", "pods"},
			expectedArgs: []string{"get", "pods", "--context", "test-ctx", "--namespace", "kube-system", "-oyaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eacher := &Eacher{
				Namespace: tt.namespace,
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
			}

			result := eacher.executeCommand(context.Background(), tt.contextName, tt.outputFormat, tt.args)

			assert.Equal(t, tt.contextName, result.Context)
			assert.Contains(t, result.Args, strings.Join(tt.expectedArgs, " "))
		})
	}
}