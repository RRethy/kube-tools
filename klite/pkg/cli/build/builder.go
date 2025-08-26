package build

import (
	"context"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kube-tools/klite/pkg/hydrate"
	"github.com/RRethy/kube-tools/klite/pkg/printer"
)

type Builder struct {
	IoStreams genericiooptions.IOStreams
	Hydrator  hydrate.Hydrator
	Printer   printer.Printer
}

func (b *Builder) Build(ctx context.Context, path string) error {
	nodes, err := b.Hydrator.Hydrate(ctx, path)
	if err != nil {
		return err
	}

	return b.Printer.Print(nodes)
}
