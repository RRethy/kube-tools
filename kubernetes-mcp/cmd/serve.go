package cmd

import (
	"github.com/spf13/cobra"

	"github.com/RRethy/utils/kubernetes-mcp/pkg/mcp"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server in stdio mode",
	Long: `Start the kubernetes-mcp server in stdio mode for Model Context Protocol (MCP) communication.
This allows LLMs to interact with your Kubernetes cluster through readonly operations.

The server exposes tools for:
- Getting Kubernetes resources (pods, deployments, services, etc.)
- Describing resources for detailed information
- Fetching pod logs
- Viewing cluster events
- Explaining resource types
- Getting cluster information

Usage:
	kubernetes-mcp serve

Examples:
	# Start in stdio mode (typical usage)
	kubernetes-mcp serve`,
	RunE: func(_ *cobra.Command, _ []string) error {
		server := mcp.NewServer()
		return server.Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
