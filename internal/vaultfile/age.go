package vaultfile

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultcrypto"
)

func openVault(content []byte, identityPaths []string) (vault.Data, error) {
	if bytes.HasPrefix(bytes.TrimSpace(content), []byte("{")) {
		return vault.Data{}, errors.New("active vault file is plaintext JSON; run migration before using encrypted vault mode")
	}
	if !bytes.HasPrefix(content, []byte(vaultHeader)) {
		if bytes.HasPrefix(content, []byte("shelf-vault/")) {
			line, _, _ := bytes.Cut(content, []byte("\n"))
			return vault.Data{}, fmt.Errorf("unsupported vault format %q", string(line))
		}
		return vault.Data{}, errors.New("unsupported vault format: missing shelf-vault/v1 header")
	}
	plain, err := vaultcrypto.DecryptAge(content[len(vaultHeader):], identityPaths)
	if err != nil {
		return vault.Data{}, err
	}
	data, err := decodeStore(plain)
	if err != nil {
		return vault.Data{}, fmt.Errorf("invalid decrypted store: %w", err)
	}
	return data, nil
}

func sealVault(plain []byte, recipients []string) ([]byte, error) {
	ciphertext, err := vaultcrypto.EncryptAge(plain, recipients)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(vaultHeader)+len(ciphertext))
	out = append(out, vaultHeader...)
	out = append(out, ciphertext...)
	return out, nil
}
