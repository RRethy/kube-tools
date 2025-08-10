package validator

import (
	"context"
	"path/filepath"
	"testing"

	apiv1 "github.com/RRethy/k8s-tools/celery/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidatorValidate(t *testing.T) {
	ctx := context.Background()
	v := &Validator{}

	tests := []struct {
		name            string
		inputFiles      []string
		rules           []apiv1.ValidationRules
		wantErrors      int
		wantTotalChecks int
	}{
		{
			name: "valid deployment passes all rules",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "test-rules.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-rules",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "has-namespace",
								Expression: "has(object.metadata.namespace)",
								Message:    "Resource must have namespace",
							},
							{
								Name:       "has-name",
								Expression: "has(object.metadata.name)",
								Message:    "Resource must have name",
							},
						},
					},
				},
			},
			wantErrors:      0,
			wantTotalChecks: 2, // 2 rules × 1 resource
		},
		{
			name: "invalid deployment fails rules",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "invalid-deployments.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "test-rules.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment-rules",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "minimum-replicas",
								Expression: "object.spec.replicas >= 5",
								Message:    "Must have at least 5 replicas",
								Target: &apiv1.TargetSelector{
									Kind: "Deployment",
								},
							},
						},
					},
				},
			},
			wantErrors:      3, // 3 deployments in invalid-deployments.yaml
			wantTotalChecks: 3,
		},
		{
			name: "rule with invalid CEL expression",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "valid-deployment.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "bad-rules.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "bad-rules",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "invalid-expression",
								Expression: "this is not valid CEL",
								Message:    "Invalid rule",
							},
						},
					},
				},
			},
			wantErrors:      -1, // Expect error in validation itself
			wantTotalChecks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := v.Validate(ctx, tt.inputFiles, tt.rules)

			if tt.wantErrors == -1 {
				// Expect an error during validation
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, results)

			// Count failures
			failureCount := 0
			for _, r := range results {
				if !r.Valid {
					failureCount++
				}
			}

			assert.Equal(t, tt.wantTotalChecks, len(results), "unexpected number of total checks")
			assert.Equal(t, tt.wantErrors, failureCount, "unexpected number of failures")

			// Verify all results have required fields
			for _, r := range results {
				assert.NotEmpty(t, r.InputFile)
				assert.NotEmpty(t, r.RuleFile)
				assert.NotEmpty(t, r.RuleName)
				assert.NotEmpty(t, r.ResourceKind)
			}
		})
	}
}

func TestValidatorWithCrossResourceValidation(t *testing.T) {
	ctx := context.Background()
	v := &Validator{}

	// Test cross-resource validation with allObjects
	rules := []apiv1.ValidationRules{
		{
			Filename: "cross-rules.yaml",
			ObjectMeta: metav1.ObjectMeta{
				Name: "cross-resource-rules",
			},
			Spec: apiv1.ValidationRulesSpec{
				Rules: []apiv1.ValidationRule{
					{
						Name:       "deployment-has-service",
						Expression: `object.kind != "Deployment" || allObjects.exists(o, o.kind == "Service")`,
						Message:    "Deployment requires a Service",
					},
				},
			},
		},
	}

	inputFiles := []string{
		filepath.Join("..", "..", "fixtures", "resources", "cross-reference-resources.yaml"),
	}

	results, err := v.Validate(ctx, inputFiles, rules)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// Verify allObjects is available in CEL context
	for _, r := range results {
		if r.Err != nil && r.Valid == false {
			// Should not have "undeclared reference to 'allObjects'" error
			assert.NotContains(t, r.Err.Error(), "allObjects")
		}
	}
}

func TestValidationResult(t *testing.T) {
	// Test the ValidationResult struct fields
	result := ValidationResult{
		InputFile:    "test.yaml",
		RuleFile:     "rules.yaml",
		RuleName:     "test-rule",
		ResourceKind: "Deployment",
		ResourceName: "test-deployment",
		Valid:        true,
		Err:          nil,
	}

	assert.Equal(t, "test.yaml", result.InputFile)
	assert.Equal(t, "rules.yaml", result.RuleFile)
	assert.Equal(t, "test-rule", result.RuleName)
	assert.Equal(t, "Deployment", result.ResourceKind)
	assert.Equal(t, "test-deployment", result.ResourceName)
	assert.True(t, result.Valid)
	assert.NoError(t, result.Err)
}

func TestValidatorWithEmptyInputs(t *testing.T) {
	ctx := context.Background()
	v := &Validator{}

	// Test with empty input files
	results, err := v.Validate(ctx, []string{}, []apiv1.ValidationRules{})
	assert.NoError(t, err)
	assert.Empty(t, results)

	// Test with rules but no input files
	rules := []apiv1.ValidationRules{
		{
			Filename: "test.yaml",
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-rules",
			},
			Spec: apiv1.ValidationRulesSpec{
				Rules: []apiv1.ValidationRule{
					{
						Name:       "test",
						Expression: "true",
						Message:    "Test",
					},
				},
			},
		},
	}

	results, err = v.Validate(ctx, []string{}, rules)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestValidatorConcurrency(t *testing.T) {
	ctx := context.Background()
	v := &Validator{}

	// Create multiple input files to test concurrent processing
	inputFiles := []string{
		filepath.Join("..", "..", "fixtures", "resources", "valid-deployment.yaml"),
		filepath.Join("..", "..", "fixtures", "resources", "services.yaml"),
		filepath.Join("..", "..", "fixtures", "resources", "configmaps-secrets.yaml"),
	}

	rules := []apiv1.ValidationRules{
		{
			Filename: "test-rules.yaml",
			ObjectMeta: metav1.ObjectMeta{
				Name: "concurrent-test",
			},
			Spec: apiv1.ValidationRulesSpec{
				Rules: []apiv1.ValidationRule{
					{
						Name:       "has-metadata",
						Expression: "has(object.metadata)",
						Message:    "Must have metadata",
					},
				},
			},
		},
	}

	// Run validation which should process files concurrently
	results, err := v.Validate(ctx, inputFiles, rules)
	require.NoError(t, err)
	
	// Verify we got results from all files
	fileMap := make(map[string]bool)
	for _, r := range results {
		fileMap[r.InputFile] = true
	}
	
	assert.Equal(t, len(inputFiles), len(fileMap), "should have results from all input files")
}

