# k8s-tools

A collection of Kubernetes CLI tools for enhanced cluster management, resource validation, and context switching.

## Tools

This workspace contains four independent tools:

### kubectl-x
Kubectl plugin that provides convenient context and namespace switching utilities for Kubernetes. Features interactive fuzzy search capabilities and maintains command history for quick navigation.

### kubernetes-mcp
A readonly MCP (Model Context Protocol) server that exposes Kubernetes cluster information through a standardized interface. Enables AI assistants and other tools to interact with Kubernetes clusters safely.

**Key Features:**
- Output filtering with `grep`, `jq`, and `yq` expressions
- Intelligent pagination with configurable limits and clear feedback
- Comprehensive debugging prompts for cluster and namespace analysis
- Full kubectl context and namespace management support

### celery
CEL-based Kubernetes resource validator that allows you to define and enforce custom validation rules using the Common Expression Language (CEL). Validate YAML manifests against complex policies before applying them to your cluster.

### kustomizelite
A lightweight, simplified implementation of Kustomize that handles basic resource composition and patching. Focuses on essential functionality with minimal dependencies for environments where full Kustomize might be too heavy.

## Installation

Install any or all tools directly from source:

```bash
# kubectl-x - Context and namespace switching
go install github.com/RRethy/k8s-tools/kubectl-x@latest

# kubernetes-mcp - MCP server for Kubernetes
go install github.com/RRethy/k8s-tools/kubernetes-mcp@latest

# celery - CEL-based resource validator
go install github.com/RRethy/k8s-tools/celery@latest

# kustomizelite - Lightweight Kustomize implementation
go install github.com/RRethy/k8s-tools/kustomizelite@latest
```

## Development

This is a Go workspace with multiple modules. Use the provided Makefile for common development tasks:

```bash
# Complete development cycle (build, test, lint-fix)
make all

# Build all tools
make build

# Build individual tools
make build-kubectl-x
make build-kubernetes-mcp
make build-celery
make build-kustomizelite

# Run tests
make test

# Lint and format
make lint
make lint-fix
make fmt

# See all available commands
make help
```

## kubectl-x Documentation

### `kubectl x ctx`

Switch context with interactive fuzzy search.

```bash
kubectl x ctx                        # Interactive context selection
kubectl x ctx my-context             # Switch to context with partial match
kubectl x ctx my-context my-namespace # Switch context and namespace
kubectl x ctx -                      # Switch to previous context/namespace
```

### `kubectl x ns`

Switch namespace with interactive fuzzy search.

```bash
kubectl x ns                # Interactive namespace selection
kubectl x ns my-namespace   # Switch to namespace with partial match
kubectl x ns -              # Switch to previous namespace
```

### `kubectl x cur`

Print current context and namespace.

```bash
kubectl x cur
```

### `kubectl x shell`

Shell into a pod with resource resolution.

```bash
kubectl x shell my-pod                    # Shell into a specific pod
kubectl x shell deployment/nginx          # Shell into a pod from deployment
kubectl x shell deploy nginx              # Shell into a pod from deployment (space separator)
kubectl x shell my-pod -c container-name  # Shell into specific container
kubectl x shell my-pod --command=/bin/bash # Use different shell
```

## kubernetes-mcp Documentation

Start the MCP server to expose Kubernetes cluster information:

```bash
# Start the server
kubernetes-mcp serve
```

The MCP server is readonly and safe to use in production environments.

### Available Tools

- **Resource Management:**
  - `get` - List resources with filtering (grep, jq, yq) and pagination
  - `describe` - Get detailed resource information
  - `api-resources` - List available API resources
  - `explain` - Explain resource types and fields

- **Monitoring & Debugging:**
  - `logs` - View pod logs with grep filtering and pagination
  - `events` - List cluster events with filtering and pagination
  - `top-pod` / `top-node` - Resource usage metrics
  - `debug-cluster` - Comprehensive cluster analysis prompt
  - `debug-namespace` - Focused namespace troubleshooting prompt

- **Cluster Information:**
  - `cluster-info` - Get cluster and services information
  - `version` - Get kubectl and cluster version info
  - `config-get-contexts` - List available contexts
  - `current-context` / `current-namespace` - Show current settings
  - `use-context` - Switch kubectl context

### Output Filtering Examples

```json
// Find pods with issues using grep
{"resource_type": "pods", "grep": "Error|CrashLoop"}

// Extract specific fields with jq
{"resource_type": "pods", "output": "json", "jq": ".items[].metadata.name"}

// Filter YAML output with yq
{"resource_type": "deployments", "output": "yaml", "yq": ".spec.replicas"}
```

### Pagination Examples

```json
// Default: First 50 lines
{"resource_type": "pods"}

// Custom limits
{"resource_type": "events", "head_limit": 20}  // First 20 lines
{"resource_type": "events", "tail_limit": 30}  // Last 30 lines

// Disable pagination
{"resource_type": "pods", "head_limit": 0}  // All results
```

Pagination feedback shows: `[Showing first 50 lines of 200 total lines]`

## celery Documentation

Validate Kubernetes resources using CEL expressions:

```bash
# Validate resources against rules
celery validate -f rules.yaml resources.yaml

# Validate with specific target resources
celery validate -f rules.yaml --target pod/my-pod resources.yaml

# Create validation rules
cat > rules.yaml <<EOF
apiVersion: celery.io/v1
kind: ValidationRules
metadata:
  name: security-rules
spec:
  rules:
    - name: require-non-root
      target:
        apiVersion: "apps/v1"
        kind: "Deployment"
      expression: |
        object.spec.template.spec.?securityContext.?runAsNonRoot.orValue(false) == true
      message: "Deployments must run as non-root user"
EOF

# Run validation
celery validate -f rules.yaml deployment.yaml
```

### CEL Expression Context

When writing CEL expressions, you have access to:
- `object` - The Kubernetes resource being validated
- `allObjects` - List of all resources in the current validation batch (for cross-resource validation)

## Requirements

- Go 1.24.4 or later (for building from source)
- kubectl configured with cluster access
- For kubectl-x: kubectl plugin support

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright © 2024 Adam Regasz-Rethy (adam.regaszrethy@gmail.com)