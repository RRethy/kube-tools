package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"

	"github.com/RRethy/kube-tools/celery/cmd"
)

func main() {
	if err := fang.Execute(context.Background(), cmd.GetRootCmd()); err != nil {
		os.Exit(1)
	}
}
