package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInjectsProjectSecretsIntoChild(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.json")
	if _, err := runShelfWithInput(t, "secret\n", "--vault", data, "secret", "create", "app-token", "--id", "item_app", "--env", "APP_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create secret: %v", err)
	}
	manifest := `{"version":1,"bindings":[{"item":"item_app","field":"value","env":"APP_TOKEN"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "run", "--", "sh", "-c", "printf %s \"$APP_TOKEN\"")
	if err != nil {
		t.Fatalf("run command: %v\n%s", err, out)
	}
	if out != "secret" {
		t.Fatalf("unexpected child output: %q", out)
	}
}

func TestRunDryRunReportsParentEnvOverride(t *testing.T) {
	t.Setenv("APP_TOKEN", "parent")
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.json")
	if _, err := runShelfWithInput(t, "secret\n", "--vault", data, "secret", "create", "app-token", "--id", "item_app", "--env", "APP_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create secret: %v", err)
	}
	manifest := `{"version":1,"bindings":[{"item":"item_app","field":"value","env":"APP_TOKEN"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "run", "--dry-run", "--", "sh", "-c", "exit 99")
	if err != nil {
		t.Fatalf("dry-run: %v\n%s", err, out)
	}
	if !strings.Contains(out, "warn APP_TOKEN overrides existing environment variable") {
		t.Fatalf("missing override warning:\n%s", out)
	}
	if !strings.Contains(out, "inject APP_TOKEN") {
		t.Fatalf("missing dry-run inject line:\n%s", out)
	}
	if strings.Contains(out, "secret") || strings.Contains(out, "parent") {
		t.Fatalf("dry-run leaked env value:\n%s", out)
	}
}

func TestRunDoesNotExecuteWhenResolutionFails(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	marker := filepath.Join(dir, "marker")
	manifest := `{"version":1,"bindings":[{"item":"item_missing","field":"value","env":"APP_TOKEN"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	data := filepath.Join(dir, "vault.json")
	if err := os.WriteFile(data, []byte(`{"version":2,"items":[]}`), 0o600); err != nil {
		t.Fatalf("write vault: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "run", "--", "sh", "-c", "touch "+marker)
	if err == nil {
		t.Fatalf("expected run to fail")
	}
	if !strings.Contains(out, "fail item_missing.value missing item required") {
		t.Fatalf("missing resolution failure:\n%s", out)
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("command executed despite resolution failure")
	}
}

func TestRunReturnsChildExitCode(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(`{"version":1,"bindings":[]}`), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := runShelf(t, "--vault", filepath.Join(dir, "vault.json"), "project", "run", "--", "sh", "-c", "exit 7")
	if err == nil {
		t.Fatalf("expected child failure")
	}
	if code := ExitCode(err); code != 7 {
		t.Fatalf("expected exit code 7, got %d from %v", code, err)
	}
}
