package store

import (
	"bytes"
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
	Data Data
}

func Load(path string) (*Store, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Store{Data: NewData()}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return &Store{Data: NewData()}, nil
	}
	data, err := decodeStore(content)
	if err != nil {
		return nil, err
	}
	return &Store{Data: data}, nil
}

func Save(path string, st *Store) error {
	plain, err := encodeStore(st.Data)
	if err != nil {
		return err
	}
	return writeStoreFile(path, plain)
}

func encodeStore(data Data) ([]byte, error) {
	if data.Version == 0 {
		data.Version = CurrentVersion
	}
	if data.Secrets == nil {
		data.Secrets = map[string]Secret{}
	}
	if err := validateData(data); err != nil {
		return nil, err
	}
	plain, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(plain, '\n'), nil
}

func decodeStore(content []byte) (Data, error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	var data Data
	if err := dec.Decode(&data); err != nil {
		return Data{}, fmt.Errorf("invalid store JSON: %w", err)
	}
	if data.Version == 0 {
		data.Version = CurrentVersion
	}
	if data.Version != CurrentVersion {
		return Data{}, fmt.Errorf("unsupported store version %d", data.Version)
	}
	if data.Secrets == nil {
		data.Secrets = map[string]Secret{}
	}
	if err := validateData(data); err != nil {
		return Data{}, err
	}
	return data, nil
}

func validateData(data Data) error {
	for path, secret := range data.Secrets {
		if err := ValidatePath(path); err != nil {
			return err
		}
		if err := ValidateSecret(secret); err != nil {
			return fmt.Errorf("invalid secret %s: %w", path, err)
		}
	}
	return nil
}


func writeStoreFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
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
	return os.Rename(tmpName, path)
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
