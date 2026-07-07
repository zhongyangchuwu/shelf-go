package vaultmigrate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func ConvertProjectManifest(manifest project.Manifest, paths map[string]PathMapping) (ProjectResult, error) {
	if err := project.Validate(manifest); err != nil {
		return ProjectResult{}, err
	}
	result := ProjectResult{Manifest: ProjectManifest{Version: ProjectManifestVersion, Bindings: []ProjectBinding{}}}
	seen := map[string]struct{}{}
	for _, entry := range manifest.Secrets {
		var bindings []ProjectBinding
		var warnings []Warning
		var err error
		switch {
		case entry.Path != "":
			bindings, err = convertPathEntry(entry, paths)
		case entry.Prefix != "":
			bindings, warnings, err = convertPrefixEntry(entry, paths)
		case len(entry.Tags) > 0:
			bindings, warnings, err = convertTagEntry(entry, paths)
		}
		if err != nil {
			return ProjectResult{}, err
		}
		result.Warnings = append(result.Warnings, warnings...)
		for _, binding := range bindings {
			key := binding.Item + ":" + binding.Field + ":" + binding.Env
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result.Manifest.Bindings = append(result.Manifest.Bindings, binding)
		}
	}
	if err := ValidateProjectManifest(result.Manifest); err != nil {
		return ProjectResult{}, err
	}
	return result, nil
}

func convertPathEntry(entry project.Entry, paths map[string]PathMapping) ([]ProjectBinding, error) {
	mapping, ok := paths[entry.Path]
	if !ok {
		return nil, fmt.Errorf("missing migration mapping for path: %s", entry.Path)
	}
	env := entry.Env
	if env == "" {
		env = mapping.EnvName
	}
	return []ProjectBinding{bindingForEntry(entry, mapping, env)}, nil
}

func convertPrefixEntry(entry project.Entry, paths map[string]PathMapping) ([]ProjectBinding, []Warning, error) {
	mappings := mappingsForPrefix(entry.Prefix, paths)
	if len(mappings) == 0 {
		return nil, []Warning{{Selector: entry.Prefix, Message: "prefix matched no migrated secrets"}}, nil
	}
	bindings := make([]ProjectBinding, 0, len(mappings))
	for _, mapping := range mappings {
		bindings = append(bindings, bindingForEntry(entry, mapping, mapping.EnvName))
	}
	return bindings, nil, nil
}

func convertTagEntry(entry project.Entry, paths map[string]PathMapping) ([]ProjectBinding, []Warning, error) {
	mappings := mappingsForTags(entry.Tags, paths)
	selector := strings.Join(entry.Tags, ",")
	if len(mappings) == 0 {
		return nil, []Warning{{Selector: selector, Message: "tags matched no migrated secrets"}}, nil
	}
	bindings := make([]ProjectBinding, 0, len(mappings))
	for _, mapping := range mappings {
		bindings = append(bindings, bindingForEntry(entry, mapping, mapping.EnvName))
	}
	return bindings, nil, nil
}

func bindingForEntry(entry project.Entry, mapping PathMapping, env string) ProjectBinding {
	return ProjectBinding{Item: mapping.ItemID, Field: mapping.FieldID, Env: env, Required: entry.IsRequired(), Title: mapping.Title}
}

func mappingsForPrefix(prefix string, paths map[string]PathMapping) []PathMapping {
	matches := make([]PathMapping, 0)
	for path, mapping := range paths {
		if strings.HasPrefix(path, prefix) {
			matches = append(matches, mapping)
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].OldPath < matches[j].OldPath })
	return matches
}

func mappingsForTags(tags []string, paths map[string]PathMapping) []PathMapping {
	matches := make([]PathMapping, 0)
	for _, mapping := range paths {
		if hasTags(mapping.Labels, tags) {
			matches = append(matches, mapping)
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].OldPath < matches[j].OldPath })
	return matches
}

func hasTags(labels, tags []string) bool {
	for _, want := range tags {
		found := false
		for _, got := range labels {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func ValidateProjectManifest(manifest ProjectManifest) error {
	if manifest.Version != ProjectManifestVersion {
		return fmt.Errorf("unsupported v2 project manifest version: %d", manifest.Version)
	}
	seen := map[string]struct{}{}
	for i, binding := range manifest.Bindings {
		if binding.Item == "" {
			return fmt.Errorf("bindings[%d].item is empty", i)
		}
		if binding.Field == "" {
			return fmt.Errorf("bindings[%d].field is empty", i)
		}
		if binding.Env == "" {
			return fmt.Errorf("bindings[%d].env is empty", i)
		}
		if err := vault.ValidateEnvName(binding.Env); err != nil {
			return fmt.Errorf("bindings[%d].env invalid: %s", i, binding.Env)
		}
		key := binding.Item + ":" + binding.Field + ":" + binding.Env
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate binding: %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}
