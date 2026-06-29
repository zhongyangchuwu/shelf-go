package project

import (
	"fmt"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/source"
)

func Validate(manifest Manifest) error {
	if manifest.Version != CurrentVersion {
		return fmt.Errorf("unsupported project manifest version %d", manifest.Version)
	}
	if manifest.Secrets == nil {
		return fmt.Errorf("project manifest secrets array is required")
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
			if err := source.ValidatePath(entry.Path); err != nil {
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
				if !source.IsPathToken(segment) {
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
				if !source.IsPathToken(tag) {
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
			if err := source.ValidateEnvName(entry.Env); err != nil {
				return fmt.Errorf("invalid secrets[%d].env: %s", i, entry.Env)
			}
		}
	}
	return nil
}
