package kustomize

import (
	"github.com/RRethy/kube-tools/kustomizelite/pkg/exec"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/generator"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/helm"
)

// Option is a functional option for configuring a Kustomizer.
type Option func(*kustomization)

// WithHelmTemplater sets a custom helm templater (useful for testing).
func WithHelmTemplater(templater helm.Templater) Option {
	return func(k *kustomization) {
		k.helmTemplater = templater
	}
}

// WithExecWrapper sets a custom exec wrapper for running commands.
func WithExecWrapper(wrapper exec.Wrapper) Option {
	return func(k *kustomization) {
		k.execWrapper = wrapper
	}
}

// WithGenerator sets a custom generator (useful for testing).
func WithGenerator(gen generator.Generator) Option {
	return func(k *kustomization) {
		k.generator = gen
	}
}
