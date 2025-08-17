package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"k8s.io/klog/v2"

	"github.com/RRethy/kubectl-x/cmd"
)

func main() {
	defer klog.Flush()

	if err := fang.Execute(context.Background(), cmd.GetRootCmd()); err != nil {
		os.Exit(1)
	}
}
