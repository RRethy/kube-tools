package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func init() {
	configFlags.AddFlags(rootCmd.PersistentFlags())

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	rootCmd.PersistentFlags().AddGoFlag(klogFlags.Lookup("v"))
}

// GetRootCmd returns the root cobra command for kubectl-x
func GetRootCmd() *cobra.Command {
	return rootCmd
}
