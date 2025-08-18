package xdg

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("returns non-nil Interface", func(t *testing.T) {
		xdg := New()
		assert.NotNil(t, xdg)
	})

	t.Run("creates independent instances", func(t *testing.T) {
		xdg1 := New()
		xdg2 := New()
		assert.NotSame(t, xdg1, xdg2, "New() should create independent instances")
	})
}

func TestDataHome(t *testing.T) {
	tests := []struct {
		name        string
		xdgDataHome string
		homeDir     string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "uses XDG_DATA_HOME when set",
			xdgDataHome: "/custom/data",
			want:        "/custom/data",
		},
		{
			name:        "uses XDG_DATA_HOME with trailing slash",
			xdgDataHome: "/custom/data/",
			want:        "/custom/data/",
		},
		{
			name:    "uses HOME/.local/share when XDG_DATA_HOME not set",
			homeDir: "/home/user",
			want:    "/home/user/.local/share",
		},
		{
			name:        "prefers XDG_DATA_HOME over HOME",
			xdgDataHome: "/xdg/data",
			homeDir:     "/home/user",
			want:        "/xdg/data",
		},
		{
			name:        "returns error when HOME not set and XDG_DATA_HOME empty",
			wantErr:     true,
			errContains: "$HOME is not defined",
		},
		{
			name:        "handles whitespace-only XDG_DATA_HOME as valid path",
			xdgDataHome: "   ",
			want:        "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.xdgDataHome != "" {
				t.Setenv("XDG_DATA_HOME", tt.xdgDataHome)
			} else {
				t.Setenv("XDG_DATA_HOME", "")
			}

			if tt.homeDir != "" {
				t.Setenv("HOME", tt.homeDir)
			} else {
				t.Setenv("HOME", "")
			}

			xdg := New()
			got, err := xdg.DataHome()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConfigHome(t *testing.T) {
	tests := []struct {
		name          string
		xdgConfigHome string
		homeDir       string
		want          string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "uses XDG_CONFIG_HOME when set",
			xdgConfigHome: "/custom/config",
			want:          "/custom/config",
		},
		{
			name:          "uses XDG_CONFIG_HOME with trailing slash",
			xdgConfigHome: "/custom/config/",
			want:          "/custom/config/",
		},
		{
			name:    "uses HOME/.config when XDG_CONFIG_HOME not set",
			homeDir: "/home/user",
			want:    "/home/user/.config",
		},
		{
			name:          "prefers XDG_CONFIG_HOME over HOME",
			xdgConfigHome: "/xdg/config",
			homeDir:       "/home/user",
			want:          "/xdg/config",
		},
		{
			name:        "returns error when HOME not set and XDG_CONFIG_HOME empty",
			wantErr:     true,
			errContains: "$HOME is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.xdgConfigHome != "" {
				t.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else {
				t.Setenv("XDG_CONFIG_HOME", "")
			}

			if tt.homeDir != "" {
				t.Setenv("HOME", tt.homeDir)
			} else {
				t.Setenv("HOME", "")
			}

			xdg := New()
			got, err := xdg.ConfigHome()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCacheHome(t *testing.T) {
	tests := []struct {
		name         string
		xdgCacheHome string
		homeDir      string
		want         string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "uses XDG_CACHE_HOME when set",
			xdgCacheHome: "/custom/cache",
			want:         "/custom/cache",
		},
		{
			name:         "uses XDG_CACHE_HOME with trailing slash",
			xdgCacheHome: "/custom/cache/",
			want:         "/custom/cache/",
		},
		{
			name:    "uses HOME/.cache when XDG_CACHE_HOME not set",
			homeDir: "/home/user",
			want:    "/home/user/.cache",
		},
		{
			name:         "prefers XDG_CACHE_HOME over HOME",
			xdgCacheHome: "/xdg/cache",
			homeDir:      "/home/user",
			want:         "/xdg/cache",
		},
		{
			name:        "returns error when HOME not set and XDG_CACHE_HOME empty",
			wantErr:     true,
			errContains: "$HOME is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.xdgCacheHome != "" {
				t.Setenv("XDG_CACHE_HOME", tt.xdgCacheHome)
			} else {
				t.Setenv("XDG_CACHE_HOME", "")
			}

			if tt.homeDir != "" {
				t.Setenv("HOME", tt.homeDir)
			} else {
				t.Setenv("HOME", "")
			}

			xdg := New()
			got, err := xdg.CacheHome()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	// XDG should be safe for concurrent use
	xdg := New()

	// Set up environment
	homeDir, err := os.MkdirTemp("", "test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(homeDir)
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_DATA_HOME", "/data")
	t.Setenv("XDG_CONFIG_HOME", "/config")
	t.Setenv("XDG_CACHE_HOME", "/cache")

	var wg sync.WaitGroup
	iterations := 100

	// Test concurrent DataHome calls
	wg.Add(iterations)
	for range iterations {
		go func() {
			defer wg.Done()
			result, err := xdg.DataHome()
			assert.NoError(t, err)
			assert.Equal(t, "/data", result)
		}()
	}

	// Test concurrent ConfigHome calls
	wg.Add(iterations)
	for range iterations {
		go func() {
			defer wg.Done()
			result, err := xdg.ConfigHome()
			assert.NoError(t, err)
			assert.Equal(t, "/config", result)
		}()
	}

	// Test concurrent CacheHome calls
	wg.Add(iterations)
	for range iterations {
		go func() {
			defer wg.Done()
			result, err := xdg.CacheHome()
			assert.NoError(t, err)
			assert.Equal(t, "/cache", result)
		}()
	}

	wg.Wait()
}

func TestEnvironmentVariableEdgeCases(t *testing.T) {
	xdg := New()

	t.Run("handles paths with special characters", func(t *testing.T) {
		specialPath := "/path/with spaces/and-dashes/under_scores"
		t.Setenv("XDG_DATA_HOME", specialPath)

		result, err := xdg.DataHome()
		require.NoError(t, err)
		assert.Equal(t, specialPath, result)
	})

	t.Run("handles relative paths in XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "relative/path")

		result, err := xdg.DataHome()
		require.NoError(t, err)
		assert.Equal(t, "relative/path", result)
	})

	t.Run("handles empty HOME with valid XDG vars", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("XDG_DATA_HOME", "/xdg/data")
		t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		t.Setenv("XDG_CACHE_HOME", "/xdg/cache")

		data, err := xdg.DataHome()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/data", data)

		config, err := xdg.ConfigHome()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/config", config)

		cache, err := xdg.CacheHome()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/cache", cache)
	})
}

func TestRealWorldScenarios(t *testing.T) {
	t.Run("typical Linux setup", func(t *testing.T) {
		t.Setenv("HOME", "/home/alice")
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("XDG_CACHE_HOME", "")

		xdg := New()

		data, err := xdg.DataHome()
		require.NoError(t, err)
		assert.Equal(t, "/home/alice/.local/share", data)

		config, err := xdg.ConfigHome()
		require.NoError(t, err)
		assert.Equal(t, "/home/alice/.config", config)

		cache, err := xdg.CacheHome()
		require.NoError(t, err)
		assert.Equal(t, "/home/alice/.cache", cache)
	})

	t.Run("custom XDG setup", func(t *testing.T) {
		t.Setenv("HOME", "/home/bob")
		t.Setenv("XDG_DATA_HOME", "/var/lib/myapp")
		t.Setenv("XDG_CONFIG_HOME", "/etc/myapp")
		t.Setenv("XDG_CACHE_HOME", "/tmp/myapp-cache")

		xdg := New()

		data, err := xdg.DataHome()
		require.NoError(t, err)
		assert.Equal(t, "/var/lib/myapp", data)

		config, err := xdg.ConfigHome()
		require.NoError(t, err)
		assert.Equal(t, "/etc/myapp", config)

		cache, err := xdg.CacheHome()
		require.NoError(t, err)
		assert.Equal(t, "/tmp/myapp-cache", cache)
	})
}

func TestMethodConsistency(t *testing.T) {
	// All three methods should behave consistently
	xdg := New()

	t.Run("all return error when HOME unset and XDG vars empty", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("XDG_CACHE_HOME", "")

		_, dataErr := xdg.DataHome()
		_, configErr := xdg.ConfigHome()
		_, cacheErr := xdg.CacheHome()

		assert.Error(t, dataErr)
		assert.Error(t, configErr)
		assert.Error(t, cacheErr)

		// All should return the same error
		assert.Equal(t, dataErr.Error(), configErr.Error())
		assert.Equal(t, configErr.Error(), cacheErr.Error())
	})

	t.Run("all prefer XDG vars over HOME", func(t *testing.T) {
		t.Setenv("HOME", "/home/user")
		t.Setenv("XDG_DATA_HOME", "/xdg/data")
		t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		t.Setenv("XDG_CACHE_HOME", "/xdg/cache")

		data, err := xdg.DataHome()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/data", data)

		config, err := xdg.ConfigHome()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/config", config)

		cache, err := xdg.CacheHome()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/cache", cache)
	})
}

// TestBackwardCompatibility ensures that the implementation matches
// the XDG Base Directory Specification defaults
func TestXDGSpecificationCompliance(t *testing.T) {
	homeDir := "/home/testuser"
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	xdg := New()

	// According to XDG Base Directory Specification:
	// - Default for XDG_DATA_HOME is $HOME/.local/share
	// - Default for XDG_CONFIG_HOME is $HOME/.config
	// - Default for XDG_CACHE_HOME is $HOME/.cache

	data, err := xdg.DataHome()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(homeDir, ".local", "share"), data,
		"Default XDG_DATA_HOME should be $HOME/.local/share per specification")

	config, err := xdg.ConfigHome()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(homeDir, ".config"), config,
		"Default XDG_CONFIG_HOME should be $HOME/.config per specification")

	cache, err := xdg.CacheHome()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(homeDir, ".cache"), cache,
		"Default XDG_CACHE_HOME should be $HOME/.cache per specification")
}
