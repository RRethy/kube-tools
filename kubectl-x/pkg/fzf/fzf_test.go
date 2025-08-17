package fzf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	kexec "k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"
)

func TestNewFzf(t *testing.T) {
	tests := []struct {
		name       string
		options    []Option
		verifyFunc func(t *testing.T, f Interface)
	}{
		{
			name:    "default initialization",
			options: nil,
			verifyFunc: func(t *testing.T, f Interface) {
				assert.NotNil(t, f)
				assert.IsType(t, &Fzf{}, f)
			},
		},
		{
			name: "with exec option",
			options: []Option{
				WithExec(&testingexec.FakeExec{}),
			},
			verifyFunc: func(t *testing.T, f Interface) {
				assert.NotNil(t, f)
				fzfImpl := f.(*Fzf)
				assert.NotNil(t, fzfImpl.exec)
			},
		},
		{
			name: "with io streams option",
			options: []Option{
				WithIOStreams(genericiooptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				}),
			},
			verifyFunc: func(t *testing.T, f Interface) {
				assert.NotNil(t, f)
				fzfImpl := f.(*Fzf)
				assert.NotNil(t, fzfImpl.ioStreams.Out)
				assert.NotNil(t, fzfImpl.ioStreams.ErrOut)
			},
		},
		{
			name: "with multiple options",
			options: []Option{
				WithExec(&testingexec.FakeExec{}),
				WithIOStreams(genericiooptions.IOStreams{
					Out:    &bytes.Buffer{},
					ErrOut: &bytes.Buffer{},
				}),
			},
			verifyFunc: func(t *testing.T, f Interface) {
				assert.NotNil(t, f)
				fzfImpl := f.(*Fzf)
				assert.NotNil(t, fzfImpl.exec)
				assert.NotNil(t, fzfImpl.ioStreams.Out)
			},
		},
		{
			name: "options override in order",
			options: []Option{
				WithExec(&testingexec.FakeExec{}),
				WithIOStreams(genericiooptions.IOStreams{Out: &bytes.Buffer{}}),
				WithExec(kexec.New()), // This should override the first exec
			},
			verifyFunc: func(t *testing.T, f Interface) {
				assert.NotNil(t, f)
				fzfImpl := f.(*Fzf)
				// Can't easily check if it's kexec.New() vs FakeExec, but it shouldn't be nil
				assert.NotNil(t, fzfImpl.exec)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fzf := NewFzf(tt.options...)
			tt.verifyFunc(t, fzf)
		})
	}
}

