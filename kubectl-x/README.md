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
- Execute commands across multiple contexts in parallel
- Interactive preview of logs and describe output
- Resource ownership graph visualization
- Kubeconfig copy for isolated environments
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

### Kubeconfig Management

```bash
# Write merged kubeconfig to $XDG_DATA_HOME for isolated usage
kubectl x kubeconfig copy

# Merges all contexts from multiple KUBECONFIG files into a single file
# Preserves the current context setting
# Each copy has a unique filename with timestamp
# Files are stored in $XDG_DATA_HOME/kubectl-x/ (default: ~/.local/share/kubectl-x/)
```

#### Shell-Local Kubeconfig

The `localkubeconfig.sh` script provides a `klocal` function that creates an isolated kubeconfig for your current shell session:

```bash
# Source the script to enable the klocal function
source kubectl-x/localkubeconfig.sh

# Create a shell-local kubeconfig (won't affect other terminals)
klocal

# Now all kubectl commands in this shell use the isolated kubeconfig
kubectl get pods  # Uses the local kubeconfig
kubectl x ctx     # Context changes only affect this shell
```

This is useful for:
- Working with multiple clusters simultaneously in different terminals
- Testing configuration changes without affecting your main kubeconfig
- Temporary context/namespace switching that auto-reverts when you close the shell
- Running scripts that need isolated Kubernetes configurations

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

### Execute Commands Across Multiple Contexts

The `each` command allows you to run the same kubectl operation on multiple contexts simultaneously:

```bash
# Execute command on contexts matching a pattern
kubectl x each "prod-.*" -- get pods

# Execute on multiple specific contexts
kubectl x each "(staging|prod)" -- get deployments

# Interactive context selection with fzf
kubectl x each -i -- get nodes

# Output in JSON format
kubectl x each "dev-.*" -o json -- get services

# Output in YAML format  
kubectl x each ".*-east" -o yaml -- get ingresses

# With namespace override
kubectl x each "prod-.*" -n kube-system -- get pods
```

**Each Command Features:**
- Uses regex patterns to match context names
- Executes commands in parallel across all matching contexts
- Preserves the current namespace for each context unless overridden
- Supports JSON, YAML, or raw output formats
- Interactive mode (`-i`) allows selecting contexts with fzf

### Interactive Resource Preview

The `peek` command provides interactive preview of logs or describe output using fzf:

```bash
# Interactively preview logs for pods
kubectl x peek logs pod

# Preview logs for a specific pod
kubectl x peek logs pod nginx

# Preview logs for pods from a deployment
kubectl x peek logs deployment my-app

# Interactively preview describe output
kubectl x peek describe pod

# Preview describe for services
kubectl x peek describe service
```

**Peek Command Features:**
- Interactive selection of resources using fzf
- Live preview of logs or describe output
- Supports all standard Kubernetes resource types
- Works with resource name filtering
- Resolves higher-level resources (deployments, services) to their pods

### Resource Ownership Graph

The `owners` command displays the ownership chain of a Kubernetes resource:

```bash
# Show ownership graph of a pod
kubectl x owners pod my-pod

# Show ownership graph of a replicaset
kubectl x owners replicaset nginx-abc123

# Use resource/name format
kubectl x owners pod/my-pod

# Show ownership in a specific namespace
kubectl x owners pod my-pod -n production
```

**Example Output:**
```
Deployment/nginx (production)
  └─> ReplicaSet/nginx-5d4f8b9 (production)
    └─> Pod/nginx-5d4f8b9-xyz (production)
```

**Owners Command Features:**
- Traverses ownerReferences to build complete ownership chain
- Shows hierarchy from resource to root owner
- Supports all Kubernetes resource types
- Handles circular references gracefully
- Color-coded output for better readability

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