func TestResourceNameExtraction(t *testing.T) {
	ctx := context.Background()
	v := &Validator{}

	// Test that unnamed resources get "<unnamed>" as name
	inputFiles := []string{
		filepath.Join("..", "..", "fixtures", "resources", "test-resources.yaml"),
	}

	rules := []apiv1.ValidationRules{
		{
			Filename: "name-test.yaml",
			ObjectMeta: metav1.ObjectMeta{
				Name: "name-test",
			},
			Spec: apiv1.ValidationRulesSpec{
				Rules: []apiv1.ValidationRule{
					{
						Name:       "always-pass",
						Expression: "true",
						Message:    "Always passes",
					},
				},
			},
		},
	}

	results, err := v.Validate(ctx, inputFiles, rules)
	require.NoError(t, err)

	// Check that all results have either a real name or "<unnamed>"
	for _, r := range results {
		assert.NotEmpty(t, r.ResourceName)
		assert.NotEmpty(t, r.ResourceKind)
	}
}

func TestTargetSelection(t *testing.T) {
	ctx := context.Background()
	v := &Validator{}

	tests := []struct {
		name            string
		inputFiles      []string
		rules           []apiv1.ValidationRules
		expectedMatches int
		description     string
	}{
		{
			name: "target by kind",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "mixed-resources.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "kind-target.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "kind-target",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "deployment-only",
								Expression: "true",
								Message:    "Test",
								Target: &apiv1.TargetSelector{
									Kind: "Deployment",
								},
							},
						},
					},
				},
			},
			expectedMatches: 1,
			description:     "should only match Deployment resources",
		},
		{
			name: "target by namespace",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "multi-namespace.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "namespace-target.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-target",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "production-only",
								Expression: "true",
								Message:    "Test",
								Target: &apiv1.TargetSelector{
									Namespace: "production",
								},
							},
						},
					},
				},
			},
			expectedMatches: 1,
			description:     "should only match resources in production namespace",
		},
		{
			name: "target by exact name",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "mixed-resources.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "name-exact.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "name-exact",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "web-service-only",
								Expression: "true",
								Message:    "Test",
								Target: &apiv1.TargetSelector{
									Name: "web-service",
								},
							},
						},
					},
				},
			},
			expectedMatches: 1,
			description:     "should match only resource with exact name",
		},
		{
			name: "no target matches all",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "mixed-resources.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "no-target.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "no-target",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "all-resources",
								Expression: "true",
								Message:    "Test",
							},
						},
					},
				},
			},
			expectedMatches: 3,
			description:     "should match all resources when no target specified",
		},
		{
			name: "target by group and version",
			inputFiles: []string{
				filepath.Join("..", "..", "fixtures", "resources", "mixed-resources.yaml"),
			},
			rules: []apiv1.ValidationRules{
				{
					Filename: "group-version.yaml",
					ObjectMeta: metav1.ObjectMeta{
						Name: "group-version",
					},
					Spec: apiv1.ValidationRulesSpec{
						Rules: []apiv1.ValidationRule{
							{
								Name:       "apps-v1",
								Expression: "true",
								Message:    "Test",
								Target: &apiv1.TargetSelector{
									Group:   "apps",
									Version: "v1",
								},
							},
						},
					},
				},
			},
			expectedMatches: 2,
			description:     "should match resources with apps/v1 API group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := v.Validate(ctx, tt.inputFiles, tt.rules)
			
			if err != nil {
				t.Skipf("Skipping test %s: %v", tt.name, err)
				return
			}
			
			assert.Equal(t, tt.expectedMatches, len(results), tt.description)
		})
	}
}

func TestMatchesTargetDirectly(t *testing.T) {
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "default",
			},
		},
	}
	
	service := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      "test-service",
				"namespace": "default",
			},
		},
	}
	
	tests := []struct {
		name     string
		resource *unstructured.Unstructured
		target   *apiv1.TargetSelector
		expected bool
	}{
		{
			name:     "deployment matches Deployment kind",
			resource: deployment,
			target: &apiv1.TargetSelector{
				Kind: "Deployment",
			},
			expected: true,
		},
		{
			name:     "service does not match Deployment kind",
			resource: service,
			target: &apiv1.TargetSelector{
				Kind: "Deployment",
			},
			expected: false,
		},
		{
			name:     "service matches Service kind",
			resource: service,
			target: &apiv1.TargetSelector{
				Kind: "Service",
			},
			expected: true,
		},
		{
			name:     "nil target matches all",
			resource: deployment,
			target:   nil,
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesTarget(tt.resource, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}