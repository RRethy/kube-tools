package printer

import (
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type Printer interface {
	Print(nodes []*kyaml.RNode) error
}
