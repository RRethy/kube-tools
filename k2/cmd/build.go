package cmd

import (
	"github.com/RRethy/kube-tools/k2/pkg/cli/build"
	"github.com/spf13/cobra"
)

var outDir string

var buildCmd = &cobra.Command{
	Use:   "build [path...] [--out-dir DIR]",
	Short: "Build Kubernetes resources from one or more directories",
	Long: `Build generates Kubernetes resources from kustomization directories.

When building multiple paths, --out-dir is required to specify where to write the output files.
Output files are named based on the input path's parent directory name.`,
	Example: `  # Build current directory to stdout
  k2 build

  # Build specific directory to stdout
  k2 build ./overlays/production

  # Build multiple directories to output directory
  k2 build ./base ./overlays/dev ./overlays/prod --out-dir ./manifests

  # Build single directory to output directory
  k2 build ./overlays/production --out-dir ./output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return build.Build(cmd.Context(), args, outDir)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVar(&outDir, "out-dir", "", "Output directory for built manifests")
}
