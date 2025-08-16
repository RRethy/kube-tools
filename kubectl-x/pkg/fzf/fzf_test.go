package fzf

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	kexec "k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"
)

func TestNewFzf(t *testing.T) {
	fzf := NewFzf()
	
	assert.NotNil(t, fzf)
	assert.IsType(t, &Fzf{}, fzf)
}

func TestNewFzf_WithOptions(t *testing.T) {
	execInterface := &testingexec.FakeExec{}
	ioStreams := genericiooptions.IOStreams{
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	
	fzf := NewFzf(
		WithExec(execInterface),
		WithIOStreams(ioStreams),
	)
	
	assert.NotNil(t, fzf)
	fzfImpl := fzf.(*Fzf)
	assert.Equal(t, execInterface, fzfImpl.exec)
	assert.Equal(t, ioStreams, fzfImpl.ioStreams)
}

func TestWithExec(t *testing.T) {
	mockExec := &testingexec.FakeExec{}
	opt := WithExec(mockExec)
	
	fzf := &Fzf{}
	opt(fzf)
	
	assert.Equal(t, mockExec, fzf.exec)
}

func TestWithIOStreams(t *testing.T) {
	ioStreams := genericiooptions.IOStreams{
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
		In:     &bytes.Buffer{},
	}
	opt := WithIOStreams(ioStreams)
	
	fzf := &Fzf{}
	opt(fzf)
	
	assert.Equal(t, ioStreams, fzf.ioStreams)
}

// Config tests

func TestConfig_BuildArgs_Default(t *testing.T) {
	cfg := Config{}
	args := cfg.buildArgs()
	
	expected := []string{
		"--height", "30%",
		"--ansi",
		"--select-1",
		"--exit-0",
		"--color=dark",
		"--layout=reverse",
	}
	
	assert.Equal(t, expected, args)
}

func TestConfig_BuildArgs_ExactMatch(t *testing.T) {
	cfg := Config{
		ExactMatch: true,
	}
	args := cfg.buildArgs()
	
	assert.Contains(t, args, "--exact")
}

func TestConfig_BuildArgs_Multi(t *testing.T) {
	cfg := Config{
		Multi: true,
	}
	args := cfg.buildArgs()
	
	assert.Contains(t, args, "--multi")
}

func TestConfig_BuildArgs_Prompt(t *testing.T) {
	tests := []struct {
		name           string
		prompt         string
		expectedPrompt string
	}{
		{
			name:           "simple prompt",
			prompt:         "Select item",
			expectedPrompt: "Select item> ",
		},
		{
			name:           "empty prompt",
			prompt:         "",
			expectedPrompt: "",
		},
		{
			name:           "prompt with special chars",
			prompt:         "Choose [1-3]",
			expectedPrompt: "Choose [1-3]> ",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Prompt: tt.prompt,
			}
			args := cfg.buildArgs()
			
			if tt.prompt == "" {
				// Should not contain --prompt flag
				assert.NotContains(t, args, "--prompt")
			} else {
				// Find prompt flag and value
				promptIndex := -1
				for i, arg := range args {
					if arg == "--prompt" {
						promptIndex = i
						break
					}
				}
				
				require.NotEqual(t, -1, promptIndex, "Should contain --prompt flag")
				require.Less(t, promptIndex+1, len(args), "Should have value after --prompt")
				assert.Equal(t, tt.expectedPrompt, args[promptIndex+1])
			}
		})
	}
}

func TestConfig_BuildArgs_Query(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedQuery string
	}{
		{
			name:          "simple query",
			query:         "search term",
			expectedQuery: "search term",
		},
		{
			name:          "empty query",
			query:         "",
			expectedQuery: "",
		},
		{
			name:          "query with special chars",
			query:         "test.*regex",
			expectedQuery: "test.*regex",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Query: tt.query,
			}
			args := cfg.buildArgs()
			
			if tt.query == "" {
				// Should not contain --query flag
				assert.NotContains(t, args, "--query")
			} else {
				// Find query flag and value
				queryIndex := -1
				for i, arg := range args {
					if arg == "--query" {
						queryIndex = i
						break
					}
				}
				
				require.NotEqual(t, -1, queryIndex, "Should contain --query flag")
				require.Less(t, queryIndex+1, len(args), "Should have value after --query")
				assert.Equal(t, tt.expectedQuery, args[queryIndex+1])
			}
		})
	}
}

