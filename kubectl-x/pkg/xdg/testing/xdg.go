// Package testing provides mock implementations for testing XDG operations
package testing

import (
	"errors"

	"github.com/RRethy/kubectl-x/pkg/xdg"
)

var _ xdg.Interface = &FakeXDG{}

// FakeXDG is a mock implementation of xdg.Interface for testing
type FakeXDG struct {
	DataHomePath    string
	DataHomeError   error
	ConfigHomePath  string
	ConfigHomeError error
	CacheHomePath   string
	CacheHomeError  error
}

// NewFakeXDG creates a new fake XDG with the given paths
func NewFakeXDG(dataHome string) *FakeXDG {
	return &FakeXDG{
		DataHomePath:   dataHome,
		ConfigHomePath: dataHome + "-config",
		CacheHomePath:  dataHome + "-cache",
	}
}

// DataHome returns the configured data home path or error
func (f *FakeXDG) DataHome() (string, error) {
	if f.DataHomeError != nil {
		return "", f.DataHomeError
	}
	if f.DataHomePath == "" {
		return "", errors.New("data home not configured")
	}
	return f.DataHomePath, nil
}

// ConfigHome returns the configured config home path or error
func (f *FakeXDG) ConfigHome() (string, error) {
	if f.ConfigHomeError != nil {
		return "", f.ConfigHomeError
	}
	if f.ConfigHomePath == "" {
		return "", errors.New("config home not configured")
	}
	return f.ConfigHomePath, nil
}

// CacheHome returns the configured cache home path or error
func (f *FakeXDG) CacheHome() (string, error) {
	if f.CacheHomeError != nil {
		return "", f.CacheHomeError
	}
	if f.CacheHomePath == "" {
		return "", errors.New("cache home not configured")
	}
	return f.CacheHomePath, nil
}

// WithDataHomeError sets an error to be returned by DataHome
func (f *FakeXDG) WithDataHomeError(err error) *FakeXDG {
	f.DataHomeError = err
	return f
}
