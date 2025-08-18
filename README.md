# k8s-tools

A collection of Kubernetes CLI tools for improved cluster management and development workflows.

## Tools

### kubectl-x
Context and namespace switching for kubectl with interactive selection and command history.

- Interactive fuzzy search for contexts and namespaces
- Switch context and namespace in a single command
- Command history with quick access to previous selections
- Shell access to pods and deployments
- Isolated kubeconfig generation for testing

### kubernetes-mcp
MCP (Model Context Protocol) server that provides read-only access to Kubernetes cluster information for AI assistants and automation tools.

- Read-only operations for safe production use
- Output filtering with grep, jq, and yq
- Automatic pagination for large outputs
- Built-in debugging prompts for common scenarios

### celery
Kubernetes manifest validator using CEL (Common Expression Language) for custom validation rules.

- CEL-based validation rules
- Cross-resource validation support
- CI/CD pipeline integration
- Detailed validation error messages

### kustomizelite (Experimental)
Lightweight implementation of Kustomize for basic resource composition and patching. This tool is experimental and not recommended for production use.

- Core Kustomize functionality only
- Minimal dependencies
- Compatible with existing kustomization files

## Installation

```bash
# Install all tools
go install github.com/RRethy/k8s-tools/...@latest

# Install individually
go install github.com/RRethy/k8s-tools/kubectl-x@latest
go install github.com/RRethy/k8s-tools/kubernetes-mcp@latest
go install github.com/RRethy/k8s-tools/celery@latest
go install github.com/RRethy/k8s-tools/kustomizelite@latest
```

## Usage

### kubectl-x

```bash
# Interactive context switching
kubectl x ctx

# Switch with partial match
kubectl x ctx prod

# Switch context and namespace
kubectl x ctx staging my-namespace

# Previous context
kubectl x ctx -

# Interactive namespace switching
kubectl x ns

# Shell into pods
kubectl x shell my-pod
kubectl x shell deploy/nginx

# Copy kubeconfig for isolated use
kubectl x kubeconfig copy
```

### kubernetes-mcp

```bash
# Start the MCP server
kubernetes-mcp serve

# The server exposes tools for:
# - Listing and describing resources
# - Viewing logs and events
# - Cluster health analysis
# - Debugging with pre-built prompts
```

### celery

```yaml
# rules.yaml
apiVersion: celery.io/v1
kind: ValidationRules
metadata:
  name: production-rules
spec:
  rules:
    - name: require-resource-limits
      target:
        kind: Deployment
      expression: |
        object.spec.template.spec.containers.all(c, 
          has(c.resources.limits))
      message: "Containers must have resource limits"
```

```bash
# Validate manifests
celery validate -f rules.yaml deployment.yaml

# CI/CD integration
celery validate -f rules/ manifests/
```

### kustomizelite

```yaml
# kustomization.yaml
resources:
  - deployment.yaml
  - service.yaml

patches:
  - target:
      kind: Deployment
      name: my-app
    patch: |-
      - op: replace
        path: /spec/replicas
        value: 3
```

```bash
# Build configuration
kustomizelite build . | kubectl apply -f -
```

## Development

```bash
make all        # Build, test, and lint
make test       # Run tests
make build      # Build all tools
make help       # Show available commands
```

## Requirements

- Go 1.24.4+ (for building from source)
- kubectl configured with cluster access
- fzf (optional, for kubectl-x interactive mode)

## License

MIT License - see [LICENSE](LICENSE) file

Copyright © 2024 Adam Regasz-Rethy (adam.regaszrethy@gmail.com)