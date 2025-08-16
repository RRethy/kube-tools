package namespace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd/api"
	
	kubeconfigtesting "github.com/RRethy/kubectl-x/pkg/kubeconfig/testing"
)

func TestResolver_ResolveNamespace(t *testing.T) {
	tests := []struct {
		name              string
		contextName       string
		namespaceFlag     string
		contextNamespace  string
		expectedNamespace string
	}{
		{
			name:              "uses flag override when provided",
			contextName:       "test-context",
			namespaceFlag:     "override-ns",
			contextNamespace:  "context-ns",
			expectedNamespace: "override-ns",
		},
		{
			name:              "uses context namespace when no flag",
			contextName:       "test-context",
			namespaceFlag:     "",
			contextNamespace:  "context-ns",
			expectedNamespace: "context-ns",
		},
		{
			name:              "defaults to 'default' when no namespace set",
			contextName:       "test-context",
			namespaceFlag:     "",
			contextNamespace:  "",
			expectedNamespace: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contexts := map[string]*api.Context{
				tt.contextName: {
					Cluster:   tt.contextName,
					Namespace: tt.contextNamespace,
				},
			}
			mockKubeconfig := kubeconfigtesting.NewFakeKubeConfig(contexts, tt.contextName, tt.contextNamespace)

			var configFlags *genericclioptions.ConfigFlags
			if tt.namespaceFlag != "" {
				configFlags = genericclioptions.NewConfigFlags(false)
				configFlags.Namespace = &tt.namespaceFlag
			}

			resolver := NewResolver(mockKubeconfig, configFlags)
			namespace := resolver.Resolve(tt.contextName)

			assert.Equal(t, tt.expectedNamespace, namespace)
		})
	}
}