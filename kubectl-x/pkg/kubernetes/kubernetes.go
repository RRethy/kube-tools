package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List retrieves and type-asserts Kubernetes resources of type T
func List[T metav1.Object](ctx context.Context, client Interface) ([]T, error) {
	var obj T
	kind := reflect.TypeOf(obj).Elem().Name()
	kind = strings.ToLower(kind)

	objects, err := client.List(ctx, kind)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", kind, err)
	}

	typedObjects := make([]T, 0, len(objects))
	for i, object := range objects {
		typedObject, ok := object.(T)
		if !ok {
			return nil, fmt.Errorf("object %d is not a %s", i, kind)
		}
		typedObjects = append(typedObjects, typedObject)
	}
	return typedObjects, nil
}
