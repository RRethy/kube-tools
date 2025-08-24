module github.com/RRethy/kube-tools/kustomizelite

go 1.24.4

require (
	github.com/RRethy/krepe/jsonpatch v0.0.0-20250410043855-bb4c24df4a21
	github.com/fatih/color v1.18.0
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/cli-runtime v0.33.3
)

require github.com/RRethy/krepe/deepishcopy v0.0.0-00010101000000-000000000000

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
)

replace github.com/RRethy/krepe/deepishcopy => github.com/RRethy/krepe/deepishcopy v0.0.0-20250410043855-bb4c24df4a21

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/sys v0.34.0 // indirect
	k8s.io/apimachinery v0.33.3
)
