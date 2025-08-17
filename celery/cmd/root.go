package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "celery",
	Short: "CEL expression evaluator for Kubernetes YAML",
	Long: `Celery evaluates Common Expression Language (CEL) expressions against 
Kubernetes Resource Model (KRM) YAML files.

Use it to validate resources with custom rules, check compliance with policies,
or query resource properties using CEL's expression language.`,
	Version: version,
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
