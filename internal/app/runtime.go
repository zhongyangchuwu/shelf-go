package app

import (
	"fmt"

	gopassadapter "github.com/zhongyangchuwu/shelf-go/internal/adapters/gopass"
	"github.com/zhongyangchuwu/shelf-go/internal/adapters/shelfvault"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/source"
)

type Runtime = config.Runtime

func ResolveRuntime(configPathFlag, vaultPathFlag string) (Runtime, error) {
	return config.Resolve(configPathFlag, vaultPathFlag)
}

func DefaultVaultPath() string {
	return config.DefaultVaultPath
}

func LoadVault(configPathFlag, vaultPathFlag string) (Runtime, *shelfvault.Vault, error) {
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return Runtime{}, nil, err
	}
	v, err := shelfvault.NewVault(runtime.VaultPath, shelfvault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
	if err != nil {
		return Runtime{}, nil, err
	}
	return runtime, v, nil
}

func LoadRuntime(configPathFlag, vaultPathFlag string) (Runtime, *shelfvault.Store, error) {
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
	runtime, err := config.Resolve(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	switch runtime.Source.Type {
	case config.SourceShelfVault:
		v, err := shelfvault.NewVault(runtime.VaultPath, shelfvault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
		if err != nil {
			return nil, err
		}
		st, err := v.Load()
		if err != nil {
			return nil, err
		}
		return shelfvault.NewReader(st), nil
	case config.SourceGopass:
		return gopassadapter.NewReader(runtime.Source.GopassCommand), nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", runtime.Source.Type)
	}
}

func UpdateVault(configPathFlag, vaultPathFlag string, fn func(*shelfvault.Store) error) error {
	_, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return v.Update(fn)
}

func ReadVault(configPathFlag, vaultPathFlag string, fn func(*shelfvault.Store) error) error {
	_, v, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	return v.Read(fn)
}
