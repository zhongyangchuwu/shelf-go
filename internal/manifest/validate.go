package manifest

import (
	"fmt"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/store"
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
	for i, entry := range manifest.Secrets {
		hasPath := entry.Path != ""
		hasPrefix := entry.Prefix != ""
		if !hasPath && !hasPrefix {
			return fmt.Errorf("secrets[%d]: path or prefix is required", i)
		}
		if hasPath && hasPrefix {
			return fmt.Errorf("secrets[%d]: path and prefix are mutually exclusive", i)
		}
		if hasPath {
			if err := store.ValidatePath(entry.Path); err != nil {
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
				if !store.IsPathToken(segment) {
					return fmt.Errorf("invalid secrets[%d].prefix: unsupported characters in segment %q", i, segment)
				}
			}
			if _, ok := seenPrefix[entry.Prefix]; ok {
				return fmt.Errorf("duplicate secrets entry prefix: %s", entry.Prefix)
			}
			seenPrefix[entry.Prefix] = struct{}{}
		}
		if entry.Env != "" && hasPrefix {
			return fmt.Errorf("secrets[%d]: prefix entries must not carry env", i)
		}
		if entry.Env != "" {
			if err := store.ValidateEnvName(entry.Env); err != nil {
				return fmt.Errorf("invalid secrets[%d].env: %s", i, entry.Env)
			}
		}
	}
	return nil
}
