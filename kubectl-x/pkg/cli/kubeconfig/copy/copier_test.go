package copy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
	xdgtesting "github.com/RRethy/kubectl-x/pkg/xdg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd/api"
	testingclock "k8s.io/utils/clock/testing"
)

// testRand returns a new random generator with the given seed for testing
func testRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func TestCopier_Copy(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)

	t.Run("copies kubeconfig to XDG_DATA_HOME when set", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()

		testContent := "test kubeconfig content"
		_, err = sourceFile.WriteString(testContent)
		require.NoError(t, err)
		_ = sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		output := strings.TrimSpace(outputBuf.String())
		assert.True(t, strings.HasPrefix(output, filepath.Join(xdgDataHome, "kubectl-x", "kubeconfig-")))

		_, err = os.Stat(output)
		require.NoError(t, err)

		filename := filepath.Base(output)
		matched, err := regexp.MatchString(`^kubeconfig-20240115-103045-\d+$`, filename)
		require.NoError(t, err)
		assert.True(t, matched, "filename should match pattern kubeconfig-YYYYMMDD-HHMMSS-{unique-id}")
	})

	t.Run("uses default XDG_DATA_HOME when not set", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()

		testContent := "test kubeconfig content"
		_, err = sourceFile.WriteString(testContent)
		require.NoError(t, err)
		_ = sourceFile.Close()

		homeDir, err := os.MkdirTemp("", "home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(homeDir) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(filepath.Join(homeDir, ".local", "share")),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		output := strings.TrimSpace(outputBuf.String())
		expectedPrefix := filepath.Join(homeDir, ".local", "share", "kubectl-x", "kubeconfig-")
		assert.True(t, strings.HasPrefix(output, expectedPrefix))

		_, err = os.Stat(output)
		require.NoError(t, err)
	})

	t.Run("creates kubectl-x directory if it doesn't exist", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()
		_ = sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		kubectlXDir := filepath.Join(xdgDataHome, "kubectl-x")
		_, err = os.Stat(kubectlXDir)
		assert.True(t, os.IsNotExist(err))

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		info, err := os.Stat(kubectlXDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("works even when source kubeconfig doesn't exist", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/non/existent/kubeconfig")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		output := strings.TrimSpace(outputBuf.String())
		_, err = os.Stat(output)
		require.NoError(t, err)
	})

	t.Run("generates unique filenames on multiple runs", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()
		_ = sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputs []string
		for i := range 3 {
			var outputBuf bytes.Buffer
			fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
				map[string]*api.Context{"test": {Cluster: "test"}},
				"test",
				"default",
			).WithKubeconfigPath(sourceFile.Name())

			copier := &Copier{
				KubeConfig: fakeKubeConfig,
				IoStreams: genericiooptions.IOStreams{
					Out: &outputBuf,
				},
				XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
				Rand:  testRand(int64(12345 + i)), // Different seed for each iteration
				Clock: testingclock.NewFakeClock(fixedTime),
			}

			err = copier.Copy(ctx)
			require.NoError(t, err)

			output := strings.TrimSpace(outputBuf.String())
			outputs = append(outputs, output)
		}

		assert.Len(t, outputs, 3)
		assert.NotEqual(t, outputs[0], outputs[1])
		assert.NotEqual(t, outputs[1], outputs[2])
		assert.NotEqual(t, outputs[0], outputs[2])
	})

	t.Run("returns error when output write fails", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()
		_ = sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &failingWriter{},
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "writing output:")
	})

	t.Run("returns error when HOME is not set and XDG_DATA_HOME is empty", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()
		_ = sourceFile.Close()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG("").WithDataHomeError(errors.New("$HOME is not defined")),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "getting XDG data home")
	})

	t.Run("returns error when cannot create target directory", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()
		_ = sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		conflictingFile := filepath.Join(xdgDataHome, "kubectl-x")
		err = os.WriteFile(conflictingFile, []byte("conflict"), 0644)
		require.NoError(t, err)

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("returns error when cannot write destination file", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(sourceFile.Name()) }()
		_ = sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		kubectlXDir := filepath.Join(xdgDataHome, "kubectl-x")
		err = os.MkdirAll(kubectlXDir, 0555)
		require.NoError(t, err)

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath(sourceFile.Name())

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("writes valid merged kubeconfig", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{
				"test1": {Cluster: "cluster1", AuthInfo: "user1"},
				"test2": {Cluster: "cluster2", AuthInfo: "user2"},
			},
			"test1",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		output := strings.TrimSpace(outputBuf.String())

		writtenContent, err := os.ReadFile(output)
		require.NoError(t, err)
		assert.Contains(t, string(writtenContent), "apiVersion: v1")
		assert.Contains(t, string(writtenContent), "kind: Config")
		assert.Contains(t, string(writtenContent), "current-context: test1")
	})

	t.Run("handles XDG ConfigHome error gracefully", func(t *testing.T) {
		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		fakeXDG := &xdgtesting.FakeXDG{
			ConfigHomeError: errors.New("config home error"),
		}

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   fakeXDG,
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err := copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "getting XDG data home")
	})

	t.Run("handles XDG CacheHome error gracefully", func(t *testing.T) {
		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		fakeXDG := &xdgtesting.FakeXDG{
			CacheHomeError: errors.New("cache home error"),
		}

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   fakeXDG,
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err := copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "getting XDG data home")
	})

	t.Run("returns error when WriteToFile fails", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path").WithWriteError(errors.New("write failed"))

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "writing kubeconfig")
		assert.Contains(t, err.Error(), "write failed")
	})

	t.Run("handles empty XDG DataHome path", func(t *testing.T) {
		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		fakeXDG := &xdgtesting.FakeXDG{
			DataHomePath: "",
		}

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   fakeXDG,
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err := copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "getting XDG data home")
		assert.Contains(t, err.Error(), "data home not configured")
	})

	t.Run("handles XDG paths with special characters", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data home-with spaces-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		output := strings.TrimSpace(outputBuf.String())
		assert.True(t, strings.HasPrefix(output, filepath.Join(xdgDataHome, "kubectl-x", "kubeconfig-")))

		_, err = os.Stat(output)
		require.NoError(t, err)
	})

	t.Run("concurrent copy operations use unique filenames", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		numGoroutines := 5
		results := make(chan string, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := range numGoroutines {
			go func() {
				var outputBuf bytes.Buffer
				fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
					map[string]*api.Context{"test": {Cluster: "test"}},
					"test",
					"default",
				).WithKubeconfigPath("/fake/path")

				copier := &Copier{
					KubeConfig: fakeKubeConfig,
					IoStreams: genericiooptions.IOStreams{
						Out: &outputBuf,
					},
					XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:  testRand(int64(i)),
					Clock: testingclock.NewFakeClock(fixedTime),
				}

				if err := copier.Copy(ctx); err != nil {
					errors <- err
					return
				}
				results <- strings.TrimSpace(outputBuf.String())
			}()
		}

		uniquePaths := make(map[string]bool)
		for range numGoroutines {
			select {
			case err := <-errors:
				t.Fatalf("Unexpected error: %v", err)
			case path := <-results:
				if uniquePaths[path] {
					t.Errorf("Duplicate path generated: %s", path)
				}
				uniquePaths[path] = true
			}
		}

		assert.Len(t, uniquePaths, numGoroutines, "All paths should be unique")
	})
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, os.ErrPermission
}

