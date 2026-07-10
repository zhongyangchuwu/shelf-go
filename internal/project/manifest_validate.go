package project

import (
	"fmt"
	"regexp"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func Validate(manifest Manifest) error {
	if manifest.Version != CurrentVersion {
		return fmt.Errorf("unsupported project manifest version %d", manifest.Version)
	}
	if manifest.Bindings == nil {
		return fmt.Errorf("project manifest bindings array is required")
	}
	seen := make(map[string]struct{}, len(manifest.Bindings))
	for i, binding := range manifest.Bindings {
		if binding.Item == "" {
			return fmt.Errorf("bindings[%d].item is required", i)
		}
		if binding.Field == "" {
			return fmt.Errorf("bindings[%d].field is required", i)
		}
		if binding.Env == "" {
			return fmt.Errorf("bindings[%d].env is required", i)
		}
		if !envNamePattern.MatchString(binding.Env) {
			return fmt.Errorf("invalid bindings[%d].env: %s", i, binding.Env)
		}
		key := binding.Key()
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate binding item/field: %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}
