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

## Architecture

```
kubernetes-mcp/
├── cmd/
│   ├── root.go       # Root command setup
│   └── serve.go      # MCP server command
├── pkg/
│   └── mcp/
│       ├── server.go # Server implementation and kubectl wrapper
│       └── tools.go  # MCP tool definitions and handlers
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

### Tools (`pkg/mcp/tools.go`)
Implements MCP tools for Kubernetes operations:
- `get` - Get resources with filtering options
- `describe` - Detailed resource information
- `logs` - Pod log retrieval
- `events` - Cluster events
- `explain` - Resource type documentation
- `version` - kubectl and cluster version
- `cluster-info` - Cluster information

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
1. Define the tool schema in `tools.go`
2. Implement the handler in `tools.go`
3. Register the tool in `server.go`
4. Test the kubectl command execution
5. Update README.md with tool documentation

### Modifying Security Rules
1. Update `isBlockedResource()` in `server.go`
2. Add tests for blocked resources
3. Document security changes in README.md

## Dependencies

```go
github.com/mark3labs/mcp-go v0.6.0  // MCP protocol implementation
github.com/spf13/cobra v1.8.1       // CLI framework
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

### MCP Protocol
- Follow MCP specification for tool responses
- Use appropriate content types (text, error)
- Include helpful error messages for debugging