package validate

import (
	"context"
	"fmt"
	"path/filepath"

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

	for _, ruleFilePattern := range ruleFiles {
		matches, err := filepath.Glob(ruleFilePattern)
		if err != nil {
			return fmt.Errorf("expanding glob pattern %s: %w", ruleFilePattern, err)
		}

		if len(matches) == 0 {
			matches = []string{ruleFilePattern}
		}

		for _, ruleFile := range matches {
			loadedRules, err := yaml.ParseYAMLFileToValidationRules(ruleFile)
			if err != nil {
				return fmt.Errorf("loading validation rules from %s: %w", ruleFile, err)
			}
			ruless = append(ruless, loadedRules...)
		}
	}

	if len(ruless) == 0 {
		return fmt.Errorf("no validation rules provided")
	}

	val := &validator.Validator{}
	results, err := val.Validate(ctx, files, ruless)
	if err != nil {
		return fmt.Errorf("validating input: %w", err)
	}

	return displayResults(results)
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

func displayResults(results []validator.ValidationResult) error {
	type ruleFailure struct {
		RuleName     string
		ResourceKind string
		ResourceName string
		Err          error
	}

	groupedFailures := make(map[string]map[string][]ruleFailure)
	hasFailures := false

	for _, result := range results {
		if !result.Valid {
			hasFailures = true
			if groupedFailures[result.InputFile] == nil {
				groupedFailures[result.InputFile] = make(map[string][]ruleFailure)
			}
			groupedFailures[result.InputFile][result.RuleFile] = append(
				groupedFailures[result.InputFile][result.RuleFile],
				ruleFailure{
					RuleName:     result.RuleName,
					ResourceKind: result.ResourceKind,
					ResourceName: result.ResourceName,
					Err:          result.Err,
				},
			)
		}
	}

	if !hasFailures {
		return nil
	}

	errorCount := 0
	for inputFile, ruleFiles := range groupedFailures {
		fmt.Printf("\n❌ %s:\n", inputFile)
		for ruleFile, failures := range ruleFiles {
			fmt.Printf("  From %s:\n", ruleFile)
			for _, failure := range failures {
				fmt.Printf("    - [%s] %s/%s: %v\n", failure.RuleName, failure.ResourceKind, failure.ResourceName, failure.Err)
				errorCount++
			}
		}
	}

	return fmt.Errorf("\nvalidation failed: %d errors", errorCount)
}
