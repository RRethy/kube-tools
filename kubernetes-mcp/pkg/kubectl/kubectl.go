// Package kubectl provides an interface for executing kubectl commands
package kubectl

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// Kubectl defines the interface for executing kubectl commands
type Kubectl interface {
	Execute(ctx context.Context, args ...string) (stdout string, stderr string, err error)
}

type kubectl struct{}

// New creates a new Kubectl instance that executes real kubectl commands
func New() Kubectl {
	return &kubectl{}
}

// Execute runs kubectl with the provided arguments
func (k *kubectl) Execute(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}