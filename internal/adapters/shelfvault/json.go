package shelfvault

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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
