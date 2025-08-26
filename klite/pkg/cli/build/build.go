package build

import (
	"context"
	"os"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kube-tools/klite/pkg/hydrate"
	"github.com/RRethy/kube-tools/klite/pkg/printer"
)

func Build(ctx context.Context, path string) error {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	hydrator := hydrate.NewHydrator()
	p := printer.NewYAMLPrinter(ioStreams.Out)
	return (&Builder{
		IoStreams: ioStreams,
		Hydrator:  hydrator,
		Printer:   p,
	}).Build(ctx, path)
}
