package vaultmigrate

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

const primaryFieldID = "value"

func ConvertVault(data vault.Data) (VaultResult, error) {
	if data.Version != vault.CurrentVersion {
		return VaultResult{}, fmt.Errorf("unsupported v1 vault version: %d", data.Version)
	}
	paths := make([]string, 0, len(data.Secrets))
	for path := range data.Secrets {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	result := VaultResult{
		Vault: vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Items: make([]vaultmodel.Item, 0, len(paths))},
		Paths: make(map[string]PathMapping, len(paths)),
	}
	for _, path := range paths {
		secret := data.Secrets[path]
		if err := vault.ValidatePath(path); err != nil {
			return VaultResult{}, fmt.Errorf("invalid v1 secret path %s: %w", path, err)
		}
		if err := vault.ValidateSecret(secret); err != nil {
			return VaultResult{}, fmt.Errorf("invalid v1 secret %s: %w", path, err)
		}
		itemID := itemIDForPath(path)
		resolvedEnv, err := vault.EnvName(path, secret)
		if err != nil {
			return VaultResult{}, err
		}
		fields := []vaultmodel.Field{
			{
				ID:          primaryFieldID,
				Label:       "Value",
				Type:        vaultmodel.FieldTypeConcealed,
				Sensitivity: vaultmodel.SensitivitySecret,
				Purpose:     vaultmodel.PurposePrimarySecret,
				Value:       append([]byte(nil), secret.Value...),
			},
		}
		envField := vaultmodel.Field{
			ID:          primaryFieldID + "_env",
			Label:       "Value Env",
			Type:        vaultmodel.FieldTypeString,
			Sensitivity: vaultmodel.SensitivityPublic,
			Purpose:     vaultmodel.PurposeEnvName,
		}
		if secret.Env != "" {
			envField.Value = func() json.RawMessage { v, _ := json.Marshal(secret.Env); return v }()
		}
		fields = append(fields, envField)
		item := vaultmodel.Item{
			ID:       itemID,
			Title:    path,
			Template: vaultmodel.TemplateGenericSecret,
			Fields:   fields,
			Aliases:  []string{path},
			Notes:    secret.Description,
			Labels:   append([]string(nil), secret.Tags...),
		}
		result.Vault.Items = append(result.Vault.Items, item)
		result.Paths[path] = PathMapping{OldPath: path, ItemID: itemID, FieldID: primaryFieldID, EnvName: resolvedEnv, Title: path, Labels: append([]string(nil), secret.Tags...)}
	}
	if err := vaultmodel.Validate(result.Vault); err != nil {
		return VaultResult{}, err
	}
	return result, nil
}

func itemIDForPath(path string) string {
	sum := sha256.Sum256([]byte(path))
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum[:16])
	return "item_" + strings.ToLower(encoded)
}
