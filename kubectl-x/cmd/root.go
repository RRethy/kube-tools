package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	configFlags          = genericclioptions.NewConfigFlags(true).WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
	exactMatch           bool
	resourceBuilderFlags = func() *genericclioptions.ResourceBuilderFlags {
		builder := genericclioptions.NewResourceBuilderFlags().
			WithLabelSelector("").
			WithFieldSelector("").
			WithAllNamespaces(false).
			WithAll(false)
		builder.FileNameFlags = nil
		return builder
	}()
	rootCmd = &cobra.Command{
		Use: "kubectl-x",
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl x",
		},
		Short: "Fast context and namespace switching for kubectl",
		Long: `kubectl-x provides efficient context and namespace switching capabilities
for kubectl with interactive fuzzy search powered by fzf.

It maintains a history of recent contexts and namespaces for quick switching
and provides a streamlined interface for common Kubernetes operations.`,
		Example: `  # Switch context interactively
  kubectl x ctx
  
  # Switch namespace in current context
  kubectl x ns
  
  # Show current context and namespace
  kubectl x cur`,
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(cmd.Help())
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	configFlags.AddFlags(rootCmd.PersistentFlags())
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func initConfig() {
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error:"), err)
		os.Exit(1)
	}
}
