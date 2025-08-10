package validate

import (
	"context"
	"fmt"

	apiv1 "github.com/RRethy/utils/celery/api/v1"
	"github.com/RRethy/utils/celery/pkg/validator"
	"github.com/RRethy/utils/celery/pkg/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Validater struct{}

func (v *Validater) Validate(
	files []string,
	celExpression string,
	ruleFiles []string,
	verbose bool,
	maxWorkers int,
	targetGroup string,
	targetVersion string,
	targetKind string,
	targetName string,
	targetNamespace string,
	targetLabelSelector string,
	targetAnnotationSelector string,
) error {
	ctx := context.Background()

	var ruless []apiv1.ValidationRules
	if celExpression != "" {
		ruless = append(ruless, createInlineValidationRule(celExpression, targetGroup, targetVersion, targetKind, targetName, targetNamespace, targetLabelSelector, targetAnnotationSelector))
	}

	for _, ruleFile := range ruleFiles {
		loadedRules, err := yaml.ParseYAMLFileToValidationRules(ruleFile)
		if err != nil {
			return fmt.Errorf("loading validation rules from %s: %w", ruleFile, err)
		}
		ruless = append(ruless, loadedRules...)
	}

	if len(ruless) == 0 {
		return fmt.Errorf("no validation rules provided")
	}

	val := &validator.Validator{}
	results, err := val.Validate(ctx, files, ruless)
	if err != nil {
		return fmt.Errorf("validating input: %w", err)
	}

	return displayResults(results, verbose)
}

func createInlineValidationRule(expression string, targetGroup string, targetVersion string, targetKind string, targetName string, targetNamespace string, targetLabelSelector string, targetAnnotationSelector string) apiv1.ValidationRules {
	rule := apiv1.ValidationRules{
		Filename: "<inline>",
		ObjectMeta: metav1.ObjectMeta{
			Name: "inline-expression",
		},
		Spec: apiv1.ValidationRulesSpec{
			Rules: []apiv1.ValidationRule{
				{
					Name:       "inline",
					Expression: expression,
					Message:    "Validation failed",
					Target: &apiv1.TargetSelector{
						Group:              targetGroup,
						Version:            targetVersion,
						Kind:               targetKind,
						Name:               targetName,
						Namespace:          targetNamespace,
						LabelSelector:      targetLabelSelector,
						AnnotationSelector: targetAnnotationSelector,
					},
				},
			},
		},
	}

	return rule
}

func displayResults(results []validator.ValidationResult, verbose bool) error {
	var hasFailures bool
	var failureMessages []string

	for _, result := range results {
		if !result.Valid {
			hasFailures = true
			msg := fmt.Sprintf("❌ [%s] %s: %v", result.RuleName, result.InputFile, result.Err)
			failureMessages = append(failureMessages, msg)
			fmt.Println(msg)
		} else if verbose {
			fmt.Printf("✅ [%s] %s: PASSED\n", result.RuleName, result.InputFile)
		}
	}

	if !hasFailures {
		fmt.Println("✅ All validations passed")
		return nil
	}

	return fmt.Errorf("validation failed: %d errors", len(failureMessages))
}
