package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Kustomization struct {
	APIVersion string             `yaml:"apiVersion" json:"apiVersion"`
	Kind       string             `yaml:"kind" json:"kind"`
	Metadata   *metav1.ObjectMeta `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Resources  []string           `yaml:"resources,omitempty" json:"resources,omitempty"`
}
