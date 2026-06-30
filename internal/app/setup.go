package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhongyangchuwu/shelf-go/internal/config"
	"github.com/zhongyangchuwu/shelf-go/internal/util"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type InitConfig struct {
	VaultPath     string
	Recipients    []string
	IdentityPaths []string
}

func ResolveInitConfigPath(flag string) (string, error) {
	if flag != "" {
		return filepath.Abs(flag)
	}
	if env := os.Getenv("SHELF_CONFIG"); env != "" {
		return filepath.Abs(env)
	}
	return ExpandInitPath(config.DefaultConfigPath)
}

func (a *App) EnsureInitIdentity(path string) (vault.Identity, error) {
	return a.vaults.ReadOrCreateIdentity(path)
}

func (a *App) EnsureVaultForRuntime(runtime Runtime) (bool, error) {
	return a.EnsureVaultFile(a.vaultOptions(runtime))
}

func (a *App) EnsureVaultFile(options vault.Options) (bool, error) {
	vaultPath := options.Path
	if _, err := os.Stat(vaultPath); err == nil {
		return false, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	repo, err := a.vaults.Open(options)
	if err != nil {
		return false, err
	}
	if err := repo.Save(&vault.Store{Data: vault.NewData()}); err != nil {
		return false, err
	}
	return true, nil
}

func EnsureConfigFile(configPath string, cfg InitConfig, force bool) (bool, error) {
	if _, err := os.Stat(configPath); err == nil && !force {
		return false, nil
	}
	vaultPath, err := RelativeIfDescendant(cfg.VaultPath, filepath.Dir(configPath))
	if err != nil {
		vaultPath = cfg.VaultPath
	}
	identityPaths := make([]string, 0, len(cfg.IdentityPaths))
	for _, path := range cfg.IdentityPaths {
		rel, err := RelativeIfDescendant(path, filepath.Dir(configPath))
		if err != nil {
			rel = path
		}
		identityPaths = append(identityPaths, rel)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "version: 1\nvault_path: %s\nrecipients:\n", vaultPath)
	for _, recipient := range cfg.Recipients {
		fmt.Fprintf(&b, "  - %s\n", recipient)
	}
	b.WriteString("identity_paths:\n")
	for _, path := range identityPaths {
		fmt.Fprintf(&b, "  - %s\n", path)
	}
	return true, util.AtomicWrite(configPath, []byte(b.String()), util.AtomicWriteOptions{FileMode: 0o600, DirMode: 0o700})
}

func ExpandInitPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	path = os.ExpandEnv(path)
	return filepath.Abs(path)
}

func RelativeIfDescendant(target, base string) (string, error) {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("not descendant")
	}
	return rel, nil
}
