package vaultfile

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func encodeStore(data vault.Data) ([]byte, error) {
	if data.Version == 0 {
		data.Version = vault.CurrentVersion
	}
	if data.Secrets == nil {
		data.Secrets = map[string]vault.Secret{}
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

func decodeStore(content []byte) (vault.Data, error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	var data vault.Data
	if err := dec.Decode(&data); err != nil {
		return vault.Data{}, fmt.Errorf("invalid store JSON: %w", err)
	}
	if data.Version == 0 {
		data.Version = vault.CurrentVersion
	}
	if data.Version != vault.CurrentVersion {
		return vault.Data{}, fmt.Errorf("unsupported store version %d", data.Version)
	}
	if data.Secrets == nil {
		data.Secrets = map[string]vault.Secret{}
	}
	if err := validateData(data); err != nil {
		return vault.Data{}, err
	}
	return data, nil
}

func validateData(data vault.Data) error {
	for path, secret := range data.Secrets {
		if err := vault.ValidatePath(path); err != nil {
			return err
		}
		if err := vault.ValidateSecret(secret); err != nil {
			return fmt.Errorf("invalid secret %s: %w", path, err)
		}
	}
	return nil
}
