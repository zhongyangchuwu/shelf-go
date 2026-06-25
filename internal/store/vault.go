package store

import (
	"bytes"
	"errors"
	"os"
)

const vaultHeader = "shelf-vault/v1\n"

type FileFormat string

const (
	FileFormatMissing          FileFormat = "missing"
	FileFormatEmpty            FileFormat = "empty"
	FileFormatEncryptedVault   FileFormat = "encrypted-vault"
	FileFormatPlaintextStore   FileFormat = "plaintext-store"
	FileFormatUnsupportedVault FileFormat = "unsupported-vault"
	FileFormatUnsupported      FileFormat = "unsupported"
)

func DetectFileFormat(path string) (FileFormat, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return FileFormatMissing, nil
	}
	if err != nil {
		return "", err
	}
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) == 0 {
		return FileFormatEmpty, nil
	}
	if bytes.HasPrefix(content, []byte(vaultHeader)) {
		return FileFormatEncryptedVault, nil
	}
	if bytes.HasPrefix(content, []byte("shelf-vault/")) {
		return FileFormatUnsupportedVault, nil
	}
	if bytes.HasPrefix(trimmed, []byte("{")) {
		if _, err := decodeStore(content); err == nil {
			return FileFormatPlaintextStore, nil
		}
	}
	return FileFormatUnsupported, nil
}

type VaultOptions struct {
	Recipients    []string
	IdentityPaths []string
}

type Vault struct {
	path string
	opts VaultOptions
}

func NewVault(path string, opts VaultOptions) (*Vault, error) {
	if path == "" {
		return nil, errors.New("vault path is required")
	}
	return &Vault{path: path, opts: opts}, nil
}

func (v *Vault) Path() string {
	return v.path
}

func (v *Vault) Lock() (*Lock, error) {
	return LockFile(v.path)
}

func (v *Vault) Load() (*Store, error) {
	content, err := os.ReadFile(v.path)
	if errors.Is(err, os.ErrNotExist) {
		return &Store{Data: NewData()}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return &Store{Data: NewData()}, nil
	}
	data, err := openVault(content, v.opts.IdentityPaths)
	if err != nil {
		return nil, err
	}
	return &Store{Data: data}, nil
}

func (v *Vault) Save(st *Store) error {
	plain, err := encodeStore(st.Data)
	if err != nil {
		return err
	}
	content, err := sealVault(plain, v.opts.Recipients)
	if err != nil {
		return err
	}
	return writeStoreFile(v.path, content)
}

func (v *Vault) Read(fn func(*Store) error) error {
	st, err := v.Load()
	if err != nil {
		return err
	}
	return fn(st)
}

func (v *Vault) Update(fn func(*Store) error) error {
	lock, err := v.Lock()
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	st, err := v.Load()
	if err != nil {
		return err
	}
	if err := fn(st); err != nil {
		return err
	}
	return v.Save(st)
}
