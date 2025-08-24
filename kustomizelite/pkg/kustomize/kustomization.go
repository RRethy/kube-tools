package kustomize

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/RRethy/kube-tools/kustomizelite/api/v1"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/exec"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/generator"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/helm"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/maputils"
	"github.com/RRethy/kube-tools/kustomizelite/pkg/patch"
)

type kustomization struct {
	helmTemplater helm.Templater
	execWrapper   exec.Wrapper
	generator     generator.Generator
}

func NewKustomize(globalHelmValuesFiles []string, opts ...Option) (Kustomizer, error) {
	k := &kustomization{}

	for _, opt := range opts {
		opt(k)
	}

	if k.helmTemplater == nil {
		templater, err := helm.NewTemplaterWithExec(helm.GetBinaryFromEnv(), globalHelmValuesFiles, k.execWrapper)
		if err != nil {
			return nil, fmt.Errorf("creating helm templater: %w", err)
		}
		k.helmTemplater = templater
	}

	if k.generator == nil {
		if k.execWrapper != nil {
			k.generator = generator.New(generator.WithExecWrapper(k.execWrapper))
		} else {
			k.generator = generator.New()
		}
	}

	return k, nil
}

func (k *kustomization) Kustomize(path string, globalHelmValuesFiles []string) ([]map[string]any, error) {
	var resources []map[string]any
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat'ing path %s: %w", path, err)
	}

	if info.IsDir() {
		return k.Kustomize(filepath.Join(path, "kustomization.yaml"), globalHelmValuesFiles)
	}

	base := filepath.Base(path)
	if base != "kustomization.yaml" {
		return nil, fmt.Errorf("file %s is not a kustomization file", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	var kustomizationData v1.Kustomization
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	// decoder.KnownFields(true)
	if err := decoder.Decode(&kustomizationData); err != nil {
		return nil, fmt.Errorf("parsing Kustomization YAML: %w", err)
	}

	var errs []error
	baseDir := filepath.Dir(path)

	for _, resourceItem := range kustomizationData.Resources {
		res, err := k.processResource(filepath.Join(baseDir, resourceItem), globalHelmValuesFiles)
		if err != nil {
			errs = append(errs, fmt.Errorf("processing resource %s: %w", resourceItem, err))
		} else {
			resources = append(resources, res...)
		}
	}

	for _, helmChart := range kustomizationData.HelmCharts {
		helmResources, err := k.helmTemplater.Template(baseDir, helmChart, kustomizationData.HelmGlobals)
		if err != nil {
			errs = append(errs, fmt.Errorf("processing helm chart %s: %w", helmChart.Name, err))
		} else {
			resources = append(resources, helmResources...)
		}
	}

	for _, generatorPath := range kustomizationData.Generators {
		generatorResources, err := k.generator.Generate(filepath.Join(baseDir, generatorPath))
		if err != nil {
			errs = append(errs, fmt.Errorf("processing generator %s: %w", generatorPath, err))
		} else {
			resources = append(resources, generatorResources...)
		}
	}

	for _, component := range kustomizationData.Components {
		resources, err = k.processComponent(filepath.Join(baseDir, component), resources, globalHelmValuesFiles)
		if err != nil {
			errs = append(errs, fmt.Errorf("processing component %s: %w", component, err))
		}
	}

	if kustomizationData.Namespace != "" {
		for _, resource := range resources {
			if err := maputils.Set(resource, "metadata.namespace", kustomizationData.Namespace); err != nil {
				errs = append(errs, fmt.Errorf("setting namespace on resource: %w", err))
			}
		}
	}

	if len(kustomizationData.CommonLabels) > 0 {
		for _, resource := range resources {
			if err := maputils.MergeStringMap(resource, "metadata.labels", kustomizationData.CommonLabels); err != nil {
				errs = append(errs, fmt.Errorf("merging common labels on resource: %w", err))
			}
		}
	}

	if len(kustomizationData.CommonAnnotations) > 0 {
		for _, resource := range resources {
			if err := maputils.MergeStringMap(resource, "metadata.annotations", kustomizationData.CommonAnnotations); err != nil {
				errs = append(errs, fmt.Errorf("merging common annotations on resource: %w", err))
			}
		}
	}

	if kustomizationData.NamePrefix != "" {
		for _, resource := range resources {
			if err := applyNamePrefix(resource, kustomizationData.NamePrefix); err != nil {
				errs = append(errs, fmt.Errorf("applying name prefix: %w", err))
			}
		}
	}

	if kustomizationData.NameSuffix != "" {
		for _, resource := range resources {
			if err := applyNameSuffix(resource, kustomizationData.NameSuffix); err != nil {
				errs = append(errs, fmt.Errorf("applying name suffix: %w", err))
			}
		}
	}

	for _, patch := range kustomizationData.Patches {
		patchObject, err := k.getPatchObject(baseDir, &patch)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for i, resource := range resources {
			if k.resourceMatchesTarget(resource, patch.Target) {
				patchedResource, err := patchObject.Apply(resource)
				if err != nil {
					errs = append(errs, fmt.Errorf("applying patch to resource %d: %w", i, err))
				} else {
					resources[i] = patchedResource
				}
			}
		}
	}

	return resources, errors.Join(errs...)
}

func (k *kustomization) processResource(path string, globalHelmValuesFiles []string) ([]map[string]any, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat'ing resource %s: %w", path, err)
	}

	if info.IsDir() {
		resources, err := k.Kustomize(path, globalHelmValuesFiles)
		if err != nil {
			return resources, fmt.Errorf("kustomizing directory %s: %w", path, err)
		}
		return resources, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var resource map[string]any
	if err := yaml.Unmarshal(content, &resource); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	return []map[string]any{resource}, nil
}

func (k *kustomization) processComponent(path string, resources []map[string]any, globalHelmValuesFiles []string) ([]map[string]any, error) {
	info, err := os.Stat(path)
	if err != nil {
		return resources, fmt.Errorf("stat'ing resource %s: %w", path, err)
	}

	if !info.IsDir() {
		return resources, fmt.Errorf("component %s is not a directory", path)
	}

	componentResources, err := k.Kustomize(path, globalHelmValuesFiles)
	if err != nil {
		return nil, fmt.Errorf("kustomizing component directory %s: %w", path, err)
	}

	// Append component resources to existing resources
	resources = append(resources, componentResources...)
	return resources, nil
}

func applyNamePrefix(resource map[string]any, prefix string) error {
	return applyNameTransform(resource, func(name string) string { return prefix + name })
}

func applyNameSuffix(resource map[string]any, suffix string) error {
	return applyNameTransform(resource, func(name string) string { return name + suffix })
}

func applyNameTransform(resource map[string]any, transform func(string) string) error {
	if !maputils.Has(resource, "metadata") {
		return nil
	}

	for _, field := range []string{"metadata.name", "metadata.generateName"} {
		if !maputils.Has(resource, field) {
			continue
		}

		name, err := maputils.Get[string](resource, field)
		if err != nil {
			return fmt.Errorf("getting %s: %w", field, err)
		}

		if name != "" {
			if err := maputils.Set(resource, field, transform(name)); err != nil {
				return fmt.Errorf("setting %s: %w", field, err)
			}
		}
	}

	return nil
}

func (k *kustomization) resourceMatchesTarget(resource map[string]any, target *v1.PatchTarget) bool {
	if target == nil {
		return true
	}

	if target.Kind != "" {
		kind, _ := maputils.Get[string](resource, "kind")
		if kind != target.Kind {
			return false
		}
	}

	if target.Name != "" {
		name, _ := maputils.Get[string](resource, "metadata.name")
		if name != target.Name {
			return false
		}
	}

	if target.Namespace != "" {
		namespace, _ := maputils.Get[string](resource, "metadata.namespace")
		if namespace != target.Namespace {
			return false
		}
	}

	if target.Group != "" || target.Version != "" {
		apiVersion, _ := maputils.Get[string](resource, "apiVersion")
		if apiVersion != "" {
			gv, err := schema.ParseGroupVersion(apiVersion)
			if err != nil {
				return false
			}

			if target.Group != "" && gv.Group != target.Group {
				return false
			}

			if target.Version != "" && gv.Version != target.Version {
				return false
			}
		} else {
			if target.Group != "" {
				return false
			}
		}
	}

	if target.LabelSelector != "" {
		if !k.matchesLabelSelector(resource, target.LabelSelector) {
			return false
		}
	}

	if target.AnnotationSelector != "" {
		if !k.matchesAnnotationSelector(resource, target.AnnotationSelector) {
			return false
		}
	}

	return true
}

func (k *kustomization) matchesLabelSelector(resource map[string]any, labelSelector string) bool {
	resourceLabels, err := maputils.GetStringMap(resource, "metadata.labels")
	if err != nil && !maputils.Has(resource, "metadata.labels") {
		resourceLabels = make(map[string]string)
	} else if err != nil {
		return false
	}

	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return false
	}

	return selector.Matches(labels.Set(resourceLabels))
}

func (k *kustomization) matchesAnnotationSelector(resource map[string]any, annotationSelector string) bool {
	resourceAnnotations, err := maputils.GetStringMap(resource, "metadata.annotations")
	if err != nil && !maputils.Has(resource, "metadata.annotations") {
		resourceAnnotations = make(map[string]string)
	} else if err != nil {
		return false
	}

	selector, err := labels.Parse(annotationSelector)
	if err != nil {
		return false
	}

	return selector.Matches(labels.Set(resourceAnnotations))
}

func (k *kustomization) getPatchObject(baseDir string, p *v1.Patch) (*patch.Object, error) {
	if p.Path == "" {
		patchObject, err := patch.Parse(p.Patch)
		if err != nil {
			return nil, fmt.Errorf("parsing patch: %w", err)
		}

		return patchObject, nil
	}

	patchObject, err := patch.ParseFile(filepath.Join(baseDir, p.Path))
	if err != nil {
		return nil, fmt.Errorf("loading patch from file %s: %w", p.Path, err)
	}
	return patchObject, nil
}
