# k8s-tools

A collection of Kubernetes CLI tools for enhanced cluster management, resource validation, and context switching.

## Tools

This workspace contains three independent tools:

### kubectl-x
Kubectl plugin that provides convenient context and namespace switching utilities for Kubernetes. Features interactive fuzzy search capabilities and maintains command history for quick navigation.

### kubernetes-mcp
A readonly MCP (Model Context Protocol) server that exposes Kubernetes cluster information through a standardized interface. Enables AI assistants and other tools to interact with Kubernetes clusters safely.

### celery
CEL-based Kubernetes resource validator that allows you to define and enforce custom validation rules using the Common Expression Language (CEL). Validate YAML manifests against complex policies before applying them to your cluster.

## Installation

Install any or all tools directly from source:

```bash
# kubectl-x - Context and namespace switching
go install github.com/RRethy/k8s-tools/kubectl-x@latest

# kubernetes-mcp - MCP server for Kubernetes
go install github.com/RRethy/k8s-tools/kubernetes-mcp@latest

# celery - CEL-based resource validator
go install github.com/RRethy/k8s-tools/celery@latest
```

## Development

This is a Go workspace with multiple modules. Use the provided Makefile for common development tasks:

```bash
# Build all tools
make build

# Build individual tools
make build-kubectl-x
make build-kubernetes-mcp
make build-celery

# Run tests
make test

# Lint and format
make lint
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

## kubernetes-mcp Documentation

Start the MCP server to expose Kubernetes cluster information:

```bash
# Start the server
kubernetes-mcp serve

# The server provides tools for:
# - Getting cluster information
# - Listing and describing resources
# - Viewing pod logs
# - Watching events
# - Explaining resource types
```

The MCP server is readonly and safe to use in production environments. It exposes the following capabilities:
- `cluster-info` - Get cluster and services information
- `get` - List Kubernetes resources
- `describe` - Get detailed resource information
- `logs` - View pod container logs
- `events` - List and filter cluster events
- `explain` - Explain resource types and fields
- `version` - Get kubectl and cluster version info

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
- `oldObject` - Previous version (for updates)
- `request` - Admission request details
- `params` - Rule parameters
- `namespaceObject` - Namespace object (for namespaced resources)
- `authorizer` - For authorization checks

## Requirements

- Go 1.24.4 or later (for building from source)
- kubectl configured with cluster access
- For kubectl-x: kubectl plugin support

## License

See individual module directories for license information.