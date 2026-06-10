package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Store struct {
	Path string
	Data Data
}

func Load(path string) (*Store, error) {
	bytes, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Store{Path: path, Data: NewData()}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(bytes))) == 0 {
		return &Store{Path: path, Data: NewData()}, nil
	}
	var data Data
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("invalid store JSON: %w", err)
	}
	if data.Version == 0 {
		data.Version = CurrentVersion
	}
	if data.Version != CurrentVersion {
		return nil, fmt.Errorf("unsupported store version %d", data.Version)
	}
	if data.Secrets == nil {
		data.Secrets = map[string]Secret{}
	}
	for path, secret := range data.Secrets {
		if err := ValidatePath(path); err != nil {
			return nil, err
		}
		if err := ValidateSecret(secret); err != nil {
			return nil, fmt.Errorf("invalid secret %s: %w", path, err)
		}
	}
	return &Store{Path: path, Data: data}, nil
}

func (s *Store) Save() error {
	if s.Data.Version == 0 {
		s.Data.Version = CurrentVersion
	}
	if s.Data.Secrets == nil {
		s.Data.Secrets = map[string]Secret{}
	}
	for path, secret := range s.Data.Secrets {
		if err := ValidatePath(path); err != nil {
			return err
		}
		if err := ValidateSecret(secret); err != nil {
			return fmt.Errorf("invalid secret %s: %w", path, err)
		}
	}
	bytes, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(s.Path); err == nil {
		if err := copyFile(s.Path, s.Path+".bak"); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.Path), filepath.Base(s.Path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(bytes); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.Path)
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
