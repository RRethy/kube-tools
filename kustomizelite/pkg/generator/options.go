package generator

import "github.com/RRethy/kube-tools/kustomizelite/pkg/exec"

// Option is a functional option for configuring a Generator.
type Option func(*generator)

// WithExecWrapper sets a custom exec wrapper (useful for testing).
func WithExecWrapper(wrapper exec.Wrapper) Option {
	return func(g *generator) {
		g.execWrapper = wrapper
	}
}
