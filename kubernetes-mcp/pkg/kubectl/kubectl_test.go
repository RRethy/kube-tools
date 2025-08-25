package kubectl

import (
	"context"
	"strings"
	"testing"
)

func TestContainsShellMetacharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Clean string", "get-pods", false},
		{"Clean with dots", "app.kubernetes.io/name", false},
		{"Clean with equals", "app=nginx", false},
		{"Clean with colon", "status.phase:Running", false},
		{"Dollar sign", "test$value", true},
		{"Backtick", "test`cmd`", true},
		{"Semicolon", "test;ls", true},
		{"Pipe", "test|grep", true},
		{"Ampersand at end", "test&", true},
		{"Double ampersand", "test&&other", true},
		{"Double pipe", "test||other", true},
		{"Greater than redirect", ">file", true},
		{"Less than redirect", "<file", true},
		{"Double greater than", "test>>file", true},
		{"Double less than", "test<<EOF", true},
		{"Newline", "test\nls", true},
		{"Carriage return", "test\rls", true},
		{"Parentheses open", "test(", true},
		{"Parentheses close", "test)", true},
		{"Curly brace open", "test{", true},
		{"Curly brace close", "test}", true},
		{"Square bracket open", "test[", true},
		{"Square bracket close", "test]", true},
		{"Asterisk", "test*", true},
		{"Question mark", "test?", true},
		{"Tilde", "~test", true},
		{"Exclamation", "test!", true},
		{"Backslash", "test\\cmd", true},
		{"Double quote", "test\"value", true},
		{"Single quote", "test'value", true},
		{"Null byte", "test\x00value", true},
		{"Tab", "test\tvalue", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsShellMetacharacters(tt.input)
			if result != tt.expected {
				t.Errorf("containsShellMetacharacters(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsPathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Clean path", "/home/user/file", false},
		{"Clean Windows path", "C:\\Users\\file", false},
		{"Kubernetes resource", "default/my-pod", false},
		{"Dot notation", "app.version", false},
		{"Parent directory slash", "../test", true},
		{"Parent directory", "..", true},
		{"Parent directory backslash", "..\\test", true},
		{"Parent at end", "test/..", true},
		{"Parent Windows", "test\\..", true},
		{"Multiple dots", "....test", true},
		{"Etc directory", "/etc/passwd", true},
		{"Proc directory", "/proc/self", true},
		{"Sys directory", "/sys/class", true},
		{"Dev directory", "/dev/null", true},
		{"Windows etc", "\\etc\\config", true},
		{"Windows proc", "\\proc\\info", true},
		{"Complex traversal", "../../etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsPathTraversal(tt.input)
			if result != tt.expected {
				t.Errorf("containsPathTraversal(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateArgument(t *testing.T) {
	tests := []struct {
		name      string
		arg       string
		wantError bool
		errorMsg  string
	}{
		{"Empty string", "", false, ""},
		{"Valid resource", "pods", false, ""},
		{"Valid namespace", "default", false, ""},
		{"Valid label selector", "app=nginx", false, ""},
		{"Valid output format", "json", false, ""},
		{"Shell injection dollar", "$USER", true, "dangerous characters"},
		{"Shell injection backtick", "`whoami`", true, "dangerous characters"},
		{"Command chaining semicolon", "pods;ls", true, "dangerous characters"},
		{"Path traversal", "../etc/passwd", true, "path traversal"},
		{"System path", "/etc/shadow", true, "path traversal"},
		{"Null byte", "test\x00value", true, "dangerous characters"},
		{"Very long argument", strings.Repeat("a", 4097), true, "exceeds maximum length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgument(tt.arg)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateArgument(%q) = nil, want error containing %q", tt.arg, tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateArgument(%q) error = %v, want error containing %q", tt.arg, err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateArgument(%q) = %v, want nil", tt.arg, err)
				}
			}
		})
	}
}

func TestExecuteValidation(t *testing.T) {
	k := New()
	ctx := context.Background()

	tests := []struct {
		name      string
		args      []string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Shell injection in resource type",
			args:      []string{"get", "pods;ls"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Path traversal in namespace",
			args:      []string{"get", "pods", "-n", "../../../etc"},
			wantError: true,
			errorMsg:  "path traversal",
		},
		{
			name:      "Command substitution",
			args:      []string{"get", "pods", "--context", "$(whoami)"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Command chaining with &&",
			args:      []string{"get", "pods", "&&", "ls"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Command chaining with ||",
			args:      []string{"get", "pods", "||", "echo", "failed"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Subshell with $()",
			args:      []string{"logs", "$(echo pod)"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Variable expansion",
			args:      []string{"get", "pods", "${HOME}"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Shell redirection >>",
			args:      []string{"get", "pods", ">>", "output.txt"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Shell redirection <<",
			args:      []string{"apply", "-f", "-", "<<", "EOF"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Pipe character",
			args:      []string{"get", "pods", "|", "grep", "running"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
		{
			name:      "Background execution",
			args:      []string{"get", "pods", "&"},
			wantError: true,
			errorMsg:  "dangerous characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := k.Execute(ctx, tt.args...)
			if tt.wantError {
				if err == nil {
					t.Errorf("Execute(%v) = nil, want error containing %q", tt.args, tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Execute(%v) error = %v, want error containing %q", tt.args, err, tt.errorMsg)
				}
			}
		})
	}
}
