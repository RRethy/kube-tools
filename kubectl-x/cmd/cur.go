package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/cur"
)

var curPrompt bool

var curCmd = &cobra.Command{
	Use:   "cur",
	Short: "Print current context and namespace.",
	Long: `Print current context and namespace.

Usage:
  kubectl x cur
  kubectl x cur --prompt

Example:
  kubectl x cur           # Default: --context <ctx> --namespace <ns>
  kubectl x cur --prompt  # Compact: <ctx>/<ns>`,
	Run: func(cmd *cobra.Command, args []string) {
		checkErr(cur.CurWithPrompt(context.Background(), curPrompt))
	},
}

func init() {
	curCmd.Flags().BoolVarP(&curPrompt, "prompt", "p", false, "Output in compact format for prompts (<ctx>/<ns>)")
	rootCmd.AddCommand(curCmd)
}
