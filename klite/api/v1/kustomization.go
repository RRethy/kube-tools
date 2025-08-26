package v1

type Kustomization struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Resources  []string `yaml:"resources,omitempty" json:"resources,omitempty"`
}
