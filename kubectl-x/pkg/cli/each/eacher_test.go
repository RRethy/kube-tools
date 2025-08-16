package each

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd/api"

	fzftesting "github.com/RRethy/kubectl-x/pkg/fzf/testing"
	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
)

func TestEacher_Each(t *testing.T) {
	tests := []struct {
		name             string
		contextPattern   string
		outputFormat     string
		interactive      bool
		args             []string
		namespace        string
		setupContexts    map[string]*api.Context
		fzfResults       []string
		fzfError         error
		expectedError    string
		expectedOutput   string
		verifyFzfConfig  func(*testing.T, *fzftesting.FakeFzf)
	}{
		{
			name:           "invalid output format",
			contextPattern: "",
			outputFormat:   "invalid",
			interactive:    false,
			args:           []string{"get", "pods"},
			expectedError:  "invalid output format",
		},
		{
			name:           "no contexts matched pattern",
			contextPattern: "nonexistent",
			outputFormat:   "raw",
			interactive:    false,
			args:           []string{"get", "pods"},
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			expectedError: "no contexts matched pattern",
		},
		{
			name:           "invalid regex pattern",
			contextPattern: "[invalid",
			outputFormat:   "raw",
			interactive:    false,
			args:           []string{"get", "pods"},
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			expectedError: "invalid regex",
		},
		{
			name:           "interactive selection with fzf",
			contextPattern: "ctx",
			outputFormat:   "raw",
			interactive:    true,
			args:           []string{"get", "pods"},
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
				"ctx2": {Cluster: "ctx2"},
				"ctx3": {Cluster: "ctx3"},
			},
			fzfResults:     []string{"ctx1", "ctx3"},
			expectedOutput: "ctx1:",
			verifyFzfConfig: func(t *testing.T, fzf *fzftesting.FakeFzf) {
				assert.True(t, fzf.LastConfig.Multi)
				assert.Equal(t, "Select context", fzf.LastConfig.Prompt)
				assert.Equal(t, "ctx", fzf.LastConfig.Query)
				assert.True(t, fzf.LastConfig.Sorted)
				assert.False(t, fzf.LastConfig.ExactMatch)
			},
		},
		{
			name:           "interactive selection error",
			contextPattern: "ctx",
			outputFormat:   "raw",
			interactive:    true,
			args:           []string{"get", "pods"},
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			fzfError:      errors.New("fzf cancelled"),
			expectedError: "interactive selection",
		},
		{
			name:           "pattern selection with all contexts",
			contextPattern: "",
			outputFormat:   "raw",
			interactive:    false,
			args:           []string{"get", "pods"},
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
				"ctx2": {Cluster: "ctx2"},
			},
			expectedOutput: "ctx1:",
		},
		{
			name:           "pattern selection with regex",
			contextPattern: "^ctx[12]$",
			outputFormat:   "json",
			interactive:    false,
			args:           []string{"get", "pods"},
			namespace:      "test-ns",
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
				"ctx2": {Cluster: "ctx2"},
				"ctx3": {Cluster: "ctx3"},
			},
			expectedOutput: `"apiVersion": "v1"`,
		},
		{
			name:           "yaml output format",
			contextPattern: "ctx1",
			outputFormat:   "yaml",
			interactive:    false,
			args:           []string{"get", "pods"},
			setupContexts: map[string]*api.Context{
				"ctx1": {Cluster: "ctx1"},
			},
			expectedOutput: "apiVersion: v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup contexts
			contexts := tt.setupContexts
			if contexts == nil {
				contexts = make(map[string]*api.Context)
			}
			kubeConfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "ctx1", "default")

			// Setup IO streams
			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			ioStreams := genericclioptions.IOStreams{Out: outBuf, ErrOut: errBuf}

			// Setup fzf
			fzf := fzftesting.NewFakeFzf(tt.fzfResults, tt.fzfError)

			// Create eacher
			eacher := &Eacher{
				IOStreams:  ioStreams,
				Kubeconfig: kubeConfig,
				Namespace:  tt.namespace,
				Fzf:        fzf,
			}

			// Execute
			err := eacher.Each(context.Background(), tt.contextPattern, tt.outputFormat, tt.interactive, tt.args)

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

			// Verify fzf config
			if tt.verifyFzfConfig != nil {
				tt.verifyFzfConfig(t, fzf)
			}
		})
	}
}

