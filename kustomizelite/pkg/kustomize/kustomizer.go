package kustomize

type Kustomizer interface {
	Kustomize(path string, globalHelmValuesFiles []string) ([]map[string]any, error)
}
