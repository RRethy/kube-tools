package serve

import (
	"context"
	"os"

	mcpserver "github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/server"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func Serve(ctx context.Context) error {
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	
	mcpServer := mcpserver.NewMCPServer("kubernetes-mcp", "1.0.0")
	
	s := &Server{
		IOStreams: ioStreams,
		MCPServer: mcpServer,
	}
	return s.Serve(ctx)
}