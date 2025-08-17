package validate

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RRethy/k8s-tools/celery/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestValidaterValidate(t *testing.T) {
	tests := []struct {
		name              string
		files             []string
		celExpression     string
		ruleFiles         []string
		verbose           bool
		targetKind        string
		expectError       bool
		expectInOutput    []string
		notExpectInOutput []string
	}{
		{
			name: "validate with inline expression",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			celExpression: "object.spec.replicas >= 1",
			expectError:   false,
		},
		{
			name: "validate with rule file",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			ruleFiles: []string{
				filepath.Join("..", "..", "..", "fixtures", "rules", "basic-validation.yaml"),
			},
			expectError: false,
		},
		{
			name: "validate with failures",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "invalid-deployments.yaml"),
			},
			celExpression: "object.spec.replicas >= 10", // Will fail
			expectError:   true,
			expectInOutput: []string{
				"❌",
			},
		},
		{
			name: "verbose mode shows passes",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			celExpression: "true",
			verbose:       true,
			expectError:   false,
			expectInOutput: []string{
				"✅",
			},
		},
		{
			name: "non-verbose mode with all passes shows nothing",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			celExpression: "true",
			verbose:       false,
			expectError:   false,
			notExpectInOutput: []string{
				"✅",
				"❌",
			},
		},
		{
			name: "glob pattern in rule files",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			ruleFiles: []string{
				filepath.Join("..", "..", "..", "fixtures", "rules", "basic-*.yaml"),
			},
			expectError: false,
		},
		{
			name: "multiple rule files",
			files: []string{
				filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			ruleFiles: []string{
				filepath.Join("..", "..", "..", "fixtures", "rules", "basic-validation.yaml"),
				filepath.Join("..", "..", "..", "fixtures", "rules", "deployment-standards.yaml"),
			},
			expectError: false,
		},
		{
			name:        "no expression or rule file",
			files:       []string{"test.yaml"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create IOStreams with buffers to capture output
			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}

			v := &Validater{
				IOStreams: genericiooptions.IOStreams{
					In:     strings.NewReader(""),
					Out:    out,
					ErrOut: errOut,
				},
			}

			err := v.Validate(
				tt.files,
				tt.celExpression,
				tt.ruleFiles,
				tt.verbose,
				128, // maxWorkers
				"",  // targetGroup
				"",  // targetVersion
				tt.targetKind,
				"", // targetName
				"", // targetNamespace
				"", // targetLabelSelector
				"", // targetAnnotationSelector
			)

			if tt.expectError {
				assert.Error(t, err)
				// Check that error contains "validation failed" if we have failures
				if err != nil && tt.name == "validate with failures" {
					assert.Contains(t, err.Error(), "validation failed")
				}
				// Check that error contains "no validation rules" if that's the issue
				if err != nil && tt.name == "no expression or rule file" {
					assert.Contains(t, err.Error(), "no validation rules provided")
				}
			} else {
				assert.NoError(t, err)
			}

			output := out.String()

			// Check expected output
			for _, expected := range tt.expectInOutput {
				assert.Contains(t, output, expected, "expected to find '%s' in output", expected)
			}

			// Check not expected output
			for _, notExpected := range tt.notExpectInOutput {
				assert.NotContains(t, output, notExpected, "did not expect to find '%s' in output", notExpected)
			}
		})
	}
}

func TestDisplayResults(t *testing.T) {
	tests := []struct {
		name           string
		results        []validator.ValidationResult
		verbose        bool
		expectError    bool
		expectInOutput []string
	}{
		{
			name: "display failures",
			results: []validator.ValidationResult{
				{
					InputFile:    "test.yaml",
					RuleFile:     "rules.yaml",
					RuleName:     "test-rule",
					ResourceKind: "Deployment",
					ResourceName: "test-deploy",
					Valid:        false,
					Err:          assert.AnError,
				},
			},
			verbose:     false,
			expectError: true,
			expectInOutput: []string{
				"❌",
				"test-rule",
				"Deployment/test-deploy",
				"validation failed: 1/1 checks failed (100.0% failure rate)",
			},
		},
		{
			name: "display success in verbose mode",
			results: []validator.ValidationResult{
				{
					InputFile:    "test.yaml",
					RuleFile:     "rules.yaml",
					RuleName:     "test-rule",
					ResourceKind: "Service",
					ResourceName: "test-svc",
					Valid:        true,
				},
			},
			verbose:     true,
			expectError: false,
			expectInOutput: []string{
				"✅",
				"test-rule",
				"Service/test-svc",
			},
		},
		{
			name: "mixed results with percentage",
			results: []validator.ValidationResult{
				{
					InputFile:    "test.yaml",
					RuleFile:     "rules.yaml",
					RuleName:     "rule1",
					ResourceKind: "Deployment",
					ResourceName: "deploy1",
					Valid:        true,
				},
				{
					InputFile:    "test.yaml",
					RuleFile:     "rules.yaml",
					RuleName:     "rule2",
					ResourceKind: "Deployment",
					ResourceName: "deploy1",
					Valid:        false,
					Err:          assert.AnError,
				},
			},
			verbose:     true,
			expectError: true,
			expectInOutput: []string{
				"✅",
				"❌",
				"validation failed: 1/2 checks failed (50.0% failure rate)",
			},
		},
		{
			name:           "no results",
			results:        []validator.ValidationResult{},
			verbose:        false,
			expectError:    false,
			expectInOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			v := &Validater{
				IOStreams: genericiooptions.IOStreams{
					Out: out,
				},
			}

			err := v.displayResults(tt.results, tt.verbose)

			if tt.expectError {
				assert.Error(t, err)
				if err != nil {
					// Check error message contains expected text
					for _, expected := range tt.expectInOutput {
						if strings.Contains(expected, "validation failed") {
							assert.Contains(t, err.Error(), expected)
						}
					}
				}
			} else {
				assert.NoError(t, err)
			}

			output := out.String()
			for _, expected := range tt.expectInOutput {
				if !strings.Contains(expected, "validation failed") {
					assert.Contains(t, output, expected)
				}
			}
		})
	}
}

