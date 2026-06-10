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
	DefaultDataPath   = "~/.local/share/shelf/secrets.json"
)

type Config struct {
	Version int    `yaml:"version"`
	Data    string `yaml:"data"`
	Editor  string `yaml:"editor"`
}

type Runtime struct {
	ConfigPath string
	DataPath   string
	Editor     string
}

func Resolve(configPathFlag, dataPathFlag string) (Runtime, error) {
	configPath, err := resolveConfigPath(configPathFlag)
	if err != nil {
		return Runtime{}, err
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return Runtime{}, err
	}

	dataPath, err := resolveDataPath(dataPathFlag, cfg.Data, configPath)
	if err != nil {
		return Runtime{}, err
	}

	editor := os.ExpandEnv(cfg.Editor)
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	return Runtime{ConfigPath: configPath, DataPath: dataPath, Editor: editor}, nil
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

func resolveDataPath(flag, configData, configPath string) (string, error) {
	if flag != "" {
		return expandPath(flag, "")
	}
	if env := os.Getenv("SHELF_DATA"); env != "" {
		return expandPath(env, "")
	}
	if configData != "" {
		return expandPath(configData, filepath.Dir(configPath))
	}
	return expandPath(DefaultDataPath, "")
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
