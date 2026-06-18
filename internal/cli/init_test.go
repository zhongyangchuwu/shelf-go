package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesFilesAndReportsStatus(t *testing.T) {
	dir := t.TempDir()
	vault := filepath.Join(dir, "secrets.vault")
	cfg := filepath.Join(dir, "shelf.yaml")

	out, err := runShelf(t, "--config", cfg, "--vault", vault, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, want := range []string{vault, "(created)", cfg, "(created)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("init output missing %q: %s", want, out)
		}
	}

	out2, err := runShelf(t, "--config", cfg, "--vault", vault, "init")
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	for _, want := range []string{vault, "(exists)", cfg, "(exists)"} {
		if !strings.Contains(out2, want) {
			t.Fatalf("second init output missing %q: %s", want, out2)
		}
	}

	if _, err := runShelf(t, "--config", cfg, "--vault", vault, "init", "--minimal"); err == nil {
		t.Fatalf("expected removed --minimal flag to fail")
	}
}

func TestInitForcePreservesExistingVault(t *testing.T) {
	dir := t.TempDir()
	vault := filepath.Join(dir, "secrets.vault")
	cfg := filepath.Join(dir, "config.yaml")
	if _, err := runShelf(t, "--config", cfg, "--vault", vault, "init"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--vault", vault, "secret", "set", "app:token", "val"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--vault", vault, "init", "--force"); err != nil {
		t.Fatalf("force init: %v", err)
	}
	out, err := runShelf(t, "--config", cfg, "--vault", vault, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get after force init: %v", err)
	}
	if out != "val\n" {
		t.Fatalf("force init should preserve existing vault, got: %s", out)
	}
}

func TestVersionFlagPrintsVersion(t *testing.T) {
	out, err := runShelf(t, "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.HasPrefix(out, "shelf") {
		t.Fatalf("unexpected version output: %q", out)
	}
}