func TestCreateInlineValidationRule(t *testing.T) {
	rule := createInlineValidationRule(
		"object.spec.replicas >= 3",
		"apps",          // targetGroup
		"v1",            // targetVersion
		"Deployment",    // targetKind
		"test-deploy",   // targetName
		"production",    // targetNamespace
		"app=test",      // targetLabelSelector
		"critical=true", // targetAnnotationSelector
	)

	assert.Equal(t, "<inline>", rule.Filename)
	assert.Equal(t, "inline-expression", rule.Name)
	assert.Len(t, rule.Spec.Rules, 1)

	r := rule.Spec.Rules[0]
	assert.Equal(t, "inline", r.Name)
	assert.Equal(t, "object.spec.replicas >= 3", r.Expression)
	assert.Equal(t, "Validation failed", r.Message)

	assert.NotNil(t, r.Target)
	assert.Equal(t, "apps", r.Target.Group)
	assert.Equal(t, "v1", r.Target.Version)
	assert.Equal(t, "Deployment", r.Target.Kind)
	assert.Equal(t, "test-deploy", r.Target.Name)
	assert.Equal(t, "production", r.Target.Namespace)
	assert.Equal(t, "app=test", r.Target.LabelSelector)
	assert.Equal(t, "critical=true", r.Target.AnnotationSelector)
}

func TestValidaterGlobExpansion(t *testing.T) {
	// This test verifies that glob patterns are expanded correctly
	out := &bytes.Buffer{}
	v := &Validater{
		IOStreams: genericiooptions.IOStreams{
			Out: out,
		},
	}

	// Test with a glob that matches no files (should treat as literal)
	err := v.Validate(
		[]string{filepath.Join("..", "..", "..", "fixtures", "resources", "valid-deployment.yaml")},
		"",
		[]string{"/nonexistent/path/*.yaml"}, // Should be treated as literal filename
		false,
		128,
		"", "", "", "", "", "", "",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loading validation rules")
}

func TestValidaterSortedOutput(t *testing.T) {
	// Test that output is sorted deterministically
	results := []validator.ValidationResult{
		{
			InputFile:    "b.yaml",
			RuleFile:     "rules2.yaml",
			RuleName:     "rule1",
			ResourceKind: "Service",
			ResourceName: "svc1",
			Valid:        false,
			Err:          assert.AnError,
		},
		{
			InputFile:    "a.yaml",
			RuleFile:     "rules1.yaml",
			RuleName:     "rule2",
			ResourceKind: "Deployment",
			ResourceName: "deploy1",
			Valid:        false,
			Err:          assert.AnError,
		},
		{
			InputFile:    "a.yaml",
			RuleFile:     "rules2.yaml",
			RuleName:     "rule3",
			ResourceKind: "ConfigMap",
			ResourceName: "cm1",
			Valid:        false,
			Err:          assert.AnError,
		},
	}

	out := &bytes.Buffer{}
	v := &Validater{
		IOStreams: genericiooptions.IOStreams{
			Out: out,
		},
	}

	err := v.displayResults(results, false)
	require.Error(t, err) // Should error because there are failures

	output := out.String()

	// Verify files are sorted (a.yaml should come before b.yaml)
	aIndex := strings.Index(output, "a.yaml")
	bIndex := strings.Index(output, "b.yaml")
	assert.True(t, aIndex < bIndex, "a.yaml should appear before b.yaml in output")

	// Verify rule files are sorted within each input file
	lines := strings.Split(output, "\n")
	var currentFile string
	var ruleFiles []string

	for _, line := range lines {
		if strings.HasSuffix(line, ".yaml:") {
			currentFile = line
			ruleFiles = []string{}
		} else if strings.Contains(line, "From ") {
			ruleFile := strings.TrimSpace(strings.TrimPrefix(line, "From "))
			ruleFile = strings.TrimSuffix(ruleFile, ":")
			if len(ruleFiles) > 0 {
				// Check that this rule file comes after the previous one alphabetically
				assert.True(t, ruleFile >= ruleFiles[len(ruleFiles)-1],
					"rule files should be sorted within %s", currentFile)
			}
			ruleFiles = append(ruleFiles, ruleFile)
		}
	}
}
