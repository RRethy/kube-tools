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
	return ListInNamespace[T](ctx, client, "")
}

func ListObjectInNamespace(ctx context.Context, client Interface, kind, namespace string) ([]metav1.Object, error) {
	var objects []any
	var err error
	if namespace != "" {
		objects, err = client.ListInNamespace(ctx, kind, namespace)
	} else {
		objects, err = client.List(ctx, kind)
	}
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", kind, err)
	}

	typedObjects := make([]metav1.Object, 0, len(objects))
	for i, object := range objects {
		typedObject, ok := object.(metav1.Object)
		if !ok {
			return nil, fmt.Errorf("object %d is not a %s", i, kind)
		}
		typedObjects = append(typedObjects, typedObject)
	}
	return typedObjects, nil
}

// ListInNamespace retrieves and type-asserts resources of type T in the given namespace
func ListInNamespace[T metav1.Object](ctx context.Context, client Interface, namespace string) ([]T, error) {
	var obj T
	kind := reflect.TypeOf(obj).Elem().Name()
	kind = strings.ToLower(kind)

	var objects []any
	var err error
	if namespace != "" {
		objects, err = client.ListInNamespace(ctx, kind, namespace)
	} else {
		objects, err = client.List(ctx, kind)
	}
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
