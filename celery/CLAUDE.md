# CLAUDE.md - Celery Module

This document provides guidance to Claude Code when working with the celery module of this repository.

## Module Overview

Celery is a Kubernetes Resource Model (KRM) YAML validator using Common Expression Language (CEL) validation rules. It validates Kubernetes resources against CEL expressions to ensure they meet specified requirements.

## Key Features

- CEL expression validation for Kubernetes resources
- Support for inline expressions and ValidationRules files
- Kustomize-style target selectors
- Parallel validation of multiple files
- Verbose output mode for showing passing validations
- Glob pattern support for rule files
- Cross-resource validation with `allObjects`
- IOStreams support for testable I/O

## Architecture

```
celery/
├── cmd/
│   ├── root.go       # Root command setup
│   └── validate.go   # Validate command implementation
├── pkg/
│   ├── cli/
│   │   └── validate/
│   │       ├── validate.go    # CLI entry point with IOStreams
│   │       └── validater.go   # Validation orchestration
│   ├── validator/
│   │   └── validator.go       # Core CEL validation logic
│   └── yaml/
│       └── yaml.go            # YAML parsing utilities
├── api/
│   └── v1/
│       └── validation_policy.go  # ValidationRules types
├── fixtures/         # Test fixtures and example validation policies
│   ├── rules/       # Example ValidationRules
│   └── resources/   # Example Kubernetes resources
├── go.mod
└── main.go
```

## Development Commands

```bash
# Build the binary
make build-celery

# Run tests
cd celery && go test ./...

# Format code
cd celery && go fmt ./...

# Run linter
cd celery && golangci-lint run
```

## Key Components

### ValidationRules (KRM Resource)
- Kubernetes Resource Model format
- Contains validation rules with CEL expressions
- Supports target selectors for selective validation (implemented)

### Validator Package
- Parses YAML resources into unstructured objects
- Compiles and evaluates CEL expressions
- Supports Kustomize-style target matching

### Target Selectors
- group: API group (e.g., "apps", "batch")
- version: API version (e.g., "v1", "v1beta1")
- kind: Resource kind
- name: Resource name
- namespace: Resource namespace
- labelSelector: Label selector string
- annotationSelector: Annotation selector string

## CEL Expression Context

Available variables in expressions:
- `object`: Current resource being validated
- `allObjects`: List of all resources in the current validation batch (for cross-resource validation)

Note: `oldObject`, `request`, `params`, `namespaceObject`, `authorizer` are reserved for future Kubernetes admission webhook compatibility

## Testing

When adding new features:
1. Add unit tests for validator logic
2. Add example files demonstrating the feature
3. Test with various Kubernetes resource types
4. Ensure parallel validation works correctly

## Common Tasks

### Adding a New Validation Feature
1. Update api/v1/validation_policy.go if new fields are needed
2. Implement logic in pkg/validator/validator.go
3. Add command flags in cmd/validate.go if needed
4. Update pkg/cli/validate/validater.go for CLI orchestration
5. Add test fixtures in fixtures/ demonstrating the feature
6. Add tests for the new functionality

### Updating Dependencies
1. Update go.mod with new dependencies
2. Run `go mod tidy` from the celery directory
3. Run `go work sync` from workspace root