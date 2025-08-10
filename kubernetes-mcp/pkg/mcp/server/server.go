// Package server provides the MCP server implementation for Kubernetes operations
package server

import (
	"github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/tools"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer defines the interface for serving MCP requests
type MCPServer interface {
	Serve(opts ...ServerOption) error
}

type mcpServer struct {
	name    string
	version string
}

type serverOptions struct {
	includeReadonlyTools bool
}

// ServerOption configures server behavior
type ServerOption func(*serverOptions)

// WithIncludeReadonlyTools configures whether to include readonly Kubernetes tools
func WithIncludeReadonlyTools(include bool) ServerOption {
	return func(o *serverOptions) {
		o.includeReadonlyTools = include
	}
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(name string, version string) MCPServer {
	return &mcpServer{
		name:    name,
		version: version,
	}
}

// Serve starts the MCP server with optional configuration
func (m *mcpServer) Serve(opts ...ServerOption) error {
	options := &serverOptions{
		includeReadonlyTools: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	s := server.NewMCPServer(
		m.name,
		m.version,
		server.WithToolCapabilities(false),
		server.WithPromptCapabilities(false),
	)

	if options.includeReadonlyTools {
		t := tools.New()
		s.AddTools(
			server.ServerTool{Tool: t.CreateGetTool(), Handler: t.HandleGet},
			server.ServerTool{Tool: t.CreateGetScalingOverridesTool(), Handler: t.HandleGetScalingOverrides},
			server.ServerTool{Tool: t.CreateDescribeTool(), Handler: t.HandleDescribe},
			server.ServerTool{Tool: t.CreateLogsTool(), Handler: t.HandleLogs},
			server.ServerTool{Tool: t.CreateEventsTool(), Handler: t.HandleEvents},
			server.ServerTool{Tool: t.CreateAPIResourcesTool(), Handler: t.HandleAPIResources},
			server.ServerTool{Tool: t.CreateAPIVersionsTool(), Handler: t.HandleAPIVersions},
			server.ServerTool{Tool: t.CreateAuthCanITool(), Handler: t.HandleAuthCanI},
			server.ServerTool{Tool: t.CreateTopTool(), Handler: t.HandleTop},
			server.ServerTool{Tool: t.CreateExplainTool(), Handler: t.HandleExplain},
			server.ServerTool{Tool: t.CreateVersionTool(), Handler: t.HandleVersion},
			server.ServerTool{Tool: t.CreateConfigViewTool(), Handler: t.HandleConfigView},
			server.ServerTool{Tool: t.CreateConfigGetContextsTool(), Handler: t.HandleConfigGetContexts},
			server.ServerTool{Tool: t.CreateClusterInfoTool(), Handler: t.HandleClusterInfo},
			server.ServerTool{Tool: t.CreateCurrentContextTool(), Handler: t.HandleCurrentContext},
			server.ServerTool{Tool: t.CreateCurrentNamespaceTool(), Handler: t.HandleCurrentNamespace},
			server.ServerTool{Tool: t.CreateUseContextTool(), Handler: t.HandleUseContext},
		)
	}

	return server.ServeStdio(s)
}
