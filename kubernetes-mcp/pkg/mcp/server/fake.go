package server

// FakeMCPServer implements MCPServer interface for testing
type FakeMCPServer struct {
	ServeError   error
	ServeCalled  bool
	OptionsUsed  []ServerOption
}

func (f *FakeMCPServer) Serve(opts ...ServerOption) error {
	f.ServeCalled = true
	f.OptionsUsed = opts
	return f.ServeError
}

func NewFakeMCPServer(serveError error) *FakeMCPServer {
	return &FakeMCPServer{
		ServeError: serveError,
	}
}