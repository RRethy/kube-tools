package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/kubeconfig/copy"
)

var kubeconfigCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy current kubeconfig to $XDG_DATA_HOME",
	Long: `Copy the current kubeconfig file to $XDG_DATA_HOME/kubectl-x/kubeconfig and print the location.
	
This is useful for shell scripts that need to update $KUBECONFIG for the current session
without modifying the original kubeconfig file.`,
	Example: `  # Copy kubeconfig and get the path
  kubectl x kubeconfig copy
  
  # Use in a shell script to set $KUBECONFIG for current session
  export KUBECONFIG=$(kubectl x kubeconfig copy)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copy.Copy(context.Background(), configFlags)
	},
}

func init() {
	kubeconfigCmd.AddCommand(kubeconfigCopyCmd)
}
