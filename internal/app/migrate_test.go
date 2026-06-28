package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestMigratePlaintextStorePreservesSourceAndEncryptsTarget(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "secrets.json")
	vaultPath := filepath.Join(dir, "vault.age")
	plaintext := []byte("{\"version\":1,\"secrets\":{\"app:token\":{\"value\":\"migrated-secret\",\"env\":\"TOKEN\"}}}\n")
	if err := os.WriteFile(sourcePath, plaintext, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	identity, err := EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	targetVault, err := vault.NewVault(vaultPath, vault.VaultOptions{Recipients: []string{identity.Recipient().String()}, IdentityPaths: []string{filepath.Join(dir, "identity.txt")}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	if err := MigratePlaintextStore(sourcePath, targetVault, false); err != nil {
		t.Fatalf("migrate plaintext store: %v", err)
	}
	after, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if !bytes.Equal(after, plaintext) {
		t.Fatalf("source changed during migration")
	}
	vaultBytes, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if bytes.Contains(vaultBytes, []byte("migrated-secret")) || bytes.Contains(vaultBytes, []byte("app:token")) {
		t.Fatalf("vault contains plaintext source data")
	}
}

func TestMigratePlaintextStoreRefusesExistingTargetWithoutForce(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "secrets.json")
	vaultPath := filepath.Join(dir, "vault.age")
	plaintext := []byte("{\"version\":1,\"secrets\":{\"app:token\":{\"value\":\"new-secret\"}}}\n")
	if err := os.WriteFile(sourcePath, plaintext, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	identity, err := EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	targetVault, err := vault.NewVault(vaultPath, vault.VaultOptions{Recipients: []string{identity.Recipient().String()}, IdentityPaths: []string{filepath.Join(dir, "identity.txt")}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	if err := targetVault.Save(&vault.Store{Data: vault.NewData()}); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	err = MigratePlaintextStore(sourcePath, targetVault, false)
	if err == nil || !strings.Contains(err.Error(), "target vault already exists") {
		t.Fatalf("expected existing target error, got %v", err)
	}
}
