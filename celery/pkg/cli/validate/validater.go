package validate

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	apiv1 "github.com/RRethy/kube-tools/celery/api/v1"
	"github.com/RRethy/kube-tools/celery/pkg/validator"
	"github.com/RRethy/kube-tools/celery/pkg/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type Validater struct {
	IOStreams genericiooptions.IOStreams
}

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
		return fmt.Errorf("validation failed: %w", err)
	}

	return v.displayResults(results, verbose)
}

func createInlineValidationRule(expression string, targetGroup string, targetVersion string, targetKind string, targetName string, targetNamespace string, targetLabelSelector string, targetAnnotationSelector string) apiv1.ValidationRules {
	var target *apiv1.TargetSelector
	if targetGroup != "" || targetVersion != "" || targetKind != "" || targetName != "" ||
		targetNamespace != "" || targetLabelSelector != "" || targetAnnotationSelector != "" {
		target = &apiv1.TargetSelector{
			Group:              targetGroup,
			Version:            targetVersion,
			Kind:               targetKind,
			Name:               targetName,
			Namespace:          targetNamespace,
			LabelSelector:      targetLabelSelector,
			AnnotationSelector: targetAnnotationSelector,
		}
	}

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
					Target:     target,
				},
			},
		},
	}

	return rule
}

func (v *Validater) displayResults(results []validator.ValidationResult, verbose bool) error {
	type ruleResult struct {
		RuleName     string
		ResourceKind string
		ResourceName string
		Valid        bool
		Err          error
	}

	groupedResults := make(map[string]map[string][]ruleResult)
	hasFailures := false
	totalValidations := len(results)
	failureCount := 0

	for _, result := range results {
		if !result.Valid {
			hasFailures = true
			failureCount++
		}

		if !verbose && result.Valid {
			continue
		}

		if groupedResults[result.InputFile] == nil {
			groupedResults[result.InputFile] = make(map[string][]ruleResult)
		}
		groupedResults[result.InputFile][result.RuleFile] = append(
			groupedResults[result.InputFile][result.RuleFile],
			ruleResult{
				RuleName:     result.RuleName,
				ResourceKind: result.ResourceKind,
				ResourceName: result.ResourceName,
				Valid:        result.Valid,
				Err:          result.Err,
			},
		)
	}

	if !hasFailures && !verbose {
		return nil
	}

	var inputFiles []string
	for inputFile := range groupedResults {
		inputFiles = append(inputFiles, inputFile)
	}
	sort.Strings(inputFiles)

	for _, inputFile := range inputFiles {
		ruleFiles := groupedResults[inputFile]
		fmt.Fprintf(v.IOStreams.Out, "\n%s:\n", inputFile)

		var ruleFileNames []string
		for ruleFile := range ruleFiles {
			ruleFileNames = append(ruleFileNames, ruleFile)
		}
		sort.Strings(ruleFileNames)

		for _, ruleFile := range ruleFileNames {
			results := ruleFiles[ruleFile]
			fmt.Fprintf(v.IOStreams.Out, "  From %s:\n", ruleFile)
			for _, result := range results {
				if result.Valid {
					fmt.Fprintf(v.IOStreams.Out, "    ✅ [%s] %s/%s\n", result.RuleName, result.ResourceKind, result.ResourceName)
				} else {
					fmt.Fprintf(v.IOStreams.Out, "    ❌ [%s] %s/%s: %v\n", result.RuleName, result.ResourceKind, result.ResourceName, result.Err)
				}
			}
		}
	}

	if hasFailures {
		failurePercentage := float64(failureCount) / float64(totalValidations) * 100
		return fmt.Errorf("\nvalidation failed: %d/%d checks failed (%.1f%% failure rate)", failureCount, totalValidations, failurePercentage)
	}
	return nil
}
