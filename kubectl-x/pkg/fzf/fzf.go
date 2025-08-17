// Package fzf provides integration with the fzf fuzzy finder
package fzf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"
	kexec "k8s.io/utils/exec"
)

const binaryName = "fzf"

// Interface defines methods for fuzzy finding operations
type Interface interface {
	// Run executes fzf with the given items and configuration
	Run(ctx context.Context, items []string, cfg Config) ([]string, error)
}

// Option allows customizing Fzf behavior
type Option func(*Fzf)

// WithExec sets a custom exec interface for testing
func WithExec(exec kexec.Interface) Option {
	return func(f *Fzf) {
		f.exec = exec
	}
}

// WithIOStreams sets custom IO streams for fzf interaction
func WithIOStreams(ioStreams genericiooptions.IOStreams) Option {
	return func(f *Fzf) {
		f.ioStreams = ioStreams
	}
}

// Fzf provides fuzzy finder functionality using the external fzf binary
type Fzf struct {
	exec      kexec.Interface
	ioStreams genericiooptions.IOStreams
}

// NewFzf creates a new fuzzy finder with the given options
func NewFzf(opts ...Option) Interface {
	f := &Fzf{
		exec:      kexec.New(),
		ioStreams: genericiooptions.IOStreams{},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *Fzf) Run(ctx context.Context, items []string, cfg Config) ([]string, error) {
	klog.V(5).Infof("Running fzf with %d items, prompt=%s, query=%s", len(items), cfg.Prompt, cfg.Query)
	args := cfg.buildArgs()
	cmd := f.exec.CommandContext(ctx, binaryName, args...)
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		defer pipeWriter.Close()

		sortedItems := items
		if cfg.Sorted {
			sortedItems = make([]string, len(items))
			copy(sortedItems, items)
			sort.Strings(sortedItems)
		}

		if _, err := fmt.Fprint(pipeWriter, strings.Join(sortedItems, "\n")); err != nil {
			panic(err)
		}
	}()

	cmd.SetStdin(pipeReader)

	var out bytes.Buffer
	cmd.SetStdout(&out)
	cmd.SetStderr(f.ioStreams.ErrOut)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("running fzf: %w", err)
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return nil, fmt.Errorf("no item selected")
	}

	results := strings.Split(output, "\n")
	klog.V(5).Infof("Fzf completed successfully, selected %d items", len(results))
	return results, nil
}
