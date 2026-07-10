package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const vaultConfigTemplate = `version: 1
vault_path: {{VAULT}}
editor: test-editor
`

func TestResolveVaultConfigTemplate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := strings.NewReplacer(
		"{{VAULT}}", "vaults/main.json",
	).Replace(vaultConfigTemplate)
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	runtime, err := Resolve(configPath, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if runtime.VaultPath != filepath.Join(dir, "vaults/main.json") {
		t.Fatalf("unexpected vault path: %s", runtime.VaultPath)
	}
	if runtime.Editor != "test-editor" {
		t.Fatalf("unexpected editor: %s", runtime.Editor)
	}
}
