package strategicmerge

import "github.com/RRethy/krepe/deepishcopy"

func Apply(resource, patch map[string]any) map[string]any {
	return mergeMaps(resource, patch)
}

func mergeMaps(resource, patch map[string]any) map[string]any {
	patch = deepishcopy.Copy(patch).(map[string]any)
	for key, value := range patch {
		if value == nil {
			delete(resource, key)
			continue
		}

		if existing, exists := resource[key]; exists {
			switch existing := existing.(type) {
			case map[string]any:
				if patchValue, ok := value.(map[string]any); ok {
					mergeMaps(existing, patchValue)
				} else {
					resource[key] = value
				}
			case []any:
				if patchValue, ok := value.([]any); ok {
					resource[key] = mergeSlices(existing, patchValue)
				} else {
					resource[key] = value
				}
			default:
				resource[key] = value
			}
		} else {
			resource[key] = value
		}
	}
	return resource
}

func mergeSlices(existing, patch []any) []any {
	if len(existing) == 0 {
		return patch
	}
	if len(patch) == 0 {
		return existing
	}

	if !haveSameElementTypes(existing, patch) {
		return patch
	}

	if _, isMap := existing[0].(map[string]any); isMap {
		return mergeObjectLists(existing, patch)
	}

	seen := make(map[any]struct{})
	result := make([]any, 0, len(existing)+len(patch))

	for _, item := range existing {
		if _, exists := seen[item]; !exists {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}

	for _, item := range patch {
		if _, exists := seen[item]; !exists {
			result = append(result, item)
			seen[item] = struct{}{}
		}
	}

	return result
}

func haveSameElementTypes(slice1, slice2 []any) bool {
	if len(slice1) == 0 || len(slice2) == 0 {
		return true
	}

	type1 := getElementType(slice1[0])
	type2 := getElementType(slice2[0])

	if type1 != type2 {
		return false
	}

	for _, element := range slice1 {
		if getElementType(element) != type1 {
			return false
		}
	}

	for _, element := range slice2 {
		if getElementType(element) != type2 {
			return false
		}
	}

	return true
}

func getElementType(item any) string {
	switch item.(type) {
	case map[string]any:
		return "map"
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	default:
		return "other"
	}
}

func mergeObjectLists(existing, patch []any) []any {
	if len(existing) == 0 {
		return patch
	}
	if len(patch) == 0 {
		return existing
	}

	mergeKey, found := findMergeKey(existing, patch)
	if !found {
		return patch
	}

	result := make([]any, 0, len(existing)+len(patch))
	seen := make(map[string]int)
	for i, item := range existing {
		itemMap := item.(map[string]any)
		if keyValue, exists := itemMap[mergeKey]; exists {
			if keyStr, ok := keyValue.(string); ok {
				seen[keyStr] = i
			}
		}
		result = append(result, itemMap)
	}

	for _, item := range patch {
		itemMap := item.(map[string]any)
		if keyValue, hasKey := itemMap[mergeKey]; hasKey {
			if keyStr, ok := keyValue.(string); ok {
				if origIdx, exists := seen[keyStr]; exists {
					result[origIdx] = mergeMaps(result[origIdx].(map[string]any), itemMap)
					continue
				}
			}
		}

		result = append(result, itemMap)
	}

	return result
}

func findMergeKey(existing, patch []any) (string, bool) {
	if len(existing) == 0 || len(patch) == 0 {
		return "", false
	}

	commonKeys := getCommonKeys(append(existing, patch...))
	if len(commonKeys) == 0 {
		return "", false
	}

	preferredKeys := []string{"name", "key", "type", "kind", "mountPath", "containerPort", "devicePath", "ip", "topologyKey"}
	for _, key := range preferredKeys {
		if commonKeys[key] {
			return key, true
		}
	}

	for key := range commonKeys {
		return key, true
	}
	return "", false
}

func getCommonKeys(items []any) map[string]bool {
	if len(items) == 0 {
		return nil
	}

	first, ok := items[0].(map[string]any)
	if !ok {
		return nil
	}

	common := make(map[string]bool)
	for key := range first {
		common[key] = true
	}

	for _, item := range items[1:] {
		m, ok := item.(map[string]any)
		if !ok {
			return nil
		}
		for key := range common {
			if _, exists := m[key]; !exists {
				delete(common, key)
			}
		}
	}

	return common
}
