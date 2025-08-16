package testing

import (
	"context"
	"errors"

	"github.com/RRethy/kubectl-x/pkg/fzf"
)

var _ fzf.Interface = &FakeFzf{}

// FakeFzf is a fake implementation of fzf.Interface for testing
type FakeFzf struct {
	// Configuration for the fake
	ReturnItems []string
	ReturnError error
	
	// Capture what was called
	LastConfig   fzf.Config
	CallCount    int
	AllConfigs   []fzf.Config
	AllItems     [][]string
	
	// For simulating multiple calls with different results
	CallResults []CallResult
	callIndex   int
}

// CallResult represents the result of a single fzf call
type CallResult struct {
	Items []string
	Error error
}

// NewFakeFzf creates a new fake fzf with simple return values
func NewFakeFzf(items []string, err error) *FakeFzf {
	return &FakeFzf{
		ReturnItems: items,
		ReturnError: err,
	}
}

// NewFakeFzfWithMultipleCalls creates a fake fzf that returns different results for each call
func NewFakeFzfWithMultipleCalls(results []CallResult) *FakeFzf {
	return &FakeFzf{
		CallResults: results,
	}
}

// Run implements fzf.Interface
func (f *FakeFzf) Run(ctx context.Context, items []string, cfg fzf.Config) ([]string, error) {
	// Record the call
	f.CallCount++
	f.LastConfig = cfg
	f.AllConfigs = append(f.AllConfigs, cfg)
	f.AllItems = append(f.AllItems, items)
	
	// If using multiple call results
	if len(f.CallResults) > 0 {
		if f.callIndex >= len(f.CallResults) {
			return nil, errors.New("unexpected fzf call: no more results configured")
		}
		result := f.CallResults[f.callIndex]
		f.callIndex++
		return result.Items, result.Error
	}
	
	// Otherwise use simple return values
	return f.ReturnItems, f.ReturnError
}

// Reset clears the captured data
func (f *FakeFzf) Reset() {
	f.CallCount = 0
	f.LastConfig = fzf.Config{}
	f.AllConfigs = nil
	f.AllItems = nil
	f.callIndex = 0
}

// GetConfig returns the config from a specific call (0-indexed)
func (f *FakeFzf) GetConfig(callNum int) (fzf.Config, bool) {
	if callNum < 0 || callNum >= len(f.AllConfigs) {
		return fzf.Config{}, false
	}
	return f.AllConfigs[callNum], true
}

// GetItems returns the items from a specific call (0-indexed)
func (f *FakeFzf) GetItems(callNum int) ([]string, bool) {
	if callNum < 0 || callNum >= len(f.AllItems) {
		return nil, false
	}
	return f.AllItems[callNum], true
}

// WasCalled returns true if Run was called at least once
func (f *FakeFzf) WasCalled() bool {
	return f.CallCount > 0
}

// WasCalledTimes returns true if Run was called exactly n times
func (f *FakeFzf) WasCalledTimes(n int) bool {
	return f.CallCount == n
}