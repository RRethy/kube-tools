package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/k8s-tools/kustomizelite/pkg/kustomize"
)

type Builder struct {
	IOStreams          genericiooptions.IOStreams
	kustomizer         kustomize.Kustomizer
	useKustomizeFormat bool
}

func (b *Builder) Build(_ context.Context, path string) error {
	resources, err := b.kustomizer.Kustomize(path, nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// If kustomize formatting is disabled, output directly
	if !b.useKustomizeFormat {
		return b.directOutput(resources)
	}

	// Create a temporary directory for the resources
	tmpDir, err := os.MkdirTemp("", "adamize-build-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write resources to temporary files
	var resourceFiles []string
	for i, resource := range resources {
		filename := filepath.Join(tmpDir, fmt.Sprintf("resource-%d.yaml", i))
		resourceFiles = append(resourceFiles, filepath.Base(filename))

		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(resource); err != nil {
			return fmt.Errorf("marshaling resource %d: %w", i, err)
		}

		if err := os.WriteFile(filename, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("writing resource %d: %w", i, err)
		}
	}

	// Create a kustomization.yaml that references all resources
	kustomization := map[string]interface{}{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"resources":  resourceFiles,
	}

	kustomizationBytes, err := yaml.Marshal(kustomization)
	if err != nil {
		return fmt.Errorf("marshaling kustomization.yaml: %w", err)
	}

	kustomizationPath := filepath.Join(tmpDir, "kustomization.yaml")
	if err := os.WriteFile(kustomizationPath, kustomizationBytes, 0o644); err != nil {
		return fmt.Errorf("writing kustomization.yaml: %w", err)
	}

	// Run kubectl kustomize on the temp directory
	cmd := exec.Command("kubectl", "kustomize", tmpDir)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubectl kustomize failed: %w\nstderr: %s", err, exitErr.Stderr)
		}
		return fmt.Errorf("running kubectl kustomize: %w", err)
	}

	// Write the formatted output
	_, err = io.Copy(b.IOStreams.Out, bytes.NewReader(output))
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func (b *Builder) directOutput(resources []map[string]any) error {
	for i, resource := range resources {
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(resource); err != nil {
			return fmt.Errorf("marshaling resource %d: %w", i, err)
		}

		if i > 0 {
			_, _ = fmt.Fprintln(b.IOStreams.Out, "---")
		}
		_, _ = fmt.Fprint(b.IOStreams.Out, buf.String())
	}

	return nil
}
