package store

import (
	"fmt"
	"strings"
)

func ValidatePath(path string) error {
	if strings.Count(path, ":") != 1 {
		return fmt.Errorf("secret path must contain exactly one colon: %s", path)
	}
	parts := strings.SplitN(path, ":", 2)
	namespace, key := parts[0], parts[1]
	if namespace == "" {
		return fmt.Errorf("secret namespace is empty")
	}
	if key == "" {
		return fmt.Errorf("secret key is empty")
	}
	if strings.Contains(key, "/") {
		return fmt.Errorf("secret key must not contain '/': %s", key)
	}
	for _, segment := range strings.Split(namespace, "/") {
		if segment == "" {
			return fmt.Errorf("secret namespace contains an empty segment: %s", path)
		}
		if !isPathToken(segment) {
			return fmt.Errorf("secret namespace segment contains unsupported characters: %s", segment)
		}
	}
	if !isPathToken(key) {
		return fmt.Errorf("secret key contains unsupported characters: %s", key)
	}
	return nil
}

func isPathToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}
