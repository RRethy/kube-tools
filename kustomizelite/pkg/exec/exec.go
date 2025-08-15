package exec

import (
	"os"
	"os/exec"
)

// Wrapper provides an interface for executing commands.
type Wrapper interface {
	Command(name string, arg ...string) *exec.Cmd
}

// defaultWrapper uses os/exec directly.
type defaultWrapper struct{}

func (d *defaultWrapper) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// envWrapper executes commands with additional environment variables.
type envWrapper struct {
	env map[string]string
}

func (e *envWrapper) Command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	// Start with current environment
	cmd.Env = os.Environ()
	// Add/override with custom environment variables
	for k, v := range e.env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	return cmd
}

// New returns the default exec wrapper.
func New() Wrapper {
	return &defaultWrapper{}
}

// NewWithEnv returns an exec wrapper that includes additional environment variables.
func NewWithEnv(env map[string]string) Wrapper {
	return &envWrapper{env: env}
}
