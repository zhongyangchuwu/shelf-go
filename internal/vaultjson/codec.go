package vaultjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type Codec interface {
	Decode([]byte) (vaultmodel.Vault, error)
	Encode(vaultmodel.Vault) ([]byte, error)
}

type JSONCodec struct{}

func (JSONCodec) Decode(content []byte) (vaultmodel.Vault, error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	var vault vaultmodel.Vault
	if err := dec.Decode(&vault); err != nil {
		return vaultmodel.Vault{}, fmt.Errorf("invalid v2 vault JSON: %w", err)
	}
	if err := rejectTrailingJSON(dec); err != nil {
		return vaultmodel.Vault{}, err
	}
	if err := vaultmodel.Validate(vault); err != nil {
		return vaultmodel.Vault{}, err
	}
	return vault, nil
}

func (JSONCodec) Encode(vault vaultmodel.Vault) ([]byte, error) {
	if err := vaultmodel.Validate(vault); err != nil {
		return nil, err
	}
	content, err := json.MarshalIndent(vault, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

func rejectTrailingJSON(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("invalid v2 vault JSON trailing data: %w", err)
	}
	return fmt.Errorf("invalid v2 vault JSON: trailing data")
}
