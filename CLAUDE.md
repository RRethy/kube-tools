# CLAUDE.md - Workspace Root

This file provides workspace-level guidance to Claude Code (claude.ai/code) when working with this repository.

**IMPORTANT**: When making workspace-level changes (new modules, workspace structure, build processes), update this file. For module-specific changes, update the module's CLAUDE.md file.

## Development Commands

Use the provided Makefile for common development tasks:
```bash
make all                   - Build, test, and lint-fix (full development cycle)
make build                 - Build all binaries
make build-kubectl-x       - Build the kubectl-x binary
make build-kubernetes-mcp  - Build the kubernetes-mcp binary
make build-celery          - Build the celery binary
make build-klite   - Build the klite binary
make test                  - Run all tests
make lint                  - Run golangci-lint
make lint-fix              - Run golangci-lint with auto-fix
make fmt                   - Format Go code
make vet                   - Run go vet
make tidy                  - Run go mod tidy
make help                  - Show all available targets
```

### Workspace-Specific Commands
```bash
# Sync workspace dependencies
go work sync

# Install from source
go install github.com/RRethy/kubectl-x@latest                    # kubectl-x CLI
go install github.com/RRethy/kube-tools/kubernetes-mcp@latest     # kubernetes-mcp CLI
go install github.com/RRethy/kube-tools/celery@latest             # celery CLI
go install github.com/RRethy/kube-tools/klite@latest      # klite CLI
```

## Architecture

### Go Workspace Structure
This is a Go workspace with four modules:
- `kubectl-x/` - Kubernetes context and namespace switching CLI
- `kubernetes-mcp/` - Readonly MCP (Model Context Protocol) server for Kubernetes
- `celery/` - CEL-based Kubernetes resource validator
- `klite/` - Lightweight Kustomize-like tool for Kubernetes resource management

The root contains `go.work` for workspace configuration.

## Workspace Development Notes

### Go Workspace Configuration
- Uses Go 1.24.5
- Multi-module workspace with `kubectl-x/`, `kubernetes-mcp/`, `celery/`, and `klite/` modules
- Use `go work sync` to synchronize dependencies across workspace
- All modules include golangci-lint as tool dependency

### Module Structure
```
.
├── go.work              # Workspace configuration
├── Makefile            # Multi-module build targets
├── kubectl-x/          # Context/namespace switching CLI
│   ├── CLAUDE.md       # Module-specific documentation
│   ├── README.md       # User documentation
│   ├── cmd/            # CLI commands
│   ├── pkg/            # Package implementations
│   └── main.go
├── kubernetes-mcp/     # MCP server for Kubernetes
│   ├── CLAUDE.md       # Module-specific documentation
│   ├── README.md       # User documentation
│   ├── cmd/            # CLI commands (root, serve)
│   ├── pkg/mcp/        # MCP server implementation
│   └── main.go
├── celery/             # CEL validator for Kubernetes resources
│   ├── CLAUDE.md       # Module-specific documentation
│   ├── README.md       # User documentation
│   ├── cmd/            # CLI commands
│   ├── pkg/            # Core packages (validator, yaml, cli)
│   ├── api/            # API types (ValidationRules)
│   ├── fixtures/       # Test fixtures and examples
│   └── main.go
└── klite/      # Lightweight Kustomize-like tool
    ├── CLAUDE.md       # Module-specific documentation
    ├── README.md       # User documentation
    ├── cmd/            # CLI commands (root, build)
    ├── pkg/cli/build/  # Build command implementation
    └── main.go
```

### Adding New Modules
When adding additional modules to the workspace:
1. Create new module directory with `go mod init`
2. Update `go.work` to include the new module path
3. Add build targets to root `Makefile`
4. Run `go work sync` to update workspace dependencies

### Cross-Module Dependencies
Modules are currently independent. If adding cross-module dependencies:
- Use workspace-relative imports
- Ensure proper module versioning
- Test from workspace root with `go test ./...`

## Code Style Guidelines

### CLI Development Patterns

#### Command Pattern
Each CLI command follows a consistent architectural pattern:
- `cmd/[command].go`: Cobra command definition that calls the corresponding function in `pkg/cli/[command]/`
- `pkg/cli/[command]/[command].go`: Main function that sets up dependencies and calls the business logic
- `pkg/cli/[command]/[command]er.go`: Interface and implementation containing the actual business logic

This separation ensures:
- Clean command definitions with minimal logic
- Testable business logic separated from CLI framework
- Consistent dependency injection patterns
- Interface-based design for better testing and modularity
