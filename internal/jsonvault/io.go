package jsonvault

import (
	"bytes"
	"errors"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func Load(path string) (*vault.Store, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &vault.Store{Data: vault.NewData()}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return &vault.Store{Data: vault.NewData()}, nil
	}
	data, err := decodeStore(content)
	if err != nil {
		return nil, err
	}
	return &vault.Store{Data: data}, nil
}

func Save(path string, st *vault.Store) error {
	plain, err := encodeStore(st.Data)
	if err != nil {
		return err
	}
	return writeStoreFile(path, plain)
}

func writeStoreFile(path string, content []byte) error {
	return util.AtomicWrite(path, content, util.AtomicWriteOptions{FileMode: 0o600, DirMode: 0o700, Sync: true, Backup: true})
}
