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
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
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
func TestRunInjectsPrefixSecretsWithDerivedEnvNames(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app/api:token", "secret"); err != nil {
		t.Fatalf("set token: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"prefix":"app/api"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "run", "--", "sh", "-c", "printf %s \"$APP_API_TOKEN\"")
	if err != nil {
		t.Fatalf("run command: %v\n%s", err, out)
	}
	if out != "secret" {
		t.Fatalf("unexpected child output: %q", out)
	}
}
func TestRunDryRunReportsParentEnvOverride(t *testing.T) {
	t.Setenv("APP_TOKEN", "parent")
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
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

func TestRunDryRunUsesEncryptedVaultWithoutPlaintextSideData(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	t.Setenv("APP_TOKEN", "parent-secret")
	data := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "run-secret-value", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "run", "--dry-run", "--", "sh", "-c", "printf %s \"$APP_TOKEN\"")
	if err != nil {
		t.Fatalf("dry-run: %v\n%s", err, out)
	}
	for _, want := range []string{"warn APP_TOKEN overrides existing environment variable", "inject APP_TOKEN"} {
		if !strings.Contains(out, want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "run-secret-value") || strings.Contains(out, "parent-secret") {
		t.Fatalf("dry-run leaked env value:\n%s", out)
	}
	content, err := os.ReadFile(data)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if strings.Contains(string(content), "run-secret-value") || strings.Contains(string(content), "app:token") || strings.Contains(string(content), "APP_TOKEN") {
		t.Fatalf("encrypted vault contains plaintext run data")
	}
}

func TestRunDoesNotExecuteWhenResolutionFails(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	marker := filepath.Join(dir, "marker")
	manifest := `{"version":1,"secrets":[{"path":"app:missing"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", filepath.Join(dir, "secrets.json"), "project", "run", "--", "sh", "-c", "touch "+marker)
	if err == nil {
		t.Fatalf("expected run to fail")
	}
	if !strings.Contains(out, "fail app:missing missing required") {
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
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(`{"version":1,"secrets":[]}`), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := runShelf(t, "--vault", filepath.Join(dir, "secrets.json"), "project", "run", "--", "sh", "-c", "exit 7")
	if err == nil {
		t.Fatalf("expected child failure")
	}
	if code := ExitCode(err); code != 7 {
		t.Fatalf("expected exit code 7, got %d from %v", code, err)
	}
}
