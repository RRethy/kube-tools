package kubeconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestKubeConfig_Contexts(t *testing.T) {
	kubeConfig := KubeConfig{
		APIConfig: &api.Config{
			Contexts: map[string]*api.Context{
				"context1": {},
				"context2": {},
			},
		},
	}

	contexts := kubeConfig.Contexts()
	assert.ElementsMatch(t, []string{"context1", "context2"}, contexts)
}

func TestKubeConfig_UseContext(t *testing.T) {
	kubeConfig := KubeConfig{
		APIConfig: &api.Config{
			Contexts: map[string]*api.Context{
				"context1": {},
				"context2": {
					Namespace: "namespace2",
				},
			},
		},
	}

	err := kubeConfig.SetContext("context1")
	require.Nil(t, err)
	assert.Equal(t, "context1", kubeConfig.APIConfig.CurrentContext)
	assert.Equal(t, "default", kubeConfig.APIConfig.Contexts["context1"].Namespace)

	err = kubeConfig.SetContext("context2")
	require.Nil(t, err)
	assert.Equal(t, "context2", kubeConfig.APIConfig.CurrentContext)
	assert.Equal(t, "namespace2", kubeConfig.APIConfig.Contexts["context2"].Namespace)
}

func TestKubeConfig_UseNamespace(t *testing.T) {
	kubeConfig := KubeConfig{
		APIConfig: &api.Config{
			Contexts: map[string]*api.Context{
				"context1": {},
			},
		},
	}

	err := kubeConfig.SetContext("context1")
	require.Nil(t, err)
	err = kubeConfig.SetNamespace("namespace1")
	require.Nil(t, err)
	assert.Equal(t, "namespace1", kubeConfig.APIConfig.Contexts["context1"].Namespace)
}

func TestKubeConfig_CurrentContext(t *testing.T) {
	tests := []struct {
		name      string
		APIConfig *api.Config
		expected  string
		err       bool
		errMsg    string
	}{
		{
			name: "correct context when set",
			APIConfig: &api.Config{
				CurrentContext: "context1",
				Contexts: map[string]*api.Context{
					"context1": {},
					"context2": {},
				},
			},
			expected: "context1",
			err:      false,
			errMsg:   "",
		},
		{
			name: "error when context not set",
			APIConfig: &api.Config{
				CurrentContext: "",
				Contexts: map[string]*api.Context{
					"context1": {},
					"context2": {},
				},
			},
			expected: "",
			err:      true,
			errMsg:   "current context not set",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, err := KubeConfig{APIConfig: test.APIConfig}.GetCurrentContext()
			if test.err {
				require.NotNil(t, err)
				assert.Equal(t, test.errMsg, err.Error())
			} else {
				require.Nil(t, err)
				assert.Equal(t, test.expected, context)
			}
		})
	}
}

func TestKubeConfig_CurrentNamespace(t *testing.T) {
	kubeConfig := KubeConfig{
		APIConfig: &api.Config{
			CurrentContext: "context1",
			Contexts: map[string]*api.Context{
				"context1": {
					Namespace: "namespace1",
				},
			},
		},
	}

	namespace, err := kubeConfig.GetCurrentNamespace()
	require.Nil(t, err)
	assert.Equal(t, "namespace1", namespace)
}

func TestNewKubeConfig_UsesKUBECONFIG(t *testing.T) {
	// Create a temporary kubeconfig file
	tmpFile, err := os.CreateTemp("", "test-kubeconfig-*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write minimal valid kubeconfig
	validConfig := `apiVersion: v1
kind: Config
clusters: []
contexts: []
users: []
current-context: ""
`
	_, err = tmpFile.WriteString(validConfig)
	require.NoError(t, err)
	tmpFile.Close()

	t.Setenv("KUBECONFIG", tmpFile.Name())

	kubeConfig, err := NewKubeConfig()

	require.NotNil(t, kubeConfig)
	require.Nil(t, err)
}
