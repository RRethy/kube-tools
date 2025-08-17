# kubectl-x

Fast Kubernetes context and namespace switching with fuzzy search.

## Features

- Interactive context switching with fuzzy search
- Interactive namespace switching with fuzzy search  
- Quick switching with partial name matching
- Previous context/namespace support (`-` flag)
- Command history tracking
- Display current context and namespace
- Shell into pods with resource resolution
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

### Debugging and Logging

kubectl-x follows kubectl's standard logging practices with verbose output levels:

```bash
# Basic verbose output (shows important operations)
kubectl x ctx -v=1

# Debug level output (shows detailed operations) 
kubectl x ctx -v=4

# Trace level output (shows all operations including fzf interactions)
kubectl x ns -v=5

# API request level (shows Kubernetes API calls)
kubectl x shell my-pod -v=6

# Full HTTP debugging (shows complete request/response details)
kubectl x ctx -v=8
```

**Verbosity Levels:**
- `--v=0`: Errors and warnings only (default)
- `--v=1`: Basic information about operations
- `--v=2`: Useful steady state information
- `--v=4`: Debug level verbosity
- `--v=5`: Trace level verbosity (user interactions)
- `--v=6`: Resource operations and API calls
- `--v=8`: HTTP request/response content
- `--v=9`: Full HTTP content without truncation

### Shell into Pods

```bash
# Shell into a pod directly
kubectl x shell my-pod

# Shell into a pod from a deployment
kubectl x shell deployment/nginx
kubectl x shell deploy nginx

# Shell into a pod from other resources
kubectl x shell statefulset/database
kubectl x shell service/web-service

# Shell into a specific container
kubectl x shell my-pod -c container-name

# Use a different shell command
kubectl x shell my-pod --command=/bin/bash
```

### Debug Containers

Use `--debug` to run a debug container instead of exec'ing into the existing pod:

```bash
# Debug mode - run a debug container as sidecar
kubectl x shell my-pod --debug

# Debug mode with custom image
kubectl x shell my-pod --debug --image=ubuntu:latest

# Debug mode targeting a specific container (shares process namespace)
kubectl x shell my-pod --debug -c=app-container

# Debug with custom command
kubectl x shell my-pod --debug --command=/bin/bash --image=busybox
```

**Debug Mode Features:**
- Uses `kubectl debug` under the hood for safe debugging
- Runs as ephemeral sidecar container by default
- Can target specific containers to share process namespace
- Supports custom debug images with debugging tools
- Does not affect the running application

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

### KUBECONFIG Support

kubectl-x fully supports the `KUBECONFIG` environment variable, following the same rules as kubectl:

```bash
# Use a specific kubeconfig file
export KUBECONFIG=/path/to/my-kubeconfig
kubectl x ctx

# Use multiple kubeconfig files (merged)
export KUBECONFIG=/path/to/config1:/path/to/config2:/path/to/config3
kubectl x ctx

# Temporarily use a different kubeconfig
KUBECONFIG=/path/to/staging-config kubectl x ctx staging-cluster
```

The `KUBECONFIG` variable supports:
- Single file paths
- Multiple file paths separated by colons (Linux/macOS) or semicolons (Windows)
- Relative and absolute paths
- Files are merged in the order they appear

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