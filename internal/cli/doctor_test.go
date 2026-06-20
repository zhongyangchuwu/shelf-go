package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorReportsHealthyStore(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--config", configPath, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   config resolves", "ok   version", "ok   vault file exists", "ok   vault format", "ok   vault loads", "ok   vault file mode"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}
func TestDoctorFailsInvalidStore(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	if err := os.WriteFile(data, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write invalid data: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", data, "doctor")
	if err == nil {
		t.Fatalf("expected doctor to fail invalid store")
	}
	if !strings.Contains(out, "fail vault format") {
		t.Fatalf("doctor output missing format failure:\n%s", out)
	}
}
func TestDoctorChecksCompletionFromFpath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	completionDir := filepath.Join(dir, "zfunc")
	if err := os.MkdirAll(completionDir, 0o700); err != nil {
		t.Fatalf("mkdir completion dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(completionDir, "_shelf"), []byte("#compdef shelf\n"), 0o600); err != nil {
		t.Fatalf("write completion: %v", err)
	}
	t.Setenv("FPATH", filepath.Join(dir, "missing")+":"+completionDir)
	if _, err := runShelf(t, "--config", configPath, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	want := "ok   completion installed (" + filepath.Join(completionDir, "_shelf") + ")"
	if !strings.Contains(out, want) {
		t.Fatalf("doctor did not use FPATH completion path %q:\n%s", want, out)
	}
}

func TestDoctorFailsTrackedPlaintextStore(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "secrets.json")
	plaintext := []byte("{\n  \"version\": 1,\n  \"secrets\": {\n    \"app:token\": {\"value\": \"tracked-secret\"}\n  }\n}\n")
	if err := os.WriteFile(vaultPath, plaintext, 0o600); err != nil {
		t.Fatalf("write plaintext: %v", err)
	}
	if _, err := runGit(t, "add", "secrets.json"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "doctor")
	if err == nil {
		t.Fatalf("expected doctor to fail tracked plaintext store")
	}
	for _, want := range []string{"fail vault format", "fail git tracking", "tracked plaintext secret store is unsafe"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorConfirmsTrackedEncryptedVault(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	configPath := filepath.Join(dir, "config.yaml")
	vaultPath := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "secret", "set", "app:token", "safe-secret"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runGit(t, "add", "vault.age"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--vault", vaultPath, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ok   git tracking (tracked vault is encrypted: vault.age)") {
		t.Fatalf("doctor did not confirm encrypted tracked vault:\n%s", out)
	}
}
