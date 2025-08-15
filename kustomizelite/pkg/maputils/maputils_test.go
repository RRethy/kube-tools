package maputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		path    string
		want    any
		wantErr bool
		errMsg  string
	}{
		{
			name: "simple string",
			m: map[string]any{
				"key": "value",
			},
			path: "key",
			want: "value",
		},
		{
			name: "nested string",
			m: map[string]any{
				"metadata": map[string]any{
					"namespace": "default",
				},
			},
			path: "metadata.namespace",
			want: "default",
		},
		{
			name: "deeply nested",
			m: map[string]any{
				"spec": map[string]any{
					"template": map[string]any{
						"metadata": map[string]any{
							"labels": map[string]any{
								"app": "nginx",
							},
						},
					},
				},
			},
			path: "spec.template.metadata.labels.app",
			want: "nginx",
		},
		{
			name: "array access",
			m: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "container1"},
						map[string]any{"name": "container2"},
					},
				},
			},
			path: "spec.containers[1].name",
			want: "container2",
		},
		{
			name:    "nil map",
			m:       nil,
			path:    "key",
			wantErr: true,
			errMsg:  "map is nil",
		},
		{
			name: "key not found",
			m: map[string]any{
				"key": "value",
			},
			path:    "missing",
			wantErr: true,
			errMsg:  `key "missing" not found`,
		},
		{
			name: "type mismatch",
			m: map[string]any{
				"key": 123,
			},
			path:    "key",
			wantErr: true,
			errMsg:  "is int, not string",
		},
		{
			name: "array index out of range",
			m: map[string]any{
				"items": []any{"a", "b"},
			},
			path:    "items[5]",
			wantErr: true,
			errMsg:  "index 5 out of range",
		},
		{
			name: "empty path returns whole map",
			m: map[string]any{
				"key": "value",
			},
			path: "",
			want: map[string]any{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == nil {
				return
			}

			switch expected := tt.want.(type) {
			case string:
				got, err := Get[string](tt.m, tt.path)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errMsg)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expected, got)
				}
			case map[string]any:
				got, err := Get[map[string]any](tt.m, tt.path)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expected, got)
				}
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		path    string
		value   any
		want    map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:  "simple set",
			m:     map[string]any{},
			path:  "key",
			value: "value",
			want: map[string]any{
				"key": "value",
			},
		},
		{
			name:  "nested set with creation",
			m:     map[string]any{},
			path:  "metadata.namespace",
			value: "default",
			want: map[string]any{
				"metadata": map[string]any{
					"namespace": "default",
				},
			},
		},
		{
			name: "overwrite existing",
			m: map[string]any{
				"metadata": map[string]any{
					"namespace": "old",
				},
			},
			path:  "metadata.namespace",
			value: "new",
			want: map[string]any{
				"metadata": map[string]any{
					"namespace": "new",
				},
			},
		},
		{
			name: "set in array",
			m: map[string]any{
				"containers": []any{
					map[string]any{"name": "old"},
				},
			},
			path:  "containers[0].name",
			value: "new",
			want: map[string]any{
				"containers": []any{
					map[string]any{"name": "new"},
				},
			},
		},
		{
			name:    "nil map",
			m:       nil,
			path:    "key",
			value:   "value",
			wantErr: true,
			errMsg:  "map is nil",
		},
		{
			name:    "empty path",
			m:       map[string]any{},
			path:    "",
			value:   "value",
			wantErr: true,
			errMsg:  "empty path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Set(tt.m, tt.path, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tt.m)
			}
		})
	}
}

