package v1

// Kustomization represents the structure of a kustomization.yaml file.
type Kustomization struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`

	// Resources specifies relative paths to files for the kustomization.
	Resources []string `yaml:"resources,omitempty"`

	// Namespace to add to all resources
	Namespace string `yaml:"namespace,omitempty"`

	// NamePrefix is a prefix appended to resources for Kustomize apps
	NamePrefix string `yaml:"namePrefix,omitempty"`

	// NameSuffix is a suffix appended to resources for Kustomize apps
	NameSuffix string `yaml:"nameSuffix,omitempty"`

	// Labels to add to all resources and selectors
	CommonLabels map[string]string `yaml:"commonLabels,omitempty"`

	// Annotations to add to all resources
	CommonAnnotations map[string]string `yaml:"commonAnnotations,omitempty"`

	// Patches is a list of patches, each of which can be either a strategic merge patch or a JSON patch
	Patches []Patch `yaml:"patches,omitempty"`

	// Components is a list of component paths
	Components []string `yaml:"components,omitempty"`

	// HelmGlobals contains global configuration for Helm charts
	HelmGlobals *HelmGlobals `yaml:"helmGlobals,omitempty"`

	// HelmCharts is a list of Helm charts to inflate
	HelmCharts []HelmChart `yaml:"helmCharts,omitempty"`

	// Generators is a list of generator configuration files
	Generators []string `yaml:"generators,omitempty"`
}

// Patch represents a patch to be applied.
type Patch struct {
	Path    string          `yaml:"path,omitempty"`
	Patch   string          `yaml:"patch,omitempty"`
	Target  *PatchTarget    `yaml:"target,omitempty"`
	Options map[string]bool `yaml:"options,omitempty"`
}

// PatchTarget represents a target for a patch.
type PatchTarget struct {
	Group              string `yaml:"group,omitempty"`
	Version            string `yaml:"version,omitempty"`
	Kind               string `yaml:"kind,omitempty"`
	Name               string `yaml:"name,omitempty"`
	Namespace          string `yaml:"namespace,omitempty"`
	LabelSelector      string `yaml:"labelSelector,omitempty"`
	AnnotationSelector string `yaml:"annotationSelector,omitempty"`
}

// HelmGlobals contains global configuration for Helm charts.
type HelmGlobals struct {
	ChartHome string `yaml:"chartHome,omitempty"`
}

// HelmChart represents a Helm chart configuration.
type HelmChart struct {
	Name                  string         `yaml:"name"`
	Version               string         `yaml:"version,omitempty"`
	ReleaseName           string         `yaml:"releaseName,omitempty"`
	Namespace             string         `yaml:"namespace,omitempty"`
	ValuesFile            string         `yaml:"valuesFile,omitempty"`
	ValuesInline          map[string]any `yaml:"valuesInline,omitempty"`
	AdditionalValuesFiles []string       `yaml:"additionalValuesFiles,omitempty"`
	IncludeCRDs           bool           `yaml:"includeCRDs,omitempty"`
}
