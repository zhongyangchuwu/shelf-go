package vault

import (
	"bytes"
	"errors"
	"os"
)

func Load(path string) (*Store, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Store{Data: NewData()}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return &Store{Data: NewData()}, nil
	}
	data, err := decodeStore(content)
	if err != nil {
		return nil, err
	}
	return &Store{Data: data}, nil
}

func Save(path string, st *Store) error {
	plain, err := encodeStore(st.Data)
	if err != nil {
		return err
	}
	return writeStoreFile(path, plain)
}

func writeStoreFile(path string, content []byte) error {
	return Write(path, content, Options{FileMode: 0o600, DirMode: 0o700, Sync: true, Backup: true})
}
