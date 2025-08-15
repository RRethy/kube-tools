package patch

import (
	"fmt"
	"os"

	"github.com/RRethy/krepe/jsonpatch"
	"github.com/RRethy/k8s-tools/kustomizelite/pkg/strategicmerge"
	"gopkg.in/yaml.v3"
)

type Object struct {
	IsJSON       bool
	jsonPatchOps []jsonpatch.JsonPatch
	mergePatch   map[string]any
}

func ParseFile(path string) (*Object, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading patch file: %w", err)
	}
	return Parse(string(content))
}

func Parse(patchString string) (*Object, error) {
	if patchString == "" {
		return nil, fmt.Errorf("patch string is empty")
	}

	var patchOpsData []map[string]any
	if err := yaml.Unmarshal([]byte(patchString), &patchOpsData); err == nil {
		var patchOps []jsonpatch.JsonPatch
		for _, patchOpData := range patchOpsData {
			op, opOk := patchOpData["op"].(string)
			path, pathOk := patchOpData["path"].(string)
			if !opOk || !pathOk {
				break
			}

			value := patchOpData["value"]
			from, _ := patchOpData["from"].(string)

			patchOp, err := jsonpatch.NewJsonPatch(op, from, path, value)
			if err != nil {
				break
			}
			patchOps = append(patchOps, patchOp)
		}

		if len(patchOps) == len(patchOpsData) {
			return &Object{
				IsJSON:       true,
				jsonPatchOps: patchOps,
			}, nil
		}
	}

	var mergePatch map[string]any
	if err := yaml.Unmarshal([]byte(patchString), &mergePatch); err != nil {
		return nil, fmt.Errorf("parsing patch YAML as neither JSON patch nor merge patch: %w", err)
	}

	return &Object{
		IsJSON:     false,
		mergePatch: mergePatch,
	}, nil
}

func (p *Object) Apply(resource map[string]any) (map[string]any, error) {
	if p.IsJSON {
		return p.applyJSONPatches(resource)
	}
	return strategicmerge.Apply(resource, p.mergePatch), nil
}

func (p *Object) applyJSONPatches(resource map[string]any) (map[string]any, error) {
	current := any(resource)

	for _, patchOp := range p.jsonPatchOps {
		result, err := patchOp.Apply(current)
		if err != nil {
			return nil, fmt.Errorf("applying JSON patch: %w", err)
		}
		current = result
	}

	result, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("patch result is not a map[string]any, got %T", current)
	}

	return result, nil
}
