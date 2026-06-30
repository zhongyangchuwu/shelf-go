package app

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
)

func MigratePlaintextStoreForRuntime(configPathFlag, vaultPathFlag, sourcePath, targetPath string, force bool) (string, error) {
	runtime, configuredVault, err := LoadVault(configPathFlag, vaultPathFlag)
	if err != nil {
		return "", err
	}
	targetVault := configuredVault
	if targetPath != "" && targetPath != runtime.VaultPath {
		targetVault, err = jsonvault.NewVault(targetPath, jsonvault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
		if err != nil {
			return "", err
		}
	}
	if err := MigratePlaintextStore(sourcePath, targetVault, force); err != nil {
		return "", err
	}
	return targetVault.Path(), nil
}

func MigratePlaintextStore(sourcePath string, targetVault *jsonvault.Vault, force bool) error {
	before, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read plaintext source: %w", err)
	}
	st, err := jsonvault.Load(sourcePath)
	if err != nil {
		return fmt.Errorf("load plaintext source: %w", err)
	}
	format, err := jsonvault.DetectFileFormat(targetVault.Path())
	if err != nil {
		return fmt.Errorf("inspect target vault: %w", err)
	}
	if format != jsonvault.FileFormatMissing && format != jsonvault.FileFormatEmpty && !force {
		return fmt.Errorf("target vault already exists; pass --force to replace %s", targetVault.Path())
	}
	if format == jsonvault.FileFormatPlaintextStore {
		return fmt.Errorf("target vault is plaintext JSON; choose a different --to path or move it before migration")
	}
	if err := targetVault.Save(st); err != nil {
		return fmt.Errorf("write encrypted vault: %w", err)
	}
	verified, err := targetVault.Load()
	if err != nil {
		return fmt.Errorf("verify encrypted vault: %w", err)
	}
	if len(verified.Data.Secrets) != len(st.Data.Secrets) {
		return fmt.Errorf("verify encrypted vault: secret count mismatch")
	}
	after, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("re-read plaintext source: %w", err)
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("plaintext source changed during migration")
	}
	return nil
}
