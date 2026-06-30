package vault

import (
	"github.com/zhongyangchuwu/shelf-go/internal/source"
)

type Reader struct {
	Store *Store
}

func NewReader(st *Store) Reader {
	return Reader{Store: st}
}

func (r Reader) Get(path string) (source.Secret, error) {
	if r.Store == nil {
		return source.Secret{}, source.ErrNotFound
	}
	secret, ok := r.Store.Get(path)
	if !ok {
		return source.Secret{}, source.ErrNotFound
	}
	value, err := source.ValueString(secret.Value)
	if err != nil {
		return source.Secret{}, err
	}
	return source.Secret{Path: path, Value: value, Env: secret.Env, Description: secret.Description, Tags: append([]string(nil), secret.Tags...)}, nil
}

func (r Reader) List(prefix string) ([]string, error) {
	if r.Store == nil {
		return nil, source.ErrNotFound
	}
	return r.Store.List(prefix), nil
}

func (r Reader) ListByTags(prefix string, tags []string) ([]string, error) {
	if r.Store == nil {
		return nil, source.ErrNotFound
	}
	return r.Store.ListByTags(prefix, tags), nil
}
