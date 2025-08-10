# Celery

A Kubernetes Resource Model (KRM) YAML validator using Common Expression Language (CEL) validation rules.

## Features

- Validate Kubernetes resources against CEL expressions
- Support for both file input and stdin
- Inline CEL expressions or KRM-based ValidationPolicy files
- Selective validation using resource selectors (kind, apiVersion, labels)
- Verbose output mode for detailed validation results

## Installation

```bash
# From the workspace root
make build-celery

# Or from the celery directory
go build -o bin/celery .
```

## Usage

### Basic validation with inline expression

```bash
# Validate a file
celery validate deployment.yaml --expression "object.spec.replicas >= 3"

# Validate from stdin
cat deployment.yaml | celery validate --expression "object.spec.replicas >= 3"
```

### Using a rules file

```bash
celery validate deployment.yaml --rule-file validation-policy.yaml
```

### Examples

See the `examples/` directory for complete working examples including:
- Basic validation policies
- Multi-document policies
- Advanced selectors (labels, annotations, regex)
- Sample Kubernetes resources for testing

### ValidationRules Format (KRM)

Validation rules are defined as Kubernetes Resource Model (KRM) resources using the `ValidationRules` kind:

```yaml
apiVersion: celery.rrethy.io/v1alpha1
kind: ValidationRules
metadata:
  name: production-standards
  namespace: default
spec:
  rules:
    - name: minimum-replicas
      expression: "object.spec.replicas >= 3"
      message: "Deployments must have at least 3 replicas for high availability"
      target:
        kind: Deployment
    
    - name: required-labels
      expression: "has(object.metadata.labels) && has(object.metadata.labels.app)"
      message: "Resources must have an 'app' label"
```

#### Multiple Rules

You can define multiple ValidationRules resources in a single file using YAML document separators:

```yaml
apiVersion: celery.rrethy.io/v1alpha1
kind: ValidationRules
metadata:
  name: deployment-standards
spec:
  rules:
    - name: minimum-replicas
      expression: "object.spec.replicas >= 3"
      message: "Deployments must have at least 3 replicas"
      target:
        kind: Deployment
---
apiVersion: celery.rrethy.io/v1alpha1
kind: ValidationRules
metadata:
  name: service-standards
spec:
  rules:
    - name: service-port-check
      expression: "object.spec.ports.all(p, p.port > 0 && p.port <= 65535)"
      message: "Service ports must be valid"
      target:
        kind: Service
```

### Target Selectors (Kustomize-style)

The `target` field uses the same format as Kustomize patches, allowing precise targeting of resources:

```yaml
target:
  group: <optional API group>              # e.g., "apps", "batch"
  version: <optional version>               # e.g., "v1", "v1beta1"
  kind: <optional kind>                     # e.g., "Deployment", "Service"
  name: <optional name/regex>               # e.g., "my-app" or "web-.*"
  namespace: <optional namespace>           # e.g., "production"
  labelSelector: <optional label selector>  # e.g., "app=nginx,tier=frontend"
  annotationSelector: <optional selector>   # e.g., "criticality=high"
```

#### Examples:

```yaml
# Target all Deployments
target:
  kind: Deployment

# Target by API group/version/kind
target:
  group: apps
  version: v1
  kind: Deployment

# Target by label selector
target:
  kind: Deployment
  labelSelector: "environment=production,tier in (frontend,backend)"

# Target by name pattern (regex)
target:
  kind: Service
  name: "^web-.*"  # Matches web-app, web-service, etc.

# Target multiple kinds with regex
target:
  kind: "(Deployment|StatefulSet|DaemonSet)"

# Target by namespace
target:
  kind: Deployment
  namespace: production
```

## CEL Expression Context

The following variables are available in CEL expressions:

- `object`: The current Kubernetes resource being validated
- `oldObject`: Previous version of the object (for updates, currently null)
- `request`: Admission request context (currently null)
- `params`: Validation parameters (currently null)
- `namespaceObject`: Namespace object (currently null)
- `authorizer`: Authorization context (currently null)

## Common CEL Functions

- `has()`: Check if a field exists
- `size()`: Get the size of a list or map
- `matches()`: Regular expression matching
- `all()`, `exists()`: List predicates
- Standard operators: `==`, `!=`, `>`, `>=`, `<`, `<=`, `&&`, `||`, `!`

## Examples

```bash
# Check minimum replicas
--expression "object.spec.replicas >= 3"

# Ensure namespace is set
--expression "has(object.metadata.namespace) && object.metadata.namespace != ''"

# No latest tags in container images
--expression "!object.spec.template.spec.containers.exists(c, c.image.endsWith(':latest'))"

# Check resource limits
--expression "object.spec.template.spec.containers.all(c, has(c.resources.limits.memory))"
```

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Build binary
go build -o bin/celery .
```
