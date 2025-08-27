package build

import (
	"context"
	"os"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kube-tools/k2/pkg/hydrate"
)

func Build(ctx context.Context, paths []string, outDir string) error {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	hydrator := hydrate.NewHydrator()
	builder := &Builder{
		IoStreams: ioStreams,
		Hydrator:  hydrator,
	}
	return builder.Build(ctx, paths, outDir)
}
