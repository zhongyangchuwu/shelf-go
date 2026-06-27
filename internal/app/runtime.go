package app

import (
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func LoadVault(configPathFlag, vaultPathFlag string) (config.Runtime, *vault.Vault, error) {
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	v, err := vault.NewVault(runtime.VaultPath, vault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, v, nil
}

func LoadRuntime(configPathFlag, vaultPathFlag string) (config.Runtime, *vault.Store, error) {
	runtime, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	st, err := v.Load()
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, st, nil
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
