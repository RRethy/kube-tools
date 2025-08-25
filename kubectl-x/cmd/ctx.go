package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/ctx"
)

var ctxCmd = &cobra.Command{
	Use:   "ctx [context] [namespace]",
	Short: "Switch context",
	Long: `Switch kubeconfig context and optionally namespace using interactive fuzzy search.

Both arguments are optional and act as initial filters for the fuzzy search interface.
Use "-" as the context to switch to the previous context/namespace combination.`,
	Example: `  # Browse all contexts interactively
  kubectl x ctx
  
  # Filter contexts by partial name
  kubectl x ctx prod
  
  # Switch to specific context and namespace
  kubectl x ctx prod-cluster my-namespace
  
  # Switch to previous context/namespace
  kubectl x ctx -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var contextName string
		var namespace string
		if len(args) > 0 {
			contextName = args[0]
			if len(args) > 1 {
				namespace = args[1]
			}
		}

		return ctx.Ctx(context.Background(), configFlags, resourceBuilderFlags, contextName, namespace, exactMatch)
	},
}

func init() {
	rootCmd.AddCommand(ctxCmd)
	ctxCmd.Flags().BoolVarP(&exactMatch, "exact", "e", false, "Exact match")
	resourceBuilderFlags.AddFlags(ctxCmd.Flags())
}
