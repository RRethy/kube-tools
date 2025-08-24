package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	v1 "github.com/RRethy/kube-tools/kustomizelite/api/v1"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/exec"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/kustomize"
)

func Build(ctx context.Context, path string, helmValuesFiles []string) error {
	k, err := kustomize.NewKustomize(helmValuesFiles)
	if err != nil {
		return err
	}
	return (&Builder{
		IOStreams: genericiooptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		kustomizer:         k,
		useKustomizeFormat: true,
	}).Build(ctx, path)
}

func Batch(ctx context.Context, batchFile string, globalHelmValuesFiles []string) error {
	// Read batch file
	content, err := os.ReadFile(batchFile)
	if err != nil {
		return fmt.Errorf("reading batch file: %w", err)
	}

	var batch v1.BatchBuild
	if err := yaml.Unmarshal(content, &batch); err != nil {
		return fmt.Errorf("parsing batch file: %w", err)
	}

	// Validate
	if batch.APIVersion != "adamize.io/v1" {
		return fmt.Errorf("unsupported apiVersion: %s", batch.APIVersion)
	}
	if batch.Kind != "BatchBuild" {
		return fmt.Errorf("unsupported kind: %s", batch.Kind)
	}

	// Set global environment variables
	globalEnv := make(map[string]string)
	for _, env := range batch.Env {
		globalEnv[env.Name] = env.Value
		os.Setenv(env.Name, env.Value)
	}

	// Create error channel and wait group
	errCh := make(chan error, len(batch.Builds))
	var wg sync.WaitGroup

	// Create a semaphore to limit concurrent goroutines to 128
	const maxConcurrency = 128
	sem := make(chan struct{}, maxConcurrency)

	// Process builds in parallel with concurrency limit
	for _, build := range batch.Builds {
		wg.Add(1)
		go func(b v1.BuildConfig) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }() // Release semaphore when done

			if err := processBuild(ctx, b, globalEnv, globalHelmValuesFiles); err != nil {
				errCh <- fmt.Errorf("building %s: %w", b.Kustomization, err)
			}
		}(build)
	}

	// Wait for all builds to complete
	wg.Wait()
	close(errCh)

	// Collect errors
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch build failed with %d errors:\n%v", len(errors), errors)
	}

	return nil
}

func processBuild(ctx context.Context, build v1.BuildConfig, globalEnv map[string]string, globalHelmValuesFiles []string) error {
	// Merge environment variables (build-specific overrides global)
	env := make(map[string]string)
	for k, v := range globalEnv {
		env[k] = v
	}
	for _, e := range build.Env {
		env[e.Name] = e.Value
	}

	// Create a new exec wrapper with the merged environment
	execWrapper := exec.NewWithEnv(env)

	// Create kustomizer with the environment-aware exec wrapper
	k, err := kustomize.NewKustomize(globalHelmValuesFiles, kustomize.WithExecWrapper(execWrapper))
	if err != nil {
		return fmt.Errorf("creating kustomizer: %w", err)
	}

	// Build to a buffer first
	builder := &Builder{
		IOStreams: genericiooptions.IOStreams{
			In:     os.Stdin,
			Out:    nil, // Will be set to buffer
			ErrOut: os.Stderr,
		},
		kustomizer:         k,
		useKustomizeFormat: true,
	}

	// Create output directory if needed
	outputDir := filepath.Dir(build.Output)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(build.Output)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	// Set output to file
	builder.IOStreams.Out = f

	// Run the build
	if err := builder.Build(ctx, build.Kustomization); err != nil {
		return fmt.Errorf("building kustomization: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Built %s -> %s\n", build.Kustomization, build.Output)
	return nil
}
