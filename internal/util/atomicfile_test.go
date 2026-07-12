package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteRepairsExistingBackupPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	if err := os.WriteFile(path, []byte("previous vault"), 0o600); err != nil {
		t.Fatalf("write vault: %v", err)
	}
	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, []byte("stale backup"), 0o644); err != nil {
		t.Fatalf("write backup: %v", err)
	}
	if err := os.Chmod(backupPath, 0o644); err != nil {
		t.Fatalf("chmod backup: %v", err)
	}

	err := AtomicWrite(path, []byte("updated vault"), AtomicWriteOptions{
		FileMode: 0o600,
		DirMode:  0o700,
		Sync:     true,
		Backup:   true,
	})
	if err != nil {
		t.Fatalf("atomic write: %v", err)
	}

	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if got, want := string(backup), "previous vault"; got != want {
		t.Fatalf("backup = %q, want %q", got, want)
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("backup mode = %03o, want %03o", got, want)
	}
}