func TestHas(t *testing.T) {
	m := map[string]any{
		"metadata": map[string]any{
			"namespace": "default",
			"labels": map[string]any{
				"app": "nginx",
			},
		},
		"spec": map[string]any{
			"containers": []any{
				map[string]any{"name": "container1"},
			},
		},
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"root key exists", "metadata", true},
		{"nested key exists", "metadata.namespace", true},
		{"deeply nested exists", "metadata.labels.app", true},
		{"array element exists", "spec.containers[0].name", true},
		{"missing key", "missing", false},
		{"missing nested", "metadata.missing", false},
		{"array out of bounds", "spec.containers[5]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Has(m, tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		path    string
		want    map[string]any
		wantErr bool
	}{
		{
			name: "delete simple key",
			m: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			path: "key1",
			want: map[string]any{
				"key2": "value2",
			},
		},
		{
			name: "delete nested key",
			m: map[string]any{
				"metadata": map[string]any{
					"namespace": "default",
					"name":      "test",
				},
			},
			path: "metadata.namespace",
			want: map[string]any{
				"metadata": map[string]any{
					"name": "test",
				},
			},
		},
		{
			name: "delete non-existent key",
			m: map[string]any{
				"key": "value",
			},
			path: "missing",
			want: map[string]any{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Delete(tt.m, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tt.m)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name    string
		dest    map[string]any
		src     map[string]any
		path    string
		want    map[string]any
		wantErr bool
	}{
		{
			name: "merge at root",
			dest: map[string]any{
				"key1": "value1",
			},
			src: map[string]any{
				"key2": "value2",
			},
			path: "",
			want: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "merge into nested path",
			dest: map[string]any{
				"metadata": map[string]any{
					"namespace": "default",
				},
			},
			src: map[string]any{
				"app":     "nginx",
				"version": "1.0",
			},
			path: "metadata.labels",
			want: map[string]any{
				"metadata": map[string]any{
					"namespace": "default",
					"labels": map[string]any{
						"app":     "nginx",
						"version": "1.0",
					},
				},
			},
		},
		{
			name: "merge overwrites existing",
			dest: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app": "old",
					},
				},
			},
			src: map[string]any{
				"app": "new",
			},
			path: "metadata.labels",
			want: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app": "new",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Merge(tt.dest, tt.src, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, tt.dest)
			}
		})
	}
}

func TestGetOrDefault(t *testing.T) {
	m := map[string]any{
		"key": "value",
	}

	got := GetOrDefault(m, "key", "default")
	assert.Equal(t, "value", got)

	got = GetOrDefault(m, "missing", "default")
	assert.Equal(t, "default", got)
}

func TestMustGet(t *testing.T) {
	m := map[string]any{
		"key": "value",
	}

	got := MustGet[string](m, "key")
	assert.Equal(t, "value", got)

	assert.Panics(t, func() {
		MustGet[string](m, "missing")
	})
}

func TestGetStringMap(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		path    string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "get string map",
			m: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app":     "nginx",
						"version": "1.0",
					},
				},
			},
			path: "metadata.labels",
			want: map[string]string{
				"app":     "nginx",
				"version": "1.0",
			},
		},
		{
			name: "non-string values",
			m: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app":  "nginx",
						"port": 80,
					},
				},
			},
			path:    "metadata.labels",
			wantErr: true,
		},
		{
			name: "root level string map",
			m: map[string]any{
				"app":     "nginx",
				"version": "1.0",
			},
			path: "",
			want: map[string]string{
				"app":     "nginx",
				"version": "1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStringMap(tt.m, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMergeStringMap(t *testing.T) {
	tests := []struct {
		name   string
		m      map[string]any
		path   string
		values map[string]string
		want   map[string]any
	}{
		{
			name: "merge string map",
			m: map[string]any{
				"metadata": map[string]any{},
			},
			path: "metadata.labels",
			values: map[string]string{
				"app":     "nginx",
				"version": "1.0",
			},
			want: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app":     "nginx",
						"version": "1.0",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeStringMap(tt.m, tt.path, tt.values)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.m)
		})
	}
}

func TestGetSlice(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		path    string
		want    any
		wantErr bool
	}{
		{
			name: "get string slice",
			m: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			path: "items",
			want: []string{"a", "b", "c"},
		},
		{
			name: "get map slice",
			m: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "nginx"},
						map[string]any{"name": "redis"},
					},
				},
			},
			path: "spec.containers",
			want: []map[string]any{
				{"name": "nginx"},
				{"name": "redis"},
			},
		},
		{
			name: "type mismatch in slice",
			m: map[string]any{
				"items": []any{"a", 123, "c"},
			},
			path:    "items",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.want.(type) {
			case []string:
				got, err := GetSlice[string](tt.m, tt.path)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expected, got)
				}
			case []map[string]any:
				got, err := GetSlice[map[string]any](tt.m, tt.path)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expected, got)
				}
			}
		})
	}
}

