package generator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/RRethy/k8s-tools/kustomizelite/pkg/exec"
	"github.com/RRethy/k8s-tools/kustomizelite/pkg/maputils"
)

type Generator interface {
	Generate(path string) ([]map[string]any, error)
}

type generator struct {
	execWrapper exec.Wrapper
}

func New(opts ...Option) Generator {
	g := &generator{
		execWrapper: exec.New(),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func (g *generator) Generate(path string) ([]map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading generator file: %w", err)
	}

	var generatorConfig map[string]any
	if err := yaml.Unmarshal(content, &generatorConfig); err != nil {
		return nil, fmt.Errorf("parsing generator YAML: %w", err)
	}

	annotations, err := maputils.Get[map[string]any](generatorConfig, "metadata.annotations")
	if err != nil {
		return nil, fmt.Errorf("getting annotations: %w", err)
	}

	execPath, ok := annotations["config.kubernetes.io/function"].(string)
	if !ok {
		return nil, fmt.Errorf("no config.kubernetes.io/function annotation found")
	}

	var execConfig struct {
		Exec struct {
			Path string `yaml:"path"`
		} `yaml:"exec"`
	}
	if err := yaml.Unmarshal([]byte(execPath), &execConfig); err != nil {
		return nil, fmt.Errorf("parsing exec configuration: %w", err)
	}

	if execConfig.Exec.Path == "" {
		return nil, fmt.Errorf("no exec path specified in generator configuration")
	}

	baseDir := filepath.Dir(path)
	executablePath := filepath.Join(baseDir, execConfig.Exec.Path)

	absExecPath, err := filepath.Abs(executablePath)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path for executable: %w", err)
	}

	if _, err := os.Stat(absExecPath); err != nil {
		return nil, fmt.Errorf("executable not found at %s: %w", absExecPath, err)
	}

	cmd := g.execWrapper.Command(absExecPath)
	cmd.Dir = baseDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("executing generator %s: %w (stderr: %s)", executablePath, err, stderr.String())
	}

	output := stdout.String()
	if output == "" {
		return []map[string]any{}, nil
	}

	var resources []map[string]any
	decoder := yaml.NewDecoder(strings.NewReader(output))

	for {
		var resource map[string]any
		err := decoder.Decode(&resource)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("parsing generator output: %w", err)
		}
		if resource != nil {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}
