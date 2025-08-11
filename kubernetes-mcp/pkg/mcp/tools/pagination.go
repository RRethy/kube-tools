package tools

import "strings"

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

	// Extract head parameters
	if val, ok := args["head_limit"].(float64); ok {
		params.HeadLimit = int(val)
	}
	if val, ok := args["head_offset"].(float64); ok {
		params.HeadOffset = int(val)
	}

	// Extract tail parameters
	if val, ok := args["tail_limit"].(float64); ok {
		params.TailLimit = int(val)
		params.HeadLimit = 0 // Disable head if tail is specified
	}
	if val, ok := args["tail_offset"].(float64); ok {
		params.TailOffset = int(val)
	}

	// Validate parameters
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

// ApplyPagination applies pagination parameters to limit output size
func ApplyPagination(output string, params PaginationParams) string {
	if output == "" {
		return output
	}

	lines := strings.Split(output, "\n")

	// Apply tail pagination if specified (takes precedence)
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
		} else {
			return ""
		}
	} else if params.HeadLimit > 0 {
		// Apply head pagination
		start := params.HeadOffset
		end := params.HeadOffset + params.HeadLimit
		if start > len(lines) {
			return ""
		}
		if end > len(lines) {
			end = len(lines)
		}
		lines = lines[start:end]
	} else if params.HeadLimit == 0 && params.TailLimit == 0 {
		// User explicitly wants all results (head_limit=0)
		return output
	}

	return strings.Join(lines, "\n")
}