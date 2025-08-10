package validator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	apiv1 "github.com/RRethy/k8s-tools/celery/api/v1"
	"github.com/RRethy/k8s-tools/celery/pkg/yaml"
	"github.com/google/cel-go/cel"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

type ValidationResult struct {
	InputFile    string
	RuleFile     string
	RuleName     string
	ResourceKind string
	ResourceName string
	Valid        bool
	Err          error
}

type Rule struct {
	Filename string
	Name     string
	Message  string
	Program  cel.Program
	Target   *apiv1.TargetSelector
}

type Validator struct{}

func (v *Validator) Validate(ctx context.Context, inputFiles []string, ruless []apiv1.ValidationRules) ([]ValidationResult, error) {
	env, err := cel.NewEnv(
		cel.Variable("object", cel.DynType),
		cel.Variable("allObjects", cel.ListType(cel.DynType)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating CEL environment: %w", err)
	}

	var parsedRules []Rule
	var parseErrs []error
	for _, rules := range ruless {
		for _, rule := range rules.Spec.Rules {
			ast, issues := env.Compile(rule.Expression)
			if issues != nil && issues.Err() != nil {
				parseErrs = append(parseErrs, fmt.Errorf("compiling rule %s in file %s: %w", rule.Name, rules.Filename, issues.Err()))
			}

			prg, err := env.Program(ast)
			if err != nil {
				parseErrs = append(parseErrs, fmt.Errorf("creating converting parsed rule %s in file %s to CEL program: %w", rule.Name, rules.Filename, err))
			}

			parsedRules = append(parsedRules, Rule{
				Filename: rules.Filename,
				Name:     rule.Name,
				Message:  rule.Message,
				Program:  prg,
				Target:   rule.Target,
			})
		}
	}
	if len(parseErrs) > 0 {
		return nil, errors.Join(parseErrs...)
	}

	results := make(chan []ValidationResult, len(inputFiles))
	var wg sync.WaitGroup
	for _, file := range inputFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			results <- v.ValidateFile(ctx, f, parsedRules)
		}(file)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var collected []ValidationResult
	for fileResults := range results {
		collected = append(collected, fileResults...)
	}
	return collected, nil
}

func (v *Validator) ValidateFile(ctx context.Context, file string, rules []Rule) []ValidationResult {
	resources, err := yaml.ParseYAMLFileToUnstructured(file)
	if err != nil {
		return []ValidationResult{{
			InputFile: file,
			RuleFile:  "",
			Valid:     false,
			Err:       fmt.Errorf("reading resources from file: %w", err),
		}}
	}

	allObjects := make([]map[string]any, 0, len(resources))
	for _, resource := range resources {
		allObjects = append(allObjects, resource.Object)
	}

	type ruleEvaluation struct {
		resource *unstructured.Unstructured
		rule     Rule
	}

	var evaluations []ruleEvaluation
	for _, resource := range resources {
		for _, rule := range rules {
			if !matchesTarget(resource, rule.Target) {
				continue
			}
			evaluations = append(evaluations, ruleEvaluation{
				resource: resource,
				rule:     rule,
			})
		}
	}

	results := make(chan ValidationResult, len(evaluations))
	var wg sync.WaitGroup

	for _, eval := range evaluations {
		wg.Add(1)
		go func(r *unstructured.Unstructured, rule Rule) {
			defer wg.Done()

			resourceKind := r.GetKind()
			resourceName := r.GetName()
			if resourceName == "" {
				resourceName = "<unnamed>"
			}

			validationResult := ValidationResult{
				InputFile:    file,
				RuleFile:     rule.Filename,
				RuleName:     rule.Name,
				ResourceKind: resourceKind,
				ResourceName: resourceName,
			}

			out, _, err := rule.Program.ContextEval(ctx, map[string]any{
				"object":     r.Object,
				"allObjects": allObjects,
			})
			if err != nil {
				validationResult.Valid = false
				validationResult.Err = fmt.Errorf("evaluating rule: %w", err)
			} else if out == nil || out.Type() != cel.BoolType {
				validationResult.Valid = false
				validationResult.Err = errors.New("expression did not return a boolean")
			} else if out.Value().(bool) {
				validationResult.Valid = true
				validationResult.Err = nil
			} else {
				validationResult.Valid = false
				validationResult.Err = fmt.Errorf("%s", rule.Message)
			}

			results <- validationResult
		}(eval.resource, eval.rule)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var collected []ValidationResult
	for result := range results {
		collected = append(collected, result)
	}

	return collected
}

func matchesTarget(resource *unstructured.Unstructured, target *apiv1.TargetSelector) bool {
	if target == nil {
		return true
	}

	if target.Group != "" {
		gv := resource.GroupVersionKind()
		if gv.Group != target.Group {
			return false
		}
	}

	if target.Version != "" {
		gv := resource.GroupVersionKind()
		if gv.Version != target.Version {
			return false
		}
	}

	if target.Kind != "" {
		if resource.GetKind() != target.Kind {
			return false
		}
	}

	if target.Name != "" {
		if resource.GetName() != target.Name {
			return false
		}
	}

	if target.Namespace != "" {
		if resource.GetNamespace() != target.Namespace {
			return false
		}
	}

	if target.LabelSelector != "" {
		selector, err := labels.Parse(target.LabelSelector)
		if err != nil {
			return false
		}
		resourceLabels := labels.Set(resource.GetLabels())
		if !selector.Matches(resourceLabels) {
			return false
		}
	}

	if target.AnnotationSelector != "" {
		selector, err := labels.Parse(target.AnnotationSelector)
		if err != nil {
			return false
		}
		resourceAnnotations := labels.Set(resource.GetAnnotations())
		if !selector.Matches(resourceAnnotations) {
			return false
		}
	}

	return true
}
