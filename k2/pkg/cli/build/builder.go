package build

import (
	"context"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"sigs.k8s.io/kustomize/kyaml/kio"

	"github.com/RRethy/kube-tools/k2/pkg/hydrate"
)

type Builder struct {
	IoStreams genericiooptions.IOStreams
	Hydrator  hydrate.Hydrator
}

func (b *Builder) Build(ctx context.Context, path string) error {
	nodes, err := b.Hydrator.Hydrate(ctx, path)
	if err != nil {
		return err
	}

	writer := &kio.ByteWriter{
		Writer: b.IoStreams.Out,
	}

	return writer.Write(nodes)
}
