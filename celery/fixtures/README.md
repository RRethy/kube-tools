# Celery Test Fixtures

This directory contains test fixtures including example ValidationRules and test resources for the Celery validator.

## Directory Structure

```
fixtures/
├── rules/                    # ValidationRules definitions
│   ├── basic-validation.yaml
│   ├── deployment-standards.yaml
│   ├── service-rules.yaml
│   └── ...
├── resources/               # Test Kubernetes resources
│   ├── valid-deployment.yaml
│   ├── invalid-deployments.yaml
│   ├── services.yaml
│   └── ...
└── README.md
```

## ValidationRules Files (`rules/`)

### Basic Validation
- `basic-validation.yaml` - Simple checks for labels, namespaces, and annotations
- `deployment-standards.yaml` - Best practices for Deployments (replicas, resource limits, probes)
- `service-rules.yaml` - Service port validation and selector requirements
- `ingress-validation.yaml` - TLS requirements and host format validation

### Security Policies
- `pod-security-policies.yaml` - Container security requirements (non-root, read-only filesystem)
- `configmap-secrets-validation.yaml` - Size limits and password detection

### Advanced Validation
- `cross-resource-validation.yaml` - Validates relationships between resources
- `statefulset-rules.yaml` - StatefulSet-specific requirements
- `namespace-policies.yaml` - Namespace-based rule targeting
- `api-version-validation.yaml` - Enforces preferred API versions
- `multi-rule-example.yaml` - Multiple ValidationRules in one file
- `deployment-replicas.yaml` - Environment-based replica requirements

## Test Resources (`resources/`)

### Individual Resource Types
- `valid-deployment.yaml` - Fully compliant deployment example
- `invalid-deployments.yaml` - Deployments with various validation failures
- `services.yaml` - Valid and invalid Service configurations
- `configmaps-secrets.yaml` - ConfigMaps and Secrets for testing
- `statefulsets.yaml` - StatefulSet examples
- `ingresses.yaml` - Ingress resources with TLS and routing rules

### Mixed Resources
- `test-resources.yaml` - Mix of valid and invalid resources
- `cross-reference-resources.yaml` - Resources with volume mounts and cross-references

## Usage Examples

### Validate a single file with inline expression
```bash
celery validate resources/valid-deployment.yaml --expression "object.spec.replicas >= 3"
```

### Use a validation rules file
```bash
celery validate resources/invalid-deployments.yaml --rule-file rules/deployment-standards.yaml
```

### Validate multiple files with multiple rule files
```bash
celery validate resources/*.yaml \
  --rule-file rules/deployment-standards.yaml \
  --rule-file rules/service-rules.yaml
```

### Target specific resources
```bash
# Only validate Deployments
celery validate resources/*.yaml --rule-file rules/deployment-standards.yaml --target-kind Deployment

# Only validate resources in production namespace
celery validate resources/*.yaml --rule-file rules/namespace-policies.yaml --target-namespace production

# Target by labels
celery validate resources/*.yaml --rule-file rules/pod-security-policies.yaml --target-labels "security=strict"
```

### Cross-resource validation
```bash
# Requires passing all resources together for cross-referencing
celery validate resources/cross-reference-resources.yaml --rule-file rules/cross-resource-validation.yaml
```


## Writing Custom Rules

### Basic Structure
```yaml
apiVersion: celery.rrethy.io/v1
kind: ValidationRules
metadata:
  name: my-rules
spec:
  rules:
    - name: rule-name
      expression: "CEL expression here"
      message: "Error message when validation fails"
      target:  # Optional targeting
        kind: Deployment
        namespace: production
```

### Common CEL Patterns

Check field exists:
```cel
has(object.metadata.labels)
```

Check all items in array:
```cel
object.spec.containers.all(c, has(c.resources.limits))
```

Check any item matches:
```cel
object.spec.containers.exists(c, c.image.endsWith(':latest'))
```

Regex matching:
```cel
object.metadata.name.matches('^[a-z0-9-]+$')
```

Cross-resource validation:
```cel
allObjects.exists(r, r.kind == 'Service' && r.metadata.name == object.metadata.name)
```

## Target Selectors

Selectors can be combined to precisely target resources:

- `kind`: Resource type
- `namespace`: Specific namespace
- `group`: API group (e.g., "apps")
- `version`: API version (e.g., "v1")
- `name`: Resource name
- `labelSelector`: Kubernetes label selector syntax
- `annotationSelector`: Annotation selector syntax
