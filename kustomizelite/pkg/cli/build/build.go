package build

import (
	"context"
	"os"

	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func Build(ctx context.Context, directory string) error {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return (&Builder{IoStreams: ioStreams}).Build(ctx, directory)
}
