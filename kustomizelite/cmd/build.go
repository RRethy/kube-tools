package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kube-tools/kustomizelite/pkg/cli/build"
)

var (
	helmValuesFiles []string
	batchFile       string
)

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build and display kustomization",
	Long: `Build and display the resources from a kustomization.

This command processes a kustomization.yaml file and displays the resulting resources.
It applies all kustomize transformations including:
- Resource inclusion
- Namespace injection
- Common labels and annotations
- Name prefix/suffix
- Patches and strategic merges
- Components
- Helm chart inflation

If no path is provided, defaults to the current directory.

Example:
  # Build kustomization in current directory
  adamize build

  # Build a specific kustomization.yaml file
  adamize build /path/to/kustomization.yaml

  # Build a directory containing kustomization.yaml
  adamize build ./overlays/prod/

  # Build with additional helm values files
  adamize build --helm-values-file values-prod.yaml --helm-values-file values-secrets.yaml

  # Build multiple kustomizations in parallel using a batch file
  adamize build -f batch.yaml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if batchFile != "" {
			checkErr(build.Batch(context.Background(), batchFile, helmValuesFiles))
		} else {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			checkErr(build.Build(context.Background(), path, helmValuesFiles))
		}
	},
}

func init() {
	buildCmd.Flags().StringSliceVar(&helmValuesFiles, "helm-values-file", []string{}, "Additional values files to apply to all Helm charts (can be specified multiple times)")
	buildCmd.Flags().StringVarP(&batchFile, "file", "f", "", "Batch build configuration file")
	rootCmd.AddCommand(buildCmd)
}
