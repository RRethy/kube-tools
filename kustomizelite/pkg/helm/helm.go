package helm

import (
	"errors"
	"fmt"
	"os"
	stdexec "os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "github.com/RRethy/k8s-tools/kustomizelite/api/v1"
	"github.com/RRethy/k8s-tools/kustomizelite/pkg/exec"
)

// GetBinaryFromEnv returns the helm binary path from HELM_BINARY_PATH env var, or "helm" if not set.
func GetBinaryFromEnv() string {
	if helmPath := os.Getenv("HELM_BINARY_PATH"); helmPath != "" {
		return helmPath
	}
	return "helm"
}

// Templater inflates Helm charts using the helm CLI.
type Templater interface {
	Template(baseDir string, chart v1.HelmChart, globals *v1.HelmGlobals) ([]map[string]any, error)
}

type templater struct {
	helmBinary        string
	globalValuesFiles []string
	execWrapper       exec.Wrapper
}

// NewTemplater creates a new Helm templater with the specified helm binary path and global values files.
func NewTemplater(helmBinary string, globalValuesFiles []string) (Templater, error) {
	return NewTemplaterWithExec(helmBinary, globalValuesFiles, nil)
}

// NewTemplaterWithExec creates a new Helm templater with custom exec wrapper.
func NewTemplaterWithExec(helmBinary string, globalValuesFiles []string, wrapper exec.Wrapper) (Templater, error) {
	if helmBinary == "" {
		helmBinary = "helm"
	}

	// Resolve the helm binary path to ensure it exists and is executable
	resolvedBinary, err := stdexec.LookPath(helmBinary)
	if err != nil {
		return nil, fmt.Errorf("helm binary not found: %w", err)
	}

	absoluteGlobalValuesFiles := make([]string, len(globalValuesFiles))
	for i, file := range globalValuesFiles {
		absPath, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("resolving absolute path for %s: %w", file, err)
		}

		if _, err := os.Stat(absPath); err != nil {
			return nil, fmt.Errorf("global values file %s does not exist: %w", absPath, err)
		}

		absoluteGlobalValuesFiles[i] = absPath
	}

	if wrapper == nil {
		wrapper = exec.New()
	}

	return &templater{
		helmBinary:        resolvedBinary,
		globalValuesFiles: absoluteGlobalValuesFiles,
		execWrapper:       wrapper,
	}, nil
}

func (t *templater) Template(baseDir string, chart v1.HelmChart, globals *v1.HelmGlobals) ([]map[string]any, error) {
	args := []string{"template"}

	if chart.ReleaseName != "" {
		args = append(args, chart.ReleaseName)
	} else {
		args = append(args, chart.Name)
	}

	chartPath := filepath.Join("charts", chart.Name)
	if globals != nil && globals.ChartHome != "" {
		chartPath = filepath.Join(globals.ChartHome, chart.Name)
	}
	args = append(args, chartPath)

	if chart.ValuesFile != "" {
		args = append(args, "--values", chart.ValuesFile)
	}

	for _, additionalValuesFile := range chart.AdditionalValuesFiles {
		args = append(args, "--values", additionalValuesFile)
	}

	for _, globalValuesFile := range t.globalValuesFiles {
		args = append(args, "--values", globalValuesFile)
	}

	cmd := t.execWrapper.Command(t.helmBinary, args...)
	cmd.Dir = baseDir

	output, err := cmd.Output()
	if err != nil {
		var exitErr *stdexec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("helm template failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("executing helm template: %w", err)
	}

	return parseHelmOutput(output)
}

func parseHelmOutput(output []byte) ([]map[string]any, error) {
	documents := strings.Split(string(output), "\n---\n")
	var resources []map[string]any

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" || doc == "---" {
			continue
		}

		var resource map[string]any
		if err := yaml.Unmarshal([]byte(doc), &resource); err != nil {
			return nil, fmt.Errorf("parsing helm output document: %w", err)
		}

		if len(resource) > 0 {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}
