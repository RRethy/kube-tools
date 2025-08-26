package build

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type Builder struct {
	IoStreams genericiooptions.IOStreams
}

func (b *Builder) Build(ctx context.Context, directory string) error {
	fmt.Fprintln(b.IoStreams.Out, "Hello, World!")
	return nil
}
