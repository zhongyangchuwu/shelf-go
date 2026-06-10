package store

import (
	"fmt"
	"strings"
)

type SecretID struct {
	GroupPath string
	Key       string
}

func ParseSecretID(path string) (SecretID, error) {
	if strings.Count(path, ":") != 1 {
		return SecretID{}, fmt.Errorf("secret path must contain exactly one colon: %s", path)
	}
	parts := strings.SplitN(path, ":", 2)
	id := SecretID{GroupPath: parts[0], Key: parts[1]}
	if err := ValidateSecretID(id); err != nil {
		return SecretID{}, err
	}
	return id, nil
}

func (id SecretID) Path() string {
	return id.GroupPath + ":" + id.Key
}

func ValidatePath(path string) error {
	_, err := ParseSecretID(path)
	return err
}

func ValidateSecretID(id SecretID) error {
	if id.GroupPath == "" {
		return fmt.Errorf("secret group path is empty")
	}
	if id.Key == "" {
		return fmt.Errorf("secret key is empty")
	}
	if strings.Contains(id.Key, "/") {
		return fmt.Errorf("secret key must not contain '/': %s", id.Key)
	}
	if strings.Contains(id.GroupPath, ":") || strings.Contains(id.Key, ":") {
		return fmt.Errorf("secret id parts must not contain ':'")
	}
	for _, segment := range strings.Split(id.GroupPath, "/") {
		if segment == "" {
			return fmt.Errorf("secret group path contains an empty segment: %s", id.GroupPath)
		}
		if !isPathToken(segment) {
			return fmt.Errorf("secret group path segment contains unsupported characters: %s", segment)
		}
	}
	if !isPathToken(id.Key) {
		return fmt.Errorf("secret key contains unsupported characters: %s", id.Key)
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