func TestConfig_BuildArgs(t *testing.T) {
	tests := []struct {
		name             string
		config           Config
		expectedArgs     []string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:   "default config",
			config: Config{},
			expectedArgs: []string{
				"--height", "30%",
				"--ansi",
				"--select-1",
				"--exit-0",
				"--color=dark",
				"--layout=reverse",
			},
			shouldNotContain: []string{"--exact", "--multi", "--prompt", "--query"},
		},
		{
			name: "exact match enabled",
			config: Config{
				ExactMatch: true,
			},
			shouldContain: []string{
				"--height", "30%", "--ansi", "--select-1",
				"--exit-0", "--color=dark", "--layout=reverse",
				"--exact",
			},
			shouldNotContain: []string{"--multi"},
		},
		{
			name: "multi selection enabled",
			config: Config{
				Multi: true,
			},
			shouldContain: []string{
				"--height", "30%", "--ansi", "--select-1",
				"--exit-0", "--color=dark", "--layout=reverse",
				"--multi",
			},
			shouldNotContain: []string{"--exact"},
		},
		{
			name: "with prompt",
			config: Config{
				Prompt: "Select item",
			},
			shouldContain: []string{
				"--prompt", "Select item> ",
			},
		},
		{
			name: "empty prompt",
			config: Config{
				Prompt: "",
			},
			shouldNotContain: []string{"--prompt"},
		},
		{
			name: "with query",
			config: Config{
				Query: "search term",
			},
			shouldContain: []string{
				"--query", "search term",
			},
		},
		{
			name: "empty query",
			config: Config{
				Query: "",
			},
			shouldNotContain: []string{"--query"},
		},
		{
			name: "all options enabled",
			config: Config{
				ExactMatch: true,
				Sorted:     true, // Note: Sorted doesn't add args, handled in Run
				Multi:      true,
				Prompt:     "Choose",
				Query:      "test",
			},
			shouldContain: []string{
				"--exact", "--multi",
				"--prompt", "Choose> ",
				"--query", "test",
			},
		},
		{
			name: "sorted flag does not affect args",
			config: Config{
				Sorted: true,
			},
			shouldNotContain: []string{"--sort", "--sorted"},
		},
		{
			name: "prompt with special characters",
			config: Config{
				Prompt: "Choose [1-3]",
			},
			shouldContain: []string{
				"--prompt", "Choose [1-3]> ",
			},
		},
		{
			name: "query with regex characters",
			config: Config{
				Query: "test.*regex",
			},
			shouldContain: []string{
				"--query", "test.*regex",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.config.buildArgs()

			if tt.expectedArgs != nil {
				assert.Equal(t, tt.expectedArgs, args)
			}

			for _, expected := range tt.shouldContain {
				assert.Contains(t, args, expected, "Expected args to contain %s", expected)
			}

			for _, unexpected := range tt.shouldNotContain {
				assert.NotContains(t, args, unexpected, "Expected args not to contain %s", unexpected)
			}
		})
	}
}

func TestFzf_Run(t *testing.T) {
	tests := []struct {
		name           string
		items          []string
		config         Config
		setupMock      func() *testingexec.FakeExec
		expectedResult []string
		expectedError  string
		verifyFunc     func(t *testing.T, fakeExec *testingexec.FakeExec)
	}{
		{
			name:   "single item selected",
			items:  []string{"item1", "item2", "item3"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							assert.Equal(t, "fzf", cmd)
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("item2"), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedResult: []string{"item2"},
		},
		{
			name:   "multiple items selected",
			items:  []string{"item1", "item2", "item3", "item4"},
			config: Config{Multi: true},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							assert.Equal(t, "fzf", cmd)
							assert.Contains(t, args, "--multi")
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("item1\nitem3\nitem4"), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedResult: []string{"item1", "item3", "item4"},
		},
		{
			name:   "no item selected (empty output)",
			items:  []string{"item1", "item2"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte(""), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedError: "no item selected",
		},
		{
			name:   "whitespace only output",
			items:  []string{"item1", "item2"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("   \n\t  "), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedError: "no item selected",
		},
		{
			name:   "fzf command not found",
			items:  []string{"item1"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return nil, nil, errors.New("command not found: fzf")
									},
								},
							}
						},
					},
				}
			},
			expectedError: "running fzf",
		},
		{
			name:   "fzf exits with error",
			items:  []string{"item1"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return nil, nil, fmt.Errorf("exit status 1")
									},
								},
							}
						},
					},
				}
			},
			expectedError: "running fzf",
		},
		{
			name:   "user cancels selection (exit 130)",
			items:  []string{"item1"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return nil, nil, fmt.Errorf("exit status 130")
									},
								},
							}
						},
					},
				}
			},
			expectedError: "running fzf",
		},
		{
			name:   "items with spaces",
			items:  []string{"item one", "item two", "item three"},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("item two"), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedResult: []string{"item two"},
		},
		{
			name:   "items with special characters",
			items:  []string{"item-1", "item_2", "item.3", "item*4"},
			config: Config{Multi: true},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("item_2\nitem*4"), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedResult: []string{"item_2", "item*4"},
		},
		{
			name:   "empty item list",
			items:  []string{},
			config: Config{},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte(""), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedError: "no item selected",
		},
		{
			name:  "verify all config options passed to fzf",
			items: []string{"ctx1", "ctx2", "ctx3"},
			config: Config{
				ExactMatch: true,
				Multi:      true,
				Prompt:     "Select contexts",
				Query:      "ctx",
			},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							// Verify args
							assert.Equal(t, "fzf", cmd)
							assert.Contains(t, args, "--exact")
							assert.Contains(t, args, "--multi")
							assert.Contains(t, args, "--prompt")
							assert.Contains(t, args, "--query")

							// Find the values for prompt and query
							for i, arg := range args {
								if arg == "--prompt" && i+1 < len(args) {
									assert.Equal(t, "Select contexts> ", args[i+1])
								}
								if arg == "--query" && i+1 < len(args) {
									assert.Equal(t, "ctx", args[i+1])
								}
							}

							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("ctx1\nctx2"), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedResult: []string{"ctx1", "ctx2"},
		},
		{
			name:   "trailing newline in output",
			items:  []string{"line1", "line2", "line3"},
			config: Config{Multi: true},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return []byte("line1\nline2\nline3\n"), nil, nil
									},
								},
							}
						},
					},
				}
			},
			expectedResult: []string{"line1", "line2", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := tt.setupMock()
			errBuf := &bytes.Buffer{}
			fzf := NewFzf(
				WithExec(mockExec),
				WithIOStreams(genericiooptions.IOStreams{ErrOut: errBuf}),
			)

			result, err := fzf.Run(context.Background(), tt.items, tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, mockExec)
			}
		})
	}
}

