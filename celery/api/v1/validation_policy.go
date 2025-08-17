package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidationRules is a KRM resource that defines CEL validation rules.
type ValidationRules struct {
	metav1.TypeMeta   `yaml:",inline"`
	metav1.ObjectMeta `yaml:"metadata,omitempty"`
	Spec              ValidationRulesSpec `yaml:"spec"`
	Filename          string
}

// ValidationRulesList is a list of ValidationRules resources.
type ValidationRulesList struct {
	metav1.TypeMeta `yaml:",inline"`
	metav1.ListMeta `yaml:"metadata,omitempty"`
	Items           []ValidationRules `yaml:"items"`
}

// ValidationRulesSpec contains the validation rules.
type ValidationRulesSpec struct {
	Rules []ValidationRule `yaml:"rules"`
}

type ValidationRule struct {
	Name       string          `yaml:"name"`
	Expression string          `yaml:"expression"`
	Message    string          `yaml:"message,omitempty"`
	Target     *TargetSelector `yaml:"target,omitempty"`
}

// TargetSelector matches Kustomize's selector format.
type TargetSelector struct {
	Group              string `yaml:"group,omitempty"`
	Version            string `yaml:"version,omitempty"`
	Kind               string `yaml:"kind,omitempty"`
	Name               string `yaml:"name,omitempty"`
	Namespace          string `yaml:"namespace,omitempty"`
	LabelSelector      string `yaml:"labelSelector,omitempty"`
	AnnotationSelector string `yaml:"annotationSelector,omitempty"`
}