func TestEacher_SelectContexts(t *testing.T) {
	tests := []struct {
		name             string
		pattern          string
		interactive      bool
		allContexts      []string
		fzfReturn        []string
		fzfError         error
		expectedContexts []string
		expectError      bool
		errorContains    string
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
			name:          "invalid regex returns error",
			pattern:       "[invalid",
			interactive:   false,
			allContexts:   []string{"ctx1"},
			expectError:   true,
			errorContains: "invalid regex",
		},
		{
			name:             "interactive mode uses fzf",
			pattern:          "ctx",
			interactive:      true,
			allContexts:      []string{"ctx1", "ctx2", "ctx3"},
			fzfReturn:        []string{"ctx1", "ctx3"},
			expectedContexts: []string{"ctx1", "ctx3"},
		},
		{
			name:          "interactive mode with fzf error",
			pattern:       "ctx",
			interactive:   true,
			allContexts:   []string{"ctx1"},
			fzfError:      errors.New("user cancelled"),
			expectError:   true,
			errorContains: "interactive selection",
		},
		{
			name:             "interactive mode with empty selection",
			pattern:          "ctx",
			interactive:      true,
			allContexts:      []string{"ctx1", "ctx2"},
			fzfReturn:        []string{},
			expectedContexts: []string{},
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
			mockFzf := fzftesting.NewFakeFzf(tt.fzfReturn, tt.fzfError)

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
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
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
		{
			name:        "complex regex with groups",
			pattern:     "(dev|prod)-\\w+",
			allContexts: []string{"dev-test", "prod-test", "staging-test"},
			expected:    []string{"dev-test", "prod-test"},
		},
		{
			name:        "case sensitive matching",
			pattern:     "Dev",
			allContexts: []string{"dev", "Dev", "DEV"},
			expected:    []string{"Dev"},
		},
		{
			name:        "wildcard pattern",
			pattern:     ".*-test.*",
			allContexts: []string{"my-test-1", "test-2", "prod"},
			expected:    []string{"my-test-1"},
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
		verifyOutput   func(*testing.T, string)
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
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, `"kind": "List"`)
				assert.Contains(t, output, `"context": "ctx1"`)
			},
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
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "kind: List")
				assert.Contains(t, output, "context: ctx1")
			},
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
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "pod1")
				assert.Contains(t, output, "pod2")
			},
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
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "command failed")
				assert.Contains(t, output, "error output")
			},
		},
		{
			name:         "unsupported format",
			outputFormat: "xml",
			items:        []clusterResult{},
			expectError:  true,
		},
		{
			name:         "empty results json",
			outputFormat: "json",
			items:        []clusterResult{},
			expectedOutput: `"items": []`,
		},
		{
			name:         "empty results yaml",
			outputFormat: "yaml",
			items:        []clusterResult{},
			expectedOutput: "items: []",
		},
		{
			name:         "empty results raw",
			outputFormat: "raw",
			items:        []clusterResult{},
			expectedOutput: "",
		},
		{
			name:         "multiple results raw",
			outputFormat: "raw",
			items: []clusterResult{
				{
					Context: "ctx1",
					Output:  "output1",
				},
				{
					Context: "ctx2",
					Error:   "error2",
				},
				{
					Context: "ctx3",
					Output:  "output3",
				},
			},
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "ctx1:")
				assert.Contains(t, output, "ctx2:")
				assert.Contains(t, output, "ctx3:")
				assert.Contains(t, output, "Error: error2")
			},
		},
		{
			name:         "raw output with non-string output",
			outputFormat: "raw",
			items: []clusterResult{
				{
					Context: "ctx1",
					Output:  map[string]string{"key": "value"},
				},
			},
			expectedOutput: "ctx1:",
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "map[key:value]")
			},
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
				output := out.String()
				if tt.expectedOutput != "" {
					assert.Contains(t, output, tt.expectedOutput)
				}
				if tt.verifyOutput != nil {
					tt.verifyOutput(t, output)
				}
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
		verifyResult func(*testing.T, clusterResult)
	}{
		{
			name:         "basic command with context and namespace",
			contextName:  "test-ctx",
			namespace:    "test-ns",
			outputFormat: "raw",
			args:         []string{"get", "pods"},
			expectedArgs: []string{"get", "pods", "--context", "test-ctx", "--namespace", "test-ns"},
			verifyResult: func(t *testing.T, result clusterResult) {
				assert.Equal(t, "test-ctx", result.Context)
				assert.Contains(t, result.Args, "get pods")
				assert.Contains(t, result.Args, "--context test-ctx")
				assert.Contains(t, result.Args, "--namespace test-ns")
			},
		},
		{
			name:         "command with json output",
			contextName:  "test-ctx",
			namespace:    "default",
			outputFormat: "json",
			args:         []string{"get", "pods"},
			expectedArgs: []string{"get", "pods", "--context", "test-ctx", "--namespace", "default", "-ojson"},
			verifyResult: func(t *testing.T, result clusterResult) {
				assert.Contains(t, result.Args, "-ojson")
			},
		},
		{
			name:         "command with yaml output",
			contextName:  "test-ctx",
			namespace:    "kube-system",
			outputFormat: "yaml",
			args:         []string{"get", "pods"},
			expectedArgs: []string{"get", "pods", "--context", "test-ctx", "--namespace", "kube-system", "-oyaml"},
			verifyResult: func(t *testing.T, result clusterResult) {
				assert.Contains(t, result.Args, "-oyaml")
			},
		},
		{
			name:         "command with empty namespace",
			contextName:  "test-ctx",
			namespace:    "",
			outputFormat: "raw",
			args:         []string{"get", "nodes"},
			verifyResult: func(t *testing.T, result clusterResult) {
				assert.Contains(t, result.Args, "--namespace ")
			},
		},
		{
			name:         "command with special characters in args",
			contextName:  "test-ctx",
			namespace:    "test-ns",
			outputFormat: "raw",
			args:         []string{"get", "pods", "-l", "app=test"},
			verifyResult: func(t *testing.T, result clusterResult) {
				assert.Contains(t, result.Args, "app=test")
			},
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
			
			if tt.verifyResult != nil {
				tt.verifyResult(t, result)
			}
		})
	}
}

