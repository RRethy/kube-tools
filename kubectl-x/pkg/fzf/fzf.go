package fzf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	kexec "k8s.io/utils/exec"
)

const binaryName = "fzf"

type Interface interface {
	Run(ctx context.Context, items []string, cfg Config) ([]string, error)
}

type Option func(*Fzf)

func WithExec(exec kexec.Interface) Option {
	return func(f *Fzf) {
		f.exec = exec
	}
}

func WithIOStreams(ioStreams genericiooptions.IOStreams) Option {
	return func(f *Fzf) {
		f.ioStreams = ioStreams
	}
}

type Fzf struct {
	exec      kexec.Interface
	ioStreams genericiooptions.IOStreams
}

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
	args := cfg.buildArgs()
	cmd := exec.CommandContext(ctx, binaryName, args...)
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

	cmd.Stdin = pipeReader

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = f.ioStreams.ErrOut

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

	return strings.Split(output, "\n"), nil
}
