// Package xdg provides utilities for XDG Base Directory Specification
package xdg

import (
	"os"
	"path/filepath"
)

// Interface defines methods for XDG Base Directory operations
type Interface interface {
	// DataHome returns the XDG_DATA_HOME directory path
	DataHome() (string, error)
	// ConfigHome returns the XDG_CONFIG_HOME directory path
	ConfigHome() (string, error)
	// CacheHome returns the XDG_CACHE_HOME directory path
	CacheHome() (string, error)
}

// XDG implements the Interface for XDG Base Directory operations
type XDG struct{}

// New creates a new XDG instance
func New() Interface {
	return &XDG{}
}

// DataHome returns the XDG_DATA_HOME directory path.
// If XDG_DATA_HOME is not set, it returns $HOME/.local/share
func (x *XDG) DataHome() (string, error) {
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return xdgDataHome, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".local", "share"), nil
}

// ConfigHome returns the XDG_CONFIG_HOME directory path.
// If XDG_CONFIG_HOME is not set, it returns $HOME/.config
func (x *XDG) ConfigHome() (string, error) {
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return xdgConfigHome, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config"), nil
}

// CacheHome returns the XDG_CACHE_HOME directory path.
// If XDG_CACHE_HOME is not set, it returns $HOME/.cache
func (x *XDG) CacheHome() (string, error) {
	if xdgCacheHome := os.Getenv("XDG_CACHE_HOME"); xdgCacheHome != "" {
		return xdgCacheHome, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".cache"), nil
}
