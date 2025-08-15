package helm

import (
	v1 "github.com/RRethy/k8s-tools/kustomizelite/api/v1"
)

// FakeTemplater is a fake implementation of Templater for testing.
type FakeTemplater struct {
	TemplateFunc func(baseDir string, chart v1.HelmChart, globals *v1.HelmGlobals) ([]map[string]any, error)
}

// Template calls the TemplateFunc if set.
func (f *FakeTemplater) Template(baseDir string, chart v1.HelmChart, globals *v1.HelmGlobals) ([]map[string]any, error) {
	if f.TemplateFunc != nil {
		return f.TemplateFunc(baseDir, chart, globals)
	}
	return nil, nil
}