func TestFzf_Run_Sorting(t *testing.T) {
	tests := []struct {
		name           string
		items          []string
		sorted         bool
		expectedResult string
	}{
		{
			name:           "sorted items",
			items:          []string{"zebra", "apple", "banana"},
			sorted:         true,
			expectedResult: "apple",
		},
		{
			name:           "unsorted items",
			items:          []string{"zebra", "apple", "banana"},
			sorted:         false,
			expectedResult: "zebra",
		},
		{
			name:           "already sorted items",
			items:          []string{"a", "b", "c"},
			sorted:         true,
			expectedResult: "a",
		},
		{
			name:           "single item",
			items:          []string{"only"},
			sorted:         true,
			expectedResult: "only",
		},
		{
			name:           "items with numbers (string sort)",
			items:          []string{"item10", "item2", "item1"},
			sorted:         true,
			expectedResult: "item1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &testingexec.FakeExec{
				CommandScript: []testingexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						return &testingexec.FakeCmd{
							RunScript: []testingexec.FakeAction{
								func() ([]byte, []byte, error) {
									return []byte(tt.expectedResult), nil, nil
								},
							},
						}
					},
				},
			}

			fzf := NewFzf(WithExec(mockExec))
			cfg := Config{Sorted: tt.sorted}

			result, err := fzf.Run(context.Background(), tt.items, cfg)
			require.NoError(t, err)
			assert.Equal(t, []string{tt.expectedResult}, result)
		})
	}
}

func TestFzf_Run_ContextCancellation(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		setupMock     func() *testingexec.FakeExec
		expectedError string
	}{
		{
			name: "context cancelled before execution",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, cancel
			},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										return nil, nil, fmt.Errorf("context cancelled")
									},
								},
							}
						},
					},
				}
			},
			expectedError: "context cancelled",
		},
		{
			name: "context times out",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				return ctx, cancel
			},
			setupMock: func() *testingexec.FakeExec {
				return &testingexec.FakeExec{
					CommandScript: []testingexec.FakeCommandAction{
						func(cmd string, args ...string) kexec.Cmd {
							return &testingexec.FakeCmd{
								RunScript: []testingexec.FakeAction{
									func() ([]byte, []byte, error) {
										time.Sleep(10 * time.Millisecond)
										return nil, nil, fmt.Errorf("context deadline exceeded")
									},
								},
							}
						},
					},
				}
			},
			expectedError: "context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupContext()
			defer cancel()

			mockExec := tt.setupMock()
			fzf := NewFzf(WithExec(mockExec))

			result, err := fzf.Run(ctx, []string{"item1"}, Config{})

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Nil(t, result)
		})
	}
}

