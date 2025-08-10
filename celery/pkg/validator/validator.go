package validator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	apiv1 "github.com/RRethy/utils/celery/api/v1"
	"github.com/RRethy/utils/celery/pkg/yaml"
	"github.com/google/cel-go/cel"
)

type ValidationResult struct {
	InputFile string
	RuleFile  string
	RuleName  string
	Valid     bool
	Err       error
}

type Rule struct {
	Filename string
	Name     string
	Program  cel.Program
}

type Validator struct{}

func (v *Validator) Validate(ctx context.Context, inputFiles []string, ruless []apiv1.ValidationRules) ([]ValidationResult, error) {
	env, err := cel.NewEnv(
		cel.Variable("object", cel.DynType),
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
				parseErrs = append(parseErrs, fmt.Errorf("compiling rule %s in file %s: %w", rule.Name, rules.Name, issues.Err()))
			}

			prg, err := env.Program(ast)
			if err != nil {
				parseErrs = append(parseErrs, fmt.Errorf("creating converting parsed rule %s in file %s to CEL program: %w", rule.Name, rules.Name, err))
			}

			parsedRules = append(parsedRules, Rule{
				Filename: rules.Name,
				Name:     rule.Name,
				Program:  prg,
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

	var results []ValidationResult
	for _, resource := range resources {
		for _, rule := range rules {
			validationResult := ValidationResult{
				InputFile: file,
				RuleFile:  rule.Filename,
				RuleName:  rule.Name,
			}
			out, _, err := rule.Program.ContextEval(ctx, map[string]any{
				"object":     resource.Object,
				"allObjects": allObjects,
			})
			if err != nil {
				validationResult.Valid = false
				validationResult.Err = fmt.Errorf("evaluating rule: %w", err)
			} else if out == nil || out.Type() != cel.BoolType || out.Value().(bool) {
				validationResult.Valid = true
				validationResult.Err = nil
			} else {
				validationResult.Valid = true
				validationResult.Err = errors.New("evaluated to false")
			}
			results = append(results, validationResult)
		}
	}

	return results
}
