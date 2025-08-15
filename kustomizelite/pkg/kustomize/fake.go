package kustomize

type FakeKustomizer struct {
	KustomizeFunc func(path string, globalHelmValuesFiles []string) ([]map[string]any, error)
}

func (f *FakeKustomizer) Kustomize(path string, globalHelmValuesFiles []string) ([]map[string]any, error) {
	if f.KustomizeFunc != nil {
		return f.KustomizeFunc(path, globalHelmValuesFiles)
	}
	return nil, nil
}
