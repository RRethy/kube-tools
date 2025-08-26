package hydrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	v1 "github.com/RRethy/kube-tools/klite/api/v1"
)

type Hydrator interface {
	Hydrate(ctx context.Context, path string) ([]*kyaml.RNode, error)
}

type hydrator struct{}

func NewHydrator() Hydrator {
	return &hydrator{}
}

func (h *hydrator) Hydrate(ctx context.Context, path string) ([]*kyaml.RNode, error) {
	kustomization, err := h.resolveKustomizationFile(path)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Kustomization: %+v\n", kustomization)

	return []*kyaml.RNode{}, nil
}

func (h *hydrator) resolveKustomizationFile(path string) (*v1.Kustomization, error) {
	kustomizationPath, err := h.resolveKustomizationPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(kustomizationPath)
	if err != nil {
		return nil, err
	}

	var kustomization v1.Kustomization
	if err := yaml.Unmarshal(data, &kustomization); err != nil {
		return nil, err
	}

	return &kustomization, nil
}

func (h *hydrator) resolveKustomizationPath(path string) (string, error) {
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	var kustomizationPath string
	if info.IsDir() {
		kustomizationPath = filepath.Join(absPath, "kustomization.yaml")
	} else if filepath.Base(absPath) == "kustomization.yaml" {
		kustomizationPath = absPath
	} else {
		return "", fmt.Errorf("file %s is not kustomization.yaml", absPath)
	}

	if _, err := os.Stat(kustomizationPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not found: %s", kustomizationPath)
		}
		return "", err
	}

	return kustomizationPath, nil
}
