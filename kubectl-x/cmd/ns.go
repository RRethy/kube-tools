package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/ns"
)

var nsCmd = &cobra.Command{
	Use:   "ns [namespace]",
	Short: "Switch namespace",
	Long: `Switch Kubernetes namespace within the current context using interactive fuzzy search.

The namespace argument is optional and acts as an initial filter for the fuzzy search.
Use "-" as the namespace to switch to the previous namespace.`,
	Example: `  # Browse all namespaces interactively
  kubectl x ns
  
  # Filter namespaces by partial name
  kubectl x ns kube
  
  # Switch to specific namespace
  kubectl x ns my-namespace
  
  # Switch to previous namespace
  kubectl x ns -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var namespace string
		if len(args) > 0 {
			namespace = args[0]
		}

		return ns.Ns(context.Background(), configFlags, resourceBuilderFlags, namespace, exactMatch)
	},
}

func init() {
	rootCmd.AddCommand(nsCmd)
	nsCmd.Flags().BoolVarP(&exactMatch, "exact", "e", false, "Exact match")
	resourceBuilderFlags.AddFlags(nsCmd.Flags())
}
