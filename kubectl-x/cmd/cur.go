package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/cur"
)

var curPrompt bool

var curCmd = &cobra.Command{
	Use:   "cur",
	Short: "Print current context and namespace",
	Long: `Display the current Kubernetes context and namespace.

By default, outputs in kubectl flag format for easy copying.
Use --prompt for a compact format suitable for shell prompts.`,
	Example: `  # Show current context and namespace in kubectl format
  kubectl x cur
  # Output: --context my-context --namespace my-namespace
  
  # Show in compact format for shell prompts
  kubectl x cur --prompt
  # Output: my-context/my-namespace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cur.CurWithPrompt(context.Background(), configFlags, curPrompt)
	},
}

func init() {
	curCmd.Flags().BoolVarP(&curPrompt, "prompt", "p", false, "Output in compact format for prompts (<ctx>/<ns>)")
	rootCmd.AddCommand(curCmd)
}
