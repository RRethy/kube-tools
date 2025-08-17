// Package server provides the MCP server implementation for Kubernetes operations
package server

// FakeMCPServer provides a test implementation of MCPServer
type FakeMCPServer struct {
	ServeError  error
	ServeCalled bool
	OptionsUsed []ServerOption
}

// Serve records the call and returns the configured error
func (f *FakeMCPServer) Serve(opts ...ServerOption) error {
	f.ServeCalled = true
	f.OptionsUsed = opts
	return f.ServeError
}

// NewFakeMCPServer creates a new fake MCP server for testing
func NewFakeMCPServer(serveError error) *FakeMCPServer {
	return &FakeMCPServer{
		ServeError: serveError,
	}
}
