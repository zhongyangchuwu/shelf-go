package app

import (
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

func LoadVault(configPathFlag, vaultPathFlag string) (config.Runtime, *store.Vault, error) {
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	vault, err := store.NewVault(runtime.VaultPath, store.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, vault, nil
}

func LoadRuntime(configPathFlag, vaultPathFlag string) (config.Runtime, *store.Store, error) {
	runtime, vault, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return config.Runtime{}, nil, err
	}
	st, err := vault.Load()
	if err != nil {
		return config.Runtime{}, nil, err
	}
	return runtime, st, nil
}

func UpdateVault(configPathFlag, vaultPathFlag string, fn func(*store.Store) error) error {
	_, vault, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return vault.Update(fn)
}

func ReadVault(configPathFlag, vaultPathFlag string, fn func(*store.Store) error) error {
	_, vault, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return vault.Read(fn)
}
