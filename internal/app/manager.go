package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/exportfmt"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type ManagerService struct {
	vault *vault.Vault
}

type ManagerSecretInfo struct {
	Path        string   `json:"path"`
	Env         string   `json:"env,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags"`
	ValueSet    bool     `json:"value_set"`
}

type ManagerWriteSecretRequest struct {
	OldPath     string
	Path        string
	Value       *string
	Env         string
	Description string
	Tags        []string
	Force       bool
}

func NewManagerService(vault *vault.Vault) (*ManagerService, error) {
	if vault == nil {
		return nil, fmt.Errorf("vault is required")
	}
	return &ManagerService{vault: vault}, nil
}

func (s *ManagerService) SecretInfo(path string) (ManagerSecretInfo, error) {
	var info ManagerSecretInfo
	err := s.vault.Read(func(st *vault.Store) error {
		secret, ok := st.Get(path)
		if !ok {
			return fmt.Errorf("secret not found: %s", path)
		}
		info = newManagerSecretInfo(path, secret)
		return nil
	})
	return info, err
}

func (s *ManagerService) ListSecrets(query string) ([]ManagerSecretInfo, error) {
	query = strings.ToLower(query)
	var items []ManagerSecretInfo
	err := s.vault.Read(func(st *vault.Store) error {
		paths := st.List("")
		items = make([]ManagerSecretInfo, 0, len(paths))
		for _, path := range paths {
			secret, ok := st.Get(path)
			if !ok || query != "" && !matchesManagerSecret(query, path, secret) {
				continue
			}
			items = append(items, newManagerSecretInfo(path, secret))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
		return nil
	})
	return items, err
}

func (s *ManagerService) RevealSecret(path string) (string, error) {
	var value string
	err := s.vault.Read(func(st *vault.Store) error {
		secret, ok := st.Get(path)
		if !ok {
			return fmt.Errorf("secret not found: %s", path)
		}
		v, err := exportfmt.ValueString(secret.Value)
		if err != nil {
			return err
		}
		value = v
		return nil
	})
	return value, err
}

func (s *ManagerService) WriteSecret(update bool, req ManagerWriteSecretRequest) error {
	return s.vault.Update(func(st *vault.Store) error {
		secret := vault.Secret{Env: req.Env, Description: req.Description, Tags: req.Tags}
		if update {
			oldPath := req.OldPath
			if oldPath == "" {
				oldPath = req.Path
			}
			existing, ok := st.Get(oldPath)
			if !ok {
				return fmt.Errorf("secret not found: %s", oldPath)
			}
			secret.Value = existing.Value
			if req.Value != nil {
				value, err := vault.ParseValue(*req.Value)
				if err != nil {
					return err
				}
				secret.Value = value
			}
			id, err := vault.ParseSecretID(req.Path)
			if err != nil {
				return err
			}
			return st.Update(oldPath, id, secret)
		}
		value, err := vault.ParseValue(*req.Value)
		if err != nil {
			return err
		}
		secret.Value = value
		return st.Set(req.Path, secret, req.Force)
	})
}

func (s *ManagerService) DeleteSecret(path string) error {
	return s.vault.Update(func(st *vault.Store) error {
		if !st.Delete(path) {
			return fmt.Errorf("secret not found: %s", path)
		}
		return nil
	})
}

func newManagerSecretInfo(path string, secret vault.Secret) ManagerSecretInfo {
	return ManagerSecretInfo{Path: path, Env: secret.Env, Description: secret.Description, Tags: append([]string(nil), secret.Tags...), ValueSet: len(secret.Value) > 0}
}

func matchesManagerSecret(query, path string, secret vault.Secret) bool {
	if strings.Contains(strings.ToLower(path), query) || strings.Contains(strings.ToLower(secret.Env), query) || strings.Contains(strings.ToLower(secret.Description), query) {
		return true
	}
	for _, tag := range secret.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
