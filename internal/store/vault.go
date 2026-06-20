package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
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
	if len(identityPaths) == 0 {
		return Data{}, errors.New("no age identity paths configured for encrypted vault")
	}
	identities, err := loadIdentities(identityPaths)
	if err != nil {
		return Data{}, err
	}
	reader, err := age.Decrypt(bytes.NewReader(content[len(vaultHeader):]), identities...)
	if err != nil {
		var noMatch *age.NoIdentityMatchError
		if errors.As(err, &noMatch) {
			return Data{}, errors.New("could not decrypt vault: no configured age identity matched")
		}
		return Data{}, fmt.Errorf("could not decrypt vault: %w", err)
	}
	plain, err := io.ReadAll(reader)
	if err != nil {
		return Data{}, fmt.Errorf("read decrypted vault: %w", err)
	}
	data, err := decodeStore(plain)
	if err != nil {
		return Data{}, fmt.Errorf("invalid decrypted store: %w", err)
	}
	return data, nil
}

func sealVault(plain []byte, recipients []string) ([]byte, error) {
	if len(recipients) == 0 {
		return nil, errors.New("no age recipients configured for encrypted vault")
	}
	parsed := make([]age.Recipient, 0, len(recipients))
	for _, value := range recipients {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, errors.New("empty age recipient configured")
		}
		recipient, err := age.ParseX25519Recipient(value)
		if err != nil {
			return nil, fmt.Errorf("invalid age recipient %q: %w", value, err)
		}
		parsed = append(parsed, recipient)
	}
	var out bytes.Buffer
	out.WriteString(vaultHeader)
	writer, err := age.Encrypt(&out, parsed...)
	if err != nil {
		return nil, fmt.Errorf("encrypt vault: %w", err)
	}
	if _, err := writer.Write(plain); err != nil {
		writer.Close()
		return nil, fmt.Errorf("encrypt vault: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("encrypt vault: %w", err)
	}
	return out.Bytes(), nil
}

func loadIdentities(paths []string) ([]age.Identity, error) {
	identities := make([]age.Identity, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			return nil, errors.New("empty age identity path configured")
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("read age identity %s: %w", path, err)
		}
		parsed, parseErr := age.ParseIdentities(file)
		closeErr := file.Close()
		if parseErr != nil {
			return nil, fmt.Errorf("parse age identity %s: %w", path, parseErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("read age identity %s: %w", path, closeErr)
		}
		identities = append(identities, parsed...)
	}
	if len(identities) == 0 {
		return nil, errors.New("no age identities loaded from configured paths")
	}
	return identities, nil
}
