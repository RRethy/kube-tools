package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (stdout, stderr string, err error) {
	stdout, stderr, err = "", "", nil

	var stdoutBuf, stderrBuf bytes.Buffer
	root.SetOut(&stdoutBuf)
	root.SetErr(&stderrBuf)
	root.SetArgs(args)

	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	err = root.Execute()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	stdout = stdoutBuf.String() + buf.String()
	stderr = stderrBuf.String()

	return stdout, stderr, err
}

func TestBuildCommandLongHelp(t *testing.T) {
	cmd := &cobra.Command{Use: "adamize"}
	cmd.AddCommand(buildCmd)

	stdout, _, err := executeCommand(cmd, "build", "--help")
	assert.NoError(t, err)

	expectedHelp := []string{
		"Build and display the resources from a kustomization",
		"This command processes a kustomization.yaml file",
		"Resource inclusion",
		"Namespace injection",
		"Common labels and annotations",
		"Name prefix/suffix",
		"Patches and strategic merges",
		"Components",
		"If no path is provided, defaults to the current directory",
		"Example:",
		"# Build kustomization in current directory",
		"adamize build",
		"# Build a specific kustomization.yaml file",
		"adamize build /path/to/kustomization.yaml",
		"# Build a directory containing kustomization.yaml",
		"adamize build ./overlays/prod/",
	}

	for _, expected := range expectedHelp {
		assert.Contains(t, stdout, expected, "help output should contain: %s", expected)
	}
}

// Note: Error cases are tested at the package level in pkg/cli/build/
// Testing them here would cause the test process to exit due to checkErr calling os.Exit(1)
