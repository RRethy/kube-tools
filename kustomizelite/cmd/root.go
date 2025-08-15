package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kustomizelite",
	Short: "Shopify's Kustomize CLI tool",
	Run: func(cmd *cobra.Command, _ []string) {
		checkErr(cmd.Help())
	},
}

func Execute() {
	checkErr(rootCmd.Execute())
}

func init() {
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error:"), err)
		os.Exit(1)
	}
}
