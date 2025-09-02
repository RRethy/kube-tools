package build

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/RRethy/kube-tools/k2/pkg/hydrate"
	k2strings "github.com/RRethy/kube-tools/k2/pkg/strings"
)

type Builder struct {
	IoStreams genericiooptions.IOStreams
	Hydrator  hydrate.Hydrator
}

func (b *Builder) Build(ctx context.Context, paths []string, outDir string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	if len(paths) > 1 && outDir == "" {
		return fmt.Errorf("--out-dir is required when building multiple paths")
	}

	nodess, err := b.hydratePaths(ctx, paths)
	if err != nil {
		return fmt.Errorf("hydrating paths: %w", err)
	}

	if outDir == "" {
		return b.writeToStdout(nodess[0])
	}

	if err = os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	absPaths, err := b.toAbsPaths(paths)
	if err != nil {
		return err
	}

	commonPrefix := k2strings.FindCommonPrefix(absPaths)

	return b.writeNodess(nodess, absPaths, commonPrefix, outDir)
}

func (b *Builder) hydratePaths(ctx context.Context, paths []string) ([][]*kyaml.RNode, error) {
	type result struct {
		index int
		nodes []*kyaml.RNode
		err   error
	}

	results := make(chan result, len(paths))
	var wg sync.WaitGroup
	for i, p := range paths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()
			hydrated, err := b.Hydrator.Hydrate(ctx, path, nil)
			if err != nil {
				results <- result{index: index, err: fmt.Errorf("hydrating %s: %w", path, err)}
				return
			}
			results <- result{index: index, nodes: hydrated.Nodes}
		}(i, p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	nodess := make([][]*kyaml.RNode, len(paths))
	var errs []error

	for res := range results {
		if res.err != nil {
			errs = append(errs, res.err)
		} else {
			nodess[res.index] = res.nodes
		}
	}

	return nodess, errors.Join(errs...)
}

func (b *Builder) writeToStdout(nodes []*kyaml.RNode) error {
	writer := &kio.ByteWriter{
		Writer: b.IoStreams.Out,
	}

	return writer.Write(nodes)
}

func (b *Builder) toAbsPaths(paths []string) ([]string, error) {
	absPaths := make([]string, len(paths))
	for i, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("getting absolute path for %s: %w", p, err)
		}
		absPaths[i] = abs
	}
	return absPaths, nil
}

func (b *Builder) writeNodess(nodess [][]*kyaml.RNode, absPaths []string, commonPrefix, outDir string) error {
	for i, p := range absPaths {
		suffix := strings.TrimPrefix(p, commonPrefix)
		suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
		if suffix == "" {
			suffix = "manifest.yaml"
		}

		outputName := strings.ReplaceAll(suffix, string(filepath.Separator), "_")
		if !strings.HasSuffix(outputName, ".yaml") {
			outputName = outputName + ".yaml"
		}
		outputPath := filepath.Join(outDir, outputName)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("creating output file %s: %w", outputPath, err)
		}
		defer file.Close()

		writer := &kio.ByteWriter{
			Writer: file,
		}

		if err := writer.Write(nodess[i]); err != nil {
			return fmt.Errorf("writing to %s: %w", outputPath, err)
		}
	}

	return nil
}
