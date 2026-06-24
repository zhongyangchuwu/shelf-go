package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVaultRestoreEncryptedBackup(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	backupPath := vaultPath + ".bak"

	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "set", "app:token", "first-secret"); err != nil {
		t.Fatalf("set first secret: %v", err)
	}
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "set", "app:token", "second-secret", "--force"); err != nil {
		t.Fatalf("set second secret: %v", err)
	}

	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "restore", "--from", backupPath, "--force")
	if err != nil {
		t.Fatalf("restore: %v\n%s", err, out)
	}
	if !strings.Contains(out, "restored encrypted vault") || !strings.Contains(out, "shelf vault status") {
		t.Fatalf("restore output missing guidance:\n%s", out)
	}

	got, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get restored secret: %v\n%s", err, got)
	}
	if strings.TrimSpace(got) != "first-secret" {
		t.Fatalf("unexpected restored secret: %q", got)
	}
}

func TestVaultRestoreRefusesExistingTargetWithoutForce(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	backupPath := filepath.Join(dir, "backup.age")

	if _, err := runShelf(t, "--config", configPath, "--vault", backupPath, "secret", "set", "app:token", "backup-secret"); err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "set", "app:token", "current-secret"); err != nil {
		t.Fatalf("create target: %v", err)
	}

	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "restore", "--from", backupPath)
	if err == nil {
		t.Fatalf("expected restore to refuse existing target:\n%s", out)
	}
	if !strings.Contains(err.Error(), "target vault already exists") {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}

	got, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get target secret: %v\n%s", err, got)
	}
	if strings.TrimSpace(got) != "current-secret" {
		t.Fatalf("target changed without force: %q", got)
	}
}

func TestVaultRestoreRejectsPlaintextSource(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	sourcePath := filepath.Join(dir, "secrets.json")
	if err := os.WriteFile(sourcePath, []byte(`{"version":1,"secrets":{"app:token":{"value":"plain-secret"}}}`), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "restore", "--from", sourcePath)
	if err == nil {
		t.Fatalf("expected restore to reject plaintext source:\n%s", out)
	}
	if !strings.Contains(err.Error(), "use shelf vault migrate") {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
}

func TestVaultRestoreRejectsInvalidEncryptedSource(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	sourcePath := filepath.Join(dir, "bad.age")
	if err := os.WriteFile(sourcePath, []byte("shelf-vault/v1\nnot-age-ciphertext"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "restore", "--from", sourcePath)
	if err == nil {
		t.Fatalf("expected restore to reject invalid source:\n%s", out)
	}
	if !strings.Contains(err.Error(), "load restore source") {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
}

func TestVaultRestoreRejectsPlaintextTarget(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	sourcePath := filepath.Join(dir, "backup.age")
	targetPath := filepath.Join(dir, "target.json")

	if _, err := runShelf(t, "--config", configPath, "--vault", sourcePath, "secret", "set", "app:token", "backup-secret"); err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte(`{"version":1,"secrets":{"app:token":{"value":"plain-secret"}}}`), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}

	out, err := runShelf(t, "--config", configPath, "--vault", targetPath, "vault", "restore", "--from", sourcePath, "--force")
	if err == nil {
		t.Fatalf("expected restore to reject plaintext target:\n%s", out)
	}
	if !strings.Contains(err.Error(), "target vault is plaintext JSON") {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
}
