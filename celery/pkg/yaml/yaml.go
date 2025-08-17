package yaml

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	apiv1 "github.com/RRethy/k8s-tools/celery/api/v1"
	goyaml "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ParseYAMLToUnstructured parses multi-document YAML into unstructured Kubernetes objects.
// It handles both single and multi-document YAML files.
func ParseYAMLToUnstructured(data []byte) ([]*unstructured.Unstructured, error) {
	var resources []*unstructured.Unstructured

	decoder := goyaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var obj map[string]any
		err := decoder.Decode(&obj)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding YAML: %w", err)
		}

		if obj == nil {
			continue
		}

		u := &unstructured.Unstructured{Object: obj}
		resources = append(resources, u)
	}

	return resources, nil
}

// ParseYAMLToValidationRules parses multi-document YAML into ValidationRules.
// It handles both single and multi-document YAML files containing ValidationRules resources.
func ParseYAMLToValidationRules(data []byte, filename string) ([]apiv1.ValidationRules, error) {
	var ruless []apiv1.ValidationRules

	decoder := goyaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var rules apiv1.ValidationRules
		err := decoder.Decode(&rules)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding YAML: %w", err)
		}

		if rules.Kind == "ValidationRules" {
			rules.Filename = filename
			ruless = append(ruless, rules)
		}
	}

	if len(ruless) == 0 {
		return nil, fmt.Errorf("no ValidationRules resources found in file")
	}

	return ruless, nil
}

// ParseYAMLBytes is a generic YAML parser that can decode multi-document YAML
// into a slice of the specified type.
func ParseYAMLBytes[T any](data []byte) ([]T, error) {
	var items []T

	decoder := goyaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var item T
		err := decoder.Decode(&item)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding YAML: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// ParseYAMLFileToUnstructured reads a YAML file and parses it into unstructured Kubernetes objects.
func ParseYAMLFileToUnstructured(file string) ([]*unstructured.Unstructured, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return ParseYAMLToUnstructured(data)
}

// ParseYAMLFileToValidationRules reads a YAML file and parses it into ValidationRules.
func ParseYAMLFileToValidationRules(file string) ([]apiv1.ValidationRules, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return ParseYAMLToValidationRules(data, file)
}
