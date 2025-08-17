// Package kubectl provides an interface for executing kubectl commands
package kubectl

import (
	"bytes"
	"context"
	"fmt"
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

// containsShellMetacharacters detects shell injection attempts through metacharacter detection
func containsShellMetacharacters(s string) bool {
	dangerousChars := []string{
		"$", "`", ";", "|", "\n", "\r",
		"(", ")", "{", "}", "[", "]", "*", "?", "~", "!",
		"\\", "\"", "'", "\x00", "\t", "\v", "\f",
	}

	// Special handling for & and > < which might be in operators
	// Allow single & and single > or < but not doubles
	if strings.Contains(s, "&&") || strings.Contains(s, "||") {
		return true
	}
	if strings.Contains(s, ">>") || strings.Contains(s, "<<") {
		return true
	}
	// Check for shell redirection with single > or <
	if (strings.HasPrefix(s, ">") || strings.HasPrefix(s, "<")) && len(s) > 1 {
		return true
	}
	if strings.HasSuffix(s, "&") {
		return true // Background execution
	}

	for _, char := range dangerousChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

// containsPathTraversal detects directory traversal and system path access attempts
func containsPathTraversal(s string) bool {
	pathTraversalPatterns := []string{
		"../", "..", "..\\", "/..", "\\..",
		"....",                              // Multiple dots
		"/etc/", "/proc/", "/sys/", "/dev/", // System paths
		"\\etc\\", "\\proc\\", "\\sys\\", "\\dev\\",
	}

	for _, pattern := range pathTraversalPatterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

// validateArgument ensures kubectl arguments are safe from injection attacks
func validateArgument(arg string) error {
	// Check for empty argument
	if arg == "" {
		return nil // Empty arguments are allowed
	}

	// Check for shell metacharacters
	if containsShellMetacharacters(arg) {
		return fmt.Errorf("argument contains potentially dangerous characters: %q", arg)
	}

	// Check for path traversal
	if containsPathTraversal(arg) {
		return fmt.Errorf("argument contains potential path traversal: %q", arg)
	}

	// Check for excessive length (prevent buffer overflow attempts)
	if len(arg) > 4096 {
		return fmt.Errorf("argument exceeds maximum length: %d characters", len(arg))
	}

	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("argument contains null bytes")
	}

	return nil
}

// Execute safely executes kubectl commands after validating all arguments
func (k *kubectl) Execute(ctx context.Context, args ...string) (string, string, error) {
	// Validate all arguments before execution
	for i, arg := range args {
		if err := validateArgument(arg); err != nil {
			return "", "", fmt.Errorf("invalid argument at position %d: %w", i, err)
		}
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}