func TestAppendToSlice(t *testing.T) {
	tests := []struct {
		name   string
		m      map[string]any
		path   string
		values []any
		want   map[string]any
	}{
		{
			name: "append to existing slice",
			m: map[string]any{
				"items": []any{"a", "b"},
			},
			path:   "items",
			values: []any{"c", "d"},
			want: map[string]any{
				"items": []any{"a", "b", "c", "d"},
			},
		},
		{
			name:   "create new slice",
			m:      map[string]any{},
			path:   "items",
			values: []any{"a", "b"},
			want: map[string]any{
				"items": []any{"a", "b"},
			},
		},
		{
			name: "append to nested slice",
			m: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "nginx"},
					},
				},
			},
			path: "spec.containers",
			values: []any{
				map[string]any{"name": "redis"},
			},
			want: map[string]any{
				"spec": map[string]any{
					"containers": []any{
						map[string]any{"name": "nginx"},
						map[string]any{"name": "redis"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch values := tt.values[0].(type) {
			case string:
				stringValues := make([]string, len(tt.values))
				for i, v := range tt.values {
					stringValues[i] = v.(string)
				}
				err = AppendToSlice(tt.m, tt.path, stringValues...)
			case map[string]any:
				mapValues := make([]map[string]any, len(tt.values))
				for i, v := range tt.values {
					mapValues[i] = v.(map[string]any)
				}
				err = AppendToSlice(tt.m, tt.path, mapValues...)
			default:
				_ = values
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.m)
		})
	}
}

func TestPrependToSlice(t *testing.T) {
	m := map[string]any{
		"items": []any{"c", "d"},
	}

	err := PrependToSlice(m, "items", "a", "b")
	assert.NoError(t, err)

	expected := map[string]any{
		"items": []any{"a", "b", "c", "d"},
	}
	assert.Equal(t, expected, m)
}

func TestInsertIntoSlice(t *testing.T) {
	m := map[string]any{
		"items": []any{"a", "d"},
	}

	err := InsertIntoSlice(m, "items", 1, "b", "c")
	assert.NoError(t, err)

	expected := map[string]any{
		"items": []any{"a", "b", "c", "d"},
	}
	assert.Equal(t, expected, m)
}

func TestRemoveFromSlice(t *testing.T) {
	m := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	err := RemoveFromSlice(m, "items", 1)
	assert.NoError(t, err)

	expected := map[string]any{
		"items": []any{"a", "c"},
	}
	assert.Equal(t, expected, m)
}

func TestUpdateSliceElement(t *testing.T) {
	m := map[string]any{
		"containers": []any{
			map[string]any{"name": "old"},
			map[string]any{"name": "other"},
		},
	}

	err := UpdateSliceElement(m, "containers", 0, map[string]any{"name": "new"})
	assert.NoError(t, err)

	expected := map[string]any{
		"containers": []any{
			map[string]any{"name": "new"},
			map[string]any{"name": "other"},
		},
	}
	assert.Equal(t, expected, m)
}

func TestSliceLength(t *testing.T) {
	m := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	length, err := SliceLength(m, "items")
	assert.NoError(t, err)
	assert.Equal(t, 3, length)

	_, err = SliceLength(m, "missing")
	assert.Error(t, err)
}

func TestFilterSlice(t *testing.T) {
	m := map[string]any{
		"numbers": []any{1, 2, 3, 4, 5},
	}

	err := FilterSlice(m, "numbers", func(n int) bool {
		return n%2 == 0
	})
	assert.NoError(t, err)

	expected := map[string]any{
		"numbers": []any{2, 4},
	}
	assert.Equal(t, expected, m)
}

func TestMapSlice(t *testing.T) {
	m := map[string]any{
		"numbers": []any{1, 2, 3},
	}

	err := MapSlice(m, "numbers", func(n int) int {
		return n * 2
	})
	assert.NoError(t, err)

	expected := map[string]any{
		"numbers": []any{2, 4, 6},
	}
	assert.Equal(t, expected, m)
}

