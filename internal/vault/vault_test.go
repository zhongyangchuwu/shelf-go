package vault

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
)

func TestVaultSaveLoadEncryptsStoreAndBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.vault")
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	options := VaultOptions{
		Recipients:    []string{identity.Recipient().String()},
		IdentityPaths: []string{writeTestIdentity(t, identity)},
	}
	vault, err := NewVault(path, options)
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}

	st := &Store{Data: NewData()}
	if err := st.Set("app:token", Secret{Value: []byte(`"first-secret"`)}, false); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if err := vault.Save(st); err != nil {
		t.Fatalf("save vault: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if !bytes.HasPrefix(content, []byte(vaultHeader)) {
		t.Fatalf("missing vault header: %q", content[:min(len(content), len(vaultHeader))])
	}
	if bytes.Contains(content, []byte("first-secret")) || bytes.Contains(content, []byte("app:token")) {
		t.Fatalf("vault contains plaintext secret data")
	}
	loaded, err := vault.Load()
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}
	secret, ok := loaded.Get("app:token")
	if !ok || string(secret.Value) != `"first-secret"` {
		t.Fatalf("unexpected loaded secret: ok=%v value=%s", ok, secret.Value)
	}

	if err := loaded.Set("app:token", Secret{Value: []byte(`"second-secret"`)}, true); err != nil {
		t.Fatalf("replace secret: %v", err)
	}
	if err := vault.Save(loaded); err != nil {
		t.Fatalf("save replacement: %v", err)
	}
	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if !bytes.HasPrefix(backup, []byte(vaultHeader)) {
		t.Fatalf("missing backup vault header")
	}
	if bytes.Contains(backup, []byte("first-secret")) || bytes.Contains(backup, []byte("app:token")) {
		t.Fatalf("backup contains plaintext secret data")
	}
	backupPath := filepath.Join(t.TempDir(), "backup.vault")
	if err := os.WriteFile(backupPath, backup, 0o600); err != nil {
		t.Fatalf("write backup copy: %v", err)
	}
	backupVault, err := NewVault(backupPath, options)
	if err != nil {
		t.Fatalf("new backup vault: %v", err)
	}
	backupStore, err := backupVault.Load()
	if err != nil {
		t.Fatalf("load backup: %v", err)
	}
	backupSecret, ok := backupStore.Get("app:token")
	if !ok || string(backupSecret.Value) != `"first-secret"` {
		t.Fatalf("unexpected backup secret: ok=%v value=%s", ok, backupSecret.Value)
	}
}

func TestVaultModeRejectsPlaintextStore(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	path := filepath.Join(t.TempDir(), "secrets.vault")
	if err := os.WriteFile(path, []byte(`{"version":1,"secrets":{"app:token":{"value":"secret"}}}`), 0o600); err != nil {
		t.Fatalf("write plaintext: %v", err)
	}
	vault, err := NewVault(path, VaultOptions{Recipients: []string{identity.Recipient().String()}, IdentityPaths: []string{writeTestIdentity(t, identity)}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	_, err = vault.Load()
	if err == nil || !strings.Contains(err.Error(), "plaintext JSON") {
		t.Fatalf("expected plaintext rejection, got %v", err)
	}
}

func TestVaultModeReportsUnsupportedCorruptAndInvalidStore(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	identityPath := writeTestIdentity(t, identity)
	recipients := []string{identity.Recipient().String()}

	t.Run("unsupported header", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "secrets.vault")
		if err := os.WriteFile(path, []byte("shelf-vault/v2\n"), 0o600); err != nil {
			t.Fatalf("write vault: %v", err)
		}
		vault, err := NewVault(path, VaultOptions{Recipients: recipients, IdentityPaths: []string{identityPath}})
		if err != nil {
			t.Fatalf("new vault: %v", err)
		}
		_, err = vault.Load()
		if err == nil || !strings.Contains(err.Error(), "unsupported vault format") {
			t.Fatalf("expected unsupported format error, got %v", err)
		}
	})

	t.Run("corrupt ciphertext", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "secrets.vault")
		if err := os.WriteFile(path, []byte(vaultHeader+"not-age-ciphertext"), 0o600); err != nil {
			t.Fatalf("write vault: %v", err)
		}
		vault, err := NewVault(path, VaultOptions{Recipients: recipients, IdentityPaths: []string{identityPath}})
		if err != nil {
			t.Fatalf("new vault: %v", err)
		}
		_, err = vault.Load()
		if err == nil || !strings.Contains(err.Error(), "could not decrypt vault") {
			t.Fatalf("expected decrypt error, got %v", err)
		}
	})

	t.Run("invalid decrypted store", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "secrets.vault")
		content, err := sealVault([]byte("not-json"), recipients)
		if err != nil {
			t.Fatalf("seal invalid store: %v", err)
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatalf("write vault: %v", err)
		}
		vault, err := NewVault(path, VaultOptions{Recipients: recipients, IdentityPaths: []string{identityPath}})
		if err != nil {
			t.Fatalf("new vault: %v", err)
		}
		_, err = vault.Load()
		if err == nil || !strings.Contains(err.Error(), "invalid decrypted store") {
			t.Fatalf("expected invalid decrypted store error, got %v", err)
		}
	})
}

func TestVaultModeReportsMissingAndWrongIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.vault")
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	wrong, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate wrong identity: %v", err)
	}
	options := VaultOptions{Recipients: []string{identity.Recipient().String()}, IdentityPaths: []string{writeTestIdentity(t, identity)}}
	vault, err := NewVault(path, options)
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	st := &Store{Data: NewData()}
	if err := st.Set("app:token", Secret{Value: []byte(`"secret"`)}, false); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if err := vault.Save(st); err != nil {
		t.Fatalf("save vault: %v", err)
	}

	missingIdentityVault, err := NewVault(path, VaultOptions{Recipients: options.Recipients})
	if err != nil {
		t.Fatalf("new missing identity vault: %v", err)
	}
	_, err = missingIdentityVault.Load()
	if err == nil || !strings.Contains(err.Error(), "no age identity paths") {
		t.Fatalf("expected missing identity error, got %v", err)
	}
	wrongIdentityVault, err := NewVault(path, VaultOptions{Recipients: options.Recipients, IdentityPaths: []string{writeTestIdentity(t, wrong)}})
	if err != nil {
		t.Fatalf("new wrong identity vault: %v", err)
	}
	_, err = wrongIdentityVault.Load()
	if err == nil || !strings.Contains(err.Error(), "no configured age identity matched") {
		t.Fatalf("expected wrong identity error, got %v", err)
	}
}

func writeTestIdentity(t *testing.T, identity *age.X25519Identity) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "identity.txt")
	if err := os.WriteFile(path, []byte(identity.String()+"\n"), 0o600); err != nil {
		t.Fatalf("write identity: %v", err)
	}
	return path
}
