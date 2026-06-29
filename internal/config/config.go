package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "~/.config/shelf/config.yaml"
	DefaultVaultPath  = "~/.local/share/shelf/vault.age"

	SourceShelfVault = "shelfvault"
	SourceGopass     = "gopass"
)

type Config struct {
	Version       int          `yaml:"version"`
	VaultPath     string       `yaml:"vault_path"`
	Recipients    []string     `yaml:"recipients"`
	IdentityPaths []string     `yaml:"identity_paths"`
	Editor        string       `yaml:"editor"`
	Source        SourceConfig `yaml:"source"`
}

type SourceConfig struct {
	Type          string `yaml:"type"`
	GopassCommand string `yaml:"gopass_command"`
}

type Runtime struct {
	ConfigPath    string
	VaultPath     string
	Recipients    []string
	IdentityPaths []string
	Editor        string
	Source        SourceRuntime
}

type SourceRuntime struct {
	Type          string
	GopassCommand string
}

func Resolve(configPathFlag, vaultPathFlag string) (Runtime, error) {
	configPath, err := resolveConfigPath(configPathFlag)
	if err != nil {
		return Runtime{}, err
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return Runtime{}, err
	}

	vaultPath, err := resolveVaultPath(vaultPathFlag, cfg.VaultPath, configPath)
	if err != nil {
		return Runtime{}, err
	}
	identityPaths, err := resolvePathList(cfg.IdentityPaths, configPath)
	if err != nil {
		return Runtime{}, err
	}

	sourceConfig := resolveSource(cfg.Source)

	editor := os.ExpandEnv(cfg.Editor)
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	return Runtime{ConfigPath: configPath, VaultPath: vaultPath, Recipients: cfg.Recipients, IdentityPaths: identityPaths, Editor: editor, Source: sourceConfig}, nil
}

func resolveConfigPath(flag string) (string, error) {
	if flag != "" {
		return expandPath(flag, "")
	}
	if env := os.Getenv("SHELF_CONFIG"); env != "" {
		return expandPath(env, "")
	}
	return expandPath(DefaultConfigPath, "")
}

func resolveVaultPath(flag, configVaultPath, configPath string) (string, error) {
	if flag != "" {
		return expandPath(flag, "")
	}
	if env := os.Getenv("SHELF_VAULT"); env != "" {
		return expandPath(env, "")
	}
	if configVaultPath != "" {
		return expandPath(configVaultPath, filepath.Dir(configPath))
	}
	return expandPath(DefaultVaultPath, "")
}

func resolvePathList(paths []string, configPath string) ([]string, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	resolved := make([]string, 0, len(paths))
	for _, path := range paths {
		path, err := expandPath(path, filepath.Dir(configPath))
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, path)
	}
	return resolved, nil
}

func resolveSource(cfg SourceConfig) SourceRuntime {
	typ := strings.TrimSpace(cfg.Type)
	if typ == "" {
		typ = SourceShelfVault
	}
	return SourceRuntime{Type: typ, GopassCommand: strings.TrimSpace(cfg.GopassCommand)}
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	bytes, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if len(strings.TrimSpace(string(bytes))) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func expandPath(path, baseDir string) (string, error) {
	if path == "" {
		return "", nil
	}
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
	if !filepath.IsAbs(path) && baseDir != "" {
		path = filepath.Join(baseDir, path)
	}
	return filepath.Clean(path), nil
}
