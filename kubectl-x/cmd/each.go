package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/each"
)

var (
	eachOutputFormat string
	eachInteractive  bool
	eachCmd          = &cobra.Command{
		Use:   "each [contextPattern] -- [command]",
		Short: "Execute kubectl commands across multiple contexts",
		Long: `Execute kubectl commands across multiple contexts matching a pattern.

This command allows you to run the same kubectl operation on multiple contexts
simultaneously. It uses regex patterns to match context names and executes the
specified command in each matching context.

The command preserves the current namespace for each context unless overridden
with the -n flag. Results can be formatted as JSON, YAML, or raw output.`,
		Example: `  # Execute command on contexts matching pattern
  kubectl x each "prod-.*" -- get pods
  
  # Execute on multiple specific contexts
  kubectl x each "(staging|prod)" -- get deployments
  
  # Interactive context selection with fzf
  kubectl x each -i -- get nodes
  
  # Output in JSON format
  kubectl x each "dev-.*" -o json -- get services
  
  # Output in YAML format  
  kubectl x each ".*-east" -o yaml -- get ingresses
  
  # With namespace override
  kubectl x each "prod-.*" -n kube-system -- get pods`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var contextPattern string
			var commandArgs []string

			dashPos := cmd.ArgsLenAtDash()

			if dashPos == -1 {
				return errors.New("expected '--' to separate context pattern from command arguments")
			}

			if dashPos > 0 {
				contextPattern = args[0]
			}
			commandArgs = args[dashPos:]

			return each.Each(context.Background(), configFlags, resourceBuilderFlags, contextPattern, eachOutputFormat, eachInteractive, commandArgs)
		},
	}
)

func init() {
	rootCmd.AddCommand(eachCmd)
	eachCmd.Flags().StringVarP(&eachOutputFormat, "output", "o", "raw", "Output format (json, yaml, raw)")
	eachCmd.Flags().BoolVarP(&eachInteractive, "interactive", "i", false, "Interactively select contexts using fzf")
	resourceBuilderFlags.AddFlags(eachCmd.Flags())
}
