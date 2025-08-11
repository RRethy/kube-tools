package tools

import (
	"strings"
	"testing"
)

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		expected PaginationParams
	}{
		{
			name: "Default parameters",
			args: map[string]any{},
			expected: PaginationParams{
				HeadLimit:  50,
				HeadOffset: 0,
				TailLimit:  0,
				TailOffset: 0,
			},
		},
		{
			name: "Custom head limit",
			args: map[string]any{
				"head_limit": float64(10),
			},
			expected: PaginationParams{
				HeadLimit:  10,
				HeadOffset: 0,
				TailLimit:  0,
				TailOffset: 0,
			},
		},
		{
			name: "Head with offset",
			args: map[string]any{
				"head_limit":  float64(20),
				"head_offset": float64(5),
			},
			expected: PaginationParams{
				HeadLimit:  20,
				HeadOffset: 5,
				TailLimit:  0,
				TailOffset: 0,
			},
		},
		{
			name: "Tail parameters override head",
			args: map[string]any{
				"head_limit": float64(20),
				"tail_limit": float64(15),
			},
			expected: PaginationParams{
				HeadLimit:  0, // Head disabled when tail is set
				HeadOffset: 0,
				TailLimit:  15,
				TailOffset: 0,
			},
		},
		{
			name: "Tail with offset",
			args: map[string]any{
				"tail_limit":  float64(10),
				"tail_offset": float64(3),
			},
			expected: PaginationParams{
				HeadLimit:  0,
				HeadOffset: 0,
				TailLimit:  10,
				TailOffset: 3,
			},
		},
		{
			name: "Explicit zero head_limit for all results",
			args: map[string]any{
				"head_limit": float64(0),
			},
			expected: PaginationParams{
				HeadLimit:  0,
				HeadOffset: 0,
				TailLimit:  0,
				TailOffset: 0,
			},
		},
		{
			name: "Negative values are corrected to zero",
			args: map[string]any{
				"head_limit":  float64(-10),
				"head_offset": float64(-5),
				"tail_limit":  float64(-15),
				"tail_offset": float64(-3),
			},
			expected: PaginationParams{
				HeadLimit:  0,
				HeadOffset: 0,
				TailLimit:  0,
				TailOffset: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPaginationParams(tt.args)
			if result != tt.expected {
				t.Errorf("GetPaginationParams() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestApplyPagination(t *testing.T) {
	// Create test data with 20 lines
	lines := make([]string, 20)
	for i := 0; i < 20; i++ {
		lines[i] = string(rune('A' + i))
	}
	testData := strings.Join(lines, "\n")

	tests := []struct {
		name     string
		input    string
		params   PaginationParams
		expected string
	}{
		{
			name:  "Empty input",
			input: "",
			params: PaginationParams{
				HeadLimit: 10,
			},
			expected: "",
		},
		{
			name:  "Head limit only",
			input: testData,
			params: PaginationParams{
				HeadLimit: 5,
			},
			expected: "A\nB\nC\nD\nE",
		},
		{
			name:  "Head limit with offset",
			input: testData,
			params: PaginationParams{
				HeadLimit:  5,
				HeadOffset: 3,
			},
			expected: "D\nE\nF\nG\nH",
		},
		{
			name:  "Head offset beyond input",
			input: testData,
			params: PaginationParams{
				HeadLimit:  5,
				HeadOffset: 25,
			},
			expected: "",
		},
		{
			name:  "Head limit exceeds remaining lines",
			input: testData,
			params: PaginationParams{
				HeadLimit:  10,
				HeadOffset: 15,
			},
			expected: "P\nQ\nR\nS\nT",
		},
		{
			name:  "Tail limit only",
			input: testData,
			params: PaginationParams{
				TailLimit: 5,
			},
			expected: "P\nQ\nR\nS\nT",
		},
		{
			name:  "Tail limit with offset",
			input: testData,
			params: PaginationParams{
				TailLimit:  5,
				TailOffset: 3,
			},
			expected: "M\nN\nO\nP\nQ",
		},
		{
			name:  "Tail offset beyond input",
			input: testData,
			params: PaginationParams{
				TailLimit:  5,
				TailOffset: 25,
			},
			expected: "",
		},
		{
			name:  "Zero head_limit returns all",
			input: testData,
			params: PaginationParams{
				HeadLimit: 0,
			},
			expected: testData,
		},
		{
			name:  "Default pagination (50 lines but only 20 available)",
			input: testData,
			params: PaginationParams{
				HeadLimit: 50,
			},
			expected: testData,
		},
		{
			name:  "Single line",
			input: "single line",
			params: PaginationParams{
				HeadLimit: 1,
			},
			expected: "single line",
		},
		{
			name:  "Tail with large offset",
			input: testData,
			params: PaginationParams{
				TailLimit:  30,
				TailOffset: 5,
			},
			expected: "A\nB\nC\nD\nE\nF\nG\nH\nI\nJ\nK\nL\nM\nN\nO",
		},
		{
			name:  "Head and tail both zero",
			input: testData,
			params: PaginationParams{
				HeadLimit: 0,
				TailLimit: 0,
			},
			expected: testData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPagination(tt.input, tt.params)
			if result != tt.expected {
				t.Errorf("ApplyPagination() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestApplyPaginationLargeData(t *testing.T) {
	// Test with larger dataset
	lines := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		lines[i] = "Line " + string(rune('0'+i%10))
	}
	largeData := strings.Join(lines, "\n")

	tests := []struct {
		name          string
		params        PaginationParams
		expectedLines int
	}{
		{
			name: "Default 50 lines",
			params: PaginationParams{
				HeadLimit: 50,
			},
			expectedLines: 50,
		},
		{
			name: "Last 100 lines",
			params: PaginationParams{
				TailLimit: 100,
			},
			expectedLines: 100,
		},
		{
			name: "Middle section",
			params: PaginationParams{
				HeadLimit:  50,
				HeadOffset: 475,
			},
			expectedLines: 50,
		},
		{
			name: "Skip and take from end",
			params: PaginationParams{
				TailLimit:  200,
				TailOffset: 100,
			},
			expectedLines: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPagination(largeData, tt.params)
			resultLines := strings.Split(result, "\n")
			if len(resultLines) != tt.expectedLines {
				t.Errorf("ApplyPagination() returned %d lines, want %d", len(resultLines), tt.expectedLines)
			}
		})
	}
}