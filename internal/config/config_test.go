package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const vaultConfigTemplate = `version: 1
vault_path: {{VAULT}}
recipients:
  - age1example
identity_paths:
  - {{IDENTITY}}
editor: test-editor
`

func TestResolveVaultConfigTemplate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := strings.NewReplacer(
		"{{VAULT}}", "vaults/main.age",
		"{{IDENTITY}}", "keys/identity.txt",
	).Replace(vaultConfigTemplate)
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	runtime, err := Resolve(configPath, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if runtime.VaultPath != filepath.Join(dir, "vaults/main.age") {
		t.Fatalf("unexpected vault path: %s", runtime.VaultPath)
	}
	if len(runtime.IdentityPaths) != 1 || runtime.IdentityPaths[0] != filepath.Join(dir, "keys/identity.txt") {
		t.Fatalf("unexpected identity paths: %#v", runtime.IdentityPaths)
	}
	if len(runtime.Recipients) != 1 || runtime.Recipients[0] != "age1example" {
		t.Fatalf("unexpected recipients: %#v", runtime.Recipients)
	}
	if runtime.Editor != "test-editor" {
		t.Fatalf("unexpected editor: %s", runtime.Editor)
	}
	if runtime.Source.Type != SourceShelfVault {
		t.Fatalf("unexpected source type: %s", runtime.Source.Type)
	}
}

func TestResolveGopassSource(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `version: 1
source:
  type: gopass
  gopass_command: gopass-test
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	runtime, err := Resolve(configPath, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if runtime.Source.Type != SourceGopass {
		t.Fatalf("unexpected source type: %s", runtime.Source.Type)
	}
	if runtime.Source.GopassCommand != "gopass-test" {
		t.Fatalf("unexpected gopass command: %s", runtime.Source.GopassCommand)
	}
}
