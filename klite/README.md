# klite

A lightweight Kustomize-like tool for Kubernetes resource management.

## Overview

klite provides simplified Kustomize-like functionality for managing Kubernetes resources. It's designed to be a lightweight alternative that covers common use cases for resource customization and composition.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/RRethy/kube-tools.git
cd kube-tools

# Build and install
make build-klite
sudo make install

# Or install directly with go
go install github.com/RRethy/kube-tools/klite@latest
```

### Download Binary

Download the latest release from the [releases page](https://github.com/RRethy/kube-tools/releases).

## Usage

### Build Command

Build Kubernetes resources from a kustomization directory:

```bash
# Build resources from current directory
klite build

# Build resources from specific directory
klite build path/to/kustomization

# Pipe to kubectl to apply
klite build . | kubectl apply -f -
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
klite build ./overlays/production

# Apply directly to cluster
klite build ./base | kubectl apply -f -

# Preview changes
klite build ./overlays/staging | kubectl diff -f -
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
cd klite && go test ./...

# Run with verbose output
go run main.go build . -v

# Build binary
go build -o klite .
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## Related Projects

- [kubectl-x](../kubectl-x/) - Interactive Kubernetes context and namespace switcher
- [kubernetes-mcp](../kubernetes-mcp/) - MCP server for Kubernetes
- [celery](../celery/) - CEL-based Kubernetes resource validator