func TestFindInSlice(t *testing.T) {
	m := map[string]any{
		"containers": []any{
			map[string]any{"name": "nginx", "port": 80},
			map[string]any{"name": "redis", "port": 6379},
		},
	}

	container, index, err := FindInSlice[map[string]any](m, "containers", func(c map[string]any) bool {
		name, _ := c["name"].(string)
		return name == "redis"
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, index)
	assert.Equal(t, "redis", container["name"])
}

func TestContainsInSlice(t *testing.T) {
	m := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	exists, err := ContainsInSlice(m, "items", "b")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = ContainsInSlice(m, "items", "d")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMergeSlices(t *testing.T) {
	tests := []struct {
		name        string
		m           map[string]any
		path        string
		values      []string
		deduplicate bool
		want        map[string]any
	}{
		{
			name: "merge without deduplication",
			m: map[string]any{
				"items": []any{"a", "b"},
			},
			path:        "items",
			values:      []string{"b", "c"},
			deduplicate: false,
			want: map[string]any{
				"items": []any{"a", "b", "b", "c"},
			},
		},
		{
			name: "merge with deduplication",
			m: map[string]any{
				"items": []any{"a", "b"},
			},
			path:        "items",
			values:      []string{"b", "c"},
			deduplicate: true,
			want: map[string]any{
				"items": []any{"a", "b", "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeSlices(tt.m, tt.path, tt.values, tt.deduplicate)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.m)
		})
	}
}

func TestEnsurePath(t *testing.T) {
	m := map[string]any{}

	err := EnsurePath(m, "metadata.labels")
	assert.NoError(t, err)

	expected := map[string]any{
		"metadata": map[string]any{
			"labels": map[string]any{},
		},
	}
	assert.Equal(t, expected, m)

	err = EnsurePath(m, "metadata.annotations")
	assert.NoError(t, err)

	expected["metadata"].(map[string]any)["annotations"] = map[string]any{}
	assert.Equal(t, expected, m)
}

func TestComplexScenarios(t *testing.T) {
	t.Run("kubernetes deployment manipulation", func(t *testing.T) {
		deployment := map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "nginx-deployment",
			},
			"spec": map[string]any{
				"replicas": 3,
				"template": map[string]any{
					"spec": map[string]any{
						"containers": []any{
							map[string]any{
								"name":  "nginx",
								"image": "nginx:1.14.2",
								"ports": []any{
									map[string]any{
										"containerPort": 80,
									},
								},
							},
						},
					},
				},
			},
		}

		err := Set(deployment, "metadata.namespace", "production")
		require.NoError(t, err)

		err = MergeStringMap(deployment, "metadata.labels", map[string]string{
			"app":     "nginx",
			"version": "1.14.2",
		})
		require.NoError(t, err)

		err = AppendToSlice(deployment, "spec.template.spec.containers[0].ports", map[string]any{
			"containerPort": 443,
		})
		require.NoError(t, err)

		namespace, err := Get[string](deployment, "metadata.namespace")
		require.NoError(t, err)
		assert.Equal(t, "production", namespace)

		labels, err := GetStringMap(deployment, "metadata.labels")
		require.NoError(t, err)
		assert.Equal(t, "nginx", labels["app"])

		ports, err := GetSlice[map[string]any](deployment, "spec.template.spec.containers[0].ports")
		require.NoError(t, err)
		assert.Len(t, ports, 2)
	})

	t.Run("kustomization commonLabels", func(t *testing.T) {
		resources := []map[string]any{
			{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]any{
					"name": "my-service",
				},
			},
			{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name": "my-deployment",
					"labels": map[string]any{
						"component": "backend",
					},
				},
			},
		}

		commonLabels := map[string]string{
			"app":         "myapp",
			"environment": "production",
		}

		for _, resource := range resources {
			err := MergeStringMap(resource, "metadata.labels", commonLabels)
			require.NoError(t, err)
		}

		serviceLabels, err := GetStringMap(resources[0], "metadata.labels")
		require.NoError(t, err)
		assert.Equal(t, "myapp", serviceLabels["app"])
		assert.Equal(t, "production", serviceLabels["environment"])

		deploymentLabels, err := GetStringMap(resources[1], "metadata.labels")
		require.NoError(t, err)
		assert.Equal(t, "backend", deploymentLabels["component"])
		assert.Equal(t, "myapp", deploymentLabels["app"])
	})
}
