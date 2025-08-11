package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// mockKubectl is a test double that returns predictable output
type mockKubectlWithLines struct {
	lines int
}

func (m *mockKubectlWithLines) Execute(ctx context.Context, args ...string) (string, string, error) {
	// Generate output with specified number of lines
	var lines []string
	for i := 1; i <= m.lines; i++ {
		lines = append(lines, "Line "+string(rune('0'+(i%10))))
	}
	return strings.Join(lines, "\n"), "", nil
}

func TestHandleGetWithPagination(t *testing.T) {
	tests := []struct {
		name            string
		totalLines      int
		args            map[string]any
		expectedLines   int
		hasPaginationInfo bool
	}{
		{
			name:       "Default pagination limits to 50 lines",
			totalLines: 100,
			args: map[string]any{
				"resource-type": "pods",
			},
			expectedLines: 51, // 50 data lines + 1 pagination info
			hasPaginationInfo: true,
		},
		{
			name:       "Custom head_limit",
			totalLines: 100,
			args: map[string]any{
				"resource-type": "pods",
				"head_limit":    float64(20),
			},
			expectedLines: 21, // 20 data lines + 1 pagination info
			hasPaginationInfo: true,
		},
		{
			name:       "Head with offset",
			totalLines: 100,
			args: map[string]any{
				"resource-type": "pods",
				"head_limit":    float64(10),
				"head_offset":   float64(5),
			},
			expectedLines: 11, // 10 data lines + 1 pagination info
			hasPaginationInfo: true,
		},
		{
			name:       "Tail limit",
			totalLines: 100,
			args: map[string]any{
				"resource-type": "pods",
				"tail_limit":    float64(15),
			},
			expectedLines: 16, // 15 data lines + 1 pagination info
			hasPaginationInfo: true,
		},
		{
			name:       "Explicit zero for all results",
			totalLines: 100,
			args: map[string]any{
				"resource-type": "pods",
				"head_limit":    float64(0),
			},
			expectedLines: 100, // All lines, no pagination info
			hasPaginationInfo: false,
		},
		{
			name:       "Fewer lines than default",
			totalLines: 20,
			args: map[string]any{
				"resource-type": "pods",
			},
			expectedLines: 20, // All lines shown, no pagination info
			hasPaginationInfo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create tools with mock kubectl
			mockK := &mockKubectlWithLines{lines: tt.totalLines}
			tools := NewWithKubectl(mockK)

			// Create request
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "get",
					Arguments: tt.args,
				},
			}

			// Handle request
			result, err := tools.HandleGet(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleGet() error = %v", err)
			}

			// Check output lines
			if result.IsError {
				t.Fatalf("HandleGet() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleGet() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
			
			// Check if pagination info is present when expected
			if tt.hasPaginationInfo {
				lastLine := lines[len(lines)-1]
				if !strings.HasPrefix(lastLine, "[Showing") {
					t.Errorf("Expected pagination info in last line, got: %s", lastLine)
				}
			}
		})
	}
}

