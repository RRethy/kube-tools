package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd/api"

	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
)

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name            string
		contextFlag     string
		currentContext  string
		expectedContext string
	}{
		{
			name:            "uses flag override when provided",
			contextFlag:     "override-ctx",
			currentContext:  "current-ctx",
			expectedContext: "override-ctx",
		},
		{
			name:            "uses current context when no flag",
			contextFlag:     "",
			currentContext:  "current-ctx",
			expectedContext: "current-ctx",
		},
		{
			name:            "returns empty when no context set",
			contextFlag:     "",
			currentContext:  "",
			expectedContext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contexts := map[string]*api.Context{
				"current-ctx": {
					Cluster: "current-ctx",
				},
				"override-ctx": {
					Cluster: "override-ctx",
				},
			}
			mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, tt.currentContext, "default")

			var configFlags *genericclioptions.ConfigFlags
			if tt.contextFlag != "" {
				configFlags = genericclioptions.NewConfigFlags(false)
				configFlags.Context = &tt.contextFlag
			}

			resolver := NewResolver(mockKubeconfig, configFlags)
			context := resolver.Resolve()

			assert.Equal(t, tt.expectedContext, context)
		})
	}
}

func TestNewResolver(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {
			Cluster: "test-ctx",
		},
	}
	mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	configFlags := genericclioptions.NewConfigFlags(false)

	resolver := NewResolver(mockKubeconfig, configFlags)

	assert.NotNil(t, resolver)
}

func TestResolver_NilConfigFlags(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {
			Cluster: "test-ctx",
		},
	}
	mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")

	resolver := NewResolver(mockKubeconfig, nil)
	context := resolver.Resolve()

	assert.Equal(t, "test-ctx", context)
}

func TestResolver_EmptyFlagValue(t *testing.T) {
	contexts := map[string]*api.Context{
		"test-ctx": {
			Cluster: "test-ctx",
		},
	}
	mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, "test-ctx", "default")
	
	configFlags := genericclioptions.NewConfigFlags(false)
	emptyString := ""
	configFlags.Context = &emptyString

	resolver := NewResolver(mockKubeconfig, configFlags)
	context := resolver.Resolve()

	assert.Equal(t, "test-ctx", context)
}

func TestResolver_Priority(t *testing.T) {
	tests := []struct {
		name            string
		flagContext     *string
		currentContext  string
		expectedContext string
	}{
		{
			name:            "flag takes precedence over current context",
			flagContext:     stringPtr("flag-ctx"),
			currentContext:  "current-ctx",
			expectedContext: "flag-ctx",
		},
		{
			name:            "nil flag uses current context",
			flagContext:     nil,
			currentContext:  "current-ctx",
			expectedContext: "current-ctx",
		},
		{
			name:            "empty flag uses current context",
			flagContext:     stringPtr(""),
			currentContext:  "current-ctx",
			expectedContext: "current-ctx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contexts := map[string]*api.Context{
				"flag-ctx": {
					Cluster: "flag-ctx",
				},
				"current-ctx": {
					Cluster: "current-ctx",
				},
			}
			mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, tt.currentContext, "default")

			configFlags := genericclioptions.NewConfigFlags(false)
			configFlags.Context = tt.flagContext

			resolver := NewResolver(mockKubeconfig, configFlags)
			context := resolver.Resolve()

			assert.Equal(t, tt.expectedContext, context)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}