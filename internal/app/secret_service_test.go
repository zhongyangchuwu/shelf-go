package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultfile"
)

func TestSecretServiceWritesListsAndReveals(t *testing.T) {
	dir := t.TempDir()
	identity, err := EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient()}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new manager service: %v", err)
	}
	value := "secret-value"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:token", Value: &value, Env: "APP_TOKEN", Description: "primary token", Tags: []string{"api"}}); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	items, err := service.ListSecrets("api")
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if len(items) != 1 || items[0].Path != "app:token" || !items[0].ValueSet {
		t.Fatalf("unexpected list result: %#v", items)
	}
	revealed, err := service.RevealSecret("app:token")
	if err != nil {
		t.Fatalf("reveal secret: %v", err)
	}
	if revealed != "secret-value" {
		t.Fatalf("revealed = %q, want secret-value", revealed)
	}
}

func newTestSecretService(vaultPath string, recipients, identityPaths []string) (*SecretService, error) {
	if err := os.MkdirAll(filepath.Dir(vaultPath), 0o700); err != nil {
		return nil, err
	}
	v, err := vaultfile.NewVault(vaultPath, vaultfile.VaultOptions{Recipients: recipients, IdentityPaths: identityPaths})
	if err != nil {
		return nil, err
	}
	return NewSecretService(v)
}
