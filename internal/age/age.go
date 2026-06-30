package age

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	filippoage "filippo.io/age"
)

type Identity struct {
	value     string
	recipient string
}

func (i Identity) String() string {
	return i.value
}

func (i Identity) Recipient() string {
	return i.recipient
}

func ReadOrCreateIdentity(path string) (Identity, error) {
	if content, err := os.ReadFile(path); err == nil {
		identities, err := filippoage.ParseIdentities(strings.NewReader(string(content)))
		if err != nil {
			return Identity{}, fmt.Errorf("parse age identity %s: %w", path, err)
		}
		for _, identity := range identities {
			if x25519, ok := identity.(*filippoage.X25519Identity); ok {
				return Identity{value: x25519.String(), recipient: x25519.Recipient().String()}, nil
			}
		}
		return Identity{}, fmt.Errorf("age identity %s contains no X25519 identity", path)
	} else if !os.IsNotExist(err) {
		return Identity{}, fmt.Errorf("read age identity %s: %w", path, err)
	}
	identity, err := filippoage.GenerateX25519Identity()
	if err != nil {
		return Identity{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return Identity{}, err
	}
	if err := os.WriteFile(path, []byte(identity.String()+"\n"), 0o600); err != nil {
		return Identity{}, err
	}
	return Identity{value: identity.String(), recipient: identity.Recipient().String()}, nil
}

func Decrypt(content []byte, identityPaths []string) ([]byte, error) {
	if len(identityPaths) == 0 {
		return nil, errors.New("no age identity paths configured for encrypted vault")
	}
	identities, err := loadIdentities(identityPaths)
	if err != nil {
		return nil, err
	}
	reader, err := filippoage.Decrypt(bytes.NewReader(content), identities...)
	if err != nil {
		var noMatch *filippoage.NoIdentityMatchError
		if errors.As(err, &noMatch) {
			return nil, errors.New("could not decrypt vault: no configured age identity matched")
		}
		return nil, fmt.Errorf("could not decrypt vault: %w", err)
	}
	plain, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read decrypted vault: %w", err)
	}
	return plain, nil
}

func Encrypt(plain []byte, recipients []string) ([]byte, error) {
	if len(recipients) == 0 {
		return nil, errors.New("no age recipients configured for encrypted vault")
	}
	parsed := make([]filippoage.Recipient, 0, len(recipients))
	for _, value := range recipients {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, errors.New("empty age recipient configured")
		}
		recipient, err := filippoage.ParseX25519Recipient(value)
		if err != nil {
			return nil, fmt.Errorf("invalid age recipient %q: %w", value, err)
		}
		parsed = append(parsed, recipient)
	}
	var out bytes.Buffer
	writer, err := filippoage.Encrypt(&out, parsed...)
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

func loadIdentities(paths []string) ([]filippoage.Identity, error) {
	identities := make([]filippoage.Identity, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			return nil, errors.New("empty age identity path configured")
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("read age identity %s: %w", path, err)
		}
		parsed, parseErr := filippoage.ParseIdentities(file)
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
