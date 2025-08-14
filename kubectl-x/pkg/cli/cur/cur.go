package cur

import (
	"context"
	"os"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func Cur(ctx context.Context) error {
	return CurWithPrompt(ctx, false)
}

func CurWithPrompt(ctx context.Context, promptFormat bool) error {
	kubeConfig, err := kubeconfig.NewKubeConfig()
	if err != nil {
		return err
	}
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	curer := NewCurer(kubeConfig, ioStreams)
	return curer.CurWithPrompt(ctx, promptFormat)
}
