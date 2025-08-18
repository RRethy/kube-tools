// Package copy provides functionality to copy kubeconfig to $XDG_DATA_HOME
package copy

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"github.com/RRethy/kubectl-x/pkg/xdg"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/utils/clock"
)

// Copy copies the current kubeconfig file to $XDG_DATA_HOME and prints the location
func Copy(ctx context.Context, configFlags *genericclioptions.ConfigFlags) error {
	kubeConfig, err := kubeconfig.NewKubeConfig(kubeconfig.WithConfigFlags(configFlags))
	if err != nil {
		return err
	}

	return (&Copier{
		KubeConfig: kubeConfig,
		IoStreams:  genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		XDG:        xdg.New(),
		Clock:      clock.RealClock{},
		Rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}).Copy(ctx)
}
