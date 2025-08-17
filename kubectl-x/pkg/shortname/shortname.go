package shortname

var resourceKindMap = map[string]string{
	"deploy": "deployment",
	"sts":    "statefulset",
	"rs":     "replicaset",
	"ds":     "daemonset",
	"svc":    "service",
}

func Expand(shortType string) string {
	if expanded, ok := resourceKindMap[shortType]; ok {
		return expanded
	}
	return shortType
}