package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/owners"
)

var ownersCmd = &cobra.Command{
	Use:   "owners (resource-type resource-name | resource-type/resource-name)",
	Short: "Print the ownership graph of a resource",
	Long: `Display the ownership chain of a Kubernetes resource, showing all parent resources
up to the root owner. This helps understand the hierarchy and dependencies of resources.

The command traverses the ownerReferences field to build the complete ownership graph,
displaying each level of ownership from the specified resource to the top-level owner.`,
	Example: `  # Show ownership graph of a pod (space-separated)
  kubectl x owners pod my-pod
  
  # Show ownership graph of a pod (slash-separated)
  kubectl x owners pod/my-pod
  
  # Show ownership graph of a replicaset
  kubectl x owners replicaset nginx-abc123
  
  # Show ownership graph of a deployment
  kubectl x owners deployment/nginx
  
  # Show ownership in a specific namespace
  kubectl x owners pod my-pod -n production`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return owners.Owners(context.Background(), configFlags, resourceBuilderFlags, strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(ownersCmd)
	resourceBuilderFlags.AddFlags(ownersCmd.Flags())
}
