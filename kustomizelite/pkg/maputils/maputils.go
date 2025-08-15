package maputils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var arrayIndexRegex = regexp.MustCompile(`^(.+)\[(\d+)\]$`)

func parsePath(path string) []string {
	if path == "" {
		return []string{}
	}

	parts := strings.Split(path, ".")
	result := []string{}

	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

func Get[T any](m map[string]any, path string) (T, error) {
	var zero T

	value, err := getValue(m, path)
	if err != nil {
		return zero, err
	}

	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("value at path %q is %T, not %T", path, value, zero)
	}

	return typed, nil
}

func getValue(m map[string]any, path string) (any, error) {
	if m == nil {
		return nil, fmt.Errorf("map is nil")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return m, nil
	}

	current := any(m)

	for i, part := range parts {
		if matches := arrayIndexRegex.FindStringSubmatch(part); matches != nil {
			key := matches[1]
			index, _ := strconv.Atoi(matches[2])

			switch v := current.(type) {
			case map[string]any:
				sliceValue, exists := v[key]
				if !exists {
					return nil, fmt.Errorf("key %q not found at path %q", key, strings.Join(parts[:i+1], "."))
				}

				slice, ok := sliceValue.([]any)
				if !ok {
					return nil, fmt.Errorf("value at %q is not a slice", strings.Join(parts[:i], "."))
				}

				if index < 0 || index >= len(slice) {
					return nil, fmt.Errorf("index %d out of range for slice at %q (length %d)", index, strings.Join(parts[:i+1], "."), len(slice))
				}

				current = slice[index]
			default:
				return nil, fmt.Errorf("cannot index into %T at path %q", v, strings.Join(parts[:i], "."))
			}
		} else {
			switch v := current.(type) {
			case map[string]any:
				value, exists := v[part]
				if !exists {
					return nil, fmt.Errorf("key %q not found at path %q", part, strings.Join(parts[:i+1], "."))
				}
				current = value
			default:
				return nil, fmt.Errorf("cannot access key %q in %T at path %q", part, v, strings.Join(parts[:i], "."))
			}
		}
	}

	return current, nil
}

func Set(m map[string]any, path string, value any) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	current := m

	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		if matches := arrayIndexRegex.FindStringSubmatch(part); matches != nil {
			key := matches[1]
			index, _ := strconv.Atoi(matches[2])

			sliceValue, exists := current[key]
			if !exists {
				return fmt.Errorf("key %q not found", key)
			}

			slice, ok := sliceValue.([]any)
			if !ok {
				return fmt.Errorf("value at %q is not a slice", key)
			}

			if index < 0 || index >= len(slice) {
				return fmt.Errorf("index %d out of range", index)
			}

			if i == len(parts)-2 {
				lastPart := parts[len(parts)-1]
				if elem, ok := slice[index].(map[string]any); ok {
					elem[lastPart] = value
					return nil
				}
				return fmt.Errorf("element at index %d is not a map", index)
			}

			if elem, ok := slice[index].(map[string]any); ok {
				current = elem
			} else {
				return fmt.Errorf("element at index %d is not a map", index)
			}
		} else {
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}

			next, ok := current[part].(map[string]any)
			if !ok {
				return fmt.Errorf("value at %q is not a map", part)
			}
			current = next
		}
	}

	lastPart := parts[len(parts)-1]
	if matches := arrayIndexRegex.FindStringSubmatch(lastPart); matches != nil {
		key := matches[1]
		index, _ := strconv.Atoi(matches[2])

		sliceValue, exists := current[key]
		if !exists {
			return fmt.Errorf("key %q not found", key)
		}

		slice, ok := sliceValue.([]any)
		if !ok {
			return fmt.Errorf("value at %q is not a slice", key)
		}

		if index < 0 || index >= len(slice) {
			return fmt.Errorf("index %d out of range", index)
		}

		slice[index] = value
	} else {
		current[lastPart] = value
	}

	return nil
}

func Has(m map[string]any, path string) bool {
	_, err := getValue(m, path)
	return err == nil
}

func Delete(m map[string]any, path string) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	if len(parts) == 1 {
		delete(m, parts[0])
		return nil
	}

	parentPath := strings.Join(parts[:len(parts)-1], ".")
	parent, err := getValue(m, parentPath)
	if err != nil {
		return err
	}

	parentMap, ok := parent.(map[string]any)
	if !ok {
		return fmt.Errorf("parent at path %q is not a map", parentPath)
	}

	delete(parentMap, parts[len(parts)-1])
	return nil
}

