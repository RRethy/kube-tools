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
		name           string
		input          string
		params         PaginationParams
		expectedOutput string
		expectedInfo   string
	}{
		{
			name:  "Empty input",
			input: "",
			params: PaginationParams{
				HeadLimit: 10,
			},
			expectedOutput: "",
			expectedInfo:   "",
		},
		{
			name:  "Head limit only",
			input: testData,
			params: PaginationParams{
				HeadLimit: 5,
			},
			expectedOutput: "A\nB\nC\nD\nE",
			expectedInfo:   "\n[Showing first 5 lines of 20 total lines]",
		},
		{
			name:  "Head limit with offset",
			input: testData,
			params: PaginationParams{
				HeadLimit:  5,
				HeadOffset: 3,
			},
			expectedOutput: "D\nE\nF\nG\nH",
			expectedInfo:   "\n[Showing lines 4-8 of 20 total lines]",
		},
		{
			name:  "Head offset beyond input",
			input: testData,
			params: PaginationParams{
				HeadLimit:  5,
				HeadOffset: 25,
			},
			expectedOutput: "",
			expectedInfo:   "[Offset 25 exceeds 20 total lines]",
		},
		{
			name:  "Head limit exceeds remaining lines",
			input: testData,
			params: PaginationParams{
				HeadLimit:  10,
				HeadOffset: 15,
			},
			expectedOutput: "P\nQ\nR\nS\nT",
			expectedInfo:   "\n[Showing lines 16-20 of 20 total lines]",
		},
		{
			name:  "Tail limit only",
			input: testData,
			params: PaginationParams{
				TailLimit: 5,
			},
			expectedOutput: "P\nQ\nR\nS\nT",
			expectedInfo:   "\n[Showing last 5 lines of 20 total lines]",
		},
		{
			name:  "Tail limit with offset",
			input: testData,
			params: PaginationParams{
				TailLimit:  5,
				TailOffset: 3,
			},
			expectedOutput: "M\nN\nO\nP\nQ",
			expectedInfo:   "\n[Showing last 5 lines (skipped 3 from end) of 20 total lines]",
		},
		{
			name:  "Tail offset beyond input",
			input: testData,
			params: PaginationParams{
				TailLimit:  5,
				TailOffset: 25,
			},
			expectedOutput: "",
			expectedInfo:   "[No lines to display from 20 total lines]",
		},
		{
			name:  "Zero head_limit returns all",
			input: testData,
			params: PaginationParams{
				HeadLimit: 0,
			},
			expectedOutput: testData,
			expectedInfo:   "",
		},
		{
			name:  "Default pagination (50 lines but only 20 available)",
			input: testData,
			params: PaginationParams{
				HeadLimit: 50,
			},
			expectedOutput: testData,
			expectedInfo:   "", // No info when all lines are shown
		},
		{
			name:  "Single line",
			input: "single line",
			params: PaginationParams{
				HeadLimit: 1,
			},
			expectedOutput: "single line",
			expectedInfo:   "", // No info when all lines are shown
		},
		{
			name:  "Tail with large offset",
			input: testData,
			params: PaginationParams{
				TailLimit:  30,
				TailOffset: 5,
			},
			expectedOutput: "A\nB\nC\nD\nE\nF\nG\nH\nI\nJ\nK\nL\nM\nN\nO",
			expectedInfo:   "\n[Showing last 15 lines (skipped 5 from end) of 20 total lines]",
		},
		{
			name:  "Head and tail both zero",
			input: testData,
			params: PaginationParams{
				HeadLimit: 0,
				TailLimit: 0,
			},
			expectedOutput: testData,
			expectedInfo:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPagination(tt.input, tt.params)
			if result.Output != tt.expectedOutput {
				t.Errorf("ApplyPagination().Output = %q, want %q", result.Output, tt.expectedOutput)
			}
			if result.PaginationInfo != tt.expectedInfo {
				t.Errorf("ApplyPagination().PaginationInfo = %q, want %q", result.PaginationInfo, tt.expectedInfo)
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
		hasInfo       bool
	}{
		{
			name: "Default 50 lines",
			params: PaginationParams{
				HeadLimit: 50,
			},
			expectedLines: 50,
			hasInfo:       true,
		},
		{
			name: "Last 100 lines",
			params: PaginationParams{
				TailLimit: 100,
			},
			expectedLines: 100,
			hasInfo:       true,
		},
		{
			name: "Middle section",
			params: PaginationParams{
				HeadLimit:  50,
				HeadOffset: 475,
			},
			expectedLines: 50,
			hasInfo:       true,
		},
		{
			name: "Skip and take from end",
			params: PaginationParams{
				TailLimit:  200,
				TailOffset: 100,
			},
			expectedLines: 200,
			hasInfo:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPagination(largeData, tt.params)
			resultLines := strings.Split(result.Output, "\n")
			if len(resultLines) != tt.expectedLines {
				t.Errorf("ApplyPagination() returned %d lines, want %d", len(resultLines), tt.expectedLines)
			}
			if tt.hasInfo && result.PaginationInfo == "" {
				t.Errorf("ApplyPagination() expected pagination info but got none")
			}
			if result.OriginalLines != 1000 {
				t.Errorf("ApplyPagination().OriginalLines = %d, want 1000", result.OriginalLines)
			}
			if result.ReturnedLines != tt.expectedLines {
				t.Errorf("ApplyPagination().ReturnedLines = %d, want %d", result.ReturnedLines, tt.expectedLines)
			}
		})
	}
}