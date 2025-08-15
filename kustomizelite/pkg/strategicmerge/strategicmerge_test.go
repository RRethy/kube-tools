package strategicmerge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	tests := []struct {
		name     string
		resource map[string]any
		patch    map[string]any
		expected map[string]any
	}{
		{
			name: "simple merge",
			resource: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"spec": map[string]any{
					"replicas": 3,
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"replicas": 7,
				},
			},
			expected: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"spec": map[string]any{
					"replicas": 7,
				},
			},
		},
		{
			name: "nested merge with preservation",
			resource: map[string]any{
				"apiVersion": "apps/v1",
				"metadata": map[string]any{
					"name": "test-deployment",
					"labels": map[string]any{
						"version": "v1",
					},
				},
				"spec": map[string]any{
					"replicas": 3,
				},
			},
			patch: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app": "updated",
					},
				},
				"spec": map[string]any{
					"replicas": 5,
				},
			},
			expected: map[string]any{
				"apiVersion": "apps/v1",
				"metadata": map[string]any{
					"name": "test-deployment",
					"labels": map[string]any{
						"version": "v1",
						"app":     "updated",
					},
				},
				"spec": map[string]any{
					"replicas": 5,
				},
			},
		},
		{
			name: "nil value deletion",
			resource: map[string]any{
				"metadata": map[string]any{
					"name":   "test",
					"labels": map[string]any{"app": "test"},
				},
			},
			patch: map[string]any{
				"metadata": map[string]any{
					"labels": nil,
				},
			},
			expected: map[string]any{
				"metadata": map[string]any{
					"name": "test",
				},
			},
		},
		{
			name: "same type list merge with deduplication",
			resource: map[string]any{
				"spec": map[string]any{
					"finalizers": []any{"finalizer1", "finalizer2"},
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"finalizers": []any{"finalizer2", "finalizer3"},
				},
			},
			expected: map[string]any{
				"spec": map[string]any{
					"finalizers": []any{"finalizer1", "finalizer2", "finalizer3"},
				},
			},
		},
		{
			name: "different type list merge - replace",
			resource: map[string]any{
				"spec": map[string]any{
					"items": []any{"string1", "string2"},
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"items": []any{1, 2, 3},
				},
			},
			expected: map[string]any{
				"spec": map[string]any{
					"items": []any{1, 2, 3},
				},
			},
		},
		{
			name: "mixed type list in resource - replace",
			resource: map[string]any{
				"spec": map[string]any{
					"items": []any{"string1", 42, "string2"},
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"items": []any{"new1", "new2"},
				},
			},
			expected: map[string]any{
				"spec": map[string]any{
					"items": []any{"new1", "new2"},
				},
			},
		},
		{
			name: "mixed type list in patch - replace",
			resource: map[string]any{
				"spec": map[string]any{
					"items": []any{"string1", "string2"},
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"items": []any{"mixed", 42, true},
				},
			},
			expected: map[string]any{
				"spec": map[string]any{
					"items": []any{"mixed", 42, true},
				},
			},
		},
		{
			name: "object list merging with strategic keys",
			resource: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "app", "image": "app:v1", "ports": []any{8080}},
						map[string]any{"name": "sidecar", "image": "sidecar:v1"},
					},
					"volumes": []any{
						map[string]any{"name": "data", "mountPath": "/data", "size": "10Gi"},
					},
				},
			},
			patch: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "app", "image": "app:v2", "env": []any{"NEW_VAR=value"}},
						map[string]any{"name": "cache", "image": "cache:v1"},
					},
					"volumes": []any{
						map[string]any{"name": "data", "size": "20Gi"},
						map[string]any{"name": "logs", "mountPath": "/logs"},
					},
				},
			},
			expected: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "app", "image": "app:v2", "ports": []any{8080}, "env": []any{"NEW_VAR=value"}},
						map[string]any{"name": "sidecar", "image": "sidecar:v1"},
						map[string]any{"name": "cache", "image": "cache:v1"},
					},
					"volumes": []any{
						map[string]any{"name": "data", "mountPath": "/data", "size": "20Gi"},
						map[string]any{"name": "logs", "mountPath": "/logs"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Apply(tt.resource, tt.patch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeSlices(t *testing.T) {
	tests := []struct {
		name     string
		existing []any
		patch    []any
		expected []any
	}{
		{
			name:     "same type strings with deduplication",
			existing: []any{"a", "b", "c"},
			patch:    []any{"b", "c", "d"},
			expected: []any{"a", "b", "c", "d"},
		},
		{
			name:     "same type numbers with deduplication",
			existing: []any{1, 2, 3},
			patch:    []any{3, 4, 5},
			expected: []any{1, 2, 3, 4, 5},
		},
		{
			name:     "different types - string and int",
			existing: []any{"a", "b"},
			patch:    []any{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "different types - maps and strings",
			existing: []any{map[string]any{"name": "item1"}},
			patch:    []any{"string1", "string2"},
			expected: []any{"string1", "string2"},
		},
		{
			name:     "same type maps - no deduplication for maps",
			existing: []any{map[string]any{"name": "item1"}},
			patch:    []any{map[string]any{"name": "item2"}},
			expected: []any{map[string]any{"name": "item1"}, map[string]any{"name": "item2"}},
		},
		{
			name:     "empty existing",
			existing: []any{},
			patch:    []any{"a", "b"},
			expected: []any{"a", "b"},
		},
		{
			name:     "empty patch",
			existing: []any{"a", "b"},
			patch:    []any{},
			expected: []any{"a", "b"},
		},
		{
			name:     "both nil",
			existing: nil,
			patch:    nil,
			expected: nil,
		},
		{
			name:     "mixed types in existing - replace",
			existing: []any{"string", 42, true},
			patch:    []any{"new1", "new2"},
			expected: []any{"new1", "new2"},
		},
		{
			name:     "mixed types in patch - replace",
			existing: []any{"a", "b"},
			patch:    []any{"mixed", 42, true},
			expected: []any{"mixed", 42, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeSlices(tt.existing, tt.patch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHaveSameElementTypes(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []any
		slice2   []any
		expected bool
	}{
		{
			name:     "both strings",
			slice1:   []any{"a", "b"},
			slice2:   []any{"c", "d"},
			expected: true,
		},
		{
			name:     "both numbers",
			slice1:   []any{1, 2},
			slice2:   []any{3, 4},
			expected: true,
		},
		{
			name:     "both maps",
			slice1:   []any{map[string]any{"key": "value1"}},
			slice2:   []any{map[string]any{"key": "value2"}},
			expected: true,
		},
		{
			name:     "string vs int",
			slice1:   []any{"a", "b"},
			slice2:   []any{1, 2},
			expected: false,
		},
		{
			name:     "map vs string",
			slice1:   []any{map[string]any{"key": "value"}},
			slice2:   []any{"string"},
			expected: false,
		},
		{
			name:     "empty slice1",
			slice1:   []any{},
			slice2:   []any{"a", "b"},
			expected: true,
		},
		{
			name:     "empty slice2",
			slice1:   []any{"a", "b"},
			slice2:   []any{},
			expected: true,
		},
		{
			name:     "both empty",
			slice1:   []any{},
			slice2:   []any{},
			expected: true,
		},
		{
			name:     "mixed types in slice1",
			slice1:   []any{"string", 42},
			slice2:   []any{"a", "b"},
			expected: false,
		},
		{
			name:     "mixed types in slice2",
			slice1:   []any{"a", "b"},
			slice2:   []any{"string", 42},
			expected: false,
		},
		{
			name:     "mixed types in both slices",
			slice1:   []any{"string", 42},
			slice2:   []any{1, "text"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := haveSameElementTypes(tt.slice1, tt.slice2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetElementType(t *testing.T) {
	tests := []struct {
		name     string
		item     any
		expected string
	}{
		{
			name:     "string",
			item:     "hello",
			expected: "string",
		},
		{
			name:     "int",
			item:     42,
			expected: "int",
		},
		{
			name:     "float",
			item:     3.14,
			expected: "float",
		},
		{
			name:     "bool",
			item:     true,
			expected: "bool",
		},
		{
			name:     "map",
			item:     map[string]any{"key": "value"},
			expected: "map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getElementType(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeObjectLists(t *testing.T) {
	tests := []struct {
		name     string
		existing []any
		patch    []any
		expected []any
	}{
		{
			name:     "empty existing",
			existing: []any{},
			patch: []any{
				map[string]any{"name": "app"},
			},
			expected: []any{
				map[string]any{"name": "app"},
			},
		},
		{
			name: "empty patch",
			existing: []any{
				map[string]any{"name": "app"},
			},
			patch: []any{},
			expected: []any{
				map[string]any{"name": "app"},
			},
		},
		{
			name: "no common keys - replace",
			existing: []any{
				map[string]any{"foo": "bar"},
			},
			patch: []any{
				map[string]any{"baz": "qux"},
			},
			expected: []any{
				map[string]any{"baz": "qux"},
			},
		},
		{
			name: "merge by name key",
			existing: []any{
				map[string]any{"name": "app", "port": 8080},
				map[string]any{"name": "db", "port": 5432},
			},
			patch: []any{
				map[string]any{"name": "app", "port": 9090, "env": "prod"},
				map[string]any{"name": "cache", "port": 6379},
			},
			expected: []any{
				map[string]any{"name": "app", "port": 9090, "env": "prod"},
				map[string]any{"name": "db", "port": 5432},
				map[string]any{"name": "cache", "port": 6379},
			},
		},
		{
			name: "merge by containerPort key",
			existing: []any{
				map[string]any{"containerPort": "8080", "protocol": "TCP"},
				map[string]any{"containerPort": "8443", "protocol": "TCP"},
			},
			patch: []any{
				map[string]any{"containerPort": "8080", "protocol": "UDP"},
				map[string]any{"containerPort": "9090", "protocol": "TCP"},
			},
			expected: []any{
				map[string]any{"containerPort": "8080", "protocol": "UDP"},
				map[string]any{"containerPort": "8443", "protocol": "TCP"},
				map[string]any{"containerPort": "9090", "protocol": "TCP"},
			},
		},
		{
			name: "merge by mountPath key",
			existing: []any{
				map[string]any{"mountPath": "/data", "volume": "data-vol"},
				map[string]any{"mountPath": "/config", "volume": "config-vol"},
			},
			patch: []any{
				map[string]any{"mountPath": "/data", "volume": "new-data-vol", "readOnly": true},
			},
			expected: []any{
				map[string]any{"mountPath": "/data", "volume": "new-data-vol", "readOnly": true},
				map[string]any{"mountPath": "/config", "volume": "config-vol"},
			},
		},
		{
			name: "merge by non-preferred key",
			existing: []any{
				map[string]any{"id": "1", "value": "old"},
				map[string]any{"id": "2", "value": "keep"},
			},
			patch: []any{
				map[string]any{"id": "1", "value": "new"},
				map[string]any{"id": "3", "value": "add"},
			},
			expected: []any{
				map[string]any{"id": "1", "value": "new"},
				map[string]any{"id": "2", "value": "keep"},
				map[string]any{"id": "3", "value": "add"},
			},
		},
		{
			name: "non-string merge key values ignored",
			existing: []any{
				map[string]any{"port": 8080, "name": "http"},
				map[string]any{"port": 8443, "name": "https"},
			},
			patch: []any{
				map[string]any{"port": 8080, "name": "http-updated"},
			},
			expected: []any{
				map[string]any{"port": 8080, "name": "http"},
				map[string]any{"port": 8443, "name": "https"},
				map[string]any{"port": 8080, "name": "http-updated"},
			},
		},
		{
			name: "missing merge key in patch item",
			existing: []any{
				map[string]any{"name": "app", "port": 8080},
			},
			patch: []any{
				map[string]any{"port": 9090},
			},
			expected: []any{
				map[string]any{"name": "app", "port": 8080},
				map[string]any{"port": 9090},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeObjectLists(tt.existing, tt.patch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindMergeKey(t *testing.T) {
	tests := []struct {
		name        string
		existing    []any
		patch       []any
		expectedKey string
		expectedOk  bool
	}{
		{
			name:        "empty slices",
			existing:    []any{},
			patch:       []any{},
			expectedKey: "",
			expectedOk:  false,
		},
		{
			name:     "empty existing",
			existing: []any{},
			patch: []any{
				map[string]any{"name": "app"},
			},
			expectedKey: "",
			expectedOk:  false,
		},
		{
			name: "empty patch",
			existing: []any{
				map[string]any{"name": "app"},
			},
			patch:       []any{},
			expectedKey: "",
			expectedOk:  false,
		},
		{
			name: "prefer name",
			existing: []any{
				map[string]any{"name": "a", "id": "1"},
			},
			patch: []any{
				map[string]any{"name": "b", "id": "2"},
			},
			expectedKey: "name",
			expectedOk:  true,
		},
		{
			name: "prefer key",
			existing: []any{
				map[string]any{"key": "a", "id": "1"},
			},
			patch: []any{
				map[string]any{"key": "b", "id": "2"},
			},
			expectedKey: "key",
			expectedOk:  true,
		},
		{
			name: "prefer type",
			existing: []any{
				map[string]any{"type": "a", "id": "1"},
			},
			patch: []any{
				map[string]any{"type": "b", "id": "2"},
			},
			expectedKey: "type",
			expectedOk:  true,
		},
		{
			name: "prefer containerPort",
			existing: []any{
				map[string]any{"containerPort": "8080", "id": "1"},
			},
			patch: []any{
				map[string]any{"containerPort": "9090", "id": "2"},
			},
			expectedKey: "containerPort",
			expectedOk:  true,
		},
		{
			name: "no common keys",
			existing: []any{
				map[string]any{"foo": "a"},
			},
			patch: []any{
				map[string]any{"bar": "b"},
			},
			expectedKey: "",
			expectedOk:  false,
		},
		{
			name: "common key not in all items",
			existing: []any{
				map[string]any{"name": "a", "id": "1"},
				map[string]any{"id": "2"},
			},
			patch: []any{
				map[string]any{"name": "c", "id": "3"},
			},
			expectedKey: "id",
			expectedOk:  true,
		},
		{
			name: "non-map items",
			existing: []any{
				"string",
			},
			patch: []any{
				map[string]any{"name": "a"},
			},
			expectedKey: "",
			expectedOk:  false,
		},
		{
			name: "multiple common keys - prefer name",
			existing: []any{
				map[string]any{"name": "a", "type": "t1", "id": "1"},
			},
			patch: []any{
				map[string]any{"name": "b", "type": "t2", "id": "2"},
			},
			expectedKey: "name",
			expectedOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, ok := findMergeKey(tt.existing, tt.patch)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedOk, ok)
		})
	}
}

func TestGetCommonKeys(t *testing.T) {
	tests := []struct {
		name     string
		items    []any
		expected map[string]bool
	}{
		{
			name:     "empty slice",
			items:    []any{},
			expected: nil,
		},
		{
			name: "single item",
			items: []any{
				map[string]any{"a": 1, "b": 2},
			},
			expected: map[string]bool{"a": true, "b": true},
		},
		{
			name: "all common keys",
			items: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"a": 3, "b": 4},
			},
			expected: map[string]bool{"a": true, "b": true},
		},
		{
			name: "some common keys",
			items: []any{
				map[string]any{"a": 1, "b": 2, "c": 3},
				map[string]any{"a": 4, "b": 5},
				map[string]any{"a": 6, "b": 7, "d": 8},
			},
			expected: map[string]bool{"a": true, "b": true},
		},
		{
			name: "no common keys",
			items: []any{
				map[string]any{"a": 1},
				map[string]any{"b": 2},
			},
			expected: map[string]bool{},
		},
		{
			name: "non-map item",
			items: []any{
				"not a map",
			},
			expected: nil,
		},
		{
			name: "mixed types",
			items: []any{
				map[string]any{"a": 1},
				"not a map",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCommonKeys(tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}
