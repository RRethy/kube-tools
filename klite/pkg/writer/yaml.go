package writer

import (
	"fmt"
	"io"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type yaml struct {
	writer io.Writer
}

func NewYAML(w io.Writer) Writer {
	return &yaml{
		writer: w,
	}
}

func (y *yaml) Write(nodes []*kyaml.RNode) error {
	for i, node := range nodes {
		if i > 0 {
			fmt.Fprintln(y.writer, "---")
		}

		yamlStr, err := node.String()
		if err != nil {
			return fmt.Errorf("converting node to YAML: %w", err)
		}

		fmt.Fprint(y.writer, yamlStr)
	}

	return nil
}