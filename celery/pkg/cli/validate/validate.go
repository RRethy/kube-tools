package validate

import "context"

func Validate(
	ctx context.Context,
	files []string,
	celExpression string,
	ruleFiles []string,
	maxWorkers int,
	targetGroup string,
	targetVersion string,
	targetKind string,
	targetName string,
	targetNamespace string,
	targetLabelSelector string,
	targetAnnotationSelector string,
) error {
	v := &Validater{}
	return v.Validate(
		files,
		celExpression,
		ruleFiles,
		maxWorkers,
		targetGroup,
		targetVersion,
		targetKind,
		targetName,
		targetNamespace,
		targetLabelSelector,
		targetAnnotationSelector,
	)
}