func TestConfig_BuildArgs_AllOptions(t *testing.T) {
	cfg := Config{
		ExactMatch: true,
		Sorted:     true,
		Multi:      true,
		Prompt:     "Choose",
		Query:      "test",
	}
	args := cfg.buildArgs()
	
	// Verify all base options are present
	assert.Contains(t, args, "--height")
	assert.Contains(t, args, "30%")
	assert.Contains(t, args, "--ansi")
	assert.Contains(t, args, "--select-1")
	assert.Contains(t, args, "--exit-0")
	assert.Contains(t, args, "--color=dark")
	assert.Contains(t, args, "--layout=reverse")
	
	// Verify additional options
	assert.Contains(t, args, "--exact")
	assert.Contains(t, args, "--multi")
	
	// Find prompt
	promptIndex := -1
	for i, arg := range args {
		if arg == "--prompt" {
			promptIndex = i
			break
		}
	}
	require.NotEqual(t, -1, promptIndex)
	assert.Equal(t, "Choose> ", args[promptIndex+1])
	
	// Find query
	queryIndex := -1
	for i, arg := range args {
		if arg == "--query" {
			queryIndex = i
			break
		}
	}
	require.NotEqual(t, -1, queryIndex)
	assert.Equal(t, "test", args[queryIndex+1])
}

func TestConfig_BuildArgs_Order(t *testing.T) {
	cfg := Config{
		ExactMatch: true,
		Multi:      true,
		Prompt:     "Select",
		Query:      "search",
	}
	args := cfg.buildArgs()
	
	// Verify that base args come first
	assert.Equal(t, "--height", args[0])
	assert.Equal(t, "30%", args[1])
	
	// The additional flags should come after the base ones
	// Base args: --height 30% --ansi --select-1 --exit-0 --color=dark --layout=reverse
	// That's 7 items total
	baseArgsCount := 7
	
	// Find positions of optional args
	var exactPos, multiPos, promptPos, queryPos int = -1, -1, -1, -1
	for i, arg := range args {
		switch arg {
		case "--exact":
			exactPos = i
		case "--multi":
			multiPos = i
		case "--prompt":
			promptPos = i
		case "--query":
			queryPos = i
		}
	}
	
	// All optional args should be after base args
	if exactPos >= 0 {
		assert.GreaterOrEqual(t, exactPos, baseArgsCount)
	}
	if multiPos >= 0 {
		assert.GreaterOrEqual(t, multiPos, baseArgsCount)
	}
	if promptPos >= 0 {
		assert.GreaterOrEqual(t, promptPos, baseArgsCount)
	}
	if queryPos >= 0 {
		assert.GreaterOrEqual(t, queryPos, baseArgsCount)
	}
}

func TestFzf_DefaultInitialization(t *testing.T) {
	fzf := &Fzf{}
	
	// By default, exec and ioStreams should be nil/empty
	assert.Nil(t, fzf.exec)
	assert.Equal(t, genericiooptions.IOStreams{}, fzf.ioStreams)
}

func TestFzf_OptionsAreAppliedInOrder(t *testing.T) {
	// Create different exec interfaces to test ordering
	exec1 := &testingexec.FakeExec{}
	exec2 := kexec.New()
	
	outBuf1 := &bytes.Buffer{}
	outBuf2 := &bytes.Buffer{}
	
	ioStreams1 := genericiooptions.IOStreams{Out: outBuf1}
	ioStreams2 := genericiooptions.IOStreams{Out: outBuf2}
	
	// Apply options in specific order
	fzf := NewFzf(
		WithExec(exec1),
		WithIOStreams(ioStreams1),
		WithExec(exec2),       // This should override exec1
		WithIOStreams(ioStreams2), // This should override ioStreams1
	)
	
	fzfImpl := fzf.(*Fzf)
	assert.Equal(t, exec2, fzfImpl.exec)
	assert.Equal(t, ioStreams2, fzfImpl.ioStreams)
}

// Test that Config fields work correctly
func TestConfig_Fields(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		check  func(t *testing.T, args []string)
	}{
		{
			name: "exact match only",
			config: Config{
				ExactMatch: true,
			},
			check: func(t *testing.T, args []string) {
				assert.Contains(t, args, "--exact")
				assert.NotContains(t, args, "--multi")
			},
		},
		{
			name: "multi only",
			config: Config{
				Multi: true,
			},
			check: func(t *testing.T, args []string) {
				assert.Contains(t, args, "--multi")
				assert.NotContains(t, args, "--exact")
			},
		},
		{
			name: "sorted does not affect args",
			config: Config{
				Sorted: true,
			},
			check: func(t *testing.T, args []string) {
				// Sorted doesn't add any args, it's handled in Run
				assert.NotContains(t, args, "--sort")
				assert.NotContains(t, args, "--sorted")
			},
		},
		{
			name: "combination of flags",
			config: Config{
				ExactMatch: true,
				Multi:      true,
				Sorted:     true,
			},
			check: func(t *testing.T, args []string) {
				assert.Contains(t, args, "--exact")
				assert.Contains(t, args, "--multi")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.config.buildArgs()
			tt.check(t, args)
		})
	}
}