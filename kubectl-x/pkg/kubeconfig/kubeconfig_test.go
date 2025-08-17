package kubeconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestKubeConfig_Contexts(t *testing.T) {
	kubeConfig := KubeConfig{
		apiConfig: &api.Config{
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
		apiConfig: &api.Config{
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
	assert.Equal(t, "context1", kubeConfig.apiConfig.CurrentContext)
	assert.Equal(t, "default", kubeConfig.apiConfig.Contexts["context1"].Namespace)

	err = kubeConfig.SetContext("context2")
	require.Nil(t, err)
	assert.Equal(t, "context2", kubeConfig.apiConfig.CurrentContext)
	assert.Equal(t, "namespace2", kubeConfig.apiConfig.Contexts["context2"].Namespace)
}

func TestKubeConfig_UseNamespace(t *testing.T) {
	kubeConfig := KubeConfig{
		apiConfig: &api.Config{
			Contexts: map[string]*api.Context{
				"context1": {},
			},
		},
	}

	err := kubeConfig.SetContext("context1")
	require.Nil(t, err)
	err = kubeConfig.SetNamespace("namespace1")
	require.Nil(t, err)
	assert.Equal(t, "namespace1", kubeConfig.apiConfig.Contexts["context1"].Namespace)
}

func TestKubeConfig_CurrentContext(t *testing.T) {
	tests := []struct {
		name      string
		apiConfig *api.Config
		expected  string
		err       bool
		errMsg    string
	}{
		{
			name: "correct context when set",
			apiConfig: &api.Config{
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
			apiConfig: &api.Config{
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
			context, err := KubeConfig{apiConfig: test.apiConfig}.GetCurrentContext()
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
		apiConfig: &api.Config{
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
	t.Setenv("KUBECONFIG", "/tmp/test-kubeconfig:/tmp/test-kubeconfig2")

	kubeConfig, err := NewKubeConfig()

	require.NotNil(t, kubeConfig)
	require.Nil(t, err)
}