func TestCopy(t *testing.T) {
	t.Run("successfully copies kubeconfig", func(t *testing.T) {
		// Create a temporary kubeconfig file
		tmpKubeconfig, err := os.CreateTemp("", "test-kubeconfig-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpKubeconfig.Name()) }()

		// Write a valid kubeconfig
		kubeconfigContent := `apiVersion: v1
kind: Config
current-context: test-context
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
clusters:
- name: test-cluster
  cluster:
    server: https://test-server:6443
users:
- name: test-user
  user:
    token: test-token`
		_, err = tmpKubeconfig.WriteString(kubeconfigContent)
		require.NoError(t, err)
		_ = tmpKubeconfig.Close()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Create config flags pointing to our test kubeconfig
		kubeconfigPath := tmpKubeconfig.Name()
		configFlags := genericclioptions.NewConfigFlags(true)
		configFlags.KubeConfig = &kubeconfigPath

		// Run the Copy function
		ctx := context.Background()
		err = Copy(ctx, configFlags)

		// Restore stdout and read output
		_ = w.Close()
		os.Stdout = oldStdout
		output, _ := io.ReadAll(r)

		require.NoError(t, err)

		// Verify output path was written
		outputPath := strings.TrimSpace(string(output))
		assert.Contains(t, outputPath, "kubectl-x/kubeconfig-")

		// Verify the file was created
		_, err = os.Stat(outputPath)
		require.NoError(t, err)

		// Clean up
		_ = os.Remove(outputPath)
	})

	t.Run("handles nil config flags", func(t *testing.T) {
		ctx := context.Background()
		err := Copy(ctx, nil)

		// This might succeed if there's a valid kubeconfig in the default location
		// or fail if kubeconfig cannot be loaded, both are acceptable
		_ = err
	})
}

func TestCopier_Copy_ContextCancellation(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	t.Run("respects context cancellation", func(t *testing.T) {
		// Create a context that's already cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		// The function should still complete successfully as it doesn't check context
		err = copier.Copy(ctx)
		require.NoError(t, err)
	})

	t.Run("handles context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)
	})
}

