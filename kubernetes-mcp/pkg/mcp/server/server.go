package server

import (
	"github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/tools"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer interface {
	Serve(opts ...ServerOption) error
}

type mcpServer struct {
	name    string
	version string
}

type serverOptions struct {
	excludeReadonlyTools bool
}

type ServerOption func(*serverOptions)

func WithExcludeReadonlyTools(exclude bool) ServerOption {
	return func(o *serverOptions) {
		o.excludeReadonlyTools = exclude
	}
}

func NewMCPServer(name string, version string) MCPServer {
	return &mcpServer{
		name:    name,
		version: version,
	}
}

func (m *mcpServer) Serve(opts ...ServerOption) error {
	options := &serverOptions{}
	for _, opt := range opts {
		opt(options)
	}

	s := server.NewMCPServer(
		m.name,
		m.version,
		server.WithToolCapabilities(false),
		server.WithPromptCapabilities(false),
	)

	if !options.excludeReadonlyTools {
		t := tools.New()
		s.AddTools(
			server.ServerTool{Tool: t.CreateGetTool(), Handler: t.HandleGet},
			server.ServerTool{Tool: t.CreateGetScalingOverridesTool(), Handler: t.HandleGetScalingOverrides},
			server.ServerTool{Tool: t.CreateDescribeTool(), Handler: t.HandleDescribe},
			server.ServerTool{Tool: t.CreateLogsTool(), Handler: t.HandleLogs},
			server.ServerTool{Tool: t.CreateEventsTool(), Handler: t.HandleEvents},
			server.ServerTool{Tool: t.CreateAPIResourcesTool(), Handler: t.HandleAPIResources},
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

