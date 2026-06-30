package app

import (
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/source"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultfile"
)

type Runtime = config.Runtime

func ResolveRuntime(configPathFlag, vaultPathFlag string) (Runtime, error) {
	return config.Resolve(configPathFlag, vaultPathFlag)
}

func DefaultVaultPath() string {
	return config.DefaultVaultPath
}

func LoadVault(configPathFlag, vaultPathFlag string) (Runtime, *vaultfile.Vault, error) {
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	v, err := vaultfile.NewVault(runtime.VaultPath, vaultfile.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, v, nil
}

func LoadRuntime(configPathFlag, vaultPathFlag string) (Runtime, *vault.Store, error) {
	runtime, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	st, err := v.Load()
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, st, nil
}

func LoadSecretReader(configPathFlag, vaultPathFlag string) (source.Reader, error) {
	_, st, err := LoadRuntime(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	return vault.NewReader(st), nil
}

func UpdateVault(configPathFlag, vaultPathFlag string, fn func(*vault.Store) error) error {
	_, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return v.Update(fn)
}

func ReadVault(configPathFlag, vaultPathFlag string, fn func(*vault.Store) error) error {
	_, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return v.Read(fn)
}
