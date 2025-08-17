package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// FilterParams holds filtering configuration for output processing
type FilterParams struct {
	Grep string
	JQ   string
	YQ   string
}

// GetFilterParams extracts filter parameters from request arguments
func GetFilterParams(args map[string]any) FilterParams {
	params := FilterParams{}

	if val, ok := args["grep"].(string); ok && val != "" {
		params.Grep = val
	}
	if val, ok := args["jq"].(string); ok && val != "" {
		params.JQ = val
	}
	if val, ok := args["yq"].(string); ok && val != "" {
		params.YQ = val
	}

	return params
}

// ApplyFilter applies the appropriate filter to the output based on params and format
func ApplyFilter(output string, params FilterParams, outputFormat string) (string, error) {
	if output == "" {
		return output, nil
	}

	if params.Grep != "" {
		filtered, err := applyGrepFilter(output, params.Grep)
		if err != nil {
			return output, fmt.Errorf("grep filter failed: %w", err)
		}
		output = filtered
	}

	if params.JQ != "" && outputFormat == "json" {
		filtered, err := applyJQFilter(output, params.JQ)
		if err != nil {
			return output, fmt.Errorf("jq filter failed: %w", err)
		}
		output = filtered
	}

	if params.YQ != "" && outputFormat == "yaml" {
		filtered, err := applyYQFilter(output, params.YQ)
		if err != nil {
			return output, fmt.Errorf("yq filter failed: %w", err)
		}
		output = filtered
	}

	return output, nil
}

// applyGrepFilter filters output lines matching the pattern
func applyGrepFilter(output, pattern string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return applyLiteralGrepFilter(output, pattern), nil
	}

	lines := strings.Split(output, "\n")
	var filtered []string

	for _, line := range lines {
		if re.MatchString(line) {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n"), nil
}

// applyLiteralGrepFilter filters output lines containing the literal string
func applyLiteralGrepFilter(output, pattern string) string {
	lines := strings.Split(output, "\n")
	var filtered []string

	for _, line := range lines {
		if strings.Contains(line, pattern) {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

// applyJQFilter applies a jq filter to JSON output
func applyJQFilter(output, filter string) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		return "", fmt.Errorf("invalid JSON input: %w", err)
	}

	if _, err := exec.LookPath("jq"); err != nil {
		return applySimpleJSONFilter(output, filter)
	}

	cmd := exec.Command("jq", filter)
	cmd.Stdin = strings.NewReader(output)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("jq execution failed: %s", errBuf.String())
	}

	return outBuf.String(), nil
}

// applyYQFilter applies a yq filter to YAML output
func applyYQFilter(output, filter string) (string, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(output), &yamlData); err != nil {
		return "", fmt.Errorf("invalid YAML input: %w", err)
	}

	if _, err := exec.LookPath("yq"); err != nil {
		return applySimpleYAMLFilter(output, filter)
	}

	cmd := exec.Command("yq", filter)
	cmd.Stdin = strings.NewReader(output)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yq execution failed: %s", errBuf.String())
	}

	return outBuf.String(), nil
}

// applySimpleJSONFilter provides basic JSONPath-like filtering when jq is not available
func applySimpleJSONFilter(output, filter string) (string, error) {
	if !strings.HasPrefix(filter, ".") {
		return "", fmt.Errorf("filter must start with '.'")
	}

	var data interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return "", err
	}

	parts := strings.Split(strings.TrimPrefix(filter, "."), ".")
	result := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := result.(type) {
		case map[string]interface{}:
			var ok bool
			result, ok = v[part]
			if !ok {
				return "null", nil
			}
		case []interface{}:
			if part == "[]" {
				jsonBytes, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					return "", err
				}
				return string(jsonBytes), nil
			}
			return "", fmt.Errorf("array indexing not supported in simple filter")
		default:
			return "null", nil
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// applySimpleYAMLFilter provides basic YAML filtering when yq is not available
func applySimpleYAMLFilter(output, filter string) (string, error) {
	if !strings.HasPrefix(filter, ".") {
		return "", fmt.Errorf("filter must start with '.'")
	}

	var data interface{}
	if err := yaml.Unmarshal([]byte(output), &data); err != nil {
		return "", err
	}

	parts := strings.Split(strings.TrimPrefix(filter, "."), ".")
	result := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := result.(type) {
		case map[string]interface{}:
			var ok bool
			result, ok = v[part]
			if !ok {
				return "null", nil
			}
		case []interface{}:
			if part == "[]" {
				yamlBytes, err := yaml.Marshal(v)
				if err != nil {
					return "", err
				}
				return string(yamlBytes), nil
			}
			return "", fmt.Errorf("array indexing not supported in simple filter")
		default:
			return "null", nil
		}
	}

	yamlBytes, err := yaml.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}
