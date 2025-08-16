// Package history manages command history for kubectl-x operations
package history

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const (
	maxHistoryItems = 2
)

var (
	_ Interface = &History{}

	defaultHistoryPath = filepath.Join(os.ExpandEnv("$HOME"), ".local", "share", "kubectl-x", "history.yaml")
)

// Interface defines methods for managing command history
type Interface interface {
	// Get retrieves a historical item from the specified group at the given distance
	Get(group string, distance int) (string, error)
	// Add appends an item to the specified history group
	Add(group, item string)
	// Write persists the history to storage
	Write() error
}

// ConfigOption allows customizing History configuration
type ConfigOption func(*Config)

// WithHistoryPath sets a custom path for the history file
func WithHistoryPath(path string) ConfigOption {
	return func(config *Config) {
		config.historyPath = path
	}
}

// Config holds configuration for History
type Config struct {
	historyPath string
}

// NewConfig creates a new configuration with the given options
func NewConfig(options ...ConfigOption) *Config {
	config := &Config{historyPath: defaultHistoryPath}
	for _, option := range options {
		option(config)
	}
	return config
}

// History manages persistent command history
type History struct {
	Data map[string][]string `json:"data"`

	path string
}

// NewHistory creates a history manager that loads from the configured path
func NewHistory(config *Config) (*History, error) {
	contents, err := os.ReadFile(config.historyPath)
	history := History{path: config.historyPath}
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading file: %s", err)
	} else if err == nil {
		err = yaml.Unmarshal(contents, &history)
		if err != nil {
			return nil, fmt.Errorf("unmarshalling history: %s", err)
		}
	}
	return &history, nil
}

func (h *History) Get(group string, distance int) (string, error) {
	if h.Data == nil {
		return "", fmt.Errorf("no history found")
	}

	groupHistory, ok := h.Data[group]
	if !ok {
		return "", fmt.Errorf("group '%s' not found in history", group)
	}

	if distance >= len(groupHistory) {
		return "", fmt.Errorf("unable to go back %d items in history", distance)
	}

	return groupHistory[distance], nil
}

func (h *History) Add(group, item string) {
	if h.Data == nil {
		h.Data = make(map[string][]string)
	}

	h.Data[group] = append([]string{item}, h.Data[group]...)
	if len(h.Data[group]) > maxHistoryItems {
		h.Data[group] = h.Data[group][:maxHistoryItems]
	}
}

func (h *History) Write() error {
	contents, err := yaml.Marshal(h)
	if err != nil {
		return fmt.Errorf("marshalling history: %s", err)
	}

	err = os.MkdirAll(filepath.Dir(h.path), 0o755)
	if err != nil {
		return fmt.Errorf("creating directory: %s", err)
	}

	err = os.WriteFile(h.path, contents, 0o644)
	if err != nil {
		return fmt.Errorf("writing file: %s", err)
	}

	return nil
}
