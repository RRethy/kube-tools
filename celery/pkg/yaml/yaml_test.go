package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYAMLToUnstructured(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "single deployment",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 3`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multi-document yaml",
			input: `
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "empty document",
			input: `
---
---`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "invalid yaml",
			input:     `{invalid yaml: [}`,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := ParseYAMLToUnstructured([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resources, tt.wantCount)

			// Verify resource properties if we got resources
			if len(resources) > 0 {
				for _, r := range resources {
					assert.NotEmpty(t, r.GetKind())
					assert.NotNil(t, r.Object)
				}
			}
		})
	}
}

func TestParseYAMLToValidationRules(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		filename  string
		wantCount int
		wantErr   bool
	}{
		{
			name: "single validation rules",
			input: `
apiVersion: celery.rrethy.io/v1
kind: ValidationRules
metadata:
  name: test-rules
spec:
  rules:
    - name: test-rule
      expression: "true"
      message: "Test message"`,
			filename:  "test.yaml",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple validation rules",
			input: `
apiVersion: celery.rrethy.io/v1
kind: ValidationRules
metadata:
  name: rules-1
spec:
  rules:
    - name: rule-1
      expression: "true"
---
apiVersion: celery.rrethy.io/v1
kind: ValidationRules
metadata:
  name: rules-2
spec:
  rules:
    - name: rule-2
      expression: "false"`,
			filename:  "multi.yaml",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "mixed resources with validation rules",
			input: `
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: celery.rrethy.io/v1
kind: ValidationRules
metadata:
  name: test-rules
spec:
  rules:
    - name: test-rule
      expression: "true"`,
			filename:  "mixed.yaml",
			wantCount: 1, // Only ValidationRules should be returned
			wantErr:   false,
		},
		{
			name: "no validation rules",
			input: `
apiVersion: v1
kind: Service
metadata:
  name: test-service`,
			filename:  "no-rules.yaml",
			wantCount: 0,
			wantErr:   true, // Should error when no ValidationRules found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := ParseYAMLToValidationRules([]byte(tt.input), tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, rules, tt.wantCount)

			// Verify filename is set correctly
			for _, r := range rules {
				assert.Equal(t, tt.filename, r.Filename)
				assert.Equal(t, "ValidationRules", r.Kind)
				assert.NotEmpty(t, r.Name)
			}
		})
	}
}

func TestParseYAMLBytes(t *testing.T) {
	type TestStruct struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
	}

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "parse generic structs",
			input: `
apiVersion: v1
kind: Service
metadata:
  name: service-1
---
apiVersion: v1
kind: Service
metadata:
  name: service-2`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "invalid yaml",
			input:     `{invalid: [}`,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := ParseYAMLBytes[TestStruct]([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, items, tt.wantCount)

			// Verify parsed data
			for i, item := range items {
				assert.Equal(t, "v1", item.APIVersion)
				assert.Equal(t, "Service", item.Kind)
				assert.Equal(t, fmt.Sprintf("service-%d", i+1), item.Metadata.Name)
			}
		})
	}
}

func TestParseYAMLFileToUnstructured(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Test successful parsing
	resources, err := ParseYAMLFileToUnstructured(testFile)
	require.NoError(t, err)
	assert.Len(t, resources, 1)
	assert.Equal(t, "Deployment", resources[0].GetKind())
	assert.Equal(t, "test-deployment", resources[0].GetName())

	// Test non-existent file
	_, err = ParseYAMLFileToUnstructured("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}

func TestParseYAMLFileToValidationRules(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "rules.yaml")

	content := `
apiVersion: celery.rrethy.io/v1
kind: ValidationRules
metadata:
  name: test-rules
spec:
  rules:
    - name: test-rule
      expression: "true"
      message: "Test message"`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Test successful parsing
	rules, err := ParseYAMLFileToValidationRules(testFile)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.Equal(t, testFile, rules[0].Filename)
	assert.Equal(t, "test-rules", rules[0].Name)

	// Test non-existent file
	_, err = ParseYAMLFileToValidationRules("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}

func TestParseYAMLWithEmptyDocuments(t *testing.T) {
	// Test handling of empty documents and whitespace
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name:      "empty documents with separators",
			input:     "---\n---\n---",
			wantCount: 0,
		},
		{
			name: "mixed empty and valid documents",
			input: `---
---
apiVersion: v1
kind: Service
metadata:
  name: test
---
---`,
			wantCount: 1,
		},
		{
			name:      "only whitespace",
			input:     "   \n\n   \n",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := ParseYAMLToUnstructured([]byte(tt.input))
			require.NoError(t, err)
			assert.Len(t, resources, tt.wantCount)
		})
	}
}
