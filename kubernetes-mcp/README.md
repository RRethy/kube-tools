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
- **Output filtering capabilities**:
  - `grep`: Filter output lines with regex or literal patterns
  - `jq`: Filter JSON output with jq expressions
  - `yq`: Filter YAML output with yq expressions
- **Intelligent pagination**:
  - Default 50-line limit to prevent context overflow
  - Configurable head/tail pagination with offsets
  - Clear feedback showing pagination status
- Intelligent debugging prompts:
  - debug-cluster: Comprehensive cluster analysis
  - debug-namespace: Focused namespace troubleshooting
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
go install github.com/RRethy/k8s-tools/kubernetes-mcp@latest
```

## Usage

### Start the MCP Server

```bash
# Start in stdio mode (for MCP clients)
kubernetes-mcp serve
```

### Available MCP Tools

#### get
Get Kubernetes resources from the cluster with filtering and pagination:
```json
{
  "resource_type": "pods",
  "resource_name": "my-pod",    // optional
  "namespace": "default",        // optional
  "all_namespaces": false,       // optional
  "output": "json",              // optional: json, yaml, wide
  "selector": "app=nginx",       // optional
  "context": "my-context",       // optional
  
  // Filtering options
  "grep": "Error|Running",       // optional: filter lines matching pattern
  "jq": ".items[].metadata.name", // optional: apply jq filter to JSON output
  "yq": ".metadata.labels",      // optional: apply yq filter to YAML output
  
  // Pagination options
  "head_limit": 50,              // optional: lines to show from start (default: 50, 0 for all)
  "head_offset": 0,              // optional: skip N lines from start
  "tail_limit": 20,              // optional: lines to show from end
  "tail_offset": 0               // optional: skip N lines from end
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
Fetch logs from a pod with filtering and pagination:
```json
{
  "pod_name": "my-pod",
  "namespace": "default",        // optional
  "container": "nginx",          // optional
  "tail": 100,                   // optional: kubectl-level tail
  "since": "5m",                 // optional
  "previous": false,             // optional
  "timestamps": false,           // optional
  "context": "my-context",       // optional
  
  // Filtering options
  "grep": "ERROR|WARN",          // optional: filter log lines matching pattern
  
  // Client-side pagination (applied after kubectl)
  "head_limit": 50,              // optional: lines to show from start
  "head_offset": 0,              // optional: skip N lines from start
  "tail_limit": 20,              // optional: lines to show from end
  "tail_offset": 0               // optional: skip N lines from end
}
```

#### events
View cluster events with filtering and pagination:
```json
{
  "namespace": "default",        // optional
  "all_namespaces": false,       // optional
  "for": "pod/my-pod",          // optional
  "output": "json",              // optional: json, yaml, wide
  "context": "my-context",       // optional
  
  // Filtering options
  "grep": "Warning|Error",       // optional: filter events matching pattern
  "jq": ".items[].message",      // optional: apply jq filter to JSON output
  "yq": ".items[].reason",       // optional: apply yq filter to YAML output
  
  // Pagination options
  "head_limit": 50,              // optional: lines to show (default: 50)
  "tail_limit": 20               // optional: show last N events
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

### Available MCP Prompts

#### debug-cluster
Comprehensive Kubernetes cluster debugging assistant that systematically analyzes the entire cluster for issues:
```json
{
  "context": "my-context"        // optional: defaults to current context
}
```
This prompt will:
- Check cluster health and node status
- Analyze system namespaces (kube-system)
- Review cluster-wide events
- Check resource pressure and quotas
- Verify networking and storage components
- Provide specific YAML fixes for any issues found

#### debug-namespace
Focused debugging assistant for a specific Kubernetes namespace:
```json
{
  "namespace": "my-namespace",   // optional: defaults to current namespace
  "context": "my-context"        // optional: defaults to current context
}
```
This prompt will:
- Analyze all pods in the namespace
- Check deployments and replica sets
- Review services and endpoints
- Examine recent events
- Check resource usage and limits
- Provide specific YAML fixes for any issues found

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
  - `tools/` - MCP tool implementations
    - Individual tool handlers (get.go, logs.go, events.go, etc.)
    - `filter.go` - Output filtering logic (grep, jq, yq)
    - `pagination.go` - Pagination logic with feedback
  - `prompts/` - Intelligent debugging prompts

The server wraps kubectl commands and exposes them through the MCP protocol, allowing LLMs to interact with Kubernetes clusters in a controlled, read-only manner.

## Output Filtering

The MCP server supports three types of output filtering:

### Grep Filtering
Works on any output format to filter lines matching a pattern:
```bash
# Filter for error or warning events
{"grep": "Error|Warning"}

# Filter for specific pod states
{"grep": "Running|Pending"}
```

### JQ Filtering
Applies jq expressions to JSON output:
```bash
# Extract pod names
{"output": "json", "jq": ".items[].metadata.name"}

# Get container images
{"output": "json", "jq": ".items[].spec.containers[].image"}
```

### YQ Filtering
Applies yq expressions to YAML output:
```bash
# Extract metadata
{"output": "yaml", "yq": ".metadata"}

# Get specific fields
{"output": "yaml", "yq": ".spec.replicas"}
```

## Pagination

All tools support intelligent pagination to prevent overwhelming LLM context:

- **Default limit**: 50 lines (configurable)
- **Head pagination**: Show first N lines with optional offset
- **Tail pagination**: Show last N lines with optional offset
- **Clear feedback**: Shows "[Showing first 50 lines of 200 total lines]"
- **Disable pagination**: Set `head_limit: 0` to get all results

### Examples
```json
// Get first 20 pods
{"resource_type": "pods", "head_limit": 20}

// Skip first 10, then get 20
{"resource_type": "pods", "head_offset": 10, "head_limit": 20}

// Get last 30 events
{"tail_limit": 30}

// Get all results (no pagination)
{"resource_type": "pods", "head_limit": 0}
```

## Practical Examples

### Finding Pods with Issues
```json
// Get all pods that are not running
{
  "resource_type": "pods",
  "all_namespaces": true,
  "grep": "Error|CrashLoop|Pending|ImagePull"
}

// Get pod names with their status
{
  "resource_type": "pods",
  "output": "json",
  "jq": ".items[] | {name: .metadata.name, status: .status.phase}"
}
```

### Analyzing Logs
```json
// Find errors in application logs
{
  "pod_name": "my-app-pod",
  "grep": "ERROR|Exception|Failed",
  "tail": 1000,
  "head_limit": 100  // Show first 100 matching lines
}

// Get recent warnings
{
  "pod_name": "my-app-pod",
  "since": "10m",
  "grep": "WARN|WARNING"
}
```

### Monitoring Events
```json
// Get all warning events in the cluster
{
  "all_namespaces": true,
  "grep": "Warning",
  "tail_limit": 30  // Show last 30 warnings
}

// Extract event messages for a specific deployment
{
  "for": "deployment/my-app",
  "output": "json",
  "jq": ".items[].message"
}
```

### Resource Analysis
```json
// Get container images for all pods
{
  "resource_type": "pods",
  "all_namespaces": true,
  "output": "json",
  "jq": ".items[].spec.containers[].image | unique"
}

// Find deployments with specific labels
{
  "resource_type": "deployments",
  "output": "yaml",
  "yq": ".items[] | select(.metadata.labels.environment == \"production\")"
}
```