func TestEacher_ExecuteCommands(t *testing.T) {
	tests := []struct {
		name         string
		contexts     []string
		outputFormat string
		args         []string
		namespace    string
		verifyResults func(*testing.T, []clusterResult)
	}{
		{
			name:         "single context",
			contexts:     []string{"ctx1"},
			outputFormat: "raw",
			args:         []string{"get", "pods"},
			namespace:    "default",
			verifyResults: func(t *testing.T, results []clusterResult) {
				assert.Len(t, results, 1)
				assert.Equal(t, "ctx1", results[0].Context)
			},
		},
		{
			name:         "multiple contexts sorted",
			contexts:     []string{"ctx3", "ctx1", "ctx2"},
			outputFormat: "raw",
			args:         []string{"get", "pods"},
			namespace:    "default",
			verifyResults: func(t *testing.T, results []clusterResult) {
				assert.Len(t, results, 3)
				// Results should be sorted by context name
				assert.Equal(t, "ctx1", results[0].Context)
				assert.Equal(t, "ctx2", results[1].Context)
				assert.Equal(t, "ctx3", results[2].Context)
			},
		},
		{
			name:         "many contexts with concurrency",
			contexts:     []string{"ctx1", "ctx2", "ctx3", "ctx4", "ctx5", "ctx6", "ctx7", "ctx8"},
			outputFormat: "json",
			args:         []string{"get", "pods"},
			namespace:    "test",
			verifyResults: func(t *testing.T, results []clusterResult) {
				assert.Len(t, results, 8)
				// All should have the same namespace
				for _, r := range results {
					assert.Contains(t, r.Args, "--namespace test")
					assert.Contains(t, r.Args, "-ojson")
				}
			},
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

			results := eacher.executeCommands(context.Background(), tt.contexts, tt.outputFormat, tt.args)

			if tt.verifyResults != nil {
				tt.verifyResults(t, results)
			}
		})
	}
}

func TestEacher_SelectContextsInteractively(t *testing.T) {
	tests := []struct {
		name            string
		allContexts     []string
		initialPattern  string
		fzfReturn       []string
		fzfError        error
		expectedResult  []string
		expectError     bool
		verifyFzfConfig func(*testing.T, *fzftesting.FakeFzf)
	}{
		{
			name:           "successful selection",
			allContexts:    []string{"ctx1", "ctx2", "ctx3"},
			initialPattern: "ctx",
			fzfReturn:      []string{"ctx1", "ctx3"},
			expectedResult: []string{"ctx1", "ctx3"},
			verifyFzfConfig: func(t *testing.T, fzf *fzftesting.FakeFzf) {
				assert.True(t, fzf.LastConfig.Multi)
				assert.Equal(t, "Select context", fzf.LastConfig.Prompt)
				assert.Equal(t, "ctx", fzf.LastConfig.Query)
				assert.True(t, fzf.LastConfig.Sorted)
				assert.False(t, fzf.LastConfig.ExactMatch)
			},
		},
		{
			name:           "fzf error",
			allContexts:    []string{"ctx1"},
			initialPattern: "",
			fzfError:       errors.New("user cancelled"),
			expectError:    true,
		},
		{
			name:           "empty selection",
			allContexts:    []string{"ctx1", "ctx2"},
			initialPattern: "test",
			fzfReturn:      []string{},
			expectedResult: []string{},
		},
		{
			name:           "single selection from multi",
			allContexts:    []string{"ctx1", "ctx2", "ctx3"},
			initialPattern: "",
			fzfReturn:      []string{"ctx2"},
			expectedResult: []string{"ctx2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fzf := fzftesting.NewFakeFzf(tt.fzfReturn, tt.fzfError)
			
			eacher := &Eacher{
				Fzf: fzf,
				IOStreams: genericclioptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				},
			}

			result, err := eacher.selectContextsInteractively(context.Background(), tt.allContexts, tt.initialPattern)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "interactive selection")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			if tt.verifyFzfConfig != nil {
				tt.verifyFzfConfig(t, fzf)
			}
		})
	}
}