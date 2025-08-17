// Package shortname provides Kubernetes resource shortname expansion
package shortname

var resourceKindMap = map[string]string{
	"deploy": "deployment",
	"sts":    "statefulset",
	"rs":     "replicaset",
	"ds":     "daemonset",
	"svc":    "service",
}

// Expand converts Kubernetes resource shortnames to their full forms
func Expand(shortType string) string {
	if expanded, ok := resourceKindMap[shortType]; ok {
		return expanded
	}
	return shortType
}