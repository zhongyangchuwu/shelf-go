package source

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	nonEnvChar     = regexp.MustCompile(`[^A-Za-z0-9]+`)
	envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

type Secret struct {
	Path        string
	Value       string
	Env         string
	Description string
	Tags        []string
}

var ErrNotFound = errors.New("secret not found")

type Reader interface {
	Get(path string) (Secret, error)
	List(prefix string) ([]string, error)
	ListByTags(prefix string, tags []string) ([]string, error)
}

func EnvName(path string, secret Secret) (string, error) {
	if secret.Env != "" {
		return secret.Env, nil
	}
	name := nonEnvChar.ReplaceAllString(path, "_")
	name = strings.Trim(name, "_")
	name = strings.ToUpper(name)
	if !IsEnvName(name) {
		return "", fmt.Errorf("derived env name for %s is invalid: %s", path, name)
	}
	return name, nil
}

func IsEnvName(name string) bool {
	return envNamePattern.MatchString(name)
}

func ValidateEnvName(name string) error {
	if !IsEnvName(name) {
		return fmt.Errorf("invalid env name: %s", name)
	}
	return nil
}

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
		if !IsPathToken(segment) {
			return fmt.Errorf("secret group path segment contains unsupported characters: %s", segment)
		}
	}
	if !IsPathToken(id.Key) {
		return fmt.Errorf("secret key contains unsupported characters: %s", id.Key)
	}
	return nil
}

func IsPathToken(s string) bool {
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
func ValueString(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", err
	}
	if compact.String() == "null" {
		return "", nil
	}
	return compact.String(), nil
}

type MemoryReader map[string]Secret

func (r MemoryReader) Get(path string) (Secret, error) {
	secret, ok := r[path]
	if !ok {
		return Secret{}, ErrNotFound
	}
	secret.Path = path
	secret.Tags = append([]string(nil), secret.Tags...)
	return secret, nil
}

func (r MemoryReader) List(prefix string) ([]string, error) {
	paths := make([]string, 0)
	for path := range r {
		if prefix == "" || path == prefix || strings.HasPrefix(path, prefix+"/") || strings.HasPrefix(path, prefix+":") {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func (r MemoryReader) ListByTags(prefix string, tags []string) ([]string, error) {
	paths := make([]string, 0)
	for path, secret := range r {
		if prefix != "" && path != prefix && !strings.HasPrefix(path, prefix+"/") && !strings.HasPrefix(path, prefix+":") {
			continue
		}
		if HasTags(secret, tags) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func HasTags(secret Secret, tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	owned := make(map[string]struct{}, len(secret.Tags))
	for _, tag := range secret.Tags {
		owned[tag] = struct{}{}
	}
	for _, tag := range tags {
		if _, ok := owned[tag]; !ok {
			return false
		}
	}
	return true
}
