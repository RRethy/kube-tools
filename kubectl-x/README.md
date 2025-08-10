# kubectl-x

Fast Kubernetes context and namespace switching with fuzzy search.

## Features

- Interactive context switching with fuzzy search
- Interactive namespace switching with fuzzy search  
- Quick switching with partial name matching
- Previous context/namespace support (`-` flag)
- Command history tracking
- Display current context and namespace
- Native kubectl plugin integration

## Installation

### From Source

```bash
# From the workspace root
make build-kubectl-x

# Or from the kubectl-x directory
go build -o kubectl-x .

# Install as kubectl plugin
cp kubectl-x /usr/local/bin/kubectl-x
```

### Using go install

```bash
go install github.com/RRethy/kubectl-x@latest
```

## Usage

kubectl-x integrates as a kubectl plugin, so you can use it with `kubectl x`:

### Context Switching

```bash
# Interactive context selection with fuzzy search
kubectl x ctx

# Switch to context by partial name match
kubectl x ctx prod

# Switch to previous context
kubectl x ctx -

# Switch context and set namespace
kubectl x ctx staging kube-system
```

### Namespace Switching

```bash
# Interactive namespace selection with fuzzy search
kubectl x ns

# Switch to namespace by partial name match  
kubectl x ns frontend

# Switch to previous namespace
kubectl x ns -
```

### Current Status

```bash
# Show current context and namespace
kubectl x cur
```

## Features

### Fuzzy Search
When selecting contexts or namespaces interactively, kubectl-x uses fzf for powerful fuzzy searching. Type any part of the name to filter the list.

### Partial Matching
When providing a context or namespace name as an argument, kubectl-x will match partial names. For example:
- `kubectl x ctx prod` matches `production-cluster`
- `kubectl x ns front` matches `frontend-apps`

### History
kubectl-x maintains a history of context and namespace switches in `~/.local/share/kubectl-x/history.yaml`. Use the `-` flag to quickly switch to the previous context or namespace.

### Exact Matching
Use the `--exact` flag to disable fuzzy/partial matching and require exact names:
```bash
kubectl x ctx --exact production-cluster
kubectl x ns --exact default
```

## Requirements

- kubectl configured with contexts
- fzf (for interactive selection)
- Go 1.24+ (for building from source)

## Configuration

kubectl-x uses your existing kubectl configuration from `~/.kube/config` or the path specified by the `KUBECONFIG` environment variable.

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Build binary
go build -o kubectl-x .
```

## Architecture

kubectl-x follows a clean architecture pattern with:
- CLI commands in `cmd/`
- Business logic in `pkg/cli/`
- Supporting packages in `pkg/` (fzf, history, kubeconfig, kubernetes)
- Mock implementations in `testing/` subdirectories for unit testing

The tool integrates with kubectl's plugin system and uses the standard Kubernetes client-go library for all cluster operations.