package kubectl

import "context"

// FakeKubectl provides a test implementation of Kubectl
type FakeKubectl struct {
	ExecuteStdout string
	ExecuteStderr string
	ExecuteError  error
	ExecuteCalled bool
	ExecuteArgs   []string
}

// NewFake creates a new fake Kubectl for testing
func NewFake(stdout, stderr string, err error) *FakeKubectl {
	return &FakeKubectl{
		ExecuteStdout: stdout,
		ExecuteStderr: stderr,
		ExecuteError:  err,
	}
}

// Execute records the call and returns the configured response
func (f *FakeKubectl) Execute(ctx context.Context, args ...string) (string, string, error) {
	f.ExecuteCalled = true
	f.ExecuteArgs = args
	return f.ExecuteStdout, f.ExecuteStderr, f.ExecuteError
}
