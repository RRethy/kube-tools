package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/RRethy/utils/celery/pkg/cli/validate"
)

const defaultWorkers = 128

var (
	celExpression string
	ruleFiles     []string
	verbose       bool
	maxWorkers    int

	targetGroup              string
	targetVersion            string
	targetKind               string
	targetName               string
	targetNamespace          string
	targetLabelSelector      string
	targetAnnotationSelector string
)

var validateCmd = &cobra.Command{
	Use:   "validate [files...]",
	Short: "Validate Kubernetes KRM resources using CEL expressions",
	Long: `Evaluate CEL expressions against Kubernetes resources to check if they meet 
specified conditions.

Input sources:
  • File paths as arguments (supports multiple files)
  • Stdin when no files specified
  • Supports multi-document YAML

Validation rules:
  • Inline via --expression flag
  • Policy files via --rule-file flag
  • Target specific resources using selectors

The command exits with status 1 if any validation fails. Multiple files are 
processed in parallel for performance.`,
	Example: `# Validate a single file
celery validate deployment.yaml --expression "spec.replicas >= 3"

# Validate multiple files
celery validate deployment.yaml service.yaml configmap.yaml --rule-file validation-rules.yaml

# Validate with multiple rule files
celery validate deployment.yaml --rule-file base-rules.yaml --rule-file prod-rules.yaml

# Validate with glob pattern (must be quoted)
celery validate deployment.yaml --rule-file "rules/*.yaml"

# Validate all YAML files in a directory
celery validate *.yaml --rule-file validation-rules.yaml

# Validate from stdin
cat deployment.yaml | celery validate --expression "spec.replicas >= 3"

# Validate only Deployments
celery validate resources.yaml -e "object.spec.replicas >= 3" --target-kind Deployment

# Validate resources with specific labels
celery validate resources.yaml -e "has(object.spec.template)" --target-labels "tier=frontend"

# Validate resources in production namespace
celery validate resources.yaml -e "object.spec.replicas >= 5" --target-namespace production

# Validate specific API group/version/kind
celery validate resources.yaml -e "has(object.spec.parallelism)" --target-group batch --target-version v1 --target-kind Job

# Combine multiple selectors (all must match)
celery validate resources.yaml -e "object.spec.replicas >= 3" --target-kind Deployment --target-labels "environment=prod"`,
	RunE: func(_ *cobra.Command, args []string) error {
		return validate.Validate(
			context.Background(),
			args,
			celExpression,
			ruleFiles,
			verbose,
			maxWorkers,
			targetGroup,
			targetVersion,
			targetKind,
			targetName,
			targetNamespace,
			targetLabelSelector,
			targetAnnotationSelector,
		)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&celExpression, "expression", "e", "", "CEL expression to validate resources")
	validateCmd.Flags().StringSliceVarP(&ruleFiles, "rule-file", "r", []string{}, "YAML files containing validation rules (supports globs when quoted, can be specified multiple times)")
	validateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all validation results including passes")
	validateCmd.Flags().IntVar(&maxWorkers, "max-workers", defaultWorkers, "Maximum number of parallel workers for multi-file validation")

	validateCmd.Flags().StringVar(&targetGroup, "target-group", "", "Target resources by API group (e.g., apps, batch)")
	validateCmd.Flags().StringVar(&targetVersion, "target-version", "", "Target resources by API version (e.g., v1, v1beta1)")
	validateCmd.Flags().StringVar(&targetKind, "target-kind", "", "Target resources by kind (supports regex)")
	validateCmd.Flags().StringVar(&targetName, "target-name", "", "Target resources by name (supports regex)")
	validateCmd.Flags().StringVar(&targetNamespace, "target-namespace", "", "Target resources in specific namespace")
	validateCmd.Flags().StringVar(&targetLabelSelector, "target-labels", "", "Target resources by label selector (e.g., 'app=nginx,tier=frontend')")
	validateCmd.Flags().StringVar(&targetAnnotationSelector, "target-annotations", "", "Target resources by annotation selector")

	validateCmd.MarkFlagsMutuallyExclusive("expression", "rule-file")
	validateCmd.MarkFlagsOneRequired("expression", "rule-file")
}
