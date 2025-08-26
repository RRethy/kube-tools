package cmd

import (
	"github.com/RRethy/kube-tools/kustomizelite/pkg/cli/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Build Kubernetes resources from a directory",
	Long:  `Build generates Kubernetes resources from a kustomization directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		return build.Build(cmd.Context(), path)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
