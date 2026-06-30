package app

import (
	"fmt"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func (a *App) MigratePlaintextStoreForRuntime(configPathFlag, vaultPathFlag, sourcePath, targetPath string, force bool) (string, error) {
	runtime, configuredVault, err := a.LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	targetVault := configuredVault
	if targetPath != "" && targetPath != runtime.VaultPath {
		options := a.vaultOptions(runtime)
		options.Path = targetPath
		targetVault, err = a.vaults.Open(options)
		if err != nil {
			return "", err
		}
	}
	if err := MigratePlaintextStore(sourcePath, targetVault, force); err != nil {
		return "", err
	}
	return targetVault.Path(), nil
}

func MigratePlaintextStore(sourcePath string, targetVault vault.Repository, force bool) error {
	migrator, ok := targetVault.(vault.PlaintextMigrator)
	if !ok {
		return fmt.Errorf("vault implementation does not support plaintext migration")
	}
	return migrator.MigratePlaintext(sourcePath, force)
}
