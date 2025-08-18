package copy

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestCopier_Copy(t *testing.T) {
	ctx := context.Background()

	t.Run("copies kubeconfig to XDG_DATA_HOME when set", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer os.Remove(sourceFile.Name())

		testContent := "test kubeconfig content"
		_, err = sourceFile.WriteString(testContent)
		require.NoError(t, err)
		sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
		}

		err = copier.Copy(ctx)
		require.NoError(t, err)

		output := strings.TrimSpace(outputBuf.String())
		assert.True(t, strings.HasPrefix(output, filepath.Join(xdgDataHome, "kubectl-x", "kubeconfig-")))

		_, err = os.Stat(output)
		require.NoError(t, err)

		filename := filepath.Base(output)
		matched, err := regexp.MatchString(`^kubeconfig-\d{8}-\d{6}-\d{4}$`, filename)
		require.NoError(t, err)
		assert.True(t, matched, "filename should match pattern kubeconfig-YYYYMMDD-HHMMSS-XXXX")
	})

	t.Run("uses default XDG_DATA_HOME when not set", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer os.Remove(sourceFile.Name())

		testContent := "test kubeconfig content"
		_, err = sourceFile.WriteString(testContent)
		require.NoError(t, err)
		sourceFile.Close()

		homeDir, err := os.MkdirTemp("", "home-*")
		require.NoError(t, err)
		defer os.RemoveAll(homeDir)

		t.Setenv("HOME", homeDir)
		t.Setenv("XDG_DATA_HOME", "")

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
		defer os.Remove(sourceFile.Name())
		sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
		defer os.RemoveAll(xdgDataHome)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
		defer os.Remove(sourceFile.Name())
		sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

		var outputs []string
		for i := 0; i < 3; i++ {
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
		defer os.Remove(sourceFile.Name())
		sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "writing output:")
	})

	t.Run("returns error when HOME is not set and XDG_DATA_HOME is empty", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer os.Remove(sourceFile.Name())
		sourceFile.Close()

		t.Setenv("HOME", "")
		t.Setenv("XDG_DATA_HOME", "")

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
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "$HOME is not defined")
	})

	t.Run("returns error when cannot create target directory", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer os.Remove(sourceFile.Name())
		sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		conflictingFile := filepath.Join(xdgDataHome, "kubectl-x")
		err = os.WriteFile(conflictingFile, []byte("conflict"), 0644)
		require.NoError(t, err)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("returns error when cannot write destination file", func(t *testing.T) {
		sourceFile, err := os.CreateTemp("", "kubeconfig-source-*")
		require.NoError(t, err)
		defer os.Remove(sourceFile.Name())
		sourceFile.Close()

		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		kubectlXDir := filepath.Join(xdgDataHome, "kubectl-x")
		err = os.MkdirAll(kubectlXDir, 0555)
		require.NoError(t, err)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
		}

		err = copier.Copy(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("writes valid merged kubeconfig", func(t *testing.T) {
		xdgDataHome, err := os.MkdirTemp("", "xdg-data-home-*")
		require.NoError(t, err)
		defer os.RemoveAll(xdgDataHome)

		t.Setenv("XDG_DATA_HOME", xdgDataHome)

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
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, os.ErrPermission
}
