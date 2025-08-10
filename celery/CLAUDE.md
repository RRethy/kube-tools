# CLAUDE.md - Celery Module

This document provides guidance to Claude Code when working with the celery module of this repository.

## Module Overview

Celery is a Kubernetes Resource Model (KRM) YAML validator using Common Expression Language (CEL) validation rules. It validates Kubernetes resources against CEL expressions to ensure they meet specified requirements.

## Key Features

- CEL expression validation for Kubernetes resources
- Support for inline expressions and ValidationPolicy files
- Kustomize-style target selectors
- Parallel validation of multiple files
- Verbose output mode

## Architecture

```
celery/
├── cmd/
│   ├── root.go       # Root command setup
│   └── validate.go   # Validate command implementation
├── pkg/
│   └── validator/
│       ├── types.go      # Type definitions
│       └── validator.go  # Core validation logic
├── examples/         # Example validation policies and resources
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
- Supports target selectors for selective validation

### Validator Package
- Parses YAML resources into unstructured objects
- Compiles and evaluates CEL expressions
- Supports Kustomize-style target matching

### Target Selectors
- group: API group (e.g., "apps", "batch")
- version: API version (e.g., "v1", "v1beta1")
- kind: Resource kind (supports regex)
- name: Resource name (supports regex)
- namespace: Resource namespace
- labelSelector: Label selector string
- annotationSelector: Annotation selector string

## CEL Expression Context

Available variables in expressions:
- `object`: Current resource being validated
- `objects`: List of all resources (for cross-reference)
- `oldObject`, `request`, `params`, `namespaceObject`, `authorizer`: Reserved for future use

## Testing

When adding new features:
1. Add unit tests for validator logic
2. Add example files demonstrating the feature
3. Test with various Kubernetes resource types
4. Ensure parallel validation works correctly

## Common Tasks

### Adding a New Validation Feature
1. Update types.go if new fields are needed
2. Implement logic in validator.go
3. Add command flags in validate.go if needed
4. Add examples demonstrating the feature

### Updating Dependencies
1. Update go.mod with new dependencies
2. Run `go mod tidy` from the celery directory
3. Run `go work sync` from workspace root