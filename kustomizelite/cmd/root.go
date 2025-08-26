package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var rootCmd = &cobra.Command{
	Use:   "kustomizelite",
	Short: "A lightweight Kustomize-like tool",
	Long:  `kustomizelite is a CLI tool that provides simplified Kustomize-like functionality`,
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	rootCmd.PersistentFlags().AddGoFlag(klogFlags.Lookup("v"))
}
