package writer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestYAML_Write(t *testing.T) {
	tests := []struct {
		name       string
		yaml       []string
		wantOutput string
	}{
		{
			name:       "empty input",
			yaml:       []string{},
			wantOutput: "",
		},
		{
			name: "single document",
			yaml: []string{
				`apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    example.com/annotation: "true"
  labels:
    app: my-app
  name: test
data:
  key: value`,
			},
			wantOutput: `apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    example.com/annotation: "true"
  labels:
    app: my-app
  name: test
data:
  key: value
`,
		},
		{
			name: "multiple documents with separator",
			yaml: []string{
				`apiVersion: v1
kind: ConfigMap
metadata:
  name: first`,
				`apiVersion: v1
kind: Service
metadata:
  name: second`,
			},
			wantOutput: `apiVersion: v1
kind: ConfigMap
metadata:
  name: first
---
apiVersion: v1
kind: Service
metadata:
  name: second
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse YAML strings to RNodes
			var nodes []*kyaml.RNode
			for _, y := range tt.yaml {
				node, err := kyaml.Parse(y)
				require.NoError(t, err)
				nodes = append(nodes, node)
			}

			// Write to buffer
			var buf bytes.Buffer
			writer := NewYAML(&buf)
			err := writer.Write(nodes)
			require.NoError(t, err)

			// Check output
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}