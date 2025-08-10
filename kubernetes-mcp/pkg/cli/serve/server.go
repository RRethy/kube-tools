package serve

import (
	"context"

	mcpserver "github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/server"
)

// Server wraps an MCP server for serving requests
type Server struct {
	MCPServer mcpserver.MCPServer
}

// Serve starts the MCP server to handle incoming requests
func (s *Server) Serve(ctx context.Context) error {
	return s.MCPServer.Serve()
}
