package app

import "github.com/zhongyangchuwu/shelf-go/internal/manager"

func (a *App) OpenManager(configPathFlag, vaultPathFlag, addr string) (*manager.Runtime, error) {
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

func managerSecretInfo(info SecretSummary) manager.SecretInfo {
	return manager.SecretInfo{
		Path:        info.Path,
		Env:         info.Env,
		Description: info.Description,
		Tags:        info.Tags,
		ValueSet:    info.ValueSet,
	}
}
