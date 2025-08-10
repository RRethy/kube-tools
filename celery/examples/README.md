# Celery Examples

This directory contains example ValidationRules and test resources for the Celery validator.

## ValidationRules Files

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

### Test Resources
- `test-deployment.yaml` - Valid deployment example
- `test-resources.yaml` - Mix of valid and invalid resources for testing
- `deployment-replicas.yaml` - Environment-based replica requirements

## Usage Examples

### Validate a single file with inline expression
```bash
celery validate test-deployment.yaml --expression "object.spec.replicas >= 3"
```

### Use a validation rules file
```bash
celery validate test-resources.yaml --rule-file deployment-standards.yaml
```

### Validate multiple files with multiple rule files
```bash
celery validate test-resources.yaml test-deployment.yaml \
  --rule-file deployment-standards.yaml \
  --rule-file service-rules.yaml
```

### Target specific resources
```bash
# Only validate Deployments
celery validate test-resources.yaml --rule-file deployment-standards.yaml --target-kind Deployment

# Only validate resources in production namespace
celery validate test-resources.yaml --rule-file namespace-policies.yaml --target-namespace production

# Target by labels
celery validate test-resources.yaml --rule-file pod-security-policies.yaml --target-labels "security=strict"
```

### Cross-resource validation
```bash
# Requires passing all resources together for cross-referencing
celery validate test-resources.yaml --rule-file cross-resource-validation.yaml
```

### Verbose output
```bash
celery validate test-resources.yaml --rule-file deployment-standards.yaml --verbose
```

## Writing Custom Rules

### Basic Structure
```yaml
apiVersion: celery.rrethy.io/v1alpha1
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

- `kind`: Resource type (supports regex)
- `namespace`: Specific namespace (supports regex)  
- `group`: API group (e.g., "apps")
- `version`: API version (e.g., "v1")
- `name`: Resource name (supports regex)
- `labelSelector`: Kubernetes label selector syntax
- `annotationSelector`: Annotation selector syntax
