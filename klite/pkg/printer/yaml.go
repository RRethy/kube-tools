package printer

import (
	"fmt"
	"io"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type yamlPrinter struct {
	Writer io.Writer
}

func NewYAMLPrinter(w io.Writer) Printer {
	return &yamlPrinter{
		Writer: w,
	}
}

func (p *yamlPrinter) Print(nodes []*kyaml.RNode) error {
	for i, node := range nodes {
		if i > 0 {
			fmt.Fprintln(p.Writer, "---")
		}

		yamlStr, err := node.String()
		if err != nil {
			return fmt.Errorf("converting node to YAML: %w", err)
		}

		fmt.Fprint(p.Writer, yamlStr)
	}

	return nil
}
