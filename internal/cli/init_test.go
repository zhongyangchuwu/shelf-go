package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesFilesAndReportsStatus(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "secrets.json")
	cfg := filepath.Join(dir, "shelf.yaml")

	out, err := runShelf(t, "--config", cfg, "--data", data, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, want := range []string{data, "(created)", cfg, "(created)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("init output missing %q: %s", want, out)
		}
	}

	out2, err := runShelf(t, "--config", cfg, "--data", data, "init")
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	for _, want := range []string{data, "(exists)", cfg, "(exists)"} {
		if !strings.Contains(out2, want) {
			t.Fatalf("second init output missing %q: %s", want, out2)
		}
	}

	out3, err := runShelf(t, "--config", cfg, "--data", data, "init", "--minimal")
	if err != nil {
		t.Fatalf("minimal init: %v", err)
	}
	if strings.Contains(out3, "config") {
		t.Fatalf("minimal init mentioned config: %s", out3)
	}
}
func TestInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "secrets.json")
	cfg := filepath.Join(dir, "config.yaml")
	if _, err := runShelf(t, "--config", cfg, "--data", data, "init"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--data", data, "secret", "set", "app:token", "val"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--data", data, "init", "--force"); err != nil {
		t.Fatalf("force init: %v", err)
	}
	out, err := runShelf(t, "--config", cfg, "--data", data, "secret", "list")
	if err != nil {
		t.Fatalf("list after force init: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty store after force init, got: %s", out)
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
