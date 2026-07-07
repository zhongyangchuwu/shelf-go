package vaultjson

import (
	"errors"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type Repository struct {
	path  string
	codec Codec
}

func NewRepository(path string) Repository {
	return Repository{path: path, codec: JSONCodec{}}
}

func NewRepositoryWithCodec(path string, codec Codec) Repository {
	if codec == nil {
		codec = JSONCodec{}
	}
	return Repository{path: path, codec: codec}
}

func (r Repository) Path() string {
	return r.path
}

func (r Repository) Load() (vaultmodel.Vault, error) {
	content, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return vaultmodel.Vault{}, err
	}
	if err != nil {
		return vaultmodel.Vault{}, err
	}
	return r.codec.Decode(content)
}

func (r Repository) Save(vault vaultmodel.Vault) error {
	content, err := r.codec.Encode(vault)
	if err != nil {
		return err
	}
	return util.AtomicWrite(r.path, content, util.AtomicWriteOptions{FileMode: 0o600, DirMode: 0o700, Sync: true, Backup: true})
}
