package build

import (
	"context"
	"os"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kube-tools/klite/pkg/hydrate"
	"github.com/RRethy/kube-tools/klite/pkg/writer"
)

func Build(ctx context.Context, path string) error {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	hydrator := hydrate.NewHydrator()
	w := writer.NewYAML(ioStreams.Out)
	return (&Builder{
		IoStreams: ioStreams,
		Hydrator:  hydrator,
		Writer: w,
	}).Build(ctx, path)
}
