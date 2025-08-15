package v1

// BatchBuild represents the structure of a batch build configuration file.
type BatchBuild struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	// Env contains global environment variables applied to all builds
	Env []EnvVar `yaml:"env,omitempty"`

	// Builds contains the list of build configurations to execute
	Builds []BuildConfig `yaml:"builds"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// BuildConfig represents a single build configuration within a batch.
type BuildConfig struct {
	// Kustomization is the path to the kustomization.yaml file
	Kustomization string `yaml:"kustomization"`

	// Output is the path where the built YAML should be written
	Output string `yaml:"output"`

	// Env contains environment variables specific to this build
	Env []EnvVar `yaml:"env,omitempty"`
}
