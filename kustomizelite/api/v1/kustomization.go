package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Kustomization struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata" yaml:"metadata"`

	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`
}
