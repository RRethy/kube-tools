package tools

import (
	"fmt"
	"strings"
)

// PaginationParams holds pagination configuration for output limiting
type PaginationParams struct {
	HeadLimit  int
	HeadOffset int
	TailLimit  int
	TailOffset int
}

// GetPaginationParams extracts pagination parameters from request arguments with sensible defaults
func GetPaginationParams(args map[string]any) PaginationParams {
	params := PaginationParams{
		HeadLimit: 50, // Default to 50 lines to prevent context overflow
	}

	if val, ok := args["head_limit"].(float64); ok {
		params.HeadLimit = int(val)
	}
	if val, ok := args["head_offset"].(float64); ok {
		params.HeadOffset = int(val)
	}

	if val, ok := args["tail_limit"].(float64); ok {
		params.TailLimit = int(val)
		params.HeadLimit = 0
	}
	if val, ok := args["tail_offset"].(float64); ok {
		params.TailOffset = int(val)
	}

	if params.HeadLimit < 0 {
		params.HeadLimit = 0
	}
	if params.HeadOffset < 0 {
		params.HeadOffset = 0
	}
	if params.TailLimit < 0 {
		params.TailLimit = 0
	}
	if params.TailOffset < 0 {
		params.TailOffset = 0
	}

	return params
}

// PaginationResult contains the paginated output and information about what was paginated
type PaginationResult struct {
	Output       string
	OriginalLines int
	ReturnedLines int
	PaginationInfo string
}

// ApplyPagination applies pagination parameters to limit output size
func ApplyPagination(output string, params PaginationParams) PaginationResult {
	if output == "" {
		return PaginationResult{Output: output}
	}

	lines := strings.Split(output, "\n")
	originalCount := len(lines)
	var paginationInfo string

	if params.TailLimit > 0 {
		start := len(lines) - params.TailLimit - params.TailOffset
		if start < 0 {
			start = 0
		}
		end := len(lines) - params.TailOffset
		if end > len(lines) {
			end = len(lines)
		}
		if end < 0 {
			end = 0
		}
		if start < end {
			lines = lines[start:end]
			if params.TailOffset > 0 {
				paginationInfo = fmt.Sprintf("\n[Showing last %d lines (skipped %d from end) of %d total lines]", len(lines), params.TailOffset, originalCount)
			} else {
				paginationInfo = fmt.Sprintf("\n[Showing last %d lines of %d total lines]", len(lines), originalCount)
			}
		} else {
			return PaginationResult{Output: "", PaginationInfo: fmt.Sprintf("[No lines to display from %d total lines]", originalCount)}
		}
	} else if params.HeadLimit > 0 {
		start := params.HeadOffset
		end := params.HeadOffset + params.HeadLimit
		if start > len(lines) {
			return PaginationResult{Output: "", PaginationInfo: fmt.Sprintf("[Offset %d exceeds %d total lines]", start, originalCount)}
		}
		if end > len(lines) {
			end = len(lines)
		}
		lines = lines[start:end]
		if params.HeadOffset > 0 {
			paginationInfo = fmt.Sprintf("\n[Showing lines %d-%d of %d total lines]", start+1, end, originalCount)
		} else if len(lines) < originalCount {
			paginationInfo = fmt.Sprintf("\n[Showing first %d lines of %d total lines]", len(lines), originalCount)
		}
	} else if params.HeadLimit == 0 && params.TailLimit == 0 {
		return PaginationResult{
			Output: output,
			OriginalLines: originalCount,
			ReturnedLines: originalCount,
		}
	}

	return PaginationResult{
		Output: strings.Join(lines, "\n"),
		OriginalLines: originalCount,
		ReturnedLines: len(lines),
		PaginationInfo: paginationInfo,
	}
}