package writer

import (
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type Writer interface {
	Write(nodes []*kyaml.RNode) error
}