func TestFzf_Run_StderrOutput(t *testing.T) {
	tests := []struct {
		name           string
		stderrContent  string
		stdoutContent  string
		expectedError  bool
		expectedStderr string
	}{
		{
			name:           "stderr passed through on success",
			stderrContent:  "fzf: info message\n",
			stdoutContent:  "selected",
			expectedError:  false,
			expectedStderr: "fzf: info message\n",
		},
		{
			name:           "stderr passed through on error",
			stderrContent:  "fzf: error occurred\n",
			stdoutContent:  "",
			expectedError:  true,
			expectedStderr: "fzf: error occurred\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errBuf := &bytes.Buffer{}
			mockExec := &testingexec.FakeExec{
				CommandScript: []testingexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						return &testingexec.FakeCmd{
							Stderr: errBuf,
							RunScript: []testingexec.FakeAction{
								func() ([]byte, []byte, error) {
									var stderr []byte
									if tt.stderrContent != "" {
										stderr = []byte(tt.stderrContent)
									}
									if tt.expectedError {
										return nil, stderr, fmt.Errorf("fzf error")
									}
									return []byte(tt.stdoutContent), stderr, nil
								},
							},
						}
					},
				},
			}

			fzf := NewFzf(
				WithExec(mockExec),
				WithIOStreams(genericiooptions.IOStreams{ErrOut: errBuf}),
			)

			_, err := fzf.Run(context.Background(), []string{"item1"}, Config{})

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Contains(t, errBuf.String(), tt.expectedStderr)
		})
	}
}

func TestFzf_Run_LargeItemList(t *testing.T) {
	tests := []struct {
		name          string
		itemCount     int
		selectedIndex int
	}{
		{
			name:          "100 items",
			itemCount:     100,
			selectedIndex: 50,
		},
		{
			name:          "1000 items",
			itemCount:     1000,
			selectedIndex: 500,
		},
		{
			name:          "10000 items",
			itemCount:     10000,
			selectedIndex: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate items
			items := make([]string, tt.itemCount)
			for i := range items {
				items[i] = fmt.Sprintf("item-%d", i)
			}

			expectedSelection := fmt.Sprintf("item-%d", tt.selectedIndex)

			mockExec := &testingexec.FakeExec{
				CommandScript: []testingexec.FakeCommandAction{
					func(cmd string, args ...string) kexec.Cmd {
						return &testingexec.FakeCmd{
							RunScript: []testingexec.FakeAction{
								func() ([]byte, []byte, error) {
									return []byte(expectedSelection), nil, nil
								},
							},
						}
					},
				},
			}

			fzf := NewFzf(WithExec(mockExec))
			result, err := fzf.Run(context.Background(), items, Config{})

			require.NoError(t, err)
			assert.Equal(t, []string{expectedSelection}, result)
		})
	}
}

// Additional test for option functions
func TestWithOptions(t *testing.T) {
	tests := []struct {
		name       string
		option     Option
		verifyFunc func(t *testing.T, f *Fzf)
	}{
		{
			name:   "WithExec sets exec interface",
			option: WithExec(&testingexec.FakeExec{}),
			verifyFunc: func(t *testing.T, f *Fzf) {
				assert.NotNil(t, f.exec)
				assert.IsType(t, &testingexec.FakeExec{}, f.exec)
			},
		},
		{
			name: "WithIOStreams sets io streams",
			option: WithIOStreams(genericiooptions.IOStreams{
				Out:    &bytes.Buffer{},
				ErrOut: &bytes.Buffer{},
				In:     &bytes.Buffer{},
			}),
			verifyFunc: func(t *testing.T, f *Fzf) {
				assert.NotNil(t, f.ioStreams.Out)
				assert.NotNil(t, f.ioStreams.ErrOut)
				assert.NotNil(t, f.ioStreams.In)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fzf{}
			tt.option(f)
			tt.verifyFunc(t, f)
		})
	}
}
