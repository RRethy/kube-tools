// Package cmd implements CLI commands for the kubernetes-mcp server
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubernetes-mcp",
	Short: "Kubernetes Model Context Protocol (MCP) server",
	Long: `A readonly MCP (Model Context Protocol) server that exposes kubectl functionality 
as tools for LLM integration. Provides safe, readonly access to Kubernetes cluster 
information including resources, logs, and events.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and handles errors
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}