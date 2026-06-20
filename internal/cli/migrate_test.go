package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigratePlaintextStoreToEncryptedVault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	sourcePath := filepath.Join(dir, "secrets.json")
	vaultPath := filepath.Join(dir, "vault.age")
	plaintext := []byte("{\n  \"version\": 1,\n  \"secrets\": {\n    \"app:token\": {\"value\": \"migrated-secret\", \"env\": \"TOKEN\"}\n  }\n}\n")
	if err := os.WriteFile(sourcePath, plaintext, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "migrate", "--from", sourcePath)
	if err != nil {
		t.Fatalf("migrate: %v\n%s", err, out)
	}
	if !strings.Contains(out, "plaintext source preserved") {
		t.Fatalf("migration output missing cleanup guidance:\n%s", out)
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
	got, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get migrated secret: %v\n%s", err, got)
	}
	if strings.TrimSpace(got) != "migrated-secret" {
		t.Fatalf("unexpected migrated secret output: %q", got)
	}
}

func TestMigrateRefusesExistingTargetWithoutForce(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	sourcePath := filepath.Join(dir, "secrets.json")
	vaultPath := filepath.Join(dir, "vault.age")
	plaintext := []byte("{\"version\":1,\"secrets\":{\"app:token\":{\"value\":\"new-secret\"}}}\n")
	if err := os.WriteFile(sourcePath, plaintext, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "set", "app:token", "old-secret"); err != nil {
		t.Fatalf("seed vault: %v", err)
	}
	before, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "migrate", "--from", sourcePath)
	if err == nil {
		t.Fatalf("expected migration to refuse existing target")
	}
	if !strings.Contains(out+err.Error(), "target vault already exists") {
		t.Fatalf("missing overwrite refusal:\n%s\n%v", out, err)
	}
	after, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault after refusal: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("vault changed without force")
	}
}

func TestMigrateForceCreatesEncryptedBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	sourcePath := filepath.Join(dir, "secrets.json")
	vaultPath := filepath.Join(dir, "vault.age")
	plaintext := []byte("{\"version\":1,\"secrets\":{\"app:token\":{\"value\":\"new-secret\"}}}\n")
	if err := os.WriteFile(sourcePath, plaintext, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "set", "app:token", "old-secret"); err != nil {
		t.Fatalf("seed vault: %v", err)
	}
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "migrate", "--from", sourcePath, "--force"); err != nil {
		t.Fatalf("force migrate: %v", err)
	}
	backup, err := os.ReadFile(vaultPath + ".bak")
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	for _, forbidden := range [][]byte{[]byte("old-secret"), []byte("new-secret"), []byte("app:token")} {
		if bytes.Contains(backup, forbidden) {
			t.Fatalf("backup contains plaintext %q", forbidden)
		}
	}
	got, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get migrated secret: %v\n%s", err, got)
	}
	if strings.TrimSpace(got) != "new-secret" {
		t.Fatalf("unexpected migrated secret output: %q", got)
	}
}
