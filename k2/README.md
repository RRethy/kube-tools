# k2

**WIP**

A lightweight Kustomize-like tool for Kubernetes resource management.

## Overview

k2 provides simplified Kustomize-like functionality for managing Kubernetes resources. It's designed to be a lightweight alternative that covers common use cases for resource customization and composition.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/RRethy/kube-tools.git
cd kube-tools

# Build and install
make build-k2
sudo make install

# Or install directly with go
go install github.com/RRethy/kube-tools/k2@latest
```

### Download Binary

Download the latest release from the [releases page](https://github.com/RRethy/kube-tools/releases).

## Usage

### Build Command

Build Kubernetes resources from a kustomization directory:

```bash
# Build resources from current directory
k2 build

# Build resources from specific directory
k2 build path/to/kustomization

# Pipe to kubectl to apply
k2 build . | kubectl apply -f -
```

## Features

### Current
- Basic build command structure
- Directory-based resource building

### Planned
- Resource merging and strategic patching
- ConfigMap and Secret generation from files/literals
- Variable substitution and templating
- Resource ordering and dependencies
- Namespace injection
- Common labels and annotations
- Image tag management
- Multi-base overlays

## Examples

### Basic Usage

```bash
# Build resources from a kustomization directory
k2 build ./overlays/production

# Apply directly to cluster
k2 build ./base | kubectl apply -f -

# Preview changes
k2 build ./overlays/staging | kubectl diff -f -
```

## Project Structure

```
kustomization/
├── base/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── kustomization.yaml
└── overlays/
    ├── production/
    │   ├── replica-patch.yaml
    │   └── kustomization.yaml
    └── staging/
        ├── resource-patch.yaml
        └── kustomization.yaml
```

## Development

```bash
# Run tests
cd k2 && go test ./...

# Run with verbose output
go run main.go build . -v

# Build binary
go build -o k2 .
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## Related Projects

- [kubectl-x](../kubectl-x/) - Interactive Kubernetes context and namespace switcher
- [kubernetes-mcp](../kubernetes-mcp/) - MCP server for Kubernetes
- [celery](../celery/) - CEL-based Kubernetes resource validator
