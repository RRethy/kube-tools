package hydrate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	v1 "github.com/RRethy/kube-tools/k2/api/v1"
)

type HydratedResult struct {
	Nodes    []*kyaml.RNode
	Metadata metav1.ObjectMeta
}

type Hydrator interface {
	Hydrate(ctx context.Context, path string) (*HydratedResult, error)
}

type hydrator struct{}

func NewHydrator() Hydrator {
	return &hydrator{}
}

func (h *hydrator) Hydrate(ctx context.Context, path string) (*HydratedResult, error) {
	kustomization, baseDir, err := h.resolveKustomizationFile(path)
	if err != nil {
		return nil, err
	}

	nodes, err := h.loadResources(kustomization.Resources, baseDir)
	if err != nil {
		return nil, err
	}

	// TODO: h.loadComponents

	result := &HydratedResult{
		Nodes: nodes,
	}
	if kustomization.Metadata != nil {
		result.Metadata = *kustomization.Metadata
	}
	return result, nil
}

func (h *hydrator) loadResources(resources []string, baseDir string) ([]*kyaml.RNode, error) {
	nodes := []*kyaml.RNode{}
	for _, resource := range resources {
		resourcePath := filepath.Join(baseDir, resource)
		resourceNodes, err := h.loadResource(resourcePath)
		if err != nil {
			return nil, fmt.Errorf("loading resource %s: %w", resource, err)
		}
		nodes = append(nodes, resourceNodes...)
	}
	return nodes, nil
}

func (h *hydrator) resolveKustomizationFile(path string) (*v1.Kustomization, string, error) {
	kustomizationPath, err := h.resolveKustomizationPath(path)
	if err != nil {
		return nil, "", err
	}

	baseDir := filepath.Dir(kustomizationPath)

	data, err := os.ReadFile(kustomizationPath)
	if err != nil {
		return nil, "", err
	}

	var kustomization v1.Kustomization
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&kustomization); err != nil {
		return nil, "", err
	}

	return &kustomization, baseDir, nil
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

func (h *hydrator) loadResource(resourcePath string) ([]*kyaml.RNode, error) {
	info, err := os.Stat(resourcePath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		result, err := h.Hydrate(context.Background(), resourcePath)
		if err != nil {
			return nil, err
		}
		return result.Nodes, nil
	}

	data, err := os.ReadFile(resourcePath)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return []*kyaml.RNode{}, nil
	}

	node, err := kyaml.Parse(string(data))
	if err != nil {
		return nil, err
	}

	return []*kyaml.RNode{node}, nil
}
