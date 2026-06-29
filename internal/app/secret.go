package app

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/zhongyangchuwu/shelf-go/internal/adapters/shelfvault"
	secretsvc "github.com/zhongyangchuwu/shelf-go/internal/secret"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
)

type AddSecretRequest struct {
	Args         []string
	In           io.Reader
	Out          io.Writer
	ReadPassword func() ([]byte, error)
}

func AddSecret(configPathFlag, vaultPathFlag string, req AddSecretRequest) (string, error) {
	var path string
	err := UpdateVault(configPathFlag, vaultPathFlag, func(st *shelfvault.Store) error {
		addedPath, err := secretsvc.Add(st, secretsvc.AddRequest{Args: req.Args, In: req.In, Out: req.Out, ReadPassword: req.ReadPassword})
		if err != nil {
			return err
		}
		path = addedPath
		return nil
	})
	return path, err
}

type SetSecretRequest struct {
	Path        string
	Value       string
	Env         string
	Description string
	Tags        []string
	Force       bool
}

func SetSecretInStore(st *shelfvault.Store, req SetSecretRequest) error {
	value, err := shelfvault.ParseValue(req.Value)
	if err != nil {
		return err
	}
	secret := shelfvault.Secret{Value: value, Env: req.Env, Description: req.Description, Tags: req.Tags}
	return st.Set(req.Path, secret, req.Force)
}

func SetSecret(configPathFlag, vaultPathFlag string, req SetSecretRequest) error {
	return UpdateVault(configPathFlag, vaultPathFlag, func(st *shelfvault.Store) error {
		return SetSecretInStore(st, req)
	})
}

func GetSecretValueFromStore(st *shelfvault.Store, path string) (string, error) {
	secret, ok := st.Get(path)
	if !ok {
		return "", fmt.Errorf("secret not found: %s", path)
	}
	return util.ValueString(secret.Value)
}

func GetSecretValue(configPathFlag, vaultPathFlag, path string) (string, error) {
	_, st, err := LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	return GetSecretValueFromStore(st, path)
}

type ListSecretsRequest struct {
	Prefix string
	Tags   []string
}

func ListSecretPathsInStore(st *shelfvault.Store, req ListSecretsRequest) []string {
	return st.ListByTags(req.Prefix, req.Tags)
}

func ListSecretPaths(configPathFlag, vaultPathFlag string, req ListSecretsRequest) ([]string, error) {
	_, st, err := LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	return ListSecretPathsInStore(st, req), nil
}

func AllSecretPaths(configPathFlag, vaultPathFlag string) ([]string, error) {
	_, st, err := LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	return st.List(""), nil
}

func SecretPaths(configPathFlag, vaultPathFlag, prefix string) ([]string, error) {
	_, st, err := LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	return st.List(prefix), nil
}
func SecretInfoJSONFromStore(st *shelfvault.Store, path string) (string, error) {
	info, ok := st.Info(path)
	if !ok {
		return "", fmt.Errorf("secret not found: %s", path)
	}
	bytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func SecretInfoJSON(configPathFlag, vaultPathFlag, path string) (string, error) {
	_, st, err := LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	return SecretInfoJSONFromStore(st, path)
}

func RemoveSecretFromStore(st *shelfvault.Store, path string) error {
	if !st.Delete(path) {
		return fmt.Errorf("secret not found: %s", path)
	}
	return nil
}

func RemoveSecret(configPathFlag, vaultPathFlag, path string) error {
	return UpdateVault(configPathFlag, vaultPathFlag, func(st *shelfvault.Store) error {
		return RemoveSecretFromStore(st, path)
	})
}

type EditSecretRequest struct {
	Path   string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func EditSecret(configPathFlag, vaultPathFlag string, req EditSecretRequest) error {
	runtime, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return v.Update(func(st *shelfvault.Store) error {
		return secretsvc.Edit(st, secretsvc.EditRequest{Path: req.Path, Editor: runtime.Editor, Stdin: req.Stdin, Stdout: req.Stdout, Stderr: req.Stderr})
	})
}
