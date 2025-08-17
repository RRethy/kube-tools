package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilterParams(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		expected FilterParams
	}{
		{
			name: "All filter parameters",
			args: map[string]any{
				"grep": "error",
				"jq":   ".items[]",
				"yq":   ".metadata.name",
			},
			expected: FilterParams{
				Grep: "error",
				JQ:   ".items[]",
				YQ:   ".metadata.name",
			},
		},
		{
			name: "Only grep",
			args: map[string]any{
				"grep": "warning",
			},
			expected: FilterParams{
				Grep: "warning",
			},
		},
		{
			name: "Empty parameters",
			args: map[string]any{
				"grep": "",
				"jq":   "",
				"yq":   "",
			},
			expected: FilterParams{},
		},
		{
			name:     "No parameters",
			args:     map[string]any{},
			expected: FilterParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilterParams(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyGrepFilter(t *testing.T) {
	testInput := `Line 1: This is an error message
Line 2: This is a warning message
Line 3: This is an info message
Line 4: Another error occurred
Line 5: Everything is fine`

	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:    "Literal string match",
			pattern: "error",
			expected: `Line 1: This is an error message
Line 4: Another error occurred`,
		},
		{
			name:    "Regex pattern",
			pattern: "^Line [13]:",
			expected: `Line 1: This is an error message
Line 3: This is an info message`,
		},
		{
			name:     "No matches",
			pattern:  "nonexistent",
			expected: "",
		},
		{
			name:     "Case sensitive",
			pattern:  "Error",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applyGrepFilter(testInput, tt.pattern)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplySimpleJSONFilter(t *testing.T) {
	jsonInput := `{
		"metadata": {
			"name": "test-pod",
			"namespace": "default"
		},
		"items": [
			{"name": "item1", "value": 10},
			{"name": "item2", "value": 20}
		],
		"status": "Running"
	}`

	tests := []struct {
		name      string
		filter    string
		expected  string
		expectErr bool
	}{
		{
			name:     "Simple field access",
			filter:   ".status",
			expected: `"Running"`,
		},
		{
			name:     "Nested field access",
			filter:   ".metadata.name",
			expected: `"test-pod"`,
		},
		{
			name:     "Non-existent field",
			filter:   ".nonexistent",
			expected: "null",
		},
		{
			name:      "Invalid filter format",
			filter:    "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applySimpleJSONFilter(jsonInput, tt.filter)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Normalize whitespace for comparison
				expected := strings.TrimSpace(tt.expected)
				actual := strings.TrimSpace(result)
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestApplySimpleYAMLFilter(t *testing.T) {
	yamlInput := `metadata:
  name: test-pod
  namespace: default
items:
  - name: item1
    value: 10
  - name: item2
    value: 20
status: Running`

	tests := []struct {
		name      string
		filter    string
		expected  string
		expectErr bool
	}{
		{
			name:     "Simple field access",
			filter:   ".status",
			expected: `Running`,
		},
		{
			name:     "Nested field access",
			filter:   ".metadata.name",
			expected: `test-pod`,
		},
		{
			name:     "Non-existent field",
			filter:   ".nonexistent",
			expected: "null",
		},
		{
			name:      "Invalid filter format",
			filter:    "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applySimpleYAMLFilter(yamlInput, tt.filter)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Normalize whitespace and newlines for comparison
				expected := strings.TrimSpace(tt.expected)
				actual := strings.TrimSpace(result)
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestApplyFilter(t *testing.T) {
	textOutput := `Line 1: Normal output
Line 2: Error message
Line 3: Warning message`

	jsonOutput := `{"name": "test", "status": "Running", "items": [1, 2, 3]}`

	yamlOutput := `name: test
status: Running
items:
  - 1
  - 2
  - 3`

	tests := []struct {
		name         string
		output       string
		params       FilterParams
		outputFormat string
		expected     string
		expectErr    bool
	}{
		{
			name:   "Grep filter on text",
			output: textOutput,
			params: FilterParams{
				Grep: "Error",
			},
			outputFormat: "",
			expected:     "Line 2: Error message",
		},
		{
			name:   "JQ filter on JSON",
			output: jsonOutput,
			params: FilterParams{
				JQ: ".status",
			},
			outputFormat: "json",
			expected:     `"Running"`,
		},
		{
			name:   "YQ filter on YAML",
			output: yamlOutput,
			params: FilterParams{
				YQ: ".status",
			},
			outputFormat: "yaml",
			expected:     "Running",
		},
		{
			name:   "Grep filter works on any format",
			output: jsonOutput,
			params: FilterParams{
				Grep: "status",
			},
			outputFormat: "json",
			expected:     jsonOutput,
		},
		{
			name:   "JQ filter ignored for non-JSON",
			output: textOutput,
			params: FilterParams{
				JQ: ".status",
			},
			outputFormat: "text",
			expected:     textOutput,
		},
		{
			name:   "YQ filter ignored for non-YAML",
			output: textOutput,
			params: FilterParams{
				YQ: ".status",
			},
			outputFormat: "text",
			expected:     textOutput,
		},
		{
			name:         "Empty output returns empty",
			output:       "",
			params:       FilterParams{Grep: "test"},
			outputFormat: "",
			expected:     "",
		},
		{
			name:         "No filters returns original",
			output:       textOutput,
			params:       FilterParams{},
			outputFormat: "",
			expected:     textOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyFilter(tt.output, tt.params, tt.outputFormat)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Normalize whitespace for comparison
				expected := strings.TrimSpace(tt.expected)
				actual := strings.TrimSpace(result)
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestApplyFilterWithCombinations(t *testing.T) {
	jsonWithErrors := `{"logs": ["Info: Starting", "Error: Failed to connect", "Info: Retrying", "Error: Timeout"], "status": "Failed"}`

	tests := []struct {
		name         string
		output       string
		params       FilterParams
		outputFormat string
		expected     string
	}{
		{
			name:   "Grep then JQ on JSON",
			output: jsonWithErrors,
			params: FilterParams{
				Grep: "status",
			},
			outputFormat: "json",
			expected:     jsonWithErrors, // Grep matches the line with "status"
		},
		{
			name:   "Multiple filters but only relevant ones apply",
			output: jsonWithErrors,
			params: FilterParams{
				Grep: "Error",
				JQ:   ".status",
				YQ:   ".invalid", // Won't apply since it's JSON
			},
			outputFormat: "json",
			expected:     `"Failed"`, // Grep matches the line (contains "Error"), then JQ extracts .status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyFilter(tt.output, tt.params, tt.outputFormat)
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.expected), strings.TrimSpace(result))
		})
	}
}