func Merge(dest map[string]any, src map[string]any, path string) error {
	if dest == nil {
		return fmt.Errorf("destination map is nil")
	}

	if src == nil {
		return nil
	}

	if path == "" {
		for k, v := range src {
			dest[k] = v
		}
		return nil
	}

	if err := EnsurePath(dest, path); err != nil {
		return err
	}

	targetValue, err := getValue(dest, path)
	if err != nil {
		return err
	}

	target, ok := targetValue.(map[string]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a map", path)
	}

	for k, v := range src {
		target[k] = v
	}

	return nil
}

func EnsurePath(m map[string]any, path string) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return nil
	}

	current := m

	for i, part := range parts {
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]any)
		}

		next, ok := current[part].(map[string]any)
		if !ok {
			return fmt.Errorf("value at %q is not a map", strings.Join(parts[:i+1], "."))
		}
		current = next
	}

	return nil
}

func GetOrDefault[T any](m map[string]any, path string, defaultValue T) T {
	value, err := Get[T](m, path)
	if err != nil {
		return defaultValue
	}
	return value
}

func MustGet[T any](m map[string]any, path string) T {
	value, err := Get[T](m, path)
	if err != nil {
		panic(fmt.Sprintf("failed to get value at path %q: %v", path, err))
	}
	return value
}

func GetStringMap(m map[string]any, path string) (map[string]string, error) {
	if path == "" {
		result := make(map[string]string)
		for k, v := range m {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result, nil
	}

	value, err := getValue(m, path)
	if err != nil {
		return nil, err
	}

	anyMap, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("value at path %q is not a map", path)
	}

	result := make(map[string]string)
	for k, v := range anyMap {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("value at key %q is not a string", k)
		}
		result[k] = str
	}

	return result, nil
}

func MergeStringMap(m map[string]any, path string, values map[string]string) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	if len(values) == 0 {
		return nil
	}

	if err := EnsurePath(m, path); err != nil {
		return err
	}

	targetValue, err := getValue(m, path)
	if err != nil {
		return err
	}

	target, ok := targetValue.(map[string]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a map", path)
	}

	for k, v := range values {
		target[k] = v
	}

	return nil
}

func GetSlice[T any](m map[string]any, path string) ([]T, error) {
	value, err := getValue(m, path)
	if err != nil {
		return nil, err
	}

	anySlice, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("value at path %q is not a slice", path)
	}

	result := make([]T, 0, len(anySlice))
	for i, v := range anySlice {
		typed, ok := v.(T)
		if !ok {
			return nil, fmt.Errorf("element at index %d is %T, not %T", i, v, *new(T))
		}
		result = append(result, typed)
	}

	return result, nil
}

