package shelfvault

import (
	"bytes"
	"errors"
	"fmt"

	agecrypt "github.com/zhongyangchuwu/shelf-go/internal/crypto/age"
)

func openVault(content []byte, identityPaths []string) (Data, error) {
	if bytes.HasPrefix(bytes.TrimSpace(content), []byte("{")) {
		return Data{}, errors.New("active vault file is plaintext JSON; run migration before using encrypted vault mode")
	}
	if !bytes.HasPrefix(content, []byte(vaultHeader)) {
		if bytes.HasPrefix(content, []byte("shelf-vault/")) {
			line, _, _ := bytes.Cut(content, []byte("\n"))
			return Data{}, fmt.Errorf("unsupported vault format %q", string(line))
		}
		return Data{}, errors.New("unsupported vault format: missing shelf-vault/v1 header")
	}
	plain, err := agecrypt.Decrypt(content[len(vaultHeader):], identityPaths)
	if err != nil {
		return Data{}, err
	}
	data, err := decodeStore(plain)
	if err != nil {
		return Data{}, fmt.Errorf("invalid decrypted store: %w", err)
	}
	return data, nil
}

func sealVault(plain []byte, recipients []string) ([]byte, error) {
	ciphertext, err := agecrypt.Encrypt(plain, recipients)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(vaultHeader)+len(ciphertext))
	out = append(out, vaultHeader...)
	out = append(out, ciphertext...)
	return out, nil
}
