package validate

import (
	"context"
	"os"

	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func Validate(
	ctx context.Context,
	files []string,
	celExpression string,
	ruleFiles []string,
	verbose bool,
	maxWorkers int,
	targetGroup string,
	targetVersion string,
	targetKind string,
	targetName string,
	targetNamespace string,
	targetLabelSelector string,
	targetAnnotationSelector string,
) error {
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	
	v := &Validater{
		IOStreams: ioStreams,
	}
	return v.Validate(
		files,
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
}
