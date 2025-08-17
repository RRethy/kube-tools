package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/kubectl-x/pkg/cli/shell"
)

var (
	shellContainer  string
	shellCommand    string
	shellDebug      bool
	shellDebugImage string
	shellCmd        = &cobra.Command{
		Use:   "shell [pod-name | resource-type resource-name | resource-type/resource-name]",
		Short: "Shell into a pod",
		Long: `Shell into a pod directly or by resolving a resource to its associated pods.

When given a pod name, it shells directly into that pod.
When given a resource type and name (separated by space or slash), it resolves 
to the associated pods and shells into one of them.

Use --debug to run a debug container instead of exec'ing into the pod.`,
		Example: `  # Shell into a pod directly
  kubectl x shell my-pod
  
  # Shell into a pod from a deployment (with slash)
  kubectl x shell deployment/nginx
  kubectl x shell deploy/nginx
  
  # Shell into a pod from a deployment (with space)
  kubectl x shell deployment nginx
  kubectl x shell deploy nginx
  
  # Shell into a pod from a statefulset
  kubectl x shell statefulset database
  kubectl x shell sts/database
  
  # Shell into a pod from a replicaset
  kubectl x shell replicaset nginx-abc123
  kubectl x shell rs/nginx-abc123
  
  # Shell into a pod from a job
  kubectl x shell job backup
  
  # Shell into a specific container
  kubectl x shell my-pod -c container-name
  
  # Use a different shell
  kubectl x shell my-pod --command=/bin/bash
  
  # Debug mode examples
  kubectl x shell my-pod --debug
  kubectl x shell my-pod --debug --image=ubuntu
  kubectl x shell my-pod --debug -c=app-container`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var target string
			if len(args) == 1 {
				target = args[0]
			} else if len(args) == 2 {
				target = args[0] + " " + args[1]
			}
			return shell.Shell(context.Background(), configFlags, resourceBuilderFlags, target, shellContainer, shellCommand, shellDebug, shellDebugImage)
		},
	}
)

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().StringVarP(&shellContainer, "container", "c", "", "Container name (for shell exec) or target container (for debug)")
	shellCmd.Flags().StringVar(&shellCommand, "command", "/bin/sh", "Shell command to execute")
	shellCmd.Flags().BoolVar(&shellDebug, "debug", false, "Use kubectl debug to run a debug container")
	shellCmd.Flags().StringVar(&shellDebugImage, "image", "", "Debug container image (when using --debug)")
}
