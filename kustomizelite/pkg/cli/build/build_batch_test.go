package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBatchInvalidConfig(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		wantErrMsg string
	}{
		{
			name: "wrong apiVersion",
			content: `apiVersion: wrong/v1
kind: BatchBuild
builds: []`,
			wantErrMsg: "unsupported apiVersion: wrong/v1",
		},
		{
			name: "wrong kind",
			content: `apiVersion: adamize.io/v1
kind: WrongKind
builds: []`,
			wantErrMsg: "unsupported kind: WrongKind",
		},
		{
			name:       "invalid yaml",
			content:    `{invalid yaml`,
			wantErrMsg: "parsing batch file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchFile := filepath.Join(tempDir, "batch.yaml")
			err := os.WriteFile(batchFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			err = Batch(t.Context(), batchFile, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestBuildBatchNonExistentFile(t *testing.T) {
	err := Batch(t.Context(), "/non/existent/file.yaml", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading batch file")
}

func TestBuildBatchWithErrors(t *testing.T) {
	tempDir := t.TempDir()

	// Create a batch config with a non-existent kustomization
	batchContent := `apiVersion: adamize.io/v1
kind: BatchBuild
builds:
  - kustomization: /non/existent/kustomization.yaml
    output: ` + filepath.Join(tempDir, "output.yaml") + `
`

	batchFile := filepath.Join(tempDir, "batch.yaml")
	err := os.WriteFile(batchFile, []byte(batchContent), 0644)
	require.NoError(t, err)

	// Execute batch build
	err = Batch(t.Context(), batchFile, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch build failed with 1 errors")
}