func TestCopier_Copy_NilFields(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	ctx := context.Background()

	t.Run("handles nil In stream", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				In:  nil, // Explicitly nil
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)
	})

	t.Run("handles nil ErrOut stream", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out:    &outputBuf,
				ErrOut: nil, // Explicitly nil
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)
	})
}

func TestCopier_Copy_FilesystemEdgeCases(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	ctx := context.Background()

	t.Run("handles deeply nested directory creation", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		// Create a deeply nested XDG path
		deepPath := filepath.Join(xdgDataHome, "level1", "level2", "level3", "level4")

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(deepPath),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		// Verify the deeply nested directory was created
		kubectlXDir := filepath.Join(deepPath, "kubectl-x")
		info, err := os.Stat(kubectlXDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("handles symlinked XDG directory", func(t *testing.T) {
		// Create actual directory
		actualDir, err := os.MkdirTemp("", "actual-xdg-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(actualDir) }()

		// Create symlink directory
		symlinkDir, err := os.MkdirTemp("", "symlink-base-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(symlinkDir) }()

		symlinkPath := filepath.Join(symlinkDir, "xdg-symlink")
		err = os.Symlink(actualDir, symlinkPath)
		require.NoError(t, err)

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(symlinkPath),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		// Verify the file was written through the symlink
		output := strings.TrimSpace(outputBuf.String())
		_, err = os.Stat(output)
		require.NoError(t, err)

		// Verify file is accessible from actual directory too
		filename := filepath.Base(output)
		actualPath := filepath.Join(actualDir, "kubectl-x", filename)
		_, err = os.Stat(actualPath)
		require.NoError(t, err)
	})

	t.Run("handles existing kubectl-x directory with files", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(xdgDataHome) }()

		// Pre-create kubectl-x directory with existing files
		kubectlXDir := filepath.Join(xdgDataHome, "kubectl-x")
		err = os.MkdirAll(kubectlXDir, 0755)
		require.NoError(t, err)

		existingFile := filepath.Join(kubectlXDir, "existing-file.txt")
		err = os.WriteFile(existingFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		var outputBuf bytes.Buffer
		fakeKubeConfig := kubeconfigtesting.NewFakeKubeConfig(
			map[string]*api.Context{"test": {Cluster: "test"}},
			"test",
			"default",
		).WithKubeconfigPath("/fake/path")

		copier := &Copier{
			KubeConfig: fakeKubeConfig,
			IoStreams: genericiooptions.IOStreams{
				Out: &outputBuf,
			},
			XDG:   xdgtesting.NewFakeXDG(xdgDataHome),
			Rand:  testRand(12345),
			Clock: testingclock.NewFakeClock(fixedTime),
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		// Verify both new and existing files are present
		output := strings.TrimSpace(outputBuf.String())
		_, err = os.Stat(output)
		require.NoError(t, err)

		_, err = os.Stat(existingFile)
		require.NoError(t, err, "Existing file should still be present")
	})
}

func TestCopier_Copy_TableDriven(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)

	tests := []struct {
		name          string
		setupCopier   func(*testing.T) (*Copier, func())
		expectedError string
		validateFunc  func(*testing.T, *Copier, string, error)
	}{
		// Success cases
		{
			name: "successful copy with minimal config",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				xdgDataHome, err := os.MkdirTemp("", "xdg-*")
				require.NoError(t, err)

				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() { _ = os.RemoveAll(xdgDataHome) }
			},
			expectedError: "",
			validateFunc: func(t *testing.T, c *Copier, output string, err error) {
				assert.Contains(t, output, "kubeconfig-20240115-103045-")
				assert.FileExists(t, output)

				// Verify file content
				content, readErr := os.ReadFile(output)
				require.NoError(t, readErr)
				assert.NotEmpty(t, content)
			},
		},
		{
			name: "creates kubectl-x subdirectory if not exists",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				xdgDataHome, err := os.MkdirTemp("", "xdg-*")
				require.NoError(t, err)

				// Ensure kubectl-x directory doesn't exist initially
				kubectlXDir := filepath.Join(xdgDataHome, "kubectl-x")
				_, err = os.Stat(kubectlXDir)
				require.True(t, os.IsNotExist(err))

				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() { _ = os.RemoveAll(xdgDataHome) }
			},
			expectedError: "",
			validateFunc: func(t *testing.T, c *Copier, output string, _ error) {
				xdgDataHome, _ := c.XDG.DataHome()
				kubectlXDir := filepath.Join(xdgDataHome, "kubectl-x")
				info, err := os.Stat(kubectlXDir)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
				assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
			},
		},
		{
			name: "generates deterministic filename with same seed and time",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				xdgDataHome, err := os.MkdirTemp("", "xdg-*")
				require.NoError(t, err)

				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(999999),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() { _ = os.RemoveAll(xdgDataHome) }
			},
			expectedError: "",
			validateFunc: func(t *testing.T, c *Copier, output string, err error) {
				// With seed 999999, the random suffix should be 8380901385198203263
				expectedSuffix := rand.New(rand.NewSource(999999)).Int63()
				expectedFilename := fmt.Sprintf("kubeconfig-20240115-103045-%d", expectedSuffix)
				assert.Contains(t, output, expectedFilename)
			},
		},

		// Error cases
		{
			name: "error when XDG DataHome returns error",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       &xdgtesting.FakeXDG{DataHomeError: errors.New("xdg error")},
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() {}
			},
			expectedError: "getting XDG data home",
			validateFunc: func(t *testing.T, c *Copier, output string, err error) {
				assert.Empty(t, output, "no output should be written on error")
			},
		},
		{
			name: "error on MkdirAll failure with invalid path",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG("/dev/null/impossible"),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() {}
			},
			expectedError: "not a directory",
		},
		{
			name: "error when kubeconfig WriteToFile fails",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				xdgDataHome, err := os.MkdirTemp("", "xdg-*")
				require.NoError(t, err)

				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					).WithWriteError(errors.New("write failed")),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() { _ = os.RemoveAll(xdgDataHome) }
			},
			expectedError: "writing kubeconfig: write failed",
			validateFunc: func(t *testing.T, c *Copier, output string, err error) {
				assert.Empty(t, output, "no output should be written on write error")
			},
		},
		{
			name: "error when output stream write fails",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				xdgDataHome, err := os.MkdirTemp("", "xdg-*")
				require.NoError(t, err)

				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &failingWriter{}},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() { _ = os.RemoveAll(xdgDataHome) }
			},
			expectedError: "writing output",
		},

		// Edge cases
		{
			name: "handles empty XDG data home path gracefully",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(""),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() {}
			},
			expectedError: "getting XDG data home",
		},
		{
			name: "handles XDG path with special characters",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				// Create a directory with spaces and special characters
				xdgDataHome, err := os.MkdirTemp("", "xdg data-home with spaces*")
				require.NoError(t, err)

				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() { _ = os.RemoveAll(xdgDataHome) }
			},
			expectedError: "",
			validateFunc: func(t *testing.T, c *Copier, output string, err error) {
				assert.Contains(t, output, "kubeconfig-20240115-103045-")
				assert.FileExists(t, output)
				assert.Contains(t, output, "with spaces")
			},
		},
		{
			name: "handles write permission error in target directory",
			setupCopier: func(t *testing.T) (*Copier, func()) {
				if os.Getuid() == 0 {
					t.Skip("Cannot test permission errors as root")
				}
				xdgDataHome, err := os.MkdirTemp("", "xdg-*")
				require.NoError(t, err)

				// Create kubectl-x directory without write permissions
				targetDir := filepath.Join(xdgDataHome, "kubectl-x")
				err = os.MkdirAll(targetDir, 0o555) // read-only
				require.NoError(t, err)

				var outputBuf bytes.Buffer
				copier := &Copier{
					KubeConfig: kubeconfigtesting.NewFakeKubeConfig(
						map[string]*api.Context{"ctx": {Cluster: "cluster"}},
						"ctx",
						"ns",
					),
					IoStreams: genericiooptions.IOStreams{Out: &outputBuf},
					XDG:       xdgtesting.NewFakeXDG(xdgDataHome),
					Rand:      testRand(12345),
					Clock:     testingclock.NewFakeClock(fixedTime),
				}
				return copier, func() {
					_ = os.Chmod(targetDir, 0o755) // Reset permissions for cleanup
					_ = os.RemoveAll(xdgDataHome)
				}
			},
			expectedError: "writing kubeconfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copier, cleanup := tt.setupCopier(t)
			defer cleanup()

			err := copier.Copy(ctx)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.validateFunc != nil {
				output := ""
				if buf, ok := copier.IoStreams.Out.(*bytes.Buffer); ok {
					output = strings.TrimSpace(buf.String())
				}
				tt.validateFunc(t, copier, output, err)
			}
		})
	}
}
