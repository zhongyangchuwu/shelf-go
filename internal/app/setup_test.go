package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultfile"
)

func TestEnsureConfigFileWritesRelativeDescendantPaths(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	cfg := InitConfig{
		VaultPath:     filepath.Join(dir, "vault.age"),
		Recipients:    []string{"age1example"},
		IdentityPaths: []string{filepath.Join(dir, "identity.txt")},
	}
	created, err := EnsureConfigFile(configPath, cfg, false)
	if err != nil {
		t.Fatalf("ensure config: %v", err)
	}
	if !created {
		t.Fatalf("expected config to be created")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	for _, want := range []string{"vault_path: vault.age", "  - age1example", "  - identity.txt"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("config missing %q:\n%s", want, content)
		}
	}
}

func TestEnsureVaultFileCreatesEncryptedVaultOnce(t *testing.T) {
	identity, err := EnsureInitIdentity(filepath.Join(t.TempDir(), "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	vaultPath := filepath.Join(t.TempDir(), "vault.age")
	created, err := EnsureVaultFile(vaultPath, vaultfile.VaultOptions{Recipients: []string{identity.Recipient()}, IdentityPaths: []string{}})
	if err != nil {
		t.Fatalf("ensure vault: %v", err)
	}
	if !created {
		t.Fatalf("expected vault to be created")
	}
	created, err = EnsureVaultFile(vaultPath, vaultfile.VaultOptions{Recipients: []string{identity.Recipient()}, IdentityPaths: []string{}})
	if err != nil {
		t.Fatalf("ensure vault again: %v", err)
	}
	if created {
		t.Fatalf("expected existing vault to be preserved")
	}
}

func TestRelativeIfDescendantRejectsOutsidePath(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(t.TempDir(), "file")
	if _, err := RelativeIfDescendant(outside, base); err == nil {
		t.Fatalf("expected outside path to be rejected")
	}
}
