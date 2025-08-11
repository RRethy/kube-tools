# CLAUDE.md - kubernetes-mcp Module

This document provides guidance to Claude Code when working with the kubernetes-mcp module of this repository.

## Module Overview

kubernetes-mcp is a Model Context Protocol (MCP) server that provides read-only access to Kubernetes clusters for LLMs. It wraps kubectl commands and exposes them through the MCP protocol, allowing safe interaction with Kubernetes resources.

## Key Features

- Read-only Kubernetes operations via MCP
- Security-focused design (blocks secrets access)
- Stdio-based MCP server implementation
- Full kubectl context support
- Comprehensive resource access tools
- Output filtering with grep, jq, and yq
- Intelligent pagination with clear feedback
- Debugging prompts for cluster and namespace analysis

## Architecture

```
kubernetes-mcp/
├── cmd/
│   ├── root.go       # Root command setup
│   └── serve.go      # MCP server command
├── pkg/
│   ├── mcp/
│   │   ├── server/
│   │   │   ├── server.go    # MCP server implementation
│   │   │   └── kubectl.go   # kubectl wrapper with security
│   │   ├── tools/
│   │   │   ├── tools.go     # Tool registry and common logic
│   │   │   ├── get.go       # Get resources tool
│   │   │   ├── describe.go  # Describe tool
│   │   │   ├── logs.go      # Logs retrieval tool
│   │   │   ├── events.go    # Events tool
│   │   │   ├── filter.go    # Output filtering (grep, jq, yq)
│   │   │   ├── pagination.go # Pagination with feedback
│   │   │   └── *_test.go    # Tool tests
│   │   └── prompts/
│   │       ├── debug_cluster.go    # Cluster debugging prompt
│   │       └── debug_namespace.go  # Namespace debugging prompt
├── go.mod
└── main.go          # Entry point
```

## Development Commands

```bash
# Build the binary
make build-kubernetes-mcp

# Run from source
go run . serve

# Run tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

## Key Components

### MCP Server (`pkg/mcp/server.go`)
- Implements the MCP server using mark3labs/mcp-go
- Wraps kubectl commands for safe execution
- Blocks access to sensitive resources (secrets)
- Handles stdio communication

### Tools (`pkg/mcp/tools/`)
Implements MCP tools for Kubernetes operations:
- `get` - Get resources with filtering options (grep, jq, yq) and pagination
- `describe` - Detailed resource information
- `logs` - Pod log retrieval with grep filtering and pagination
- `events` - Cluster events with filtering and pagination
- `explain` - Resource type documentation
- `version` - kubectl and cluster version
- `cluster-info` - Cluster information
- `api-resources` - List available API resources with filtering
- `top-pod` / `top-node` - Resource usage metrics with pagination
- `config-get-contexts` - List available contexts
- `use-context` - Switch kubectl context
- `current-context` / `current-namespace` - Show current context/namespace

### Filtering System (`pkg/mcp/tools/filter.go`)
- **grep**: Regex or literal pattern matching on any output
- **jq**: JSON filtering with jq expressions (falls back to simple JSONPath)
- **yq**: YAML filtering with yq expressions (falls back to simple YAML path)
- Filters are applied in order: grep → jq/yq
- External tools (jq/yq) used when available, with built-in fallbacks

### Pagination System (`pkg/mcp/tools/pagination.go`)
- Default 50-line limit to prevent LLM context overflow
- Head pagination: `head_limit`, `head_offset`
- Tail pagination: `tail_limit`, `tail_offset`
- Clear feedback: "[Showing first 50 lines of 200 total lines]"
- Set `head_limit: 0` to disable pagination

## Security Considerations

- **Read-only operations**: No write/modify/delete operations exposed
- **Secrets blocked**: Explicitly prevents access to Kubernetes secrets
- **Context isolation**: Uses kubectl's existing context and auth

## Testing

When adding new features:
1. Test kubectl command execution
2. Verify MCP protocol compliance
3. Ensure security restrictions work
4. Test error handling and edge cases

## Common Tasks

### Adding a New Tool
1. Create a new file in `pkg/mcp/tools/` (e.g., `newtool.go`)
2. Define the tool schema with `CreateNewTool()` method
3. Implement the handler `HandleNewTool()` method
4. Register the tool in `pkg/mcp/server/server.go`
5. Add filtering support if the tool has JSON/YAML output:
   - Extract output format from args
   - Call `GetFilterParams()` and `ApplyFilter()`
6. Add pagination support:
   - Call `GetPaginationParams()` and `ApplyPagination()`
   - Append `result.PaginationInfo` to output
7. Write tests in `newtool_test.go`
8. Update README.md with tool documentation

### Modifying Security Rules
1. Update `isBlockedResource()` in `server.go`
2. Add tests for blocked resources
3. Document security changes in README.md

## Dependencies

```go
github.com/mark3labs/mcp-go v0.32.0  // MCP protocol implementation
github.com/spf13/cobra v1.9.1        // CLI framework
github.com/stretchr/testify v1.10.0  // Testing utilities
gopkg.in/yaml.v3                     // YAML parsing for yq filtering
```

## Code Style Guidelines

### Comments
- **DO NOT add comments unless explicitly asked by the user**
- Code should be self-documenting through clear naming
- Tool descriptions should be comprehensive in the schema

### Error Handling
- Wrap kubectl errors with context
- Return user-friendly error messages
- Handle kubectl not found gracefully
- Filter errors are non-fatal (preserve original output)

### MCP Protocol
- Follow MCP specification for tool responses
- Use appropriate content types (text, error)
- Include helpful error messages for debugging

### Adding Features
- Filtering should be applied before pagination
- Always show pagination feedback when output is truncated
- Preserve original output if filters fail
- Use existing patterns from other tools

## Recent Enhancements

### Output Filtering (December 2024)
- Added grep, jq, and yq filtering capabilities
- Filters work independently or in combination
- Fallback implementations when external tools unavailable
- Non-fatal errors preserve original output

### Pagination System (December 2024)
- Added intelligent pagination with 50-line default
- Clear feedback about lines shown vs total
- Support for head/tail with offsets
- Prevents LLM context overflow