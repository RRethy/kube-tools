# Test Fixtures

This directory contains test fixtures for the adamize project.

## Directory Structure

- `valid/` - Valid kustomization.yaml files for testing successful parsing
  - `basic/` - Minimal valid kustomization with resources
  - `with-namespace/` - Kustomization with namespace and labels
  - `with-patches/` - Kustomization with patches and name transformations
  - `with-components/` - Kustomization with components and annotations
  - `empty/` - Minimal kustomization with only apiVersion and kind

- `invalid/` - Invalid YAML files for testing error handling
  - `malformed-yaml/` - YAML with incorrect indentation
  - `invalid-syntax/` - YAML with syntax errors
  - `duplicate-keys/` - YAML with duplicate keys
  - `tab-indentation/` - YAML using tabs instead of spaces

- `multi-doc/` - Multi-document YAML files
  - `multiple-documents/` - File with multiple YAML documents separated by `---`
  - `with-comments/` - Multi-document file with comments

## Usage

These fixtures are used by the test suite to ensure proper handling of various kustomization.yaml formats and edge cases.