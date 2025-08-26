package build

import (
	"context"

	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/RRethy/kube-tools/klite/pkg/hydrate"
)

type Builder struct {
	IoStreams genericiooptions.IOStreams
	Hydrator  hydrate.Hydrator
}

func (b *Builder) Build(ctx context.Context, path string) error {
	_, err := b.Hydrator.Hydrate(ctx, path)
	return err
}
