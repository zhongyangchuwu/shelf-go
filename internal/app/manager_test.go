package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestOpenManagerV2VaultSidecarGate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	v2Path := vaultPath + ".v2.json"
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	content := "version: 1\nvault_path: vault.age\nrecipients:\n  - " + identity.Recipient + "\nidentity_paths:\n  - identity.txt\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	writeV2EmptyVault(t, v2Path)
	_, err = testApp().OpenManager(configPath, "", "127.0.0.1:0")
	if err == nil {
		t.Fatal("expected gate error for v2 vault sidecar, got nil")
	}
	if !strings.Contains(err.Error(), "local manager is not available for v2 vaults yet") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), v2Path) {
		t.Fatalf("error missing v2 vault path: %v", err)
	}
}

func TestOpenManagerExplicitV2VaultPathGate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	v2VaultPath := filepath.Join(dir, "myvault.v2.json")
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	// Config vault_path is irrelevant since v2VaultPath is passed explicitly
	content := "version: 1\nvault_path: irrelevant.age\nrecipients:\n  - " + identity.Recipient + "\nidentity_paths:\n  - identity.txt\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	writeV2EmptyVault(t, v2VaultPath)
	_, err = testApp().OpenManager(configPath, v2VaultPath, "127.0.0.1:0")
	if err == nil {
		t.Fatal("expected gate error for explicit v2 vault path, got nil")
	}
	if !strings.Contains(err.Error(), "local manager is not available for v2 vaults yet") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), v2VaultPath) {
		t.Fatalf("error missing v2 vault path: %v", err)
	}
}

func TestOpenManagerSucceedsWithoutV2Vault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	content := "version: 1\nvault_path: vault.age\nrecipients:\n  - " + identity.Recipient + "\nidentity_paths:\n  - identity.txt\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	v, err := jsonvault.NewVault(vaultPath, vault.Options{Path: vaultPath, Recipients: []string{identity.Recipient}, IdentityPaths: []string{filepath.Join(dir, "identity.txt")}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	st := &vault.Store{Data: vault.NewData()}
	if err := st.Set("app:token", vault.Secret{Value: []byte(`"test"`)}, false); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if err := v.Save(st); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	// No v2 sidecar file exists → gate passes
	runtime, err := testApp().OpenManager(configPath, "", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("open manager with v1 vault: %v", err)
	}
	runtime.Close()
}
