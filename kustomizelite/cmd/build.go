package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Build Kubernetes resources from a directory",
	Long:  `Build generates Kubernetes resources from a kustomization directory`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running build command")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
