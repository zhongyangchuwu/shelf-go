package shelfvault

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func IsEnvName(name string) bool {
	return envNamePattern.MatchString(name)
}

func ValidateEnvName(name string) error {
	if !IsEnvName(name) {
		return fmt.Errorf("invalid env name: %s", name)
	}
	return nil
}

func ValidateSecret(secret Secret) error {
	if len(secret.Value) == 0 {
		return fmt.Errorf("secret value is required")
	}
	var value any
	if err := json.Unmarshal(secret.Value, &value); err != nil {
		return fmt.Errorf("secret value must be JSON-compatible: %w", err)
	}
	if secret.Env != "" {
		if err := ValidateEnvName(secret.Env); err != nil {
			return err
		}
	}
	seen := map[string]struct{}{}
	for _, tag := range secret.Tags {
		if tag == "" {
			return fmt.Errorf("tag must not be empty")
		}
		if !IsPathToken(tag) {
			return fmt.Errorf("tag contains unsupported characters: %s", tag)
		}
		if _, ok := seen[tag]; ok {
			return fmt.Errorf("duplicate tag: %s", tag)
		}
		seen[tag] = struct{}{}
	}
	return nil
}

func ParseValue(input string) (json.RawMessage, error) {
	var value any
	if err := json.Unmarshal([]byte(input), &value); err == nil {
		return json.RawMessage([]byte(input)), nil
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(bytes), nil
}
