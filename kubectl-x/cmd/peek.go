package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/peek"
)

var peekCmd = &cobra.Command{
	Use:   "peek (logs|describe) [resource-type] [resource-name]",
	Short: "Interactively preview logs or describe output for resources",
	Long:  "Peek allows you to interactively preview logs or describe output for Kubernetes resources using fzf",
	Example: `  # Interactively preview logs for pods
  kubectl x peek logs pod

  # Preview logs for a specific pod
  kubectl x peek logs pod nginx

  # Preview logs for pods from a deployment
  kubectl x peek logs deployment my-app

  # Interactively preview describe output
  kubectl x peek describe pod

  # Preview describe for services
  kubectl x peek describe service`,
	Args: cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return peek.Peek(cmd.Context(), configFlags, resourceBuilderFlags, args[0], strings.Join(args[1:], " "))
	},
}

func init() {
	rootCmd.AddCommand(peekCmd)
}
