# kubernetes-mcp

A Model Context Protocol (MCP) server that provides read-only access to Kubernetes clusters for LLMs.

## Features

- Read-only Kubernetes cluster access via MCP
- Secure by design - blocks access to secrets
- Supports multiple Kubernetes operations:
  - Get resources (pods, deployments, services, etc.)
  - Describe resources for detailed information
  - Fetch pod logs
  - View cluster events
  - Explain resource types and fields
  - Get cluster version and info
- Works with any MCP-compatible client
- Respects KUBECONFIG and kubectl context

## Installation

### From Source

```bash
# From the workspace root
make build-kubernetes-mcp

# Or from the kubernetes-mcp directory
go build -o kubernetes-mcp .
```

### Using go install

```bash
go install github.com/RRethy/utils/kubernetes-mcp@latest
```

## Usage

### Start the MCP Server

```bash
# Start in stdio mode (for MCP clients)
kubernetes-mcp serve
```

### Available MCP Tools

#### get
Get Kubernetes resources from the cluster:
```json
{
  "resource_type": "pods",
  "resource_name": "my-pod",    // optional
  "namespace": "default",        // optional
  "all_namespaces": false,       // optional
  "output": "json",              // optional: json, yaml, wide
  "selector": "app=nginx",       // optional
  "context": "my-context"        // optional
}
```

#### describe
Get detailed information about a resource:
```json
{
  "resource_type": "pod",
  "resource_name": "my-pod",
  "namespace": "default",        // optional
  "context": "my-context"        // optional
}
```

#### logs
Fetch logs from a pod:
```json
{
  "pod_name": "my-pod",
  "namespace": "default",        // optional
  "container": "nginx",          // optional
  "tail": 100,                   // optional
  "since": "5m",                 // optional
  "previous": false,             // optional
  "timestamps": false,           // optional
  "context": "my-context"        // optional
}
```

#### events
View cluster events:
```json
{
  "namespace": "default",        // optional
  "all_namespaces": false,       // optional
  "for": "pod/my-pod",          // optional
  "context": "my-context"        // optional
}
```

#### explain
Explain Kubernetes resource types:
```json
{
  "resource": "pods.spec.containers",
  "recursive": false,            // optional
  "context": "my-context"        // optional
}
```

#### version
Get kubectl and cluster version:
```json
{
  "output": "json",              // optional: json, yaml
  "context": "my-context"        // optional
}
```

#### cluster-info
Get cluster information:
```json
{
  "context": "my-context"        // optional
}
```

## Security

- **Read-only**: All operations are read-only, no modifications to the cluster
- **Secrets blocked**: Access to Kubernetes secrets is explicitly blocked
- **Context aware**: Respects kubectl context and KUBECONFIG settings

## Requirements

- kubectl configured with cluster access
- Go 1.24+ (for building from source)

## Integration with LLM Clients

kubernetes-mcp implements the Model Context Protocol (MCP) and can be used with any MCP-compatible client. Configure your MCP client to connect to the kubernetes-mcp server via stdio.

### Example with Claude Desktop

Add to your Claude Desktop configuration:
```json
{
  "mcpServers": {
    "kubernetes": {
      "command": "kubernetes-mcp",
      "args": ["serve"]
    }
  }
}
```

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Build binary
go build -o kubernetes-mcp .
```

## Architecture

The server is structured as:
- `cmd/` - CLI commands (root, serve)
- `pkg/mcp/` - MCP server implementation
  - `server.go` - Main server logic and kubectl execution
  - `tools.go` - MCP tool definitions and handlers

The server wraps kubectl commands and exposes them through the MCP protocol, allowing LLMs to interact with Kubernetes clusters in a controlled, read-only manner.