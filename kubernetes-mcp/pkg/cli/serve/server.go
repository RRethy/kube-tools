package serve

import (
	"context"

	mcpserver "github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/server"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type Server struct {
	IOStreams genericiooptions.IOStreams
	MCPServer mcpserver.MCPServer
}

func (s *Server) Serve(ctx context.Context) error {
	return s.MCPServer.Serve()
}