func AppendToSlice[T any](m map[string]any, path string, values ...T) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	if len(values) == 0 {
		return nil
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	if len(parts) == 1 {
		existing, exists := m[parts[0]]
		if !exists {
			slice := make([]any, 0, len(values))
			for _, v := range values {
				slice = append(slice, v)
			}
			m[parts[0]] = slice
			return nil
		}

		slice, ok := existing.([]any)
		if !ok {
			return fmt.Errorf("value at path %q is not a slice", path)
		}

		for _, v := range values {
			slice = append(slice, v)
		}
		m[parts[0]] = slice
		return nil
	}

	parentPath := strings.Join(parts[:len(parts)-1], ".")
	if err := EnsurePath(m, parentPath); err != nil {
		return err
	}

	parent, err := getValue(m, parentPath)
	if err != nil {
		return err
	}

	parentMap, ok := parent.(map[string]any)
	if !ok {
		return fmt.Errorf("parent at path %q is not a map", parentPath)
	}

	key := parts[len(parts)-1]
	existing, exists := parentMap[key]
	if !exists {
		slice := make([]any, 0, len(values))
		for _, v := range values {
			slice = append(slice, v)
		}
		parentMap[key] = slice
		return nil
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	for _, v := range values {
		slice = append(slice, v)
	}
	parentMap[key] = slice

	return nil
}

func PrependToSlice[T any](m map[string]any, path string, values ...T) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	if len(values) == 0 {
		return nil
	}

	existing, err := getValue(m, path)
	if err != nil {
		slice := make([]any, 0, len(values))
		for _, v := range values {
			slice = append(slice, v)
		}
		return Set(m, path, slice)
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	newSlice := make([]any, 0, len(values)+len(slice))
	for _, v := range values {
		newSlice = append(newSlice, v)
	}
	newSlice = append(newSlice, slice...)

	return Set(m, path, newSlice)
}

func InsertIntoSlice[T any](m map[string]any, path string, index int, values ...T) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	if len(values) == 0 {
		return nil
	}

	existing, err := getValue(m, path)
	if err != nil {
		return err
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	if index < 0 || index > len(slice) {
		return fmt.Errorf("index %d out of range for slice of length %d", index, len(slice))
	}

	newSlice := make([]any, 0, len(slice)+len(values))
	newSlice = append(newSlice, slice[:index]...)
	for _, v := range values {
		newSlice = append(newSlice, v)
	}
	newSlice = append(newSlice, slice[index:]...)

	return Set(m, path, newSlice)
}

func RemoveFromSlice(m map[string]any, path string, index int) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	existing, err := getValue(m, path)
	if err != nil {
		return err
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	if index < 0 || index >= len(slice) {
		return fmt.Errorf("index %d out of range for slice of length %d", index, len(slice))
	}

	newSlice := make([]any, 0, len(slice)-1)
	newSlice = append(newSlice, slice[:index]...)
	newSlice = append(newSlice, slice[index+1:]...)

	return Set(m, path, newSlice)
}

func UpdateSliceElement(m map[string]any, path string, index int, value any) error {
	slicePath := path

	existing, err := getValue(m, slicePath)
	if err != nil {
		return err
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", slicePath)
	}

	if index < 0 || index >= len(slice) {
		return fmt.Errorf("index %d out of range for slice of length %d", index, len(slice))
	}

	slice[index] = value
	return nil
}

func SliceLength(m map[string]any, path string) (int, error) {
	existing, err := getValue(m, path)
	if err != nil {
		return 0, err
	}

	slice, ok := existing.([]any)
	if !ok {
		return 0, fmt.Errorf("value at path %q is not a slice", path)
	}

	return len(slice), nil
}

func FilterSlice[T any](m map[string]any, path string, predicate func(T) bool) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	existing, err := getValue(m, path)
	if err != nil {
		return err
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	filtered := make([]any, 0)
	for _, item := range slice {
		typed, ok := item.(T)
		if !ok {
			continue
		}
		if predicate(typed) {
			filtered = append(filtered, item)
		}
	}

	return Set(m, path, filtered)
}

func MapSlice[T, R any](m map[string]any, path string, mapper func(T) R) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	existing, err := getValue(m, path)
	if err != nil {
		return err
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	mapped := make([]any, 0, len(slice))
	for i, item := range slice {
		typed, ok := item.(T)
		if !ok {
			return fmt.Errorf("element at index %d is %T, not %T", i, item, *new(T))
		}
		mapped = append(mapped, mapper(typed))
	}

	return Set(m, path, mapped)
}

func FindInSlice[T any](m map[string]any, path string, predicate func(T) bool) (T, int, error) {
	var zero T

	existing, err := getValue(m, path)
	if err != nil {
		return zero, -1, err
	}

	slice, ok := existing.([]any)
	if !ok {
		return zero, -1, fmt.Errorf("value at path %q is not a slice", path)
	}

	for i, item := range slice {
		typed, ok := item.(T)
		if !ok {
			continue
		}
		if predicate(typed) {
			return typed, i, nil
		}
	}

	return zero, -1, fmt.Errorf("no matching element found")
}

func ContainsInSlice[T comparable](m map[string]any, path string, value T) (bool, error) {
	existing, err := getValue(m, path)
	if err != nil {
		return false, err
	}

	slice, ok := existing.([]any)
	if !ok {
		return false, fmt.Errorf("value at path %q is not a slice", path)
	}

	for _, item := range slice {
		typed, ok := item.(T)
		if ok && typed == value {
			return true, nil
		}
	}

	return false, nil
}

func MergeSlices[T any](m map[string]any, path string, values []T, deduplicate bool) error {
	if m == nil {
		return fmt.Errorf("map is nil")
	}

	if len(values) == 0 {
		return nil
	}

	existing, err := getValue(m, path)
	if err != nil {
		slice := make([]any, 0, len(values))
		for _, v := range values {
			slice = append(slice, v)
		}
		return Set(m, path, slice)
	}

	slice, ok := existing.([]any)
	if !ok {
		return fmt.Errorf("value at path %q is not a slice", path)
	}

	if !deduplicate {
		for _, v := range values {
			slice = append(slice, v)
		}
		return Set(m, path, slice)
	}

	seen := make(map[string]bool)
	deduped := make([]any, 0)

	for _, item := range slice {
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, item)
		}
	}

	for _, v := range values {
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, v)
		}
	}

	return Set(m, path, deduped)
}
