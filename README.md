# kube-tools

A collection of Kubernetes CLI tools for improved cluster management and development workflows.

## Tools

### [kubectl-x](./kubectl-x/)
Interactive context and namespace switching for kubectl with fuzzy search and command history.

### [kubernetes-mcp](./kubernetes-mcp/)
MCP (Model Context Protocol) server providing read-only Kubernetes cluster access for AI assistants.

### [celery](./celery/)
Kubernetes manifest validator using CEL (Common Expression Language) for custom validation rules.

### [klite](./klite/)
Lightweight Kustomize-like tool for Kubernetes resource management and customization.

## Installation

```bash
# Install all tools
go install github.com/RRethy/kube-tools/...@latest

# Install individually
go install github.com/RRethy/kube-tools/kubectl-x@latest
go install github.com/RRethy/kube-tools/kubernetes-mcp@latest
go install github.com/RRethy/kube-tools/celery@latest
go install github.com/RRethy/kube-tools/klite@latest
```

## Development

```bash
make all        # Build, test, and lint
make test       # Run tests
make build      # Build all tools
make help       # Show available commands
```

## License

MIT License - see [LICENSE](LICENSE) file

Copyright © 2024 Adam Regasz-Rethy (adam.regaszrethy@gmail.com)