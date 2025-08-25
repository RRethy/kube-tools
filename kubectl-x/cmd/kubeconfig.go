package cmd

import (
	"github.com/spf13/cobra"
)

var kubeconfigCmd = &cobra.Command{
	Use:     "kubeconfig",
	Aliases: []string{"kc"},
	Short:   "Kubeconfig operations",
	Long:    `Perform various operations on kubeconfig files.`,
	Example: `  # Copy kubeconfig to $XDG_DATA_HOME
  kubectl x kubeconfig copy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(kubeconfigCmd)
}
