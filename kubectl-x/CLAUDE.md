# CLAUDE.md - kubectl-x Module

This file provides module-specific guidance for Claude Code when working within the kubectl-x module directory.

**IMPORTANT**: When making changes to module-specific aspects (dependencies, commands, internal packages), update this file. For workspace-level changes, update the root `CLAUDE.md`. Also keep `kubectl-x/README.md` up to date with user-facing changes.

## Usage Examples
```bash
kubectl x ctx                    # Interactive context selection
kubectl x ctx my-context        # Switch to context with partial match
kubectl x ns                     # Interactive namespace selection
kubectl x ns my-namespace       # Switch to namespace with partial match
kubectl x cur                    # Show current context and namespace
kubectl x ctx -                  # Switch to previous context/namespace
kubectl x shell my-pod          # Shell into a pod
kubectl x shell deploy/nginx    # Shell into a pod from deployment

# Verbose logging examples
kubectl x ctx -v=0              # Only errors and warnings (default)
kubectl x ctx -v=1              # Basic information
kubectl x ctx -v=2              # Useful steady state information  
kubectl x ctx -v=4              # Debug level verbosity
kubectl x ctx -v=6              # Show requested resources
kubectl x ctx -v=8              # Show HTTP request contents
```

## Module Commands

### Build and Test (from kubectl-x/ directory)
```bash
# From workspace root - complete development cycle
make all                     # Build, test, and lint-fix
make build-kubectl-x         # Build only kubectl-x

# From kubectl-x/ directory - individual operations
go build .                   # Build the application
go test ./...                # Run all module tests

# Run specific package tests
go test ./pkg/cli/ctx/
go test ./pkg/cli/ns/
go test ./pkg/cli/cur/
go test ./pkg/fzf/
go test ./pkg/history/
go test ./pkg/kubeconfig/
go test ./pkg/kubernetes/

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Benchmark tests
go test -bench=. ./...

# Update dependencies
go mod tidy
go mod download
```

## Module Architecture

### Entry Point
- `main.go` - Simple entry point that calls `cmd.Execute()`

### Command Structure (`cmd/`)
- `root.go` - Root command setup with kubectl CLI options integration
- `ctx.go` - Context switching command definition
- `ns.go` - Namespace switching command definition  
- `cur.go` - Current status display command definition
- `shell.go` - Shell command definition for pod execution

### Package Details

#### `pkg/cli/ctx/`
- `ctx.go` - Context command implementation
- `ctxer.go` - Context switching business logic
- `ctxer_test.go` - Comprehensive test suite
- **Key interfaces**: `Ctxer` for context operations

#### `pkg/cli/ns/`
- `ns.go` - Namespace command implementation
- `nser.go` - Namespace switching business logic
- `nser_test.go` - Test suite with table-driven tests
- **Key interfaces**: `Nser` for namespace operations

#### `pkg/cli/cur/`
- `cur.go` - Current status command implementation
- `curer.go` - Current status display logic
- `curer_test.go` - Status display tests
- **Key interfaces**: `Curer` for status operations

#### `pkg/cli/shell/`
- `shell.go` - Shell command implementation
- `sheller.go` - Shell execution business logic
- `sheller_test.go` - Comprehensive test suite for shell operations
- **Key interfaces**: `Sheller` for pod shell execution

#### `pkg/fzf/`
- `fzf.go` - Fuzzy finder integration with external fzf binary
- `fzf_test.go` - Tests including user cancellation scenarios
- `testing/fzf.go` - Mock implementation for testing
- **Key interfaces**: `Fzf` for interactive selection

#### `pkg/history/`
- `history.go` - Command history persistence and retrieval
- `history_test.go` - History management tests
- `testing/history.go` - Mock history implementation
- **Storage**: `~/.local/share/kubectl-x/history.yaml`
- **Key interfaces**: `History` for history operations

#### `pkg/kubeconfig/`
- `kubeconfig.go` - Kubeconfig file manipulation
- `kubeconfig_test.go` - Kubeconfig operation tests
- `testing/kubeconfig.go` - Mock kubeconfig implementation
- **Key interfaces**: `Kubeconfig` for kubeconfig operations

#### `pkg/kubernetes/`
- `client.go` - Kubernetes API client wrapper
- `kubernetes.go` - Generic resource operations
- `testing/client.go` - Mock Kubernetes client
- **Key interfaces**: `Client` for Kubernetes operations

