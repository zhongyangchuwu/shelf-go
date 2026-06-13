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
	if _, err := runShelf(t, "--config", configPath, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   config resolves", "ok   version", "ok   data file exists", "ok   store loads", "ok   data file mode"} {
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
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err == nil {
		t.Fatalf("expected doctor to fail invalid store")
	}
	if !strings.Contains(out, "fail store loads") {
		t.Fatalf("doctor output missing store failure:\n%s", out)
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
	if _, err := runShelf(t, "--config", configPath, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	want := "ok   completion installed (" + filepath.Join(completionDir, "_shelf") + ")"
	if !strings.Contains(out, want) {
		t.Fatalf("doctor did not use FPATH completion path %q:\n%s", want, out)
	}
}
