package vault

import (
	"fmt"
	"sort"
	"strings"
)

type Store struct {
	Data Data
}

func (s *Store) Set(path string, secret Secret, force bool) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	if err := ValidateSecret(secret); err != nil {
		return err
	}
	if _, exists := s.Data.Secrets[path]; exists && !force {
		return fmt.Errorf("secret already exists: %s", path)
	}
	s.Data.Secrets[path] = secret
	return nil
}

func (s *Store) Update(oldPath string, id SecretID, secret Secret) error {
	if err := ValidatePath(oldPath); err != nil {
		return err
	}
	if err := ValidateSecretID(id); err != nil {
		return err
	}
	if err := ValidateSecret(secret); err != nil {
		return err
	}
	newPath := id.Path()
	if _, exists := s.Data.Secrets[oldPath]; !exists {
		return fmt.Errorf("secret not found: %s", oldPath)
	}
	if oldPath != newPath {
		if _, exists := s.Data.Secrets[newPath]; exists {
			return fmt.Errorf("secret already exists: %s", newPath)
		}
		delete(s.Data.Secrets, oldPath)
	}
	s.Data.Secrets[newPath] = secret
	return nil
}

func (s *Store) Get(path string) (Secret, bool) {
	secret, ok := s.Data.Secrets[path]
	return secret, ok
}

func (s *Store) List(prefix string) []string {
	paths := make([]string, 0, len(s.Data.Secrets))
	for path := range s.Data.Secrets {
		if prefix == "" || strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func (s *Store) ListByTags(prefix string, tags []string) []string {
	paths := s.List(prefix)
	if len(tags) == 0 {
		return paths
	}
	filtered := paths[:0]
	for _, path := range paths {
		secret, ok := s.Data.Secrets[path]
		if ok && HasTags(secret, tags) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func HasTags(secret Secret, tags []string) bool {
	for _, want := range tags {
		found := false
		for _, got := range secret.Tags {
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

func (s *Store) Delete(path string) bool {
	_, existed := s.Data.Secrets[path]
	delete(s.Data.Secrets, path)
	return existed
}

func (s *Store) Info(path string) (Info, bool) {
	secret, ok := s.Get(path)
	if !ok {
		return Info{}, false
	}
	id, err := ParseSecretID(path)
	if err != nil {
		return Info{}, false
	}
	tags := secret.Tags
	if tags == nil {
		tags = []string{}
	}
	return Info{Path: path, GroupPath: id.GroupPath, Key: id.Key, ValueSet: len(secret.Value) > 0, Env: secret.Env, Description: secret.Description, Tags: tags}, true
}
