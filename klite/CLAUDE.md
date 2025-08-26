# CLAUDE.md - klite Module

This file provides module-specific guidance to Claude Code (claude.ai/code) when working with the klite module.

## Module Overview

klite is a lightweight Kustomize-like tool that provides simplified Kubernetes resource management functionality. It follows the same architectural patterns as kubectl-x for consistency across the workspace.

## Architecture

### Command Structure
The module follows the standard CLI pattern used across the workspace:

```
klite/
├── cmd/
│   ├── root.go        # Root command definition
│   └── build.go       # Build subcommand
├── pkg/
│   └── cli/
│       └── build/
│           ├── build.go    # Main function that sets up dependencies
│           └── builder.go  # Interface and implementation
└── main.go            # Entry point
```

### Command Pattern
Each command follows this pattern:
1. `cmd/[command].go`: Cobra command definition
2. `pkg/cli/[command]/[command].go`: Dependency setup
3. `pkg/cli/[command]/[command]er.go`: Business logic interface and implementation

## Current Features

### Build Command
- **Purpose**: Build Kubernetes resources from a kustomization directory
- **Usage**: `klite build [directory]`
- **Implementation**: Currently scaffolded with basic structure

## Development Guidelines

### Adding New Commands
When adding new commands to klite:
1. Create `cmd/[command].go` with Cobra command definition
2. Create `pkg/cli/[command]/` directory
3. Implement the pattern with `[command].go` and `[command]er.go`
4. Follow interface-based design for testability

### Testing
- Write unit tests for the business logic in `pkg/cli/[command]/[command]er_test.go`
- Use interfaces to mock dependencies
- Follow the testing patterns from kubectl-x

### Future Enhancements
Potential features to implement:
- Resource merging and patching
- ConfigMap and Secret generation
- Variable substitution
- Resource ordering
- Namespace injection
- Label and annotation management

## Dependencies
- `github.com/spf13/cobra`: CLI framework
- Additional dependencies will be added as features are implemented

## Build and Test
```bash
# Build the module
make build-klite

# Run tests
cd klite && go test ./...

# Run linting
cd klite && go run github.com/golangci/golangci-lint/cmd/golangci-lint run
```