package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/manager"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
)

func (a *App) OpenManager(configPathFlag, vaultPathFlag, addr string) (*manager.Runtime, error) {
	if err := a.ensureManagerSupportsResolvedVault(configPathFlag, vaultPathFlag); err != nil {
		return nil, err
	}
	_, vaultHandle, err := a.LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return nil, err
	}
	service, err := NewSecretService(vaultHandle)
	if err != nil {
		return nil, err
	}
	return manager.Open(manager.ServiceFuncs{
		SecretInfoFunc: func(path string) (manager.SecretInfo, error) {
			info, err := service.SecretInfo(path)
			if err != nil {
				return manager.SecretInfo{}, err
			}
			return managerSecretInfo(info), nil
		},
		ListSecretsFunc: func(query string) ([]manager.SecretInfo, error) {
			items, err := service.ListSecrets(query)
			if err != nil {
				return nil, err
			}
			out := make([]manager.SecretInfo, 0, len(items))
			for _, item := range items {
				out = append(out, managerSecretInfo(item))
			}
			return out, nil
		},
		RevealSecretFunc: service.RevealSecret,
		WriteSecretFunc: func(update bool, req manager.WriteSecretRequest) error {
			return service.WriteSecret(update, WriteSecretRequest{
				OldPath:     req.OldPath,
				Path:        req.Path,
				Value:       req.Value,
				Env:         req.Env,
				Description: req.Description,
				Tags:        req.Tags,
				Force:       req.Force,
			})
		},
		DeleteSecretFunc: service.DeleteSecret,
	}, addr)
}

func ensureManagerV2GateError(path string) error {
	return fmt.Errorf("local manager is not available for v2 vaults yet: %s; use shelf secret and shelf project commands for v2 workflows", path)
}

func (a *App) ensureManagerSupportsResolvedVault(configPathFlag, vaultPathFlag string) error {
	path, err := v2VaultPath(configPathFlag, vaultPathFlag)
	if err != nil {
		return err
	}
	v2, err := managerV2VaultExists(path)
	if err != nil {
		return err
	}
	if v2 {
		return ensureManagerV2GateError(path)
	}
	return nil
}

func managerV2VaultExists(path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("check v2 vault for manager: %w", err)
	}
	if _, err := vaultjson.NewRepository(path).Load(); err != nil {
		return false, nil
	}
	return true, nil
}

func managerSecretInfo(info SecretSummary) manager.SecretInfo {
	return manager.SecretInfo{
		Path:        info.Path,
		Env:         info.Env,
		Description: info.Description,
		Tags:        info.Tags,
		ValueSet:    info.ValueSet,
	}
}
