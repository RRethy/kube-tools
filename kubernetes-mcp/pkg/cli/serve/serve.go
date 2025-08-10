// Package serve implements the MCP server serve command logic
package serve

import (
	"context"

	mcpserver "github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/server"
)

// Serve initializes and starts the MCP server
func Serve(ctx context.Context) error {
	mcpServer := mcpserver.NewMCPServer("kubernetes-mcp", "1.0.0")
	
	s := &Server{
		MCPServer: mcpServer,
	}
	return s.Serve(ctx)
}