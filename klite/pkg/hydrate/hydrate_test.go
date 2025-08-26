package hydrate

import (
	"context"
	"strings"
	"testing"
)

func TestNewHydrator(t *testing.T) {
	h := NewHydrator()
	if h == nil {
		t.Error("NewHydrator() returned nil")
	}
}

func TestHydrate(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "basic empty kustomization with directory path",
			path:    "../../fixtures/basic-empty",
			wantErr: false,
		},
		{
			name:    "basic empty kustomization with file path",
			path:    "../../fixtures/basic-empty/kustomization.yaml",
			wantErr: false,
		},
		{
			name:    "non-existent directory",
			path:    "../../fixtures/non-existent",
			wantErr: true,
			errMsg:  "no such file or directory",
		},
		{
			name:    "directory without kustomization.yaml",
			path:    "../../fixtures",
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:    "non-kustomization file",
			path:    "hydrate.go",
			wantErr: true,
			errMsg:  "is not kustomization.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHydrator()
			ctx := context.Background()
			
			nodes, err := h.Hydrate(ctx, tt.path)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Hydrate() expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Hydrate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Hydrate() unexpected error = %v", err)
				}
				if nodes == nil {
					t.Errorf("Hydrate() returned nil nodes, want empty slice")
				}
			}
		})
	}
}

