package manifest

import (
	"fmt"
	"regexp"

	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func Validate(manifest Manifest) error {
	if manifest.Version != CurrentVersion {
		return fmt.Errorf("unsupported project manifest version %d", manifest.Version)
	}
	if manifest.Secrets == nil {
		return fmt.Errorf("project manifest secrets array is required")
	}
	seen := make(map[string]struct{}, len(manifest.Secrets))
	for i, entry := range manifest.Secrets {
		if entry.Prefix != "" {
			return fmt.Errorf("secrets[%d].prefix is not supported in v0.2", i)
		}
		if entry.Path == "" {
			return fmt.Errorf("secrets[%d].path is required", i)
		}
		if err := store.ValidatePath(entry.Path); err != nil {
			return fmt.Errorf("invalid secrets[%d].path: %w", i, err)
		}
		if _, ok := seen[entry.Path]; ok {
			return fmt.Errorf("duplicate secrets entry path: %s", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		if entry.Env != "" && !envNamePattern.MatchString(entry.Env) {
			return fmt.Errorf("invalid secrets[%d].env: %s", i, entry.Env)
		}
	}
	return nil
}
