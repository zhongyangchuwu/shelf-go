package app

import (
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type Runtime = config.Runtime

func ResolveRuntime(configPathFlag, vaultPathFlag string) (Runtime, error) {
	return config.Resolve(configPathFlag, vaultPathFlag)
}

func DefaultVaultPath() string {
	return config.DefaultVaultPath
}

func (a *App) LoadVault(configPathFlag, vaultPathFlag string) (Runtime, vault.Repository, error) {
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	repo, err := a.vaults.Open(a.vaultOptions(runtime))
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, repo, nil
}

func (a *App) LoadRuntime(configPathFlag, vaultPathFlag string) (Runtime, *vault.Store, error) {
	runtime, repo, err := a.LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	st, err := repo.Load()
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, st, nil
}

func (a *App) UpdateVault(configPathFlag, vaultPathFlag string, fn func(*vault.Store) error) error {
	_, repo, err := a.LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return repo.Update(fn)
}

func (a *App) ReadVault(configPathFlag, vaultPathFlag string, fn func(*vault.Store) error) error {
	_, repo, err := a.LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return repo.Read(fn)
}
