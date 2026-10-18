package app

import (
	"fmt"
	"io"
	"sort"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

// SetSecretRequest names a concrete mapped field. An empty Target means create
// an unmapped environment name, never resolve a newly appearing mapping.
// Values are accepted only from In, and are never included in returned errors.
type SetSecretRequest struct {
	Env    string
	Target string
	In     io.Reader
}

// SecretSetTargets returns stable item-ID.field targets, including ambiguous
// mappings. It does not initialize or write storage.
func (a *App) SecretSetTargets(configPath, vaultPath, env string) ([]string, error) {
	if _, err := newEnvironmentSecret(env); err != nil {
		return nil, err
	}
	vault, _, err := a.loadVaultForWrite(configPath, vaultPath)
	if err != nil {
		return nil, fmt.Errorf("load vault for secret set failed")
	}
	return secretSetTargets(vault, env)
}

func secretSetTargets(vault vaultmodel.Vault, env string) ([]string, error) {
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	var targets []string
	for _, item := range vault.Items {
		for _, field := range item.Fields {
			if !resolver.IsProjectableField(item, field.ID) {
				continue
			}
			name, mapped, err := resolver.EnvName(item, field.ID)
			if err != nil {
				return nil, err
			}
			if mapped && name == env {
				targets = append(targets, item.ID+"."+field.ID)
			}
		}
	}
	sort.Strings(targets)
	return targets, nil
}

// SetSecret rechecks the chosen mapping against a fresh vault after collecting
// input. Only the selected atomic field is changed; all other data is retained.
func (a *App) SetSecret(configPath, vaultPath string, req SetSecretRequest) (string, error) {
	if req.In == nil {
		return "", fmt.Errorf("secret input is required")
	}
	vault, _, err := a.loadVaultForWrite(configPath, vaultPath)
	if err != nil {
		return "", fmt.Errorf("load vault for secret set failed")
	}
	_, original, fieldID, err := secretSetItem(vault, req.Env, req.Target)
	if err != nil {
		return "", err
	}
	originalResolver := vaultmodel.NewResolverFromValidatedVault(vault)
	_, originalField, err := originalResolver.Field(original, fieldID)
	if err != nil {
		return "", err
	}
	value, err := readFieldValue(req.In, originalField.Type)
	if err != nil {
		return "", fmt.Errorf("read secret input failed")
	}

	vault, repo, err := a.loadVaultForWrite(configPath, vaultPath)
	if err != nil {
		return "", fmt.Errorf("reload vault before setting secret failed")
	}
	index, item, fieldID, err := secretSetItem(vault, req.Env, req.Target)
	if err != nil {
		return "", err
	}
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	_, field, err := resolver.Field(item, fieldID)
	if err != nil {
		return "", err
	}
	if item.Template != original.Template || field.Type != originalField.Type || field.Sensitivity != originalField.Sensitivity {
		return "", fmt.Errorf("target schema changed; retry secret set")
	}
	if err := setFieldValueIfPresent(&item, fieldID, value); err != nil {
		return "", err
	}
	if index < 0 {
		vault.Items = append(vault.Items, item)
	} else {
		vault.Items[index] = item
	}
	// Canonical validation can include public field values in its errors. Treat
	// every value collected by this workflow as secret, even for public fields.
	if err := vaultmodel.Validate(vault); err != nil {
		return "", fmt.Errorf("input is invalid for the selected field")
	}
	if err := repo.Save(vault); err != nil {
		return "", fmt.Errorf("save secret item failed")
	}
	return item.ID, nil
}

func newEnvironmentSecret(env string) (vaultmodel.Item, error) {
	if env == "" {
		return vaultmodel.Item{}, fmt.Errorf("environment variable name is required")
	}
	item, err := buildSecretItem(CreateSecretRequest{Title: env, Env: env}, vaultmodel.NewDefaultRegistry())
	if err != nil {
		return vaultmodel.Item{}, err
	}
	// Use the same env-name validation as normal item creation and saves.
	if err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Items: []vaultmodel.Item{item}}); err != nil {
		return vaultmodel.Item{}, fmt.Errorf("invalid environment variable name")
	}
	return item, nil
}

func secretSetItem(vault vaultmodel.Vault, env, target string) (int, vaultmodel.Item, string, error) {
	created, err := newEnvironmentSecret(env)
	if err != nil {
		return 0, vaultmodel.Item{}, "", err
	}
	targets, err := secretSetTargets(vault, env)
	if err != nil {
		return 0, vaultmodel.Item{}, "", err
	}
	if target == "" {
		if len(targets) > 0 {
			return 0, vaultmodel.Item{}, "", fmt.Errorf("environment mapping changed; retry secret set and choose --target ITEM.FIELD")
		}
		// Fail closed on either ID/title cross-collision, without renaming or
		// adopting any unrelated existing item.
		for _, selector := range []string{created.ID, created.Title} {
			if err := ensureSelectorAvailable(vault.Items, -1, selector); err != nil {
				return 0, vaultmodel.Item{}, "", fmt.Errorf("cannot create %s: %w; use secret create with a distinct title and --id", env, err)
			}
		}
		return -1, created, "value", nil
	}
	matches := 0
	for _, candidate := range targets {
		if candidate == target {
			matches++
		}
	}
	if matches != 1 {
		return 0, vaultmodel.Item{}, "", fmt.Errorf("--target must uniquely identify an existing item-ID.field mapping for %s; retry secret set to list current targets", env)
	}
	// Compare concrete paths, never ResolveItem: a title may equal another ID.
	resolver := vaultmodel.NewResolverFromValidatedVault(vault)
	for index, item := range vault.Items {
		for _, field := range item.Fields {
			if item.ID+"."+field.ID != target || !resolver.IsProjectableField(item, field.ID) {
				continue
			}
			name, mapped, err := resolver.EnvName(item, field.ID)
			if err != nil {
				return 0, vaultmodel.Item{}, "", err
			}
			if mapped && name == env {
				return index, item, field.ID, nil
			}
		}
	}
	return 0, vaultmodel.Item{}, "", fmt.Errorf("target mapping changed; retry secret set")
}
