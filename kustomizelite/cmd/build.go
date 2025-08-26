package cmd

import (
	"github.com/RRethy/kube-tools/kustomizelite/pkg/cli/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Build Kubernetes resources from a directory",
	Long:  `Build generates Kubernetes resources from a kustomization directory`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		directory := "."
		if len(args) > 0 {
			directory = args[0]
		}
		return build.Build(cmd.Context(), directory)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
