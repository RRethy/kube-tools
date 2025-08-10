package serve

import (
	"context"
	"errors"
	"testing"

	mcpserver "github.com/RRethy/k8s-tools/kubernetes-mcp/pkg/mcp/server"
)

func TestServer_Serve(t *testing.T) {
	tests := []struct {
		name         string
		serveError   error
		wantError    bool
		wantErrorMsg string
	}{
		{
			name:       "successful serve",
			serveError: nil,
			wantError:  false,
		},
		{
			name:         "serve returns error",
			serveError:   errors.New("serve failed"),
			wantError:    true,
			wantErrorMsg: "serve failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeServer := mcpserver.NewFakeMCPServer(tt.serveError)
			
			s := &Server{
				MCPServer: fakeServer,
			}

			ctx := context.Background()
			err := s.Serve(ctx)

			if (err != nil) != tt.wantError {
				t.Errorf("Serve() error = %v, wantError %v", err, tt.wantError)
			}

			if err != nil && tt.wantErrorMsg != "" && err.Error() != tt.wantErrorMsg {
				t.Errorf("Serve() error = %v, want %v", err.Error(), tt.wantErrorMsg)
			}

			if !fakeServer.ServeCalled {
				t.Error("Expected Serve to be called")
			}
		})
	}
}

