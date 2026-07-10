package app

import "github.com/zhongyangchuwu/shelf-go/internal/config"

type Runtime = config.Runtime

func ResolveRuntime(configPathFlag, vaultPathFlag string) (Runtime, error) {
	return config.Resolve(configPathFlag, vaultPathFlag)
}

func DefaultVaultPath() string {
	return config.DefaultVaultPath
}
