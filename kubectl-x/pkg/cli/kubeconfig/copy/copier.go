// Package copy provides functionality to copy kubeconfig files
package copy

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/RRethy/kubectl-x/pkg/kubeconfig"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"
)

// Copier handles kubeconfig copying operations
type Copier struct {
	KubeConfig kubeconfig.Interface
	IoStreams  genericiooptions.IOStreams
}

// Copy copies the current kubeconfig to $XDG_DATA_HOME and prints the location
func (c *Copier) Copy(ctx context.Context) error {
	klog.V(1).Info("Copying kubeconfig to $XDG_DATA_HOME")

	kubeconfigPath := c.KubeConfig.GetKubeconfigPath()
	klog.V(4).Infof("Source kubeconfig path: %s", kubeconfigPath)

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		xdgDataHome = filepath.Join(homeDir, ".local", "share")
	}
	klog.V(4).Infof("Using $XDG_DATA_HOME: %s", xdgDataHome)

	targetDir := filepath.Join(xdgDataHome, "kubectl-x")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102-150405")
	randomSuffix := rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000)
	filename := fmt.Sprintf("kubeconfig-%s-%04d", timestamp, randomSuffix)
	targetPath := filepath.Join(targetDir, filename)

	if err := copyFile(kubeconfigPath, targetPath); err != nil {
		return err
	}

	klog.V(1).Infof("Kubeconfig copied to: %s", targetPath)
	_, err := c.IoStreams.Out.Write([]byte(targetPath + "\n"))
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