#### `pkg/shortname/`
- `shortname.go` - Kubernetes resource shortname expansion (deploy->deployment, etc.)
- Used by shell command for resource type resolution

## Development Patterns

### Interface Implementation Pattern
Each package follows this pattern:
1. Define primary interface (e.g., `Ctxer`, `Nser`)
2. Implement concrete struct
3. Create constructor function with dependency injection
4. Provide mock implementation in `testing/` subdirectory

### Error Handling Pattern
- Wrap errors with context using `fmt.Errorf`
- Return user-friendly error messages
- Handle kubectl client errors consistently
- Use k8s.io/utils/exec for command execution to enable better error handling and testing

### Testing Pattern
- Table-driven tests for multiple scenarios
- Mock all external dependencies using k8s.io/utils/exec/testing for exec mocks
- Test both success and error conditions
- Use testify/assert for assertions
- Prefer k8s.io utilities over standard library when available

### Command Integration Pattern
- Commands are thin wrappers around package implementations
- Business logic in `pkg/cli/{command}/` packages
- Dependency injection through constructor functions
- Consistent flag handling using Cobra

## Module Dependencies

### Direct Dependencies
```go
github.com/fatih/color v1.17.0          // Terminal colors
github.com/goccy/go-yaml v1.11.3        // YAML processing
github.com/spf13/cobra v1.8.1           // CLI framework
github.com/stretchr/testify v1.9.0      // Testing utilities
k8s.io/api v0.30.2                      // Kubernetes API types
k8s.io/apimachinery v0.30.2             // Kubernetes API machinery
k8s.io/cli-runtime v0.30.2              // kubectl CLI runtime
k8s.io/client-go v0.30.2                // Kubernetes Go client
k8s.io/kubectl v0.30.2                  // kubectl utilities
k8s.io/utils v0.0.0-20240502163921-fe8a2dddb1d0 // Kubernetes utilities
```

### External Runtime Dependencies
- `fzf` binary - Required for interactive selection
- `kubectl` - Uses kubectl's configuration and patterns

### Kubernetes Utilities Usage
- Use `k8s.io/utils/exec` instead of `os/exec` for better testability
- Use `k8s.io/utils/exec/testing` for mock exec implementations in tests
- Leverage existing k8s.io utilities when they provide better abstraction or testability


## Logging System

kubectl-x uses klog (k8s.io/klog/v2) for logging, following kubectl's standard practices:

### Verbosity Levels (kubectl standard)
- `--v=0`: Only errors and warnings (default)
- `--v=1`: Basic information about operations
- `--v=2`: Useful steady state information and important log messages
- `--v=3`: Extended information about changes
- `--v=4`: Debug level verbosity
- `--v=5`: Trace level verbosity
- `--v=6`: Show requested resources
- `--v=7`: Show HTTP request headers
- `--v=8`: Show HTTP request contents
- `--v=9`: Show HTTP request contents without truncation

### Usage in Code
```go
import "k8s.io/klog/v2"

klog.V(1).Infof("Setting Kubernetes context: %s", contextName)
klog.V(4).Infof("Running fzf for selection: query=%s", query)
klog.Errorf("Failed to set context %s: %v", contextName, err)
klog.Warningf("Failed to write history: %v", err)
klog.V(6).Infof("Listing resources: type=%s namespace=%s", resourceType, namespace)
```

### Log Configuration
- Logs written to both stderr and `/tmp/kubectl-x.log`
- Use `-v` flag to control verbosity following kubectl standards
- Start debugging with `-v=6` to see API requests
- Use `-v=8` or `-v=9` for deep troubleshooting

## Code Style Guidelines

### Comments
- **DO NOT add comments unless explicitly asked by the user**
- Code should be self-documenting through clear naming and structure
- Never add comments that simply restate what the code does
- Only add comments when the user specifically requests them

### Error Handling
- When wrapping errors, don't repeat "error" or "failed" in the message
- Use descriptive context without redundant error terminology
- **Good**: `fmt.Errorf("getting kubeconfig: %w", err)`
- **Bad**: `fmt.Errorf("failed to get kubeconfig: %w", err)` or `fmt.Errorf("kubeconfig error: %w", err)`

### Commit Messages
- **NEVER add co-author credits or Claude Code attribution to commit messages**
- Keep commit messages clean and focused on the actual changes
- Follow conventional commit format when appropriate
- Do not include any AI assistance acknowledgments