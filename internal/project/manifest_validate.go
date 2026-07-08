package project

import (
	"fmt"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func Validate(manifest Manifest) error {
	switch manifest.Version {
	case CurrentVersion:
		return validateV1Manifest(manifest)
	case V2Version:
		return validateV2Manifest(manifest)
	default:
		return fmt.Errorf("unsupported project manifest version %d", manifest.Version)
	}
}

func validateV1Manifest(manifest Manifest) error {
	if manifest.Secrets == nil {
		return fmt.Errorf("project manifest secrets array is required")
	}
	if len(manifest.Bindings) > 0 {
		return fmt.Errorf("project manifest bindings require version %d", V2Version)
	}
	seenPath := make(map[string]struct{}, len(manifest.Secrets))
	seenPrefix := make(map[string]struct{}, len(manifest.Secrets))
	seenTags := make(map[string]struct{}, len(manifest.Secrets))
	for i, entry := range manifest.Secrets {
		hasPath := entry.Path != ""
		hasPrefix := entry.Prefix != ""
		hasTags := len(entry.Tags) > 0
		selectorCount := 0
		for _, ok := range []bool{hasPath, hasPrefix, hasTags} {
			if ok {
				selectorCount++
			}
		}
		if selectorCount == 0 {
			return fmt.Errorf("secrets[%d]: path, prefix, or tags is required", i)
		}
		if selectorCount > 1 {
			return fmt.Errorf("secrets[%d]: path, prefix, and tags are mutually exclusive", i)
		}
		if hasPath {
			if err := vault.ValidatePath(entry.Path); err != nil {
				return fmt.Errorf("invalid secrets[%d].path: %w", i, err)
			}
			if _, ok := seenPath[entry.Path]; ok {
				return fmt.Errorf("duplicate secrets entry path: %s", entry.Path)
			}
			seenPath[entry.Path] = struct{}{}
		}
		if hasPrefix {
			if strings.Contains(entry.Prefix, ":") {
				return fmt.Errorf("invalid secrets[%d].prefix: must not contain ':'", i)
			}
			for _, segment := range strings.Split(entry.Prefix, "/") {
				if segment == "" {
					return fmt.Errorf("invalid secrets[%d].prefix: empty segment", i)
				}
				if !vault.IsPathToken(segment) {
					return fmt.Errorf("invalid secrets[%d].prefix: unsupported characters in segment %q", i, segment)
				}
			}
			if _, ok := seenPrefix[entry.Prefix]; ok {
				return fmt.Errorf("duplicate secrets entry prefix: %s", entry.Prefix)
			}
			seenPrefix[entry.Prefix] = struct{}{}
		}
		if hasTags {
			seenEntryTags := map[string]struct{}{}
			for _, tag := range entry.Tags {
				if tag == "" {
					return fmt.Errorf("invalid secrets[%d].tags: tag must not be empty", i)
				}
				if !vault.IsPathToken(tag) {
					return fmt.Errorf("invalid secrets[%d].tags: unsupported characters in tag %q", i, tag)
				}
				if _, ok := seenEntryTags[tag]; ok {
					return fmt.Errorf("invalid secrets[%d].tags: duplicate tag %s", i, tag)
				}
				seenEntryTags[tag] = struct{}{}
			}
			key := entry.Key()
			if _, ok := seenTags[key]; ok {
				return fmt.Errorf("duplicate secrets entry tags: %s", key)
			}
			seenTags[key] = struct{}{}
		}
		if entry.Env != "" && hasPrefix {
			return fmt.Errorf("secrets[%d]: prefix entries must not carry env", i)
		}
		if entry.Env != "" && hasTags {
			return fmt.Errorf("secrets[%d]: tag entries must not carry env", i)
		}
		if entry.Env != "" {
			if err := vault.ValidateEnvName(entry.Env); err != nil {
				return fmt.Errorf("invalid secrets[%d].env: %s", i, entry.Env)
			}
		}
	}
	return nil
}

func validateV2Manifest(manifest Manifest) error {
	if len(manifest.Secrets) > 0 {
		return fmt.Errorf("project manifest secrets require version %d", CurrentVersion)
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
		if err := vault.ValidateEnvName(binding.Env); err != nil {
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
