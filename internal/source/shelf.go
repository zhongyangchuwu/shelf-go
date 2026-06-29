package source

import "github.com/zhongyangchuwu/shelf-go/internal/vault"

type VaultReader struct {
	Store *vault.Store
}

func NewVaultReader(st *vault.Store) VaultReader {
	return VaultReader{Store: st}
}

func (r VaultReader) Get(path string) (Secret, error) {
	if r.Store == nil {
		return Secret{}, ErrNotFound
	}
	secret, ok := r.Store.Get(path)
	if !ok {
		return Secret{}, ErrNotFound
	}
	value, err := ValueString(secret.Value)
	if err != nil {
		return Secret{}, err
	}
	return Secret{Path: path, Value: value, Env: secret.Env, Description: secret.Description, Tags: append([]string(nil), secret.Tags...)}, nil
}

func (r VaultReader) List(prefix string) ([]string, error) {
	if r.Store == nil {
		return nil, ErrNotFound
	}
	return r.Store.List(prefix), nil
}

func (r VaultReader) ListByTags(prefix string, tags []string) ([]string, error) {
	if r.Store == nil {
		return nil, ErrNotFound
	}
	return r.Store.ListByTags(prefix, tags), nil
}
