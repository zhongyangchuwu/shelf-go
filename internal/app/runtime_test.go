package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestLoadRuntimeUsesLocalVault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	identity, err := EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	content := "version: 1\nvault_path: vault.age\nrecipients:\n  - " + identity.Recipient() + "\nidentity_paths:\n  - identity.txt\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	v, err := jsonvault.NewVault(vaultPath, jsonvault.VaultOptions{Recipients: []string{identity.Recipient()}, IdentityPaths: []string{filepath.Join(dir, "identity.txt")}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	st := &vault.Store{Data: vault.NewData()}
	if err := st.Set("app:token", vault.Secret{Value: []byte(`"secret"`)}, false); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if err := v.Save(st); err != nil {
		t.Fatalf("save vault: %v", err)
	}

	_, loaded, err := LoadRuntime(configPath, "")
	if err != nil {
		t.Fatalf("load runtime: %v", err)
	}
	secret, ok := loaded.Get("app:token")
	if !ok {
		t.Fatalf("secret not found")
	}
	if string(secret.Value) != `"secret"` {
		t.Fatalf("value = %s, want %q", secret.Value, `"secret"`)
	}
}
