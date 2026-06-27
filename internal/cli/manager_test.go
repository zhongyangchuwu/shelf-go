package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestListenLoopbackRejectsNonLoopback(t *testing.T) {
	listener, err := listenLoopback("192.0.2.10:0")
	if err == nil {
		listener.Close()
		t.Fatalf("expected non-loopback manager address to fail")
	}
}

func TestManagerTokenIsGenerated(t *testing.T) {
	first, err := managerToken()
	if err != nil {
		t.Fatalf("manager token: %v", err)
	}
	second, err := managerToken()
	if err != nil {
		t.Fatalf("manager token: %v", err)
	}
	if first == "" || second == "" {
		t.Fatalf("manager token should not be empty")
	}
	if first == second {
		t.Fatalf("manager token should be random")
	}
}

func TestRootIncludesManagerCommand(t *testing.T) {
	cmd := NewRootCmd()
	var foundManager bool
	var vaultCmd *cobra.Command
	for _, child := range cmd.Commands() {
		if child.Name() == "manager" {
			foundManager = true
		}
		if child.Name() == "vault" {
			vaultCmd = child
		}
	}
	if !foundManager {
		t.Fatalf("root command missing manager subcommand")
	}
	if vaultCmd == nil {
		t.Fatalf("root command missing vault subcommand")
	}
	for _, child := range vaultCmd.Commands() {
		if child.Name() == "open" {
			t.Fatalf("vault command should not include open subcommand")
		}
	}
}

func TestRootExcludesPreReleaseTopLevelCommands(t *testing.T) {
	forbidden := map[string]struct{}{
		"init":    {},
		"migrate": {},
		"export":  {},
		"run":     {},
	}
	for _, child := range NewRootCmd().Commands() {
		if _, exists := forbidden[child.Name()]; exists {
			t.Fatalf("root command should not include %s", child.Name())
		}
	}
}

func TestVaultStatusReportsEncryptedLoadableVault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "status")
	if err != nil {
		t.Fatalf("vault status: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   config", "ok   vault path", "ok   vault format (encrypted shelf-vault/v1)", "ok   vault loads"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestVaultStatusGuidesPlaintextMigration(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(vaultPath, []byte(`{"version":1,"secrets":{}}`), 0o600); err != nil {
		t.Fatalf("write plaintext vault: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "status")
	if err == nil {
		t.Fatalf("expected vault status to fail for plaintext store")
	}
	if !strings.Contains(out, "run shelf vault migrate") {
		t.Fatalf("status output missing migration guidance:\n%s", out)
	}
}

func TestVaultCheckAliasReportsStatus(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "check")
	if err != nil {
		t.Fatalf("vault check: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ok   vault loads") {
		t.Fatalf("check output missing status details:\n%s", out)
	}
}

func TestVaultStatusGuidesMissingRecipients(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if err := os.WriteFile(configPath, []byte("version: 1\nvault_path: "+vaultPath+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "vault", "status")
	if err == nil {
		t.Fatalf("expected vault status to fail without recipients")
	}
	for _, want := range []string{"fail vault recipients", "shelf vault init --force --recipient", "warn vault format"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestVaultStatusFailsMissingRecipientsEvenWhenVaultLoads(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	identityPath := filepath.Join(dir, "shelf-test-identity.txt")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := "version: 1\nvault_path: " + vaultPath + "\nidentity_paths:\n  - " + identityPath + "\n"
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("rewrite config: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "vault", "status")
	if err == nil {
		t.Fatalf("expected vault status to fail without recipients:\n%s", out)
	}
	for _, want := range []string{"fail vault recipients", "ok   vault loads"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestVaultInitRequiresForceToRepairExistingConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	identityPath := filepath.Join(dir, "identity.txt")
	identity, err := readOrCreateTestIdentity(identityPath)
	if err != nil {
		t.Fatalf("identity: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("version: 1\nvault_path: "+vaultPath+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "vault", "init", "--recipient", identity.Recipient().String(), "--identity", identityPath)
	if err == nil {
		t.Fatalf("expected vault init to require --force for existing config")
	}
	if !strings.Contains(err.Error(), "rerun with --force") {
		t.Fatalf("init error missing force guidance: %v", err)
	}
	out, err = runShelf(t, "--config", configPath, "vault", "init", "--force", "--recipient", identity.Recipient().String(), "--identity", identityPath)
	if err != nil {
		t.Fatalf("vault init --force: %v\n%s", err, out)
	}
	out, err = runShelf(t, "--config", configPath, "vault", "status")
	if err != nil {
		t.Fatalf("vault status after repair: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ok   vault recipients (1 configured)") {
		t.Fatalf("status output missing repaired recipient:\n%s", out)
	}
}

func TestVaultStatusGuidesMissingIdentity(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("version: 1\nvault_path: "+vaultPath+"\nrecipients:\n  - age1example\n"), 0o600); err != nil {
		t.Fatalf("rewrite config: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "vault", "status")
	if err == nil {
		t.Fatalf("expected vault status to fail without identity paths")
	}
	for _, want := range []string{"fail vault loads", "identity_paths", "shelf vault init --identity"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestVaultStatusGuidesUnsupportedVaultFormat(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(vaultPath, []byte("shelf-vault/v2\n"), 0o600); err != nil {
		t.Fatalf("write unsupported vault: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "status")
	if err == nil {
		t.Fatalf("expected vault status to fail for unsupported format")
	}
	for _, want := range []string{"unsupported shelf vault format", "restore a compatible encrypted backup"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestVaultStatusGuidesUndecryptableVault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "setup"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(vaultPath, []byte("shelf-vault/v1\nnot-age-ciphertext"), 0o600); err != nil {
		t.Fatalf("write corrupt vault: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "vault", "status")
	if err == nil {
		t.Fatalf("expected vault status to fail for undecryptable vault")
	}
	for _, want := range []string{"could not decrypt vault", "restore a known-good encrypted backup"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}