func TestHandleEventsWithPagination(t *testing.T) {
	// Create tools with mock kubectl that returns 100 lines
	mockK := &mockKubectlWithLines{lines: 100}
	tools := NewWithKubectl(mockK)

	tests := []struct {
		name          string
		args          map[string]any
		expectedLines int
	}{
		{
			name: "Default pagination",
			args: map[string]any{},
			expectedLines: 51, // 50 + pagination info
		},
		{
			name: "Custom limit",
			args: map[string]any{
				"head_limit": float64(30),
			},
			expectedLines: 31, // 30 + pagination info
		},
		{
			name: "All results",
			args: map[string]any{
				"head_limit": float64(0),
			},
			expectedLines: 100, // All lines, no pagination info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "events",
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleEvents(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleEvents() error = %v", err)
			}

			if result.IsError {
				t.Fatalf("HandleEvents() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleEvents() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
		})
	}
}

func TestHandleAPIResourcesWithPagination(t *testing.T) {
	// Create tools with mock kubectl that returns 100 lines
	mockK := &mockKubectlWithLines{lines: 100}
	tools := NewWithKubectl(mockK)

	tests := []struct {
		name          string
		args          map[string]any
		expectedLines int
	}{
		{
			name:          "Default pagination",
			args:          map[string]any{},
			expectedLines: 51, // 50 + pagination info
		},
		{
			name: "Custom limit",
			args: map[string]any{
				"head_limit": float64(25),
			},
			expectedLines: 26, // 25 + pagination info
		},
		{
			name: "All results",
			args: map[string]any{
				"head_limit": float64(0),
			},
			expectedLines: 100, // All lines, no pagination info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "api-resources",
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleAPIResources(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleAPIResources() error = %v", err)
			}

			if result.IsError {
				t.Fatalf("HandleAPIResources() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleAPIResources() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
		})
	}
}

func TestHandleConfigGetContextsWithPagination(t *testing.T) {
	// Create tools with mock kubectl that returns 100 lines
	mockK := &mockKubectlWithLines{lines: 100}
	tools := NewWithKubectl(mockK)

	tests := []struct {
		name          string
		args          map[string]any
		expectedLines int
	}{
		{
			name:          "Default pagination",
			args:          map[string]any{},
			expectedLines: 51, // 50 + pagination info
		},
		{
			name: "Custom limit",
			args: map[string]any{
				"head_limit": float64(15),
			},
			expectedLines: 16, // 15 + pagination info
		},
		{
			name: "Tail limit",
			args: map[string]any{
				"tail_limit": float64(20),
			},
			expectedLines: 21, // 20 + pagination info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "config-get-contexts",
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleConfigGetContexts(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleConfigGetContexts() error = %v", err)
			}

			if result.IsError {
				t.Fatalf("HandleConfigGetContexts() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleConfigGetContexts() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
		})
	}
}

func TestHandleTopPodWithPagination(t *testing.T) {
	// Create tools with mock kubectl that returns 100 lines
	mockK := &mockKubectlWithLines{lines: 100}
	tools := NewWithKubectl(mockK)

	tests := []struct {
		name          string
		args          map[string]any
		expectedLines int
	}{
		{
			name:          "Default pagination",
			args:          map[string]any{},
			expectedLines: 51, // 50 + pagination info
		},
		{
			name: "Custom limit",
			args: map[string]any{
				"head_limit": float64(35),
			},
			expectedLines: 36, // 35 + pagination info
		},
		{
			name: "Head with offset",
			args: map[string]any{
				"head_limit":  float64(20),
				"head_offset": float64(10),
			},
			expectedLines: 21, // 20 + pagination info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "top-pod",
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleTopPod(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleTopPod() error = %v", err)
			}

			if result.IsError {
				t.Fatalf("HandleTopPod() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleTopPod() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
		})
	}
}

func TestHandleTopNodeWithPagination(t *testing.T) {
	// Create tools with mock kubectl that returns 100 lines
	mockK := &mockKubectlWithLines{lines: 100}
	tools := NewWithKubectl(mockK)

	tests := []struct {
		name          string
		args          map[string]any
		expectedLines int
	}{
		{
			name:          "Default pagination",
			args:          map[string]any{},
			expectedLines: 51, // 50 + pagination info
		},
		{
			name: "Custom limit",
			args: map[string]any{
				"head_limit": float64(10),
			},
			expectedLines: 11, // 10 + pagination info
		},
		{
			name: "All results with zero",
			args: map[string]any{
				"head_limit": float64(0),
			},
			expectedLines: 100, // All lines, no pagination info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "top-node",
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleTopNode(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleTopNode() error = %v", err)
			}

			if result.IsError {
				t.Fatalf("HandleTopNode() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleTopNode() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
		})
	}
}

func TestHandleLogsWithPagination(t *testing.T) {
	// Create tools with mock kubectl that returns 100 lines
	mockK := &mockKubectlWithLines{lines: 100}
	tools := NewWithKubectl(mockK)

	tests := []struct {
		name          string
		args          map[string]any
		expectedLines int
	}{
		{
			name: "Default pagination with pod-name",
			args: map[string]any{
				"pod-name": "test-pod",
			},
			expectedLines: 51, // 50 + pagination info
		},
		{
			name: "Custom head_limit overrides kubectl tail",
			args: map[string]any{
				"pod-name":   "test-pod",
				"tail":       float64(80), // kubectl level
				"head_limit": float64(20), // client-side takes precedence
			},
			expectedLines: 21, // 20 + pagination info
		},
		{
			name: "Tail_limit for client-side tail",
			args: map[string]any{
				"pod-name":   "test-pod",
				"tail_limit": float64(25),
			},
			expectedLines: 26, // 25 + pagination info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "logs",
					Arguments: tt.args,
				},
			}

			result, err := tools.HandleLogs(context.Background(), req)
			if err != nil {
				t.Fatalf("HandleLogs() error = %v", err)
			}

			if result.IsError {
				t.Fatalf("HandleLogs() returned error result: %v", result.Content)
			}

			content := result.Content[0].(mcp.TextContent).Text
			lines := strings.Split(strings.TrimSpace(content), "\n")
			
			if len(lines) != tt.expectedLines {
				t.Errorf("HandleLogs() returned %d lines, want %d", len(lines), tt.expectedLines)
			}
		})
	}
}