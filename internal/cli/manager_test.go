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

func TestRootIncludesVaultOpenCommand(t *testing.T) {
	cmd := NewRootCmd()
	var vault *cobra.Command
	for _, child := range cmd.Commands() {
		if child.Name() == "vault" {
			vault = child
		}
		if child.Name() == "manager" {
			t.Fatalf("root command should not include manager subcommand")
		}
	}
	if vault == nil {
		t.Fatalf("root command missing vault subcommand")
	}
	foundOpen := false
	for _, child := range vault.Commands() {
		if child.Name() == "open" {
			foundOpen = true
		}
	}
	if !foundOpen {
		t.Fatalf("vault command missing open subcommand")
	}
}

func TestRootExcludesPreReleaseTopLevelCommands(t *testing.T) {
	forbidden := map[string]struct{}{
		"init":    {},
		"migrate": {},
		"export":  {},
		"run":     {},
		"manager": {},
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
