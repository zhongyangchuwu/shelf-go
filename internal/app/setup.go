package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/zhongyangchuwu/shelf-go/internal/config"
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

func EnsureInitIdentity(path string) (*age.X25519Identity, error) {
	if bytes, err := os.ReadFile(path); err == nil {
		identities, err := age.ParseIdentities(strings.NewReader(string(bytes)))
		if err != nil {
			return nil, fmt.Errorf("parse age identity %s: %w", path, err)
		}
		for _, identity := range identities {
			if x25519, ok := identity.(*age.X25519Identity); ok {
				return x25519, nil
			}
		}
		return nil, fmt.Errorf("age identity %s contains no X25519 identity", path)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read age identity %s: %w", path, err)
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(identity.String()+"\n"), 0o600); err != nil {
		return nil, err
	}
	return identity, nil
}

func EnsureVaultForRuntime(runtime Runtime) (bool, error) {
	return EnsureVaultFile(runtime.VaultPath, vault.VaultOptions{Recipients: runtime.Recipients, IdentityPaths: runtime.IdentityPaths})
}

func EnsureVaultFile(vaultPath string, options vault.VaultOptions) (bool, error) {
	if _, err := os.Stat(vaultPath); err == nil {
		return false, nil
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	v, err := vault.NewVault(vaultPath, options)
	if err != nil {
		return false, err
	}
	if err := v.Save(&vault.Store{Data: vault.NewData()}); err != nil {
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
	return true, vault.Write(configPath, []byte(b.String()), vault.Options{FileMode: 0o600, DirMode: 0o700})
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
