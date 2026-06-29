package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/adapters/shelfvault"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
)

type SecretService struct {
	vault *shelfvault.Vault
}

type SecretSummary struct {
	Path        string   `json:"path"`
	Env         string   `json:"env,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags"`
	ValueSet    bool     `json:"value_set"`
}

type WriteSecretRequest struct {
	OldPath     string
	Path        string
	Value       *string
	Env         string
	Description string
	Tags        []string
	Force       bool
}

func NewSecretService(vault *shelfvault.Vault) (*SecretService, error) {
	if vault == nil {
		return nil, fmt.Errorf("vault is required")
	}
	return &SecretService{vault: vault}, nil
}

func (s *SecretService) SecretInfo(path string) (SecretSummary, error) {
	var info SecretSummary
	err := s.vault.Read(func(st *shelfvault.Store) error {
		secret, ok := st.Get(path)
		if !ok {
			return fmt.Errorf("secret not found: %s", path)
		}
		info = newSecretSummary(path, secret)
		return nil
	})
	return info, err
}

func (s *SecretService) ListSecrets(query string) ([]SecretSummary, error) {
	query = strings.ToLower(query)
	var items []SecretSummary
	err := s.vault.Read(func(st *shelfvault.Store) error {
		paths := st.List("")
		items = make([]SecretSummary, 0, len(paths))
		for _, path := range paths {
			secret, ok := st.Get(path)
			if !ok || query != "" && !matchesSecretSummary(query, path, secret) {
				continue
			}
			items = append(items, newSecretSummary(path, secret))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
		return nil
	})
	return items, err
}

func (s *SecretService) RevealSecret(path string) (string, error) {
	var value string
	err := s.vault.Read(func(st *shelfvault.Store) error {
		secret, ok := st.Get(path)
		if !ok {
			return fmt.Errorf("secret not found: %s", path)
		}
		v, err := util.ValueString(secret.Value)
		if err != nil {
			return err
		}
		value = v
		return nil
	})
	return value, err
}

func (s *SecretService) WriteSecret(update bool, req WriteSecretRequest) error {
	return s.vault.Update(func(st *shelfvault.Store) error {
		secret := shelfvault.Secret{Env: req.Env, Description: req.Description, Tags: req.Tags}
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
				value, err := shelfvault.ParseValue(*req.Value)
				if err != nil {
					return err
				}
				secret.Value = value
			}
			id, err := shelfvault.ParseSecretID(req.Path)
			if err != nil {
				return err
			}
			return st.Update(oldPath, id, secret)
		}
		value, err := shelfvault.ParseValue(*req.Value)
		if err != nil {
			return err
		}
		secret.Value = value
		return st.Set(req.Path, secret, req.Force)
	})
}

func (s *SecretService) DeleteSecret(path string) error {
	return s.vault.Update(func(st *shelfvault.Store) error {
		if !st.Delete(path) {
			return fmt.Errorf("secret not found: %s", path)
		}
		return nil
	})
}

func newSecretSummary(path string, secret shelfvault.Secret) SecretSummary {
	return SecretSummary{Path: path, Env: secret.Env, Description: secret.Description, Tags: append([]string(nil), secret.Tags...), ValueSet: len(secret.Value) > 0}
}

func matchesSecretSummary(query, path string, secret shelfvault.Secret) bool {
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
