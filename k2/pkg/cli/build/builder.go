package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	var nodess [][]*kyaml.RNode
	for _, p := range paths {
		nodes, err := b.Hydrator.Hydrate(ctx, p)
		if err != nil {
			return err
		}
		nodess = append(nodess, nodes)
	}

	if outDir == "" {
		writer := &kio.ByteWriter{
			Writer: b.IoStreams.Out,
		}

		return writer.Write(nodess[0])
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	absPaths := make([]string, len(paths))
	for i, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("getting absolute path for %s: %w", p, err)
		}
		absPaths[i] = abs
	}

	commonPrefix := k2strings.FindCommonPrefix(absPaths)

	for i, p := range absPaths {
		suffix := strings.TrimPrefix(p, commonPrefix)
		suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
		if suffix == "" {
			suffix = "_"